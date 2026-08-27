package directchat

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestDirectChatDisclosesOneExplicitFileInOneToolFreeCompletion(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "routes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "routes", "api.php"), []byte("<?php Route::get('/orders');\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := validProvider()
	service := buildService(t, directConfig(root), provider)
	result, err := service.Execute(t.Context(), Request{
		Question: "Which endpoints are declared?", File: "routes/api.php",
	})
	if err != nil {
		t.Fatalf("execute direct chat: %v", err)
	}
	if result.Content != "One orders endpoint." || result.Model != "chat-model" ||
		result.RequestedNumCtx != 4096 || result.RequestedThinking != pkgProvider.ThinkingDisabled {
		t.Fatalf("unexpected result: %#v", result)
	}
	provider.mu.Lock()
	requests := append([]pkgProvider.CompletionRequest(nil), provider.requests...)
	provider.mu.Unlock()
	if len(requests) != 1 {
		t.Fatalf("expected one completion, got %d", len(requests))
	}
	request := requests[0]
	if request.Model != "chat-model" || len(request.Tools) != 0 ||
		request.ToolChoice.Mode != pkgProvider.ToolChoiceNone ||
		request.Options.ContextWindow != 4096 || request.Options.Thinking != pkgProvider.ThinkingDisabled {
		t.Fatalf("unsafe completion request: %#v", request)
	}
	encoded := messagesText(request.Messages)
	if !strings.Contains(encoded, "routes/api.php") || !strings.Contains(encoded, "Route::get") ||
		strings.Contains(encoded, root) {
		t.Fatalf("unexpected disclosure: %q", encoded)
	}
}

func TestDirectChatWithoutFileDoesNotDiscoverContext(t *testing.T) {
	provider := validProvider()
	service := buildService(t, directConfig(t.TempDir()), provider)
	if _, err := service.Execute(t.Context(), Request{Question: "What routes exist?"}); err != nil {
		t.Fatal(err)
	}
	provider.mu.Lock()
	request := provider.requests[0]
	provider.mu.Unlock()
	text := messagesText(request.Messages)
	if !strings.Contains(text, "No workspace file was supplied") || strings.Contains(text, "BEGIN WORKSPACE FILE") {
		t.Fatalf("missing-file epistemic prompt is invalid: %q", text)
	}
}

func TestDirectChatRejectsUnsafeFilesBeforeProviderIO(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "large.php"), []byte(strings.Repeat("x", 65)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "invalid.php"), []byte{0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.php")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.php")); err != nil {
		t.Fatal(err)
	}
	config := directConfig(root)
	config.Interaction.Chat.MaxFileBytes = 64
	for _, logical := range []string{"/absolute.php", "../outside.php", `dir\file.php`, "dir", "large.php", "invalid.php", "link.php", "missing.php"} {
		t.Run(strings.ReplaceAll(logical, "/", "_"), func(t *testing.T) {
			provider := validProvider()
			service := buildService(t, config, provider)
			_, err := service.Execute(t.Context(), Request{Question: "Inspect it", File: logical})
			if !errors.Is(err, ErrFileNotAllowed) {
				t.Fatalf("unsafe file %q: %v", logical, err)
			}
			provider.mu.Lock()
			calls := provider.inspectCalls + len(provider.requests)
			provider.mu.Unlock()
			if calls != 0 {
				t.Fatalf("unsafe file reached provider: %d calls", calls)
			}
		})
	}
}

func TestDirectChatFailsClosedWithoutFallback(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		provider *chatProviderStub
		want     error
	}{
		{name: "empty", provider: providerWithResponse(pkgProvider.CompletionResponse{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant}, FinishReason: pkgProvider.FinishReasonStop}), want: ErrResponseInvalid},
		{name: "tool call", provider: providerWithResponse(pkgProvider.CompletionResponse{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "x", ToolCalls: []pkgProvider.ToolCall{{Name: "workspace_read"}}}, FinishReason: pkgProvider.FinishReasonStop}), want: ErrResponseInvalid},
		{name: "length", provider: providerWithResponse(pkgProvider.CompletionResponse{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "partial"}, FinishReason: pkgProvider.FinishReasonLength}), want: ErrResponseInvalid},
		{name: "provider", provider: &chatProviderStub{id: "ollama", completeErr: errors.New("offline"), capabilities: validCapabilities()}, want: ErrProviderUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service := buildService(t, directConfig(t.TempDir()), testCase.provider)
			_, err := service.Execute(t.Context(), Request{Question: "Answer"})
			if !errors.Is(err, testCase.want) {
				t.Fatalf("expected %v, got %v", testCase.want, err)
			}
			testCase.provider.mu.Lock()
			calls := len(testCase.provider.requests)
			testCase.provider.mu.Unlock()
			if calls != 1 {
				t.Fatalf("expected no fallback after one completion, got %d", calls)
			}
		})
	}
}

