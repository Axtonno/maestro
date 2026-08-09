package gestor

import (
	"errors"
	"testing"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestRuntimeCapabilityMappingIsComplete(t *testing.T) {
	tests := []struct {
		capability pkgRuntime.Capability
		want       pkgGestor.CapabilityID
	}{
		{pkgRuntime.CapabilityConfigure, pkgGestor.CapabilityRuntimeConfigure},
		{pkgRuntime.CapabilityInitialize, pkgGestor.CapabilityRuntimeInitialize},
		{pkgRuntime.CapabilityStart, pkgGestor.CapabilityRuntimeStart},
		{pkgRuntime.CapabilityStop, pkgGestor.CapabilityRuntimeStop},
		{pkgRuntime.CapabilityReload, pkgGestor.CapabilityRuntimeReload},
		{pkgRuntime.CapabilityHealth, pkgGestor.CapabilityRuntimeHealth},
	}

	for _, test := range tests {
		got, err := runtimeCapabilityID(test.capability)
		if err != nil {
			t.Fatalf("map runtime capability %q: %v", test.capability, err)
		}
		if got != test.want {
			t.Errorf("runtime capability %q: expected %q, got %q", test.capability, test.want, got)
		}
	}

	_, err := runtimeCapabilityID("future")
	if !errors.Is(err, pkgGestor.ErrInvalidCapabilityID) {
		t.Fatalf("unknown runtime capability: expected ErrInvalidCapabilityID, got %v", err)
	}
}

func TestProviderCapabilityMappingIsComplete(t *testing.T) {
	want := map[pkgProvider.Capability]pkgGestor.CapabilityID{
		pkgProvider.CapabilityCompletion:       pkgGestor.CapabilityProviderCompletion,
		pkgProvider.CapabilityStreaming:        pkgGestor.CapabilityProviderStreaming,
		pkgProvider.CapabilityEmbedding:        pkgGestor.CapabilityProviderEmbedding,
		pkgProvider.CapabilityModelListing:     pkgGestor.CapabilityProviderModelListing,
		pkgProvider.CapabilityModelDiscovery:   pkgGestor.CapabilityProviderModelDiscovery,
		pkgProvider.CapabilityModelLoad:        pkgGestor.CapabilityProviderModelLoad,
		pkgProvider.CapabilityModelUnload:      pkgGestor.CapabilityProviderModelUnload,
		pkgProvider.CapabilityModelPull:        pkgGestor.CapabilityProviderModelPull,
		pkgProvider.CapabilityModelRemove:      pkgGestor.CapabilityProviderModelRemove,
		pkgProvider.CapabilityStructuredOutput: pkgGestor.CapabilityProviderStructuredOutput,
		pkgProvider.CapabilityToolCalling:      pkgGestor.CapabilityProviderToolCalling,
	}

	known := pkgProvider.KnownCapabilities()
	if len(known) != len(want) {
		t.Fatalf("mapping fixture has %d capabilities, provider exposes %d", len(want), len(known))
	}
	for _, capability := range known {
		got, err := providerCapabilityID(capability)
		if err != nil {
			t.Fatalf("map provider capability %q: %v", capability, err)
		}
		if got != want[capability] {
			t.Errorf("provider capability %q: expected %q, got %q", capability, want[capability], got)
		}
	}

	_, err := providerCapabilityID("future")
	if !errors.Is(err, pkgGestor.ErrInvalidCapabilityID) {
		t.Fatalf("unknown provider capability: expected ErrInvalidCapabilityID, got %v", err)
	}
}
