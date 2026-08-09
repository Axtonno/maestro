package gestor

import "slices"

const (
	CapabilityProviderCompletion       CapabilityID = "provider.completion"
	CapabilityProviderEmbedding        CapabilityID = "provider.embedding"
	CapabilityProviderModelDiscovery   CapabilityID = "provider.model_discovery"
	CapabilityProviderModelListing     CapabilityID = "provider.model_listing"
	CapabilityProviderModelLoad        CapabilityID = "provider.model_load"
	CapabilityProviderModelPull        CapabilityID = "provider.model_pull"
	CapabilityProviderModelRemove      CapabilityID = "provider.model_remove"
	CapabilityProviderModelUnload      CapabilityID = "provider.model_unload"
	CapabilityProviderStreaming        CapabilityID = "provider.streaming"
	CapabilityProviderStructuredOutput CapabilityID = "provider.structured_output"
	CapabilityProviderToolCalling      CapabilityID = "provider.tool_calling"
	CapabilityRuntimeConfigure         CapabilityID = "runtime.configure"
	CapabilityRuntimeHealth            CapabilityID = "runtime.health"
	CapabilityRuntimeInitialize        CapabilityID = "runtime.initialize"
	CapabilityRuntimeReload            CapabilityID = "runtime.reload"
	CapabilityRuntimeStart             CapabilityID = "runtime.start"
	CapabilityRuntimeStop              CapabilityID = "runtime.stop"
)

var knownCapabilities = [...]CapabilityID{
	CapabilityProviderCompletion,
	CapabilityProviderEmbedding,
	CapabilityProviderModelDiscovery,
	CapabilityProviderModelListing,
	CapabilityProviderModelLoad,
	CapabilityProviderModelPull,
	CapabilityProviderModelRemove,
	CapabilityProviderModelUnload,
	CapabilityProviderStreaming,
	CapabilityProviderStructuredOutput,
	CapabilityProviderToolCalling,
	CapabilityRuntimeConfigure,
	CapabilityRuntimeHealth,
	CapabilityRuntimeInitialize,
	CapabilityRuntimeReload,
	CapabilityRuntimeStart,
	CapabilityRuntimeStop,
}

// KnownCapabilities returns a defensive copy in lexical order. This order is
// canonical for listing and diagnostics, never an implicit ranking.
func KnownCapabilities() []CapabilityID {
	return slices.Clone(knownCapabilities[:])
}
