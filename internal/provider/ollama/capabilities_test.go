package ollama

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func ollamaDescriptor(
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

func TestOllamaAdapterCapabilitiesRequireNoIO(t *testing.T) {
	provider := newTestProvider(t, "", func(
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
	completion := ollamaDescriptor(t, report, pkgProvider.CapabilityCompletion)
	if completion.Support != pkgProvider.CapabilitySupported ||
		completion.Availability != pkgProvider.CapabilityAvailabilityUnknown {
		t.Fatalf("unexpected completion descriptor %#v", completion)
	}
	tools := ollamaDescriptor(t, report, pkgProvider.CapabilityToolCalling)
	if tools.Support != pkgProvider.CapabilitySupported ||
		tools.Availability != pkgProvider.CapabilityAvailabilityUnknown {
		t.Fatalf("unexpected tool descriptor %#v", tools)
	}
}

func TestOllamaInstanceCapabilitiesProbeCatalog(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/tags" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		writeJSON(t, writer, tagsResponse{})
	})

	report, err := provider.InspectCapabilities(
		context.Background(),
		pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance},
	)
	if err != nil {
		t.Fatalf("inspect instance: %v", err)
	}
	listing := ollamaDescriptor(t, report, pkgProvider.CapabilityModelListing)
	if listing.Availability != pkgProvider.CapabilityAvailabilityAvailable {
		t.Fatalf("unexpected listing descriptor %#v", listing)
	}
	completion := ollamaDescriptor(t, report, pkgProvider.CapabilityCompletion)
	if completion.Availability != pkgProvider.CapabilityAvailabilityUnknown {
		t.Fatalf("unexpected instance completion descriptor %#v", completion)
	}
}

func TestOllamaModelCapabilitiesUseShowMetadata(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch request.URL.Path {
		case "/api/tags":
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{{
				Name: "qwen", Model: "qwen",
			}}})
		case "/api/show":
			if request.Method != http.MethodPost {
				t.Errorf("unexpected show method %s", request.Method)
			}
			writeJSON(t, writer, modelShowResponse{
				Capabilities: []string{"completion", "tools"},
			})
		default:
			t.Errorf("unexpected path %q", request.URL.Path)
		}
	})

	report, err := provider.InspectCapabilities(
		context.Background(),
		pkgProvider.CapabilityRequest{
			Target: pkgProvider.CapabilityTargetModel, Model: "qwen",
		},
	)
	if err != nil {
		t.Fatalf("inspect model: %v", err)
	}
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityCompletion,
		pkgProvider.CapabilityStreaming,
	} {
		if descriptor := ollamaDescriptor(t, report, capability); descriptor.Availability !=
			pkgProvider.CapabilityAvailabilityAvailable {
			t.Fatalf("unexpected %q descriptor %#v", capability, descriptor)
		}
	}
	if descriptor := ollamaDescriptor(t, report, pkgProvider.CapabilityEmbedding); descriptor.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
		t.Fatalf("unexpected embedding descriptor %#v", descriptor)
	}
	if descriptor := ollamaDescriptor(t, report, pkgProvider.CapabilityToolCalling); descriptor.Support != pkgProvider.CapabilitySupported ||
		descriptor.Availability != pkgProvider.CapabilityAvailabilityAvailable {
		t.Fatalf("unexpected tool descriptor %#v", descriptor)
	}
	if descriptor := ollamaDescriptor(t, report, pkgProvider.CapabilityStructuredOutput); descriptor.Support != pkgProvider.CapabilitySupported ||
		descriptor.Availability != pkgProvider.CapabilityAvailabilityAvailable {
		t.Fatalf("unexpected structured output descriptor %#v", descriptor)
	}
}

func TestOllamaMissingModelCapabilitiesDoNotCallShow(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path != "/api/tags" {
			t.Errorf("unexpected path %q", request.URL.Path)
		}
		writeJSON(t, writer, tagsResponse{})
	})

	report, err := provider.InspectCapabilities(
		context.Background(),
		pkgProvider.CapabilityRequest{
			Target: pkgProvider.CapabilityTargetModel, Model: "missing",
		},
	)
	if err != nil {
		t.Fatalf("inspect missing model: %v", err)
	}
	if descriptor := ollamaDescriptor(t, report, pkgProvider.CapabilityCompletion); descriptor.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
		t.Fatalf("unexpected completion descriptor %#v", descriptor)
	}
	if descriptor := ollamaDescriptor(t, report, pkgProvider.CapabilityModelPull); descriptor.Availability != pkgProvider.CapabilityAvailabilityAvailable {
		t.Fatalf("unexpected pull descriptor %#v", descriptor)
	}
}

func TestOllamaCapabilitiesReflectCatalogChangesAndAreConcurrent(t *testing.T) {
	var embedding atomic.Bool
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch request.URL.Path {
		case "/api/tags":
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{{Name: "model"}}})
		case "/api/show":
			capability := "completion"
			if embedding.Load() {
				capability = "embedding"
			}
			writeJSON(t, writer, modelShowResponse{Capabilities: []string{capability}})
		default:
			t.Errorf("unexpected path %q", request.URL.Path)
		}
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
	if ollamaDescriptor(t, first, pkgProvider.CapabilityCompletion).Availability ==
		ollamaDescriptor(t, second, pkgProvider.CapabilityCompletion).Availability {
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

func TestOllamaCapabilitiesRejectInvalidMetadata(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path == "/api/tags" {
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{{Name: "qwen"}}})
			return
		}
		writeJSON(t, writer, modelShowResponse{
			Capabilities: []string{"completion", "completion"},
		})
	})

	_, err := provider.InspectCapabilities(
		context.Background(),
		pkgProvider.CapabilityRequest{
			Target: pkgProvider.CapabilityTargetModel, Model: "qwen",
		},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}
