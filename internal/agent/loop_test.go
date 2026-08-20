package agent

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	internalTool "github.com/antonio-cafeo/maestro/internal/tool"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
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

func TestAgentLoopRecoversTextualPseudoToolCallThroughDeclaredInterface(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{
		{
			Message: provider.Message{Role: provider.RoleAssistant, Content: "The file must be inspected first.\n\nHere's a tool call:\n\n" +
				`{"name":"fixture_read","parameters":{"path":"main.go"}}`},
			FinishReason: provider.FinishReasonStop,
			Usage:        provider.Usage{InputTokens: 3, OutputTokens: 2},
		},
		{
			Message:      provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-1", Name: "fixture_read", Arguments: json.RawMessage(`{"path":"main.go"}`)}}},
			FinishReason: provider.FinishReasonToolCalls,
			Usage:        provider.Usage{InputTokens: 4, OutputTokens: 3},
		},
		{
			Message:      provider.Message{Role: provider.RoleAssistant, Content: "The file was inspected."},
			FinishReason: provider.FinishReasonStop,
			Usage:        provider.Usage{InputTokens: 5, OutputTokens: 4},
		},
	}}
	fixture := &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "fixture_read")}
	tools := allowedToolRuntime(t, fixture)
	runtime := loopRuntime(t, providers, tools, pendingPlan(t, "inspect"))
	request := requestWithTools(t, runRequest(t, "run-pseudo-call", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.read"}, false)

	result, err := runtime.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("recover pseudo-call: %v", err)
	}
	if result.Content() != "The file was inspected." || fixture.executions != 1 {
		t.Fatalf("pseudo-call was accepted or tool was not executed: result=%#v executions=%d", result, fixture.executions)
	}
	if counters := result.Session().Counters(); counters.ModelTurns != 3 || counters.ToolCalls != 1 || counters.InputTokens != 12 || counters.OutputTokens != 9 {
		t.Fatalf("unexpected counters: %#v", counters)
	}
	requests := providers.Requests()
	if len(requests) != 3 || len(requests[1].Messages) != 4 ||
		!strings.Contains(requests[1].Messages[3].Content, "described a declared tool call as text") {
		t.Fatalf("protocol correction was not sent: %#v", requests)
	}
}

func TestAgentLoopAcceptsFinalAnswerThatNamesDeclaredTool(t *testing.T) {
	providers := &generationStub{responses: []provider.CompletionResponse{{
		Message:      provider.Message{Role: provider.RoleAssistant, Content: "The workspace_read result shows that the service validates orders."},
		FinishReason: provider.FinishReasonStop,
	}}}
	fixture := &loopTool{descriptor: loopToolDescriptor(t, "fixture.read", "workspace_read")}
	runtime := loopRuntime(t, providers, allowedToolRuntime(t, fixture), pendingPlan(t, "explain"))
	request := requestWithTools(t, runRequest(t, "run-tool-name", "agent.general", "workspace", 2), []pkgTool.ID{"fixture.read"}, false)

	result, err := runtime.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("accept final answer: %v", err)
	}
	if result.Content() != "The workspace_read result shows that the service validates orders." || fixture.executions != 0 {
		t.Fatalf("final answer was not accepted: result=%#v executions=%d", result, fixture.executions)
	}
	if got := len(providers.Requests()); got != 1 {
		t.Fatalf("unexpected correction turn: got %d requests", got)
	}
}

func TestEncodeToolResultPreservesStructuredJSON(t *testing.T) {
	result, err := pkgTool.NewResult(
		pkgTool.ResultSuccess,
		`{"path":"main.go","digest":"abc"}`,
		"application/json",
		"read",
		1,
		false,
		"",
	)
	if err != nil {
		t.Fatalf("construct result: %v", err)
	}
	encoded, err := encodeToolResult(result)
	if err != nil {
		t.Fatalf("encode result: %v", err)
	}
	if !containsAll(encoded, `"content":{"path":"main.go","digest":"abc"}`, `"media_type":"application/json"`) {
		t.Fatalf("structured JSON was not preserved: %s", encoded)
	}
}

