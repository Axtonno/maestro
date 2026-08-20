package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestApplicationExecutesReferenceAgentPatchThroughConfiguredPolicy(t *testing.T) {
	root := newLaravelWorkspace(t)
	filename := filepath.Join(root, "app", "Order.php")
	original, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(original)
	arguments, _ := json.Marshal(map[string]string{
		"path": "app/Order.php", "old": "class Order", "new": "final class Order",
		"expected_digest": fmt.Sprintf("%x", digest),
	})
	provider := &fixtureProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{ID: "read-1", Name: "workspace_read", Arguments: json.RawMessage(`{"path":"app/Order.php"}`)}}}, FinishReason: pkgProvider.FinishReasonToolCalls},
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{ID: "patch-1", Name: "workspace_patch", Arguments: arguments}}}, FinishReason: pkgProvider.FinishReasonToolCalls},
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Order updated."}, FinishReason: pkgProvider.FinishReasonStop},
	}}
	config := testConfig(root)
	configured, err := Build(config, testDependencies(provider))
	if err != nil {
		t.Fatalf("build application: %v", err)
	}
	defer configured.Close(context.Background())
	result, err := configured.ExecuteWithOptions(t.Context(), "Update Order class", ExecuteOptions{
		Approver: NewTerminalApprover(strings.NewReader("once\n"), io.Discard, true),
	})
	if err != nil {
		t.Fatalf("execute reference agent: %v", err)
	}
	if result.Session().Terminal() != pkgAgent.TerminalCompleted || result.Content() != "Order updated." || result.Session().WorkspaceGeneration() != 2 || result.Session().ContextStale() {
		t.Fatalf("unexpected run result: %#v", result)
	}
	updated, err := os.ReadFile(filename)
	if err != nil || string(updated) != "<?php\nfinal class Order {}\n" {
		t.Fatalf("unexpected workspace content %q: %v", updated, err)
	}
}

func TestApplicationIndexesLaravelSourcesWithoutPublicAssets(t *testing.T) {
	root := newLaravelWorkspace(t)
	for name, content := range map[string]string{
		"README.md":    "# Workspace\n",
		"dataset.json": `{"version":"1.0.0"}`,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "public"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "public", "bundle.js"),
		bytes.Repeat([]byte("x"), maxTestAssetBytes),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "resources", "views"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "resources", "views", "large.blade.php"),
		bytes.Repeat([]byte("v"), (1<<20)+1),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	provider := &fixtureProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
		Message:      pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Read-only analysis complete."},
		FinishReason: pkgProvider.FinishReasonStop,
	}}}
	config := testConfig(root)
	config.Agent.Tools = []string{"workspace.read"}
	config.Policy.WorkspaceMutation = "deny"
	configured, err := Build(config, testDependencies(provider))
	if err != nil {
		t.Fatalf("build application: %v", err)
	}
	defer configured.Close(context.Background())
	result, err := configured.Execute(t.Context(), "Explain Order")
	if err != nil || result.Session().Terminal() != pkgAgent.TerminalCompleted {
		t.Fatalf("execute real-workspace-shaped fixture: result=%#v err=%v", result, err)
	}
	snapshot, found := configured.Runtime().ContextEngine().Snapshot("laravel")
	if !found {
		t.Fatal("Laravel context snapshot is missing")
	}
	if _, found := snapshot.Document("app/Order.php"); !found {
		t.Fatal("application source was not indexed")
	}
	if _, found := snapshot.Document("resources/views/large.blade.php"); !found {
		t.Fatal("bounded Laravel view was not indexed")
	}
	for _, path := range []pkgContext.DocumentPath{"README.md", "dataset.json"} {
		if _, found := snapshot.Document(path); !found {
			t.Fatalf("root context document %q was not indexed", path)
		}
	}
	if _, found := snapshot.Document("public/bundle.js"); found {
		t.Fatal("generated public asset was indexed")
	}
}

