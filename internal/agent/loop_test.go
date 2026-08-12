package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	internalTool "github.com/antonio-cafeo/maestro/internal/tool"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestAgentLoopExecutesCorrelatedToolCallAndCompletesPlan(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{
		{Message: provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-1", Name: "fixture_read", Arguments: json.RawMessage(`{"path":"main.go"}`)}}}, FinishReason: provider.FinishReasonToolCalls, Usage: provider.Usage{InputTokens: 3, OutputTokens: 2}},
		{Message: provider.Message{Role: provider.RoleAssistant, Content: "The file was inspected."}, FinishReason: provider.FinishReasonStop, Usage: provider.Usage{InputTokens: 4, OutputTokens: 3}},
	}}
	tools := allowedToolRuntime(t, &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "fixture_read")})
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))

	request := requestWithTools(t, runRequest(t, "run-loop", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.read"}, false)
	result, err := runtime.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("run loop: %v", err)
	}
	if result.Content() != "The file was inspected." || result.Session().Terminal() != pkgAgent.TerminalCompleted {
		t.Fatalf("unexpected result: %#v", result)
	}
	plan, _ := result.Session().Plan()
	if plan.Steps()[0].Status() != pkgAgent.StepCompleted {
		t.Fatalf("step was not completed: %#v", plan.Steps()[0])
	}
	counters := result.Session().Counters()
	if counters.ModelTurns != 2 || counters.ToolCalls != 1 || counters.InputTokens != 7 || counters.OutputTokens != 5 {
		t.Fatalf("unexpected counters: %#v", counters)
	}
	requests := providers.Requests()
	if len(requests) != 2 || len(requests[1].Messages) != 4 {
		t.Fatalf("unexpected provider conversation: %#v", requests)
	}
	toolMessage := requests[1].Messages[3]
	if toolMessage.Role != provider.RoleTool || toolMessage.ToolCallID != "call-1" || toolMessage.ToolName != "fixture_read" {
		t.Fatalf("tool correlation was lost: %#v", toolMessage)
	}
	if !json.Valid([]byte(toolMessage.Content)) || !containsAll(toolMessage.Content, `"outcome":"success"`, `"content":"ok"`) {
		t.Fatalf("tool result is not typed JSON: %s", toolMessage.Content)
	}
}

func TestAgentLoopRejectsUnknownToolWithoutInvocation(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{{
		Message:      provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-1", Name: "not_allowed", Arguments: json.RawMessage(`{}`)}}},
		FinishReason: provider.FinishReasonToolCalls,
	}}}
	fixture := &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "fixture_read")}
	tools := allowedToolRuntime(t, fixture)
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))

	request := requestWithTools(t, runRequest(t, "run-unknown", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.read"}, false)
	_, err := runtime.Run(context.Background(), request)
	if !errors.Is(err, pkgAgent.ErrToolFailed) || fixture.executions != 0 {
		t.Fatalf("unknown tool reached execution: err=%v executions=%d", err, fixture.executions)
	}
	snapshot, _ := runtime.Session("run-unknown")
	if snapshot.Terminal() != pkgAgent.TerminalToolFailure {
		t.Fatalf("unexpected terminal: %q", snapshot.Terminal())
	}
}

func TestAgentLoopExecutesMultipleCallsInProviderOrder(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{
		{Message: provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{
			{ID: "call-2", Name: "fixture_second", Arguments: json.RawMessage(`{}`)},
			{ID: "call-1", Name: "fixture_first", Arguments: json.RawMessage(`{}`)},
		}}, FinishReason: provider.FinishReasonToolCalls},
		{Message: provider.Message{Role: provider.RoleAssistant, Content: "ordered"}, FinishReason: provider.FinishReasonStop},
	}}
	order := []string{}
	second := &loopTool{descriptor: loopToolDescriptor(t, "fixture.second", "fixture_second"), label: "second", order: &order}
	first := &loopTool{descriptor: loopToolDescriptor(t, "fixture.first", "fixture_first"), label: "first", order: &order}
	tools := allowedToolRuntime(t, first, second)
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))
	request := requestWithTools(t, runRequest(t, "run-multiple", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.first", "fixture.second"}, false)

	result, err := runtime.Run(context.Background(), request)
	if err != nil || result.Session().Counters().ToolCalls != 2 {
		t.Fatalf("multiple calls failed: result=%#v err=%v", result, err)
	}
	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("provider order was not preserved: %#v", order)
	}
}

