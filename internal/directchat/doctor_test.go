package directchat

import (
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestDoctorProbesDirectChatWithoutCompletion(t *testing.T) {
	provider := validProvider()
	checks := Doctor(t.Context(), directConfig(t.TempDir()), Dependencies{ProviderFactory: fixtureFactory(provider)})
	if len(checks) != 5 {
		t.Fatalf("unexpected checks: %#v", checks)
	}
	for _, check := range checks {
		if check.Status != CheckPass {
			t.Fatalf("check did not pass: %#v", check)
		}
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if provider.inspectCalls != 1 || len(provider.requests) != 0 || len(provider.streamRequests) != 0 {
		t.Fatalf("doctor invoked generation: inspect=%d complete=%d stream=%d", provider.inspectCalls, len(provider.requests), len(provider.streamRequests))
	}
}

func TestDoctorFailsClosedForWorkspaceAndCapabilities(t *testing.T) {
	config := directConfig(t.TempDir())
	config.Workspace.Root += "/missing"
	provider := validProvider()
	checks := Doctor(t.Context(), config, Dependencies{ProviderFactory: fixtureFactory(provider)})
	if checks[1].Status != CheckFail || checks[2].Status != CheckSkip || provider.inspectCalls != 0 {
		t.Fatalf("workspace failure did not stop preflight: %#v", checks)
	}

	config = directConfig(t.TempDir())
	provider = validProvider()
	provider.capabilities = []pkgProvider.CapabilityDescriptor{{
		Capability: pkgProvider.CapabilityCompletion,
		Support:    pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityUnavailable,
	}}
	checks = Doctor(t.Context(), config, Dependencies{ProviderFactory: fixtureFactory(provider)})
	if checks[3].Status != CheckFail || checks[4].Status != CheckSkip || len(provider.requests) != 0 {
		t.Fatalf("capability failure did not fail closed: %#v", checks)
	}
}
