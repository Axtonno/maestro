package mutation

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestLiveGateExecutorsTraverseDirectReadOnlyAndControlledMutationPaths(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	dependencies := application.Dependencies{
		Getenv: func(string) string { return "" },
		ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) {
			return qualificationProvider{profile: profile}, nil
		},
		RunID: func() (pkgAgent.RunID, error) { return "run-mutation-qualification", nil },
	}
	options := LiveOptions{Profile: profile, Dependencies: dependencies}
	for _, gate := range []string{GateA, GateB} {
		report, err := RunLiveGate(context.Background(), gate, options)
		if err != nil {
			t.Fatalf("run Gate %s: %v", gate, err)
		}
		if report.State != "passed" || len(report.Samples) != report.Required {
			t.Fatalf("Gate %s did not pass: %#v", gate, report)
		}
	}
	options.Approver = func(int) pkgTool.Approver { return allowOnceApprover{} }
	report, err := RunLiveGate(context.Background(), GateC, options)
	if err != nil {
		t.Fatal(err)
	}
	if report.State != "passed" || len(report.Samples) != 3 {
		t.Fatalf("Gate C did not pass: %#v", report)
	}
	for _, sample := range report.Samples {
		if sample.FinalSHA256 != profile.Fixture.ExpectedSHA256 || sample.Approval != "allowed_once" ||
			sample.ContextFreshness != "fresh" || sample.MutationAttempts != 1 {
			t.Fatalf("Gate C evidence mismatch: %#v", sample)
		}
	}
}

type qualificationProvider struct{ profile Profile }

func (provider qualificationProvider) ID() pkgProvider.ID { return "ollama" }

func (provider qualificationProvider) Complete(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	message, finish := provider.response(request)
	return pkgProvider.CompletionResponse{
		Model: request.Model, Message: message, FinishReason: finish,
		Usage: pkgProvider.Usage{InputTokens: 10, OutputTokens: 5},
	}, nil
}

func (provider qualificationProvider) Stream(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.Stream, error) {
	message, finish := provider.response(request)
	chunk := pkgProvider.StreamChunk{
		Model: request.Model, Content: message.Content, FinishReason: finish,
		Usage: pkgProvider.Usage{InputTokens: 10, OutputTokens: 5},
	}
	for index, call := range message.ToolCalls {
		chunk.ToolCalls = append(chunk.ToolCalls, pkgProvider.ToolCallDelta{
			Index: index, ID: call.ID, Name: call.Name, Arguments: string(call.Arguments),
		})
	}
	return &qualificationStream{chunk: chunk}, nil
}

func (provider qualificationProvider) response(request pkgProvider.CompletionRequest) (pkgProvider.Message, string) {
	if hasToolResult(request.Messages, "workspace_patch") {
		return pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "The exact patch was applied and the workspace context is fresh."}, pkgProvider.FinishReasonStop
	}
	if tool, exists := requestTool(request.Tools, "workspace_patch"); exists {
		arguments, _ := json.Marshal(map[string]string{
			"path": provider.profile.Fixture.Target, "old": provider.profile.Fixture.Old,
			"new":             provider.profile.Fixture.Replacement,
			"expected_digest": provider.profile.Fixture.InitialSHA256,
		})
		return pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{
			ID: "call-patch", Name: tool.Name, Arguments: arguments,
		}}}, pkgProvider.FinishReasonToolCalls
	}
	if hasToolResult(request.Messages, "workspace_read") {
		return pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "OrderService create handles the controller request."}, pkgProvider.FinishReasonStop
	}
	if tool, exists := requestTool(request.Tools, "workspace_read"); exists {
		arguments, _ := json.Marshal(map[string]string{"path": provider.profile.Fixture.Target})
		return pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{
			ID: "call-read", Name: tool.Name, Arguments: arguments,
		}}}, pkgProvider.FinishReasonToolCalls
	}
	return pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "OrderService create handles the controller request."}, pkgProvider.FinishReasonStop
}

func requestTool(tools []pkgProvider.Tool, name string) (pkgProvider.Tool, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return pkgProvider.Tool{}, false
}

func hasToolResult(messages []pkgProvider.Message, name string) bool {
	for _, message := range messages {
		if message.Role == pkgProvider.RoleTool && message.ToolName == name {
			return true
		}
	}
	return false
}

type qualificationStream struct {
	chunk pkgProvider.StreamChunk
	done  bool
}

func (stream *qualificationStream) Recv() (pkgProvider.StreamChunk, error) {
	if stream.done {
		return pkgProvider.StreamChunk{}, io.EOF
	}
	stream.done = true
	return stream.chunk, nil
}

func (*qualificationStream) Close() error { return nil }

type allowOnceApprover struct{}

func (allowOnceApprover) Approve(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
	return pkgTool.NewApproval(pkgTool.ApprovalAllow, "qualification_allow_once", "", pkgTool.GrantOneShot)
}