func TestAgentLoopReturnsRecoverableDenyToModel(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{
		{Message: provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-1", Name: "fixture_read", Arguments: json.RawMessage(`{}`)}}}, FinishReason: provider.FinishReasonToolCalls},
		{Message: provider.Message{Role: provider.RoleAssistant, Content: "Permission was unavailable."}, FinishReason: provider.FinishReasonStop},
	}}
	tools := internalTool.NewRuntime()
	fixture := &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "fixture_read")}
	if err := tools.Register(fixture); err != nil {
		t.Fatal(err)
	}
	deny, _ := pkgTool.NewDecision(pkgTool.DecisionDeny, "not_allowed", pkgTool.DenyRecoverable, "")
	allow, _ := pkgTool.NewDecision(pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantRun)
	if err := tools.RegisterPolicy(&subjectPolicy{id: "policy.test", model: allow, tool: deny}); err != nil {
		t.Fatal(err)
	}
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))

	request := requestWithTools(t, runRequest(t, "run-denied", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.read"}, false)
	result, err := runtime.Run(context.Background(), request)
	if err != nil || result.Session().Terminal() != pkgAgent.TerminalCompleted || fixture.executions != 0 {
		t.Fatalf("recoverable deny handling: result=%#v err=%v executions=%d", result, err, fixture.executions)
	}
	requests := providers.Requests()
	if !containsAll(requests[1].Messages[3].Content, `"outcome":"denied"`, `"disposition":"recoverable"`) {
		t.Fatalf("denial not returned to model: %s", requests[1].Messages[3].Content)
	}
}

func TestStreamingAssemblerConvergesOnFragmentedToolCalls(t *testing.T) {
	providers := &generationStub{streams: []provider.Stream{
		&chunkStream{chunks: []provider.StreamChunk{
			{ToolCalls: []provider.ToolCallDelta{{Index: 0, ID: "call-1", Name: "fixture_read", Arguments: `{"pa`}}},
			{ToolCalls: []provider.ToolCallDelta{{Index: 0, Arguments: `th":"main.go"}`}}, FinishReason: provider.FinishReasonToolCalls},
		}},
		&chunkStream{chunks: []provider.StreamChunk{{Content: "Done", FinishReason: provider.FinishReasonStop, Usage: provider.Usage{InputTokens: 2, OutputTokens: 1}}}},
	}}
	tools := allowedToolRuntime(t, &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "fixture_read")})
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))
	request := requestWithTools(t, runRequest(t, "run-stream", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.read"}, true)

	result, err := runtime.Run(context.Background(), request)
	if err != nil || result.Content() != "Done" || providers.completeCalls != 0 {
		t.Fatalf("stream run did not converge: result=%#v err=%v complete_calls=%d", result, err, providers.completeCalls)
	}
}

func TestStreamingAssemblerRejectsMidStreamFailureAndCloses(t *testing.T) {
	stream := &chunkStream{chunks: []provider.StreamChunk{{Content: "partial"}}, err: errors.New("transport failed")}
	_, err := assembleStream(stream, 1024, 2)
	if err == nil || !stream.closed {
		t.Fatalf("mid-stream failure was not propagated/closed: err=%v closed=%v", err, stream.closed)
	}
}

type generationStub struct {
	mu            sync.Mutex
	responses     []provider.CompletionResponse
	streams       []provider.Stream
	requests      []provider.CompletionRequest
	completeCalls int
}

func (stub *generationStub) Complete(_ context.Context, _ provider.ID, request provider.CompletionRequest) (provider.CompletionResponse, error) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.completeCalls++
	stub.requests = append(stub.requests, request)
	if len(stub.responses) == 0 {
		return provider.CompletionResponse{}, errors.New("unexpected completion")
	}
	response := stub.responses[0]
	stub.responses = stub.responses[1:]
	return response, nil
}

