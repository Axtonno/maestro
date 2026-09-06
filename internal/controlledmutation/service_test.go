package controlledmutation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type fixtureProvider struct {
	response pkgProvider.CompletionResponse
	models   []pkgProvider.ModelInfo
	requests []pkgProvider.CompletionRequest
	unloads  []string
	complete int
}

func (provider *fixtureProvider) ID() pkgProvider.ID { return "ollama" }
func (provider *fixtureProvider) Complete(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	provider.complete++
	provider.requests = append(provider.requests, request)
	return provider.response, nil
}
func (provider *fixtureProvider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	return append([]pkgProvider.ModelInfo(nil), provider.models...), nil
}
func (provider *fixtureProvider) UnloadModel(_ context.Context, request pkgProvider.ModelUnloadRequest) error {
	provider.unloads = append(provider.unloads, request.Model)
	for index := range provider.models {
		if provider.models[index].Model.ID == request.Model {
			provider.models[index].State = pkgProvider.ModelStateAvailable
		}
	}
	return nil
}
func (provider *fixtureProvider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{Capability: capability, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable})
	}
	return pkgProvider.CapabilityReport{Provider: "ollama", Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

type approverFunc func(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error)

func (fn approverFunc) Approve(ctx context.Context, request pkgTool.PermissionRequest) (pkgTool.Approval, error) {
	return fn(ctx, request)
}

func TestExecuteAppliesOnlySelectedBytesAfterExactPreviewApproval(t *testing.T) {
	root, file := mutationWorkspace(t)
	provider := qualifiedProvider(`{"decision":"propose","new_text":"$workers = 8;"}`)
	service := buildFixtureService(t, root, provider, nil)
	approved := false
	result, err := service.Execute(context.Background(), Request{File: "app/Worker.php", StartLine: 2, EndLine: 2, Instruction: "set workers to 8", Approver: approverFunc(func(_ context.Context, request pkgTool.PermissionRequest) (pkgTool.Approval, error) {
		prepared, ok := request.Prepared()
		if !ok {
			t.Fatal("prepared invocation absent")
		}
		preview, ok := prepared.Preview()
		if !ok || !strings.Contains(preview.Body(), "-$workers = 4;") || !strings.Contains(preview.Body(), "+$workers = 8;") {
			t.Fatalf("incomplete diff: %#v", preview)
		}
		fields := map[string]string{}
		for _, field := range preview.Fields() {
			fields[field.Label()] = field.Value()
		}
		if fields["path"] != "app/Worker.php" || fields["start_line"] != "2" || fields["end_line"] != "2" || fields["single_file"] != "true" || fields["fingerprint"] == "" {
			t.Fatalf("incomplete preview fields: %v", fields)
		}
		approved = true
		return pkgTool.NewApproval(pkgTool.ApprovalAllow, "test_allow_once", "", pkgTool.GrantOneShot)
	})})
	if err != nil || !approved || result.Terminal != "applied" || result.Effect != pkgTool.EffectApplied || !result.Durable {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	content, _ := os.ReadFile(file)
	if string(content) != "<?php\n$workers = 8;\nreturn $workers;\n" {
		t.Fatalf("unexpected mutation: %q", content)
	}
	if len(provider.unloads) != 1 || provider.unloads[0] != productconfig.QualifiedDirectChatModel || len(provider.requests) != 1 || provider.requests[0].Model != productconfig.QualifiedMutationModel || provider.requests[0].ToolChoice.Mode != pkgProvider.ToolChoiceNone || provider.requests[0].Output == nil || provider.requests[0].Output.Mode != pkgProvider.StructuredOutputJSONSchema {
		t.Fatalf("routing drift: %#v %#v", provider.unloads, provider.requests)
	}
}

func TestExecuteFailsClosedForDenyStaleAndAbstain(t *testing.T) {
	for _, testCase := range []struct {
		name, raw    string
		stale, allow bool
		want         error
	}{
		{"deny", `{"decision":"propose","new_text":"$workers = 8;"}`, false, false, ErrApprovalRejected},
		{"stale", `{"decision":"propose","new_text":"$workers = 8;"}`, true, true, ErrStaleSource},
		{"abstain", `{"decision":"abstain"}`, false, true, ErrInsufficientInfo},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root, file := mutationWorkspace(t)
			provider := qualifiedProvider(testCase.raw)
			hook := func() {}
			if testCase.stale {
				hook = func() {
					if err := os.WriteFile(file, []byte("<?php\n$concurrent = true;\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			service := buildFixtureService(t, root, provider, hook)
			approvalCalls := 0
			_, err := service.Execute(context.Background(), Request{File: "app/Worker.php", StartLine: 2, EndLine: 2, Instruction: "set workers to 8", Approver: approverFunc(func(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
				approvalCalls++
				if testCase.allow {
					return pkgTool.NewApproval(pkgTool.ApprovalAllow, "test_allow_once", "", pkgTool.GrantOneShot)
				}
				return pkgTool.NewApproval(pkgTool.ApprovalDeny, "test_deny", pkgTool.DenyTerminal, "")
			})})
			if !errors.Is(err, testCase.want) {
				t.Fatalf("got %v want %v", err, testCase.want)
			}
			if testCase.name == "abstain" && approvalCalls != 0 {
				t.Fatal("abstention reached approval")
			}
			content, _ := os.ReadFile(file)
			if testCase.stale {
				if string(content) != "<?php\n$concurrent = true;\n" {
					t.Fatalf("stale write: %q", content)
				}
			} else if string(content) != "<?php\n$workers = 4;\nreturn $workers;\n" {
				t.Fatalf("failure changed file: %q", content)
			}
		})
	}
}

func TestDoctorChecksBothIdentitiesWithoutGenerationOrResidencyChange(t *testing.T) {
	root, _ := mutationWorkspace(t)
	provider := qualifiedProvider(`{"decision":"abstain"}`)
	checks := Doctor(context.Background(), fixtureConfig(root), Dependencies{ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil }}, true)
	if provider.complete != 0 || len(provider.unloads) != 0 {
		t.Fatal("doctor caused provider effects")
	}
	if len(checks) != 9 {
		t.Fatalf("checks=%#v", checks)
	}
	for _, check := range checks {
		if check.Status != CheckPass {
			t.Fatalf("failed doctor check: %#v", check)
		}
	}
}

func TestExecuteRejectsIdentityAndInvalidOutputWithoutEffects(t *testing.T) {
	for _, testCase := range []struct {
		name, raw string
		identity  bool
		want      error
	}{
		{"identity", `{"decision":"abstain"}`, true, ErrModelIdentity},
		{"schema", `{"decision":"propose","new_text":"x","path":"app/Other.php"}`, false, ErrResponseInvalid},
		{"observed model", `{"decision":"abstain"}`, false, ErrResponseInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root, file := mutationWorkspace(t)
			provider := qualifiedProvider(testCase.raw)
			if testCase.identity {
				provider.models[1].Digest = strings.Repeat("0", 64)
			}
			if testCase.name == "observed model" {
				provider.response.Model = productconfig.QualifiedDirectChatModel
			}
			service := buildFixtureService(t, root, provider, nil)
			approvalCalls := 0
			_, err := service.Execute(context.Background(), Request{File: "app/Worker.php", StartLine: 2, EndLine: 2, Instruction: "set workers to 8", Approver: approverFunc(func(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
				approvalCalls++
				return pkgTool.NewApproval(pkgTool.ApprovalAllow, "test_allow_once", "", pkgTool.GrantOneShot)
			})})
			if !errors.Is(err, testCase.want) || approvalCalls != 0 {
				t.Fatalf("err=%v approvals=%d", err, approvalCalls)
			}
			content, _ := os.ReadFile(file)
			if string(content) != "<?php\n$workers = 4;\nreturn $workers;\n" {
				t.Fatalf("failure changed file: %q", content)
			}
		})
	}
}

func mutationWorkspace(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	app := filepath.Join(root, "app")
	if err := os.Mkdir(app, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(app, "Worker.php")
	if err := os.WriteFile(file, []byte("<?php\n$workers = 4;\nreturn $workers;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, file
}

func fixtureConfig(root string) productconfig.Config {
	return productconfig.Config{Version: 4, Provider: productconfig.ProviderConfig{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: productconfig.Duration{Duration: 5 * time.Minute}}, Workspace: productconfig.WorkspaceConfig{ID: "laravel", Root: root, Framework: "laravel"}, DirectChat: productconfig.ChatProfileConfig{ProfileConfig: productconfig.ProfileConfig{Model: productconfig.QualifiedDirectChatModel, Timeout: productconfig.Duration{Duration: 5 * time.Minute}, NumCtx: 4096, Thinking: productconfig.ThinkingDisabled}, Digest: productconfig.QualifiedDirectChatDigest, NumPredict: 1024, Residency: productconfig.Duration{Duration: 5 * time.Minute}, MaxFileBytes: 1 << 20, MaxOutputBytes: 1 << 20}, ControlledMutation: productconfig.ControlledMutationProfileConfig{Enabled: true, Model: productconfig.QualifiedMutationModel, Digest: productconfig.QualifiedMutationDigest, Timeout: productconfig.Duration{Duration: 5 * time.Minute}, NumCtx: 4096, NumPredict: 1024, Thinking: productconfig.ThinkingDisabled, Residency: productconfig.Duration{Duration: 5 * time.Minute}, Prompt: productconfig.MutationPromptID, PromptSHA256: productconfig.MutationPromptSHA256, Schema: productconfig.MutationSchemaID, SchemaSHA256: productconfig.MutationSchemaSHA256, MaxOutputBytes: 1 << 20}}
}

func qualifiedProvider(raw string) *fixtureProvider {
	return &fixtureProvider{response: pkgProvider.CompletionResponse{Model: productconfig.QualifiedMutationModel, Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: raw}, FinishReason: pkgProvider.FinishReasonStop, Usage: pkgProvider.Usage{InputTokens: 10, OutputTokens: 5}}, models: []pkgProvider.ModelInfo{{Model: pkgProvider.Model{ID: productconfig.QualifiedDirectChatModel}, Digest: productconfig.QualifiedDirectChatDigest, State: pkgProvider.ModelStateLoaded}, {Model: pkgProvider.Model{ID: productconfig.QualifiedMutationModel}, Digest: productconfig.QualifiedMutationDigest, State: pkgProvider.ModelStateAvailable}}}
}

func buildFixtureService(t *testing.T, root string, provider *fixtureProvider, hook func()) *Service {
	t.Helper()
	service, err := Build(fixtureConfig(root), Dependencies{ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil }, RunID: func() (pkgTool.RunID, error) { return "mutation-test", nil }, AfterPreview: hook})
	if err != nil {
		t.Fatal(err)
	}
	return service
}