func TestApplicationExecutesPromptedMutationWithTerminalApprover(t *testing.T) {
	root := newLaravelWorkspace(t)
	filename := filepath.Join(root, "app", "Order.php")
	original, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(original)
	arguments, _ := json.Marshal(map[string]string{
		"path": "app/Order.php", "old": "class Order", "new": "final class Order",
		"expected_digest": fmt.Sprintf("%x", digest),
	})
	provider := &fixtureProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{
			{ID: "read-1", Name: "workspace_read", Arguments: json.RawMessage(`{"path":"app/Order.php"}`)},
			{ID: "patch-premature", Name: "workspace_patch", Arguments: arguments},
		}}, FinishReason: pkgProvider.FinishReasonToolCalls},
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Premature completion."}, FinishReason: pkgProvider.FinishReasonStop},
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{ID: "patch-1", Name: "workspace_patch", Arguments: arguments}}}, FinishReason: pkgProvider.FinishReasonToolCalls},
		{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Approved update."}, FinishReason: pkgProvider.FinishReasonStop},
	}}
	config := testConfig(root)
	config.Policy.WorkspaceMutation = "prompt"
	configured, err := Build(config, testDependencies(provider))
	if err != nil {
		t.Fatal(err)
	}
	defer configured.Close(context.Background())
	var approvalOutput bytes.Buffer
	result, err := configured.ExecuteWithOptions(t.Context(), "Update Order", ExecuteOptions{
		Approver: NewTerminalApprover(strings.NewReader("once\n"), &approvalOutput, true),
	})
	if err != nil || result.Session().Terminal() != pkgAgent.TerminalCompleted || result.Content() != "Approved update." {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if !strings.Contains(approvalOutput.String(), "workspace.mutate resource=app/Order.php workspace=laravel") ||
		!strings.Contains(approvalOutput.String(), "expected_sha256:") ||
		!strings.Contains(approvalOutput.String(), "-class Order {}") ||
		!strings.Contains(approvalOutput.String(), "+final class Order {}") {
		t.Fatalf("unsafe or incomplete approval output: %q", approvalOutput.String())
	}
	if strings.Count(approvalOutput.String(), "workspace.mutate resource=app/Order.php workspace=laravel") != 1 {
		t.Fatalf("premature patch reached approval: %q", approvalOutput.String())
	}
}

func TestProductPolicyRejectsTargetsAndPromptsForConfiguredMutation(t *testing.T) {
	config := testConfig(t.TempDir())
	config.Policy.WorkspaceMutation = "prompt"
	policy, err := NewProductPolicy(config)
	if err != nil {
		t.Fatal(err)
	}
	target, _ := pkgTool.NewModelTarget("ollama", "fixture-model")
	request, _ := pkgTool.NewModelPermissionRequest("policy.test", "run-1", target, nil)
	decision, err := policy.Decide(t.Context(), request)
	if err != nil || decision.Kind() != pkgTool.DecisionAllow {
		t.Fatalf("expected model allow, got %#v %v", decision, err)
	}

	invocation, _ := pkgTool.NewInvocation("workspace.patch", "call-1", "run-1", json.RawMessage(`{"path":"app/Order.php"}`))
	action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceMutate, "app/Order.php", "laravel")
	prepared, _ := pkgTool.NewPreparedInvocation(invocation, "1.0.0", invocation.Arguments(), []pkgTool.Action{action})
	toolRequest, _ := pkgTool.NewToolPermissionRequest("policy.test", prepared)
	decision, err = policy.Decide(t.Context(), toolRequest)
	if err != nil || decision.Kind() != pkgTool.DecisionPrompt {
		t.Fatalf("expected mutation prompt, got %#v %v", decision, err)
	}
	config.Policy.WorkspaceMutation = "deny"
	config.Agent.Tools = []string{"workspace.read"}
	denyPolicy, err := NewProductPolicy(config)
	if err != nil {
		t.Fatal(err)
	}
	decision, err = denyPolicy.Decide(t.Context(), toolRequest)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny {
		t.Fatalf("expected configured mutation deny, got %#v %v", decision, err)
	}

	otherTarget, _ := pkgTool.NewModelTarget("ollama", "other-model")
	otherRequest, _ := pkgTool.NewModelPermissionRequest("policy.test", "run-1", otherTarget, nil)
	decision, err = policy.Decide(t.Context(), otherRequest)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny {
		t.Fatalf("expected mismatched target deny, got %#v %v", decision, err)
	}
}