func (stub *generationStub) Stream(_ context.Context, _ provider.ID, request provider.CompletionRequest) (provider.Stream, error) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.requests = append(stub.requests, request)
	if len(stub.streams) == 0 {
		return nil, errors.New("unexpected stream")
	}
	stream := stub.streams[0]
	stub.streams = stub.streams[1:]
	return stream, nil
}

func (stub *generationStub) Requests() []provider.CompletionRequest {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return append([]provider.CompletionRequest(nil), stub.requests...)
}

type chunkStream struct {
	chunks []provider.StreamChunk
	err    error
	closed bool
}

func (stream *chunkStream) Recv() (provider.StreamChunk, error) {
	if len(stream.chunks) > 0 {
		chunk := stream.chunks[0]
		stream.chunks = stream.chunks[1:]
		return chunk, nil
	}
	if stream.err != nil {
		err := stream.err
		stream.err = nil
		return provider.StreamChunk{}, err
	}
	return provider.StreamChunk{}, io.EOF
}

func (stream *chunkStream) Close() error { stream.closed = true; return nil }

type loopTool struct {
	descriptor pkgTool.Descriptor
	executions int
	label      string
	order      *[]string
}

type subjectPolicy struct {
	id    pkgTool.PolicyID
	model pkgTool.Decision
	tool  pkgTool.Decision
}

func (policy *subjectPolicy) ID() pkgTool.PolicyID { return policy.id }
func (policy *subjectPolicy) Decide(_ context.Context, request pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	if request.Subject() == pkgTool.PermissionSubjectModel {
		return policy.model, nil
	}
	return policy.tool, nil
}

func (fixture *loopTool) Descriptor() pkgTool.Descriptor { return fixture.descriptor }
func (fixture *loopTool) Prepare(_ context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
	action, _ := pkgTool.NewAction(pkgTool.EffectLocalCompute, "fixture", "")
	return pkgTool.NewPreparedInvocation(invocation, fixture.descriptor.Version(), invocation.Arguments(), []pkgTool.Action{action})
}
func (fixture *loopTool) Execute(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	fixture.executions++
	if fixture.order != nil {
		*fixture.order = append(*fixture.order, fixture.label)
	}
	return pkgTool.NewResult(pkgTool.ResultSuccess, "ok", "text/plain", "completed", 1, false, "")
}

func loopToolDescriptor(t *testing.T, id pkgTool.ID, name pkgTool.Name) pkgTool.Descriptor {
	t.Helper()
	descriptor, err := pkgTool.NewDescriptor(id, name, "1", "Loop fixture tool.", json.RawMessage(`{"type":"object"}`), []pkgTool.Effect{pkgTool.EffectLocalCompute})
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

func allowedToolRuntime(t *testing.T, fixtures ...pkgTool.Tool) *internalTool.Runtime {
	t.Helper()
	runtime := internalTool.NewRuntime()
	for _, fixture := range fixtures {
		if err := runtime.Register(fixture); err != nil {
			t.Fatal(err)
		}
	}
	allow, _ := pkgTool.NewDecision(pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantRun)
	if err := runtime.RegisterPolicy(&decisionPolicy{id: "policy.test", decision: allow}); err != nil {
		t.Fatal(err)
	}
	return runtime
}

func loopRuntime(t *testing.T, providers generationRuntime, tools pkgTool.Runtime, plan pkgAgent.Plan) *Runtime {
	t.Helper()
	options := DefaultOptions()
	options.Providers = providers
	options.Tools = tools
	runtime, err := NewRuntimeWithOptions(&contextStub{bundle: testBundle(t, "workspace")}, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Register(newAgentFixture(t, "agent.general", plan)); err != nil {
		t.Fatal(err)
	}
	return runtime
}

func requestWithTools(t *testing.T, source pkgAgent.RunRequest, tools []pkgTool.ID, streaming bool) pkgAgent.RunRequest {
	t.Helper()
	request, err := pkgAgent.NewRunRequest(
		source.Run(), source.Agent(), source.Provider(), source.Model(), source.Workspace(), source.Policy(), source.Instruction(), source.Limits(),
		pkgAgent.RunRequestOptions{Context: source.Context(), Tools: tools, Approver: source.Approver(), Streaming: streaming},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func containsAll(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(value, fragment) {
			return false
		}
	}
	return true
}
