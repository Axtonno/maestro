package llamacpp

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestModelsMapsLoadedModels(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/models" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		writeJSON(t, writer, modelsResponse{Data: []modelData{
			{ID: "chat-model"},
			{ID: "embed-model"},
		}})
	})

	models, err := provider.Models(context.Background())
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	want := []pkgProvider.Model{
		{ID: "chat-model", Name: "chat-model"},
		{ID: "embed-model", Name: "embed-model"},
	}
	if !reflect.DeepEqual(models, want) {
		t.Fatalf("expected %#v, got %#v", want, models)
	}
}

func TestModelsAcceptsEmptyList(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelsResponse{})
	})

	models, err := provider.Models(context.Background())
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected empty list, got %#v", models)
	}
}

func TestModelsRejectsInvalidIdentity(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelsResponse{Data: []modelData{{ID: " bad"}}})
	})

	_, err := provider.Models(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestModelsHandlesMalformedResponse(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, _ = writer.Write([]byte("not-json"))
	})

	_, err := provider.Models(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestDiscoverModelsMapsRouterStatesAndMetadata(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodGet || request.URL.Path != "/models" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		writeJSON(t, writer, modelsResponse{Data: []modelData{
			{
				ID:     "loaded-model",
				Path:   "/models/loaded.gguf",
				Status: modelStatusData{Value: "loaded"},
				Meta:   modelMeta{Size: 100, ContextLen: 8192},
			},
			{ID: "available-model", Status: modelStatusData{Value: "unloaded"}},
			{ID: "sleeping-model", Status: modelStatusData{Value: "sleeping"}},
			{ID: "failed-model", Status: modelStatusData{
				Value: "unloaded", Failed: true,
			}},
			{ID: "future-state", Status: modelStatusData{Value: "new-state"}},
		}})
	})

	infos, err := provider.DiscoverModels(context.Background())
	if err != nil {
		t.Fatalf("discover models: %v", err)
	}
	want := []pkgProvider.ModelInfo{
		{
			Model:         pkgProvider.Model{ID: "loaded-model", Name: "loaded-model"},
			State:         pkgProvider.ModelStateLoaded,
			SizeBytes:     100,
			ContextLength: 8192,
			Format:        "gguf",
		},
		{
			Model: pkgProvider.Model{ID: "available-model", Name: "available-model"},
			State: pkgProvider.ModelStateAvailable,
		},
		{
			Model: pkgProvider.Model{ID: "sleeping-model", Name: "sleeping-model"},
			State: pkgProvider.ModelStateSleeping,
		},
		{
			Model: pkgProvider.Model{ID: "failed-model", Name: "failed-model"},
			State: pkgProvider.ModelStateFailed,
		},
		{
			Model: pkgProvider.Model{ID: "future-state", Name: "future-state"},
			State: pkgProvider.ModelStateUnknown,
		},
	}
	if !reflect.DeepEqual(infos, want) {
		t.Fatalf("expected %#v, got %#v", want, infos)
	}
}

func TestDiscoverModelsRejectsDuplicateIdentity(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelsResponse{Data: []modelData{
			{ID: "duplicate"},
			{ID: "duplicate"},
		}})
	})

	_, err := provider.DiscoverModels(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestDiscoverModelsRejectsNegativeMetadata(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelsResponse{Data: []modelData{{
			ID: "invalid", Meta: modelMeta{Size: -1},
		}}})
	})

	_, err := provider.DiscoverModels(context.Background())
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}
