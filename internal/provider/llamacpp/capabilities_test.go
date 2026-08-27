package llamacpp

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func llamaCPPDescriptor(
	t *testing.T,
	report pkgProvider.CapabilityReport,
	capability pkgProvider.Capability,
) pkgProvider.CapabilityDescriptor {
	t.Helper()

	for _, descriptor := range report.Capabilities {
		if descriptor.Capability == capability {
			return descriptor
		}
	}

	t.Fatalf("capability %q was not reported", capability)

	return pkgProvider.CapabilityDescriptor{}
}

func TestLlamaCPPAdapterCapabilitiesRequireNoIO(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		http.ResponseWriter,
		*http.Request,
	) {
		t.Fatal("adapter introspection performed remote I/O")
	})

	report, err := provider.InspectCapabilities(
		context.Background(),
		pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetAdapter},
	)
	if err != nil {
		t.Fatalf("inspect adapter: %v", err)
	}
	completion := llamaCPPDescriptor(t, report, pkgProvider.CapabilityCompletion)
	if completion.Support != pkgProvider.CapabilitySupported ||
		completion.Availability != pkgProvider.CapabilityAvailabilityUnknown {
		t.Fatalf("unexpected completion descriptor %#v", completion)
	}
	structured := llamaCPPDescriptor(
		t,
		report,
		pkgProvider.CapabilityStructuredOutput,
	)
	if structured.Support != pkgProvider.CapabilitySupported ||
		structured.Availability != pkgProvider.CapabilityAvailabilityUnknown {
		t.Fatalf("unexpected structured output descriptor %#v", structured)
	}
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityContextWindowControl,
		pkgProvider.CapabilityThinkingControl,
		pkgProvider.CapabilityThinking,
	} {
		descriptor := llamaCPPDescriptor(t, report, capability)
		if descriptor.Support != pkgProvider.CapabilityUnsupported ||
			descriptor.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
			t.Fatalf("unexpected unsupported control %q descriptor %#v", capability, descriptor)
		}
	}
}

func TestLlamaCPPInstanceCapabilitiesDistinguishRouterMode(t *testing.T) {
	tests := []struct {
		name       string
		models     []modelData
		management pkgProvider.CapabilityAvailability
	}{
		{
			name: "router",
			models: []modelData{{
				ID: "chat", Status: modelStatusData{Value: "unloaded"},
			}},
			management: pkgProvider.CapabilityAvailabilityAvailable,
		},
		{
			name:       "single model",
			models:     []modelData{{ID: "chat"}},
			management: pkgProvider.CapabilityAvailabilityUnavailable,
		},
		{
			name:       "empty catalog",
			management: pkgProvider.CapabilityAvailabilityUnknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "", "", func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Method != http.MethodGet || request.URL.Path != "/models" {
					t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
				}
				writeJSON(t, writer, modelsResponse{Data: test.models})
			})

			report, err := provider.InspectCapabilities(
				context.Background(),
				pkgProvider.CapabilityRequest{
					Target: pkgProvider.CapabilityTargetInstance,
				},
			)
			if err != nil {
				t.Fatalf("inspect instance: %v", err)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityModelLoad,
			); descriptor.Availability != test.management {
				t.Fatalf("unexpected load descriptor %#v", descriptor)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityModelDiscovery,
			); descriptor.Availability != pkgProvider.CapabilityAvailabilityAvailable {
				t.Fatalf("unexpected discovery descriptor %#v", descriptor)
			}
		})
	}
}

