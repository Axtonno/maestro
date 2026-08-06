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