func TestInitialMessagesAreScopedToDeclaredToolEffects(t *testing.T) {
	request := runRequest(t, "run-protocol", "agent.general", "workspace", 5)
	plan := pendingPlan(t, "inspect")
	messages := initialMessages(request, plan.Steps()[0], testBundle(t, "workspace"))
	if len(messages) != 2 || !containsAll(
		messages[0].Content,
		"at most one tool call per response",
		"invoke it through the declared tool interface",
		"declared tool set is read-only",
		"exact declared function name",
		"logical path relative to the workspace",
	) || strings.Contains(messages[0].Content, "copy expected_digest exactly") {
		t.Fatalf("read-only protocol advertises mutation: %#v", messages)
	}

	mutating := requestWithTools(t, request, []pkgTool.ID{"workspace.read", "workspace.patch"}, false)
	messages = initialMessages(mutating, plan.Steps()[0], testBundle(t, "workspace"))
	if len(messages) != 2 || !containsAll(
		messages[0].Content,
		"at most one tool call per response",
		"invoke it through the declared tool interface",
		"copy expected_digest exactly",
		"preserving whitespace and real newline characters",
	) {
		t.Fatalf("guarded mutation protocol is missing: %#v", messages)
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

func TestAgentLoopExecutesOnlyFirstCallAndReturnsRecoverableDependency(t *testing.T) {
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
	if len(order) != 1 || order[0] != "second" || first.executions != 0 {
		t.Fatalf("dependent call reached execution: order=%#v first=%d", order, first.executions)
	}
	requests := providers.Requests()
	if len(requests) != 2 || len(requests[1].Messages) != 5 ||
		!containsAll(requests[1].Messages[4].Content, `"outcome":"invalid"`, `"reason":"dependency_not_ready"`, `"recoverable":true`) {
		t.Fatalf("recoverable dependency result was not returned: %#v", requests)
	}
}

func TestToolChoreographyRequiresVerifiedReadForPatch(t *testing.T) {
	read := workspaceDescriptor(t, workspaceReadToolID, "workspace_read", pkgTool.EffectWorkspaceInspect)
	patch := workspaceDescriptor(t, workspacePatchToolID, "workspace_patch", pkgTool.EffectWorkspaceMutate)
	descriptors := map[string]pkgTool.Descriptor{"workspace_read": read, "workspace_patch": patch}
	tools := []provider.Tool{{Name: "workspace_read"}, {Name: "workspace_patch"}}
	state := &toolChoreography{}

	if available := state.toolsForTurn(tools, descriptors); len(available) != 1 || available[0].Name != "workspace_read" {
		t.Fatalf("patch was available without read evidence: %#v", available)
	}
	content := "package main\n"
	digest := sha256.Sum256([]byte(content))
	readContent, _ := json.Marshal(map[string]string{
		"path": "main.go", "digest": fmt.Sprintf("%x", digest), "content": content,
	})
	readResult, _ := pkgTool.NewResult(pkgTool.ResultSuccess, string(readContent), "application/json", "read", 1, false, "")
	state.afterCall(read, readResult)
	if available := state.toolsForTurn(tools, descriptors); len(available) != 2 {
		t.Fatalf("patch was not enabled by verified evidence: %#v", available)
	}

	valid, _ := json.Marshal(map[string]string{
		"path": "main.go", "old": "package main", "new": "package changed", "expected_digest": fmt.Sprintf("%x", digest),
	})
	if _, execute, err := state.beforeCall(patch, valid, 0); err != nil || !execute {
		t.Fatalf("valid observed patch was rejected: execute=%t err=%v", execute, err)
	}
	stale, _ := json.Marshal(map[string]string{
		"path": "main.go", "old": "package main", "new": "package changed", "expected_digest": strings.Repeat("0", 64),
	})
	result, execute, err := state.beforeCall(patch, stale, 0)
	if err != nil || execute || result.Reason() != "stale_observation" || result.Outcome() != pkgTool.ResultInvalid {
		t.Fatalf("stale patch was not recoverable: result=%#v execute=%t err=%v", result, execute, err)
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

func TestMutationMarksContextStaleRefreshesAndUsesNewGeneration(t *testing.T) {
	workspace, initial, refreshed, refreshedSnapshot := refreshFixture(t)
	contexts := &refreshContext{initial: initial, refreshed: refreshed, snapshot: refreshedSnapshot}
	providers := &generationStub{responses: []provider.CompletionResponse{
		{Message: provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-mutate", Name: "fixture_mutate", Arguments: json.RawMessage(`{}`)}}}, FinishReason: provider.FinishReasonToolCalls},
		{Message: provider.Message{Role: provider.RoleAssistant, Content: "mutation complete"}, FinishReason: provider.FinishReasonStop},
		{Message: provider.Message{Role: provider.RoleAssistant, Content: "fresh context used"}, FinishReason: provider.FinishReasonStop},
	}}
	registry := internalTool.NewWorkspaceRegistry()
	tools := allowedToolRuntime(t, &mutationTool{descriptor: mutationDescriptor(t), workspace: workspace.ID()})
	options := DefaultOptions()
	options.Providers = providers
	options.Tools = tools
	options.Workspaces = registry
	runtime, err := NewRuntimeWithOptions(contexts, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Register(newAgentFixture(t, "agent.general", pendingPlan(t, "mutate", "verify"))); err != nil {
		t.Fatal(err)
	}
	request := requestWithWorkspace(t, requestWithTools(t, runRequest(t, "run-refresh", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.mutate"}, false), workspace)

	result, err := runtime.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("mutation run: %v", err)
	}
	if result.Session().ContextStale() || result.Session().WorkspaceGeneration() != 2 || contexts.indexCalls != 1 {
		t.Fatalf("context was not refreshed: snapshot=%#v index_calls=%d", result.Session(), contexts.indexCalls)
	}
	requests := providers.Requests()
	if len(requests) != 3 || !strings.Contains(requests[2].Messages[1].Content, "fresh evidence") {
		t.Fatalf("next step did not use refreshed evidence: %#v", requests)
	}
}

func TestRefreshFailurePreservesStaleSession(t *testing.T) {
	workspace, initial, refreshed, refreshedSnapshot := refreshFixture(t)
	contexts := &refreshContext{initial: initial, refreshed: refreshed, snapshot: refreshedSnapshot, indexErr: errors.New("refresh failed")}
	providers := &generationStub{responses: []provider.CompletionResponse{{
		Message:      provider.Message{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "call-mutate", Name: "fixture_mutate", Arguments: json.RawMessage(`{}`)}}},
		FinishReason: provider.FinishReasonToolCalls,
	}}}
	registry := internalTool.NewWorkspaceRegistry()
	tools := allowedToolRuntime(t, &mutationTool{descriptor: mutationDescriptor(t), workspace: workspace.ID()})
	options := DefaultOptions()
	options.Providers, options.Tools, options.Workspaces = providers, tools, registry
	runtime, _ := NewRuntimeWithOptions(contexts, options)
	_ = runtime.Register(newAgentFixture(t, "agent.general", pendingPlan(t, "mutate")))
	request := requestWithWorkspace(t, requestWithTools(t, runRequest(t, "run-refresh-fail", "agent.general", "workspace", 5), []pkgTool.ID{"fixture.mutate"}, false), workspace)

	_, err := runtime.Run(context.Background(), request)
	if !errors.Is(err, pkgAgent.ErrToolFailed) {
		t.Fatalf("expected refresh failure, got %v", err)
	}
	snapshot, _ := runtime.Session("run-refresh-fail")
	if !snapshot.ContextStale() || snapshot.WorkspaceGeneration() != 1 || snapshot.Terminal() != pkgAgent.TerminalToolFailure {
		t.Fatalf("last valid generation was not preserved: %#v", snapshot)
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

type mutationTool struct {
	descriptor pkgTool.Descriptor
	workspace  pkgContext.WorkspaceID
}

func (fixture *mutationTool) Descriptor() pkgTool.Descriptor { return fixture.descriptor }
func (fixture *mutationTool) Prepare(_ context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
	action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceMutate, "main.go", fixture.workspace)
	return pkgTool.NewPreparedInvocation(invocation, fixture.descriptor.Version(), invocation.Arguments(), []pkgTool.Action{action})
}
func (*mutationTool) Execute(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	return pkgTool.NewResult(pkgTool.ResultSuccess, "mutated", "text/plain", "completed", 1, false, "")
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

func workspaceDescriptor(t *testing.T, id pkgTool.ID, name pkgTool.Name, effect pkgTool.Effect) pkgTool.Descriptor {
	t.Helper()
	descriptor, err := pkgTool.NewDescriptor(id, name, "1", "Workspace fixture tool.", json.RawMessage(`{"type":"object"}`), []pkgTool.Effect{effect})
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

func mutationDescriptor(t *testing.T) pkgTool.Descriptor {
	t.Helper()
	descriptor, err := pkgTool.NewDescriptor("fixture.mutate", "fixture_mutate", "1", "Mutation fixture tool.", json.RawMessage(`{"type":"object"}`), []pkgTool.Effect{pkgTool.EffectWorkspaceMutate})
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

func requestWithWorkspace(t *testing.T, source pkgAgent.RunRequest, workspace pkgContext.Workspace) pkgAgent.RunRequest {
	t.Helper()
	request, err := pkgAgent.NewRunRequest(
		source.Run(), source.Agent(), source.Provider(), source.Model(), source.Workspace(), source.Policy(), source.Instruction(), source.Limits(),
		pkgAgent.RunRequestOptions{Context: source.Context(), Tools: source.Tools(), Approver: source.Approver(), Streaming: source.Streaming(), Workspace: &workspace},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

type refreshContext struct {
	mu         sync.Mutex
	initial    pkgContext.ContextBundle
	refreshed  pkgContext.ContextBundle
	snapshot   pkgContext.Snapshot
	indexErr   error
	indexCalls int
}

func (*refreshContext) RegisterSource(pkgContext.Source) error            { return nil }
func (*refreshContext) RegisterAnalyzer(pkgContext.Analyzer) error        { return nil }
func (*refreshContext) RegisterEstimator(pkgContext.TokenEstimator) error { return nil }
func (engine *refreshContext) Index(context.Context, pkgContext.Workspace) (pkgContext.Snapshot, error) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	engine.indexCalls++
	return engine.snapshot, engine.indexErr
}
func (engine *refreshContext) Snapshot(pkgContext.WorkspaceID) (pkgContext.Snapshot, bool) {
	return engine.snapshot, true
}
func (*refreshContext) Retrieve(context.Context, pkgContext.RetrievalQuery) ([]pkgContext.RetrievalResult, error) {
	return nil, nil
}
func (engine *refreshContext) Build(context.Context, pkgContext.BuildRequest) (pkgContext.ContextBundle, error) {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if engine.indexCalls == 0 {
		return engine.initial, nil
	}
	return engine.refreshed, nil
}
func (*refreshContext) CacheStats() pkgContext.CacheStats { return pkgContext.CacheStats{} }

func refreshFixture(t *testing.T) (pkgContext.Workspace, pkgContext.ContextBundle, pkgContext.ContextBundle, pkgContext.Snapshot) {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace("workspace", t.TempDir(), pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	if err != nil {
		t.Fatal(err)
	}
	document1, _ := pkgContext.NewDocument("main.go", "text/plain", "", "old evidence")
	snapshot1, _ := pkgContext.NewSnapshot(workspace, 1, []pkgContext.Document{document1}, nil, nil)
	document2, _ := pkgContext.NewDocument("main.go", "text/plain", "", "fresh evidence")
	snapshot2, _ := pkgContext.NewSnapshot(workspace, 2, []pkgContext.Document{document2}, nil, nil)
	return workspace, mustBundle(t, snapshot1, document1), mustBundle(t, snapshot2, document2), snapshot2
}

func containsAll(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(value, fragment) {
			return false
		}
	}
	return true
}
