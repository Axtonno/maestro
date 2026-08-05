package ollama

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

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
