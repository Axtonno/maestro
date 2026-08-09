package gestor

import (
	"fmt"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func runtimeCapabilityID(capability pkgRuntime.Capability) (pkgGestor.CapabilityID, error) {
	switch capability {
	case pkgRuntime.CapabilityConfigure:
		return pkgGestor.CapabilityRuntimeConfigure, nil
	case pkgRuntime.CapabilityInitialize:
		return pkgGestor.CapabilityRuntimeInitialize, nil
	case pkgRuntime.CapabilityStart:
		return pkgGestor.CapabilityRuntimeStart, nil
	case pkgRuntime.CapabilityStop:
		return pkgGestor.CapabilityRuntimeStop, nil
	case pkgRuntime.CapabilityReload:
		return pkgGestor.CapabilityRuntimeReload, nil
	case pkgRuntime.CapabilityHealth:
		return pkgGestor.CapabilityRuntimeHealth, nil
	default:
		custom := pkgGestor.CapabilityID(capability)
		if err := custom.Validate(); err == nil {
			return custom, nil
		}

		return "", fmt.Errorf("runtime capability %q has no Gestor mapping: %w", capability, pkgGestor.ErrInvalidCapabilityID)
	}
}

func providerCapabilityID(capability pkgProvider.Capability) (pkgGestor.CapabilityID, error) {
	switch capability {
	case pkgProvider.CapabilityCompletion:
		return pkgGestor.CapabilityProviderCompletion, nil
	case pkgProvider.CapabilityStreaming:
		return pkgGestor.CapabilityProviderStreaming, nil
	case pkgProvider.CapabilityEmbedding:
		return pkgGestor.CapabilityProviderEmbedding, nil
	case pkgProvider.CapabilityModelListing:
		return pkgGestor.CapabilityProviderModelListing, nil
	case pkgProvider.CapabilityModelDiscovery:
		return pkgGestor.CapabilityProviderModelDiscovery, nil
	case pkgProvider.CapabilityModelLoad:
		return pkgGestor.CapabilityProviderModelLoad, nil
	case pkgProvider.CapabilityModelUnload:
		return pkgGestor.CapabilityProviderModelUnload, nil
	case pkgProvider.CapabilityModelPull:
		return pkgGestor.CapabilityProviderModelPull, nil
	case pkgProvider.CapabilityModelRemove:
		return pkgGestor.CapabilityProviderModelRemove, nil
	case pkgProvider.CapabilityStructuredOutput:
		return pkgGestor.CapabilityProviderStructuredOutput, nil
	case pkgProvider.CapabilityToolCalling:
		return pkgGestor.CapabilityProviderToolCalling, nil
	default:
		return "", fmt.Errorf("provider capability %q has no Gestor mapping: %w", capability, pkgGestor.ErrInvalidCapabilityID)
	}
}
