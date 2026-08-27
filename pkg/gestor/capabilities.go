package gestor

import "slices"

const (
	CapabilityAgentPlanning                CapabilityID = "agent.planning"
	CapabilityAgentRun                     CapabilityID = "agent.run"
	CapabilityAgentWorkspaceAware          CapabilityID = "agent.workspace-aware"
	CapabilityProviderCompletion           CapabilityID = "provider.completion"
	CapabilityProviderContextWindowControl CapabilityID = "provider.context_window_control"
	CapabilityProviderEmbedding            CapabilityID = "provider.embedding"
	CapabilityProviderModelDiscovery       CapabilityID = "provider.model_discovery"
	CapabilityProviderModelListing         CapabilityID = "provider.model_listing"
	CapabilityProviderModelLoad            CapabilityID = "provider.model_load"
	CapabilityProviderModelPull            CapabilityID = "provider.model_pull"
	CapabilityProviderModelRemove          CapabilityID = "provider.model_remove"
	CapabilityProviderModelUnload          CapabilityID = "provider.model_unload"
	CapabilityProviderStreaming            CapabilityID = "provider.streaming"
	CapabilityProviderStructuredOutput     CapabilityID = "provider.structured_output"
	CapabilityProviderThinking             CapabilityID = "provider.thinking"
	CapabilityProviderThinkingControl      CapabilityID = "provider.thinking_control"
	CapabilityProviderToolCalling          CapabilityID = "provider.tool_calling"
	CapabilityRuntimeConfigure             CapabilityID = "runtime.configure"
	CapabilityRuntimeHealth                CapabilityID = "runtime.health"
	CapabilityRuntimeInitialize            CapabilityID = "runtime.initialize"
	CapabilityRuntimeReload                CapabilityID = "runtime.reload"
	CapabilityRuntimeStart                 CapabilityID = "runtime.start"
	CapabilityRuntimeStop                  CapabilityID = "runtime.stop"
	CapabilityToolInvoke                   CapabilityID = "tool.invoke"
)

var knownCapabilities = [...]CapabilityID{
	CapabilityAgentPlanning,
	CapabilityAgentRun,
	CapabilityAgentWorkspaceAware,
	CapabilityProviderCompletion,
	CapabilityProviderContextWindowControl,
	CapabilityProviderEmbedding,
	CapabilityProviderModelDiscovery,
	CapabilityProviderModelListing,
	CapabilityProviderModelLoad,
	CapabilityProviderModelPull,
	CapabilityProviderModelRemove,
	CapabilityProviderModelUnload,
	CapabilityProviderStreaming,
	CapabilityProviderStructuredOutput,
	CapabilityProviderThinking,
	CapabilityProviderThinkingControl,
	CapabilityProviderToolCalling,
	CapabilityRuntimeConfigure,
	CapabilityRuntimeHealth,
	CapabilityRuntimeInitialize,
	CapabilityRuntimeReload,
	CapabilityRuntimeStart,
	CapabilityRuntimeStop,
	CapabilityToolInvoke,
}

// KnownCapabilities returns a defensive copy in lexical order. This order is
// canonical for listing and diagnostics, never an implicit ranking.
func KnownCapabilities() []CapabilityID {
	return slices.Clone(knownCapabilities[:])
}
