package controlledmutation

import (
	"context"
	"os"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckFail CheckStatus = "fail"
)

type Check struct {
	Name   string
	Status CheckStatus
	Detail string
}

// Doctor performs only read-only probes. It never loads, unloads, or invokes a model.
func Doctor(ctx context.Context, config productconfig.Config, dependencies Dependencies, terminal bool) []Check {
	checks := []Check{}
	add := func(name string, ok bool, pass, fail string) {
		status, detail := CheckFail, fail
		if ok {
			status, detail = CheckPass, pass
		}
		checks = append(checks, Check{Name: name, Status: status, Detail: detail})
	}
	valid := ctx != nil && config.ValidateMutationExecutionProfile() == nil
	add("configuration", valid, "schema_v4_profiles_separated", "configuration_invalid")
	add("mutation_prompt", PromptSHA256() == productconfig.MutationPromptSHA256, "qualified_prompt_digest", "prompt_digest_mismatch")
	add("mutation_schema", SchemaSHA256() == productconfig.MutationSchemaSHA256, "qualified_schema_digest", "schema_digest_mismatch")
	add("tty", terminal, "interactive_terminal", "tty_required")
	info, err := os.Lstat(config.Workspace.Root)
	workspaceOK := valid && err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
	add("workspace", workspaceOK, "root_available", "root_unavailable")
	if !valid {
		add("provider", false, "", "configuration_invalid")
		add("direct_chat_model", false, "", "configuration_invalid")
		add("controlled_mutation_model", false, "", "configuration_invalid")
		add("capability", false, "", "configuration_invalid")
		return checks
	}
	if dependencies.Getenv == nil {
		dependencies.Getenv = os.Getenv
	}
	if dependencies.ProviderFactory == nil {
		dependencies.ProviderFactory = defaultProvider
	}
	secret, secretErr := config.Secret(dependencies.Getenv)
	candidate, providerErr := dependencies.ProviderFactory(config, secret)
	provider, ok := candidate.(mutationProvider)
	providerOK := secretErr == nil && providerErr == nil && ok && !nilValue(provider) && provider.ID() == pkgProvider.ID(config.Provider.ID)
	add("provider", providerOK, "ollama_available", "provider_unavailable")
	if !providerOK {
		add("direct_chat_model", false, "", "provider_unavailable")
		add("controlled_mutation_model", false, "", "provider_unavailable")
		add("capability", false, "", "provider_unavailable")
		return checks
	}
	models, discoverErr := provider.DiscoverModels(ctx)
	identity := func(name, digest string) bool {
		if discoverErr != nil {
			return false
		}
		for _, model := range models {
			if model.Model.ID == name {
				return model.Digest == digest
			}
		}
		return false
	}
	add("direct_chat_model", identity(config.DirectChat.Model, config.DirectChat.Digest), "qualified_model_digest", "model_or_digest_mismatch")
	add("controlled_mutation_model", identity(config.ControlledMutation.Model, config.ControlledMutation.Digest), "qualified_model_digest", "model_or_digest_mismatch")
	report, capabilityErr := provider.InspectCapabilities(ctx, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetModel, Model: config.ControlledMutation.Model})
	capabilityOK := capabilityErr == nil
	for _, capability := range []pkgProvider.Capability{pkgProvider.CapabilityCompletion, pkgProvider.CapabilityStructuredOutput, pkgProvider.CapabilityModelDiscovery, pkgProvider.CapabilityModelUnload, pkgProvider.CapabilityContextWindowControl, pkgProvider.CapabilityThinkingControl} {
		capabilityOK = capabilityOK && capabilityAvailable(report, capability)
	}
	add("capability", capabilityOK, "host_bound_mutation_available", "required_capability_unavailable")
	return checks
}