func TestDoctorChecksProviderModelAndLaravelWithoutModelInvocation(t *testing.T) {
	provider := &fixtureProvider{id: "ollama"}
	checks := Doctor(t.Context(), testConfig(newLaravelWorkspace(t)), testDependencies(provider))
	if len(checks) != 9 {
		t.Fatalf("unexpected checks: %#v", checks)
	}
	for _, check := range checks {
		if check.Status != CheckPass {
			t.Fatalf("check did not pass: %#v", check)
		}
	}
	provider.mu.Lock()
	completionCalls := provider.completionCalls
	provider.mu.Unlock()
	if completionCalls != 0 {
		t.Fatalf("doctor invoked the model %d times", completionCalls)
	}
}

func TestDoctorContinuesLocalLaravelCheckAfterProviderProbeFailure(t *testing.T) {
	provider := &fixtureProvider{id: "ollama", inspectError: fmt.Errorf("offline")}
	checks := Doctor(t.Context(), testConfig(newLaravelWorkspace(t)), testDependencies(provider))
	statuses := make(map[string]CheckStatus, len(checks))
	for _, check := range checks {
		statuses[check.Name] = check.Status
	}
	if statuses["provider"] != CheckFail || statuses["model"] != CheckFail || statuses["laravel"] != CheckPass {
		t.Fatalf("independent checks were not preserved: %#v", checks)
	}
}

type fixtureProvider struct {
	mu              sync.Mutex
	id              pkgProvider.ID
	responses       []pkgProvider.CompletionResponse
	completionCalls int
	inspectError    error
}

const maxTestAssetBytes = 3 << 20

func (provider *fixtureProvider) ID() pkgProvider.ID { return provider.id }

func (provider *fixtureProvider) Complete(_ context.Context, _ pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.completionCalls++
	if len(provider.responses) == 0 {
		return pkgProvider.CompletionResponse{}, fmt.Errorf("no fixture response")
	}
	response := provider.responses[0]
	provider.responses = provider.responses[1:]
	return response, nil
}

func (provider *fixtureProvider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	return []pkgProvider.ModelInfo{{Model: pkgProvider.Model{ID: "fixture-model", Name: "Fixture"}, State: pkgProvider.ModelStateLoaded}}, nil
}

func (provider *fixtureProvider) Models(context.Context) ([]pkgProvider.Model, error) {
	return []pkgProvider.Model{{ID: "fixture-model", Name: "Fixture"}}, nil
}

func (provider *fixtureProvider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	if provider.inspectError != nil {
		return pkgProvider.CapabilityReport{}, provider.inspectError
	}
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{Capability: capability, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable})
	}
	return pkgProvider.CapabilityReport{Provider: provider.id, Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

func testDependencies(provider pkgProvider.Provider) Dependencies {
	return Dependencies{
		Getenv:          func(string) string { return "" },
		ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil },
		RunID:           func() (pkgAgent.RunID, error) { return "run-product", nil },
	}
}

func testConfig(root string) productconfig.Config {
	return productconfig.Config{
		Version:   productconfig.Version,
		Provider:  productconfig.ProviderConfig{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: productconfig.Duration{Duration: time.Minute}},
		Models:    productconfig.ModelsConfig{Chat: "fixture-model"},
		Workspace: productconfig.WorkspaceConfig{ID: "laravel", Root: filepath.Clean(root), Framework: "laravel"},
		Agent:     productconfig.AgentConfig{ID: "agent.reference", Tools: []string{"workspace.read", "workspace.patch"}},
		Policy:    productconfig.PolicyConfig{ID: "policy.test", Model: "allow", WorkspaceInspect: "allow", WorkspaceMutation: "prompt"},
		Limits:    productconfig.LimitsConfig{Duration: productconfig.Duration{Duration: time.Minute}, ModelTurns: 5, ToolCalls: 4, ToolCallsPerTurn: 2, PlanSteps: 3, PlanRevisions: 2, ToolResultBytes: 65536, SessionBytes: 1048576, InputTokens: 10000, OutputTokens: 10000},
		Context:   productconfig.ContextConfig{Retrieval: "lexical", TopK: 5, MaxTokens: 1024, ReservedTokens: 128, SafetyTokens: 64},
	}
}

func newLaravelWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"artisan":       "#!/usr/bin/env php\n",
		"composer.json": `{"require":{"laravel/framework":"^12.0"}}`,
		"app/Order.php": "<?php\nclass Order {}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