func TestDirectChatRequiresV2ProfileAndGenerationCapabilities(t *testing.T) {
	v1 := directConfig(t.TempDir())
	v1.Version = productconfig.Version
	v1.Models.Chat = "legacy"
	v1.Interaction = productconfig.InteractionConfig{}
	if _, err := Build(v1, Dependencies{ProviderFactory: fixtureFactory(validProvider())}); !errors.Is(err, ErrProfileRequired) {
		t.Fatalf("v1 config enabled chat: %v", err)
	}
	provider := validProvider()
	provider.capabilities = []pkgProvider.CapabilityDescriptor{{
		Capability: pkgProvider.CapabilityCompletion,
		Support:    pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable,
	}}
	service := buildService(t, directConfig(t.TempDir()), provider)
	if _, err := service.Execute(t.Context(), Request{Question: "Answer"}); !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("missing generation capabilities were ignored: %v", err)
	}
}

type chatProviderStub struct {
	mu           sync.Mutex
	id           pkgProvider.ID
	response     pkgProvider.CompletionResponse
	completeErr  error
	capabilities []pkgProvider.CapabilityDescriptor
	requests     []pkgProvider.CompletionRequest
	inspectCalls int
}

func (provider *chatProviderStub) ID() pkgProvider.ID { return provider.id }

func (provider *chatProviderStub) Complete(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.requests = append(provider.requests, request)
	return provider.response, provider.completeErr
}

func (provider *chatProviderStub) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	provider.inspectCalls++
	return pkgProvider.CapabilityReport{Provider: provider.id, Target: request.Target, Model: request.Model, Capabilities: append([]pkgProvider.CapabilityDescriptor(nil), provider.capabilities...)}, nil
}

func validProvider() *chatProviderStub {
	return providerWithResponse(pkgProvider.CompletionResponse{
		Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "One orders endpoint."},
		FinishReason: pkgProvider.FinishReasonStop, Usage: pkgProvider.Usage{InputTokens: 12, OutputTokens: 4},
	})
}

func providerWithResponse(response pkgProvider.CompletionResponse) *chatProviderStub {
	return &chatProviderStub{id: "ollama", response: response, capabilities: validCapabilities()}
}

func validCapabilities() []pkgProvider.CapabilityDescriptor {
	return []pkgProvider.CapabilityDescriptor{
		{Capability: pkgProvider.CapabilityCompletion, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable},
		{Capability: pkgProvider.CapabilityContextWindowControl, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable},
		{Capability: pkgProvider.CapabilityThinkingControl, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable},
	}
}

func directConfig(root string) productconfig.Config {
	return productconfig.Config{
		Version:   productconfig.CandidateVersion,
		Provider:  productconfig.ProviderConfig{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: productconfig.Duration{Duration: time.Minute}},
		Workspace: productconfig.WorkspaceConfig{ID: "laravel", Root: root, Framework: "laravel"},
		Interaction: productconfig.InteractionConfig{
			Chat: productconfig.ChatProfileConfig{ProfileConfig: productconfig.ProfileConfig{
				Model: "chat-model", Timeout: productconfig.Duration{Duration: time.Minute}, NumCtx: 4096, Thinking: productconfig.ThinkingDisabled,
			}, MaxFileBytes: 1 << 20, MaxOutputBytes: 1 << 20},
			Agent: productconfig.AgentProfileConfig{ProfileConfig: productconfig.ProfileConfig{
				Model: "agent-model", Timeout: productconfig.Duration{Duration: time.Minute}, NumCtx: 8192, Thinking: productconfig.ThinkingDefault,
			}},
		},
		Models:  productconfig.ModelsConfig{Chat: "agent-model"},
		Agent:   productconfig.AgentConfig{ID: "agent.reference", Tools: []string{"workspace.read"}},
		Policy:  productconfig.PolicyConfig{ID: "policy.test", Model: "allow", WorkspaceInspect: "allow", WorkspaceMutation: "deny"},
		Limits:  productconfig.LimitsConfig{Duration: productconfig.Duration{Duration: time.Minute}, ModelTurns: 5, ToolCalls: 4, ToolCallsPerTurn: 2, PlanSteps: 3, PlanRevisions: 2, ToolResultBytes: 65536, SessionBytes: 1048576, InputTokens: 10000, OutputTokens: 10000},
		Context: productconfig.ContextConfig{Retrieval: "lexical", TopK: 2, MaxTokens: 1024, ReservedTokens: 128, SafetyTokens: 64},
	}
}

func fixtureFactory(provider pkgProvider.Provider) ProviderFactory {
	return func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil }
}

func buildService(t *testing.T, config productconfig.Config, provider pkgProvider.Provider) *Service {
	t.Helper()
	service, err := Build(config, Dependencies{ProviderFactory: fixtureFactory(provider)})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func messagesText(messages []pkgProvider.Message) string {
	var result strings.Builder
	for _, message := range messages {
		result.WriteString(message.Content)
		result.WriteByte('\n')
	}
	return result.String()
}
