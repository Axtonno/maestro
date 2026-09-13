package directchat

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type v4Provider struct {
	id       pkgProvider.ID
	models   []pkgProvider.ModelInfo
	requests []pkgProvider.CompletionRequest
	unloads  []string
}

func (provider *v4Provider) ID() pkgProvider.ID {
	if provider.id == "" {
		return "ollama"
	}
	return provider.id
}
func (provider *v4Provider) Complete(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	provider.requests = append(provider.requests, request)
	return pkgProvider.CompletionResponse{Model: request.Model, Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Observed facts\nReady\n\nPossible inferences\nNone\n\nInformation not determinable\nNone"}, FinishReason: pkgProvider.FinishReasonStop}, nil
}
func (provider *v4Provider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	return append([]pkgProvider.ModelInfo(nil), provider.models...), nil
}
func (provider *v4Provider) UnloadModel(_ context.Context, request pkgProvider.ModelUnloadRequest) error {
	provider.unloads = append(provider.unloads, request.Model)
	for i := range provider.models {
		if provider.models[i].Model.ID == request.Model {
			provider.models[i].State = pkgProvider.ModelStateAvailable
		}
	}
	return nil
}
func (provider *v4Provider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	descriptors := []pkgProvider.CapabilityDescriptor{}
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{Capability: capability, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable})
	}
	return pkgProvider.CapabilityReport{Provider: provider.ID(), Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

func TestLlamaCPPGGUFChatUsesServerBoundControlsWithoutHandoff(t *testing.T) {
	root := t.TempDir()
	config := llamaCPPGGUFChatConfig(root)
	provider := &v4Provider{id: "llama.cpp", models: []pkgProvider.ModelInfo{{
		Model:  pkgProvider.Model{ID: productconfig.QualifiedLlamaCPPGGUFModel},
		Digest: productconfig.QualifiedLlamaCPPGGUFDigest, State: pkgProvider.ModelStateLoaded,
	}}}
	service, err := Build(config, Dependencies{ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Execute(context.Background(), Request{Question: "Are you ready?"}); err != nil {
		t.Fatal(err)
	}
	if len(provider.requests) != 1 || len(provider.unloads) != 0 {
		t.Fatalf("unexpected routing: requests=%#v unloads=%v", provider.requests, provider.unloads)
	}
	request := provider.requests[0]
	if request.Model != productconfig.QualifiedLlamaCPPGGUFModel || request.KeepAlive != 0 ||
		request.Options.ContextWindow != 0 || request.Options.Thinking != "" ||
		request.Options.MaxTokens != 1024 {
		t.Fatalf("llama.cpp controls drifted: %#v", request)
	}
}

func TestV4ChatUsesOnlyChatModelAndUnloadsMutationProfile(t *testing.T) {
	root := t.TempDir()
	config := v4ChatConfig(root)
	provider := &v4Provider{models: []pkgProvider.ModelInfo{{Model: pkgProvider.Model{ID: productconfig.QualifiedDirectChatModel}, Digest: productconfig.QualifiedDirectChatDigest, State: pkgProvider.ModelStateAvailable}, {Model: pkgProvider.Model{ID: productconfig.QualifiedMutationModel}, Digest: productconfig.QualifiedMutationDigest, State: pkgProvider.ModelStateLoaded}}}
	service, err := Build(config, Dependencies{ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil }})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Execute(context.Background(), Request{Question: "Are you ready?"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Model != productconfig.QualifiedDirectChatModel || len(provider.requests) != 1 || provider.requests[0].Model != productconfig.QualifiedDirectChatModel || len(provider.unloads) != 1 || provider.unloads[0] != productconfig.QualifiedMutationModel {
		t.Fatalf("routing drift: result=%#v requests=%#v unloads=%v", result, provider.requests, provider.unloads)
	}
}

func v4ChatConfig(root string) productconfig.Config {
	return productconfig.Config{Version: 4, Provider: productconfig.ProviderConfig{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: productconfig.Duration{Duration: 5 * time.Minute}}, Workspace: productconfig.WorkspaceConfig{ID: "laravel", Root: root, Framework: "laravel"}, DirectChat: productconfig.ChatProfileConfig{ProfileConfig: productconfig.ProfileConfig{Model: productconfig.QualifiedDirectChatModel, Timeout: productconfig.Duration{Duration: 5 * time.Minute}, NumCtx: 4096, Thinking: productconfig.ThinkingDisabled}, Digest: productconfig.QualifiedDirectChatDigest, NumPredict: 1024, Residency: productconfig.Duration{Duration: 5 * time.Minute}, MaxFileBytes: 1 << 20, MaxOutputBytes: 1 << 20}, ControlledMutation: productconfig.ControlledMutationProfileConfig{Enabled: true, Model: productconfig.QualifiedMutationModel, Digest: productconfig.QualifiedMutationDigest, Timeout: productconfig.Duration{Duration: 5 * time.Minute}, NumCtx: 4096, NumPredict: 1024, Thinking: productconfig.ThinkingDisabled, Residency: productconfig.Duration{Duration: 5 * time.Minute}, Prompt: productconfig.MutationPromptID, PromptSHA256: productconfig.MutationPromptSHA256, Schema: productconfig.MutationSchemaID, SchemaSHA256: productconfig.MutationSchemaSHA256, MaxOutputBytes: 1 << 20}}
}

func llamaCPPGGUFChatConfig(root string) productconfig.Config {
	config := v4ChatConfig(root)
	config.Provider = productconfig.ProviderConfig{
		ID: "llama.cpp", BaseURL: "http://127.0.0.1:18080",
		Timeout:     productconfig.Duration{Duration: 5 * time.Minute},
		ModelPath:   filepath.Join(root, "model.gguf"),
		ServerBuild: productconfig.QualifiedLlamaCPPServerBuild,
	}
	config.DirectChat.Model = productconfig.QualifiedLlamaCPPGGUFModel
	config.DirectChat.Digest = productconfig.QualifiedLlamaCPPGGUFDigest
	config.DirectChat.Thinking = productconfig.ThinkingDefault
	config.DirectChat.Residency.Duration = 0
	config.ControlledMutation.Model = productconfig.QualifiedLlamaCPPGGUFModel
	config.ControlledMutation.Digest = productconfig.QualifiedLlamaCPPGGUFDigest
	config.ControlledMutation.Thinking = productconfig.ThinkingDefault
	config.ControlledMutation.Residency.Duration = 0
	return config
}