func TestLlamaCPPModelCapabilitiesUseEffectiveArguments(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		completion pkgProvider.CapabilityAvailability
		embedding  pkgProvider.CapabilityAvailability
		tools      pkgProvider.CapabilityAvailability
	}{
		{
			name:       "embedding process",
			args:       []string{"llama-server", "--embedding"},
			completion: pkgProvider.CapabilityAvailabilityUnavailable,
			embedding:  pkgProvider.CapabilityAvailabilityAvailable,
			tools:      pkgProvider.CapabilityAvailabilityUnavailable,
		},
		{
			name:       "chat process",
			args:       []string{"llama-server", "--no-embedding", "--jinja"},
			completion: pkgProvider.CapabilityAvailabilityAvailable,
			embedding:  pkgProvider.CapabilityAvailabilityUnavailable,
			tools:      pkgProvider.CapabilityAvailabilityAvailable,
		},
		{
			name:       "unobservable mode",
			args:       []string{"llama-server", "-ctx", "4096"},
			completion: pkgProvider.CapabilityAvailabilityUnknown,
			embedding:  pkgProvider.CapabilityAvailabilityUnknown,
			tools:      pkgProvider.CapabilityAvailabilityUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "", "", func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writeJSON(t, writer, modelsResponse{Data: []modelData{{
					ID: "model",
					Status: modelStatusData{
						Value: "loaded", Args: test.args,
					},
				}}})
			})

			report, err := provider.InspectCapabilities(
				context.Background(),
				pkgProvider.CapabilityRequest{
					Target: pkgProvider.CapabilityTargetModel, Model: "model",
				},
			)
			if err != nil {
				t.Fatalf("inspect model: %v", err)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityCompletion,
			); descriptor.Availability != test.completion {
				t.Fatalf("unexpected completion descriptor %#v", descriptor)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityEmbedding,
			); descriptor.Availability != test.embedding {
				t.Fatalf("unexpected embedding descriptor %#v", descriptor)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityToolCalling,
			); descriptor.Availability != test.tools {
				t.Fatalf("unexpected tool descriptor %#v", descriptor)
			}
			if descriptor := llamaCPPDescriptor(
				t,
				report,
				pkgProvider.CapabilityStructuredOutput,
			); descriptor.Availability != test.completion {
				t.Fatalf("unexpected structured output descriptor %#v", descriptor)
			}
		})
	}
}

func TestLlamaCPPFailedAndMissingModelsAreUnavailable(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelsResponse{Data: []modelData{{
			ID: "failed",
			Status: modelStatusData{
				Value: "unloaded", Failed: true,
				Args: []string{"llama-server", "--no-embedding"},
			},
		}}})
	})

	for _, model := range []string{"failed", "missing"} {
		report, err := provider.InspectCapabilities(
			context.Background(),
			pkgProvider.CapabilityRequest{
				Target: pkgProvider.CapabilityTargetModel, Model: model,
			},
		)
		if err != nil {
			t.Fatalf("inspect model %q: %v", model, err)
		}
		if descriptor := llamaCPPDescriptor(
			t,
			report,
			pkgProvider.CapabilityCompletion,
		); descriptor.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
			t.Fatalf("model %q: unexpected completion %#v", model, descriptor)
		}
	}
}

func TestLlamaCPPCapabilitiesReflectCatalogChangesAndAreConcurrent(t *testing.T) {
	var embedding atomic.Bool
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		argument := "--no-embedding"
		if embedding.Load() {
			argument = "--embedding"
		}
		writeJSON(t, writer, modelsResponse{Data: []modelData{{
			ID: "model",
			Status: modelStatusData{
				Value: "loaded", Args: []string{"llama-server", argument},
			},
		}}})
	})
	request := pkgProvider.CapabilityRequest{
		Target: pkgProvider.CapabilityTargetModel, Model: "model",
	}

	first, err := provider.InspectCapabilities(context.Background(), request)
	if err != nil {
		t.Fatalf("first inspection: %v", err)
	}
	embedding.Store(true)
	second, err := provider.InspectCapabilities(context.Background(), request)
	if err != nil {
		t.Fatalf("second inspection: %v", err)
	}
	if llamaCPPDescriptor(t, first, pkgProvider.CapabilityEmbedding).Availability ==
		llamaCPPDescriptor(t, second, pkgProvider.CapabilityEmbedding).Availability {
		t.Fatal("capability inspection cached stale model metadata")
	}

	const workers = 12
	var wait sync.WaitGroup
	errorsChannel := make(chan error, workers)
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := provider.InspectCapabilities(context.Background(), request)
			errorsChannel <- err
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent inspection: %v", err)
		}
	}
}
