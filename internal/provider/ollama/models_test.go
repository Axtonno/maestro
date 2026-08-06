package ollama

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestModelsMapsOllamaModels(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/tags" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}

		writeJSON(t, writer, tagsResponse{Models: []modelResponse{
			{Name: "gemma4:latest", Model: "gemma4:latest"},
			{Name: "qwen3:8b"},
			{Model: "embeddinggemma"},
		}})
	})

	models, err := provider.Models(context.Background())
	if err != nil {
		t.Fatalf("list models: %v", err)
	}

	want := []pkgProvider.Model{
		{ID: "gemma4:latest", Name: "gemma4:latest"},
		{ID: "qwen3:8b", Name: "qwen3:8b"},
		{ID: "embeddinggemma", Name: "embeddinggemma"},
	}
	if !reflect.DeepEqual(models, want) {
		t.Fatalf("expected models %#v, got %#v", want, models)
	}
}

func TestModelsAcceptsEmptyList(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, tagsResponse{})
	})

	models, err := provider.Models(context.Background())
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected empty model list, got %#v", models)
	}
}

func TestModelsRejectsMissingIdentity(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, tagsResponse{
			Models: []modelResponse{{}},
		})
	})

	_, err := provider.Models(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestModelsHandlesMalformedAndPayloadErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  error
	}{
		{
			name: "malformed",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				_, _ = writer.Write([]byte("not-json"))
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "payload error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, tagsResponse{Error: "list failed"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "", test.handler)
			_, err := provider.Models(context.Background())
			if err == nil {
				t.Fatal("expected model listing error")
			}
			if test.target != nil && !errors.Is(err, test.target) {
				t.Fatalf("expected %v, got %v", test.target, err)
			}
		})
	}
}

func TestDiscoverModelsMergesAvailableAndRunningModels(t *testing.T) {
	modifiedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2026, 8, 6, 18, 0, 0, 0, time.UTC)
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch request.URL.Path {
		case "/api/tags":
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{
				{
					Name:       "qwen:latest",
					Model:      "qwen:latest",
					ModifiedAt: modifiedAt,
					Size:       100,
					Digest:     "sha256:qwen",
					Details: modelDetails{
						Format:        "gguf",
						Family:        "qwen",
						ParameterSize: "8B",
						Quantization:  "Q4_K_M",
					},
				},
				{Name: "embed:latest", Model: "embed:latest", Size: 50},
			}})
		case "/api/ps":
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{
				{
					Name:          "qwen:latest",
					Model:         "qwen:latest",
					ExpiresAt:     expiresAt,
					SizeVRAM:      80,
					ContextLength: 8192,
				},
				{
					Name:          "runtime-only",
					Model:         "runtime-only",
					Size:          25,
					SizeVRAM:      20,
					ContextLength: 4096,
				},
			}})
		default:
			t.Errorf("unexpected path %q", request.URL.Path)
		}
	})

	infos, err := provider.DiscoverModels(context.Background())
	if err != nil {
		t.Fatalf("discover models: %v", err)
	}

	want := []pkgProvider.ModelInfo{
		{
			Model:         pkgProvider.Model{ID: "qwen:latest", Name: "qwen:latest"},
			State:         pkgProvider.ModelStateLoaded,
			Digest:        "sha256:qwen",
			SizeBytes:     100,
			VRAMBytes:     80,
			ContextLength: 8192,
			Format:        "gguf",
			Family:        "qwen",
			ParameterSize: "8B",
			Quantization:  "Q4_K_M",
			ModifiedAt:    modifiedAt,
			ExpiresAt:     expiresAt,
		},
		{
			Model:     pkgProvider.Model{ID: "embed:latest", Name: "embed:latest"},
			State:     pkgProvider.ModelStateAvailable,
			SizeBytes: 50,
		},
		{
			Model:         pkgProvider.Model{ID: "runtime-only", Name: "runtime-only"},
			State:         pkgProvider.ModelStateLoaded,
			SizeBytes:     25,
			VRAMBytes:     20,
			ContextLength: 4096,
		},
	}
	if !reflect.DeepEqual(infos, want) {
		t.Fatalf("expected model info %#v, got %#v", want, infos)
	}
}

func TestDiscoverModelsRejectsDuplicateRunningModels(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path == "/api/tags" {
			writeJSON(t, writer, tagsResponse{})
			return
		}

		writeJSON(t, writer, tagsResponse{Models: []modelResponse{
			{Name: "qwen"},
			{Name: "qwen"},
		}})
	})

	_, err := provider.DiscoverModels(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestDiscoverModelsRejectsNegativeMetadata(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path == "/api/tags" {
			writeJSON(t, writer, tagsResponse{Models: []modelResponse{{
				Name: "qwen", Size: -1,
			}}})
			return
		}
		writeJSON(t, writer, tagsResponse{})
	})

	_, err := provider.DiscoverModels(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestDiscoverModelsPropagatesRunningListFailure(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path == "/api/tags" {
			writeJSON(t, writer, tagsResponse{})
			return
		}

		writer.WriteHeader(http.StatusInternalServerError)
		writeJSON(t, writer, errorResponse{Error: "ps failed"})
	})

	if _, err := provider.DiscoverModels(context.Background()); err == nil {
		t.Fatal("expected discovery error")
	}
}
