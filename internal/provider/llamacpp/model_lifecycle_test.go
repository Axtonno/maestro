package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestModelLifecycleMapsLoadAndUnloadRequests(t *testing.T) {
	paths := make([]string, 0, 2)
	models := make([]string, 0, 2)
	provider := newTestProvider(t, "default-model", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		paths = append(paths, request.URL.Path)
		decoded := modelLifecycleRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode lifecycle request: %v", err)
		}
		models = append(models, decoded.Model)
		writeJSON(t, writer, modelLifecycleResponse{Success: true})
	})

	if err := provider.LoadModel(
		context.Background(),
		pkgProvider.ModelLoadRequest{},
	); err != nil {
		t.Fatalf("load default model: %v", err)
	}
	if err := provider.UnloadModel(
		context.Background(),
		pkgProvider.ModelUnloadRequest{Model: "explicit-model"},
	); err != nil {
		t.Fatalf("unload explicit model: %v", err)
	}

	if !reflect.DeepEqual(paths, []string{"/models/load", "/models/unload"}) {
		t.Fatalf("unexpected lifecycle paths: %#v", paths)
	}
	if !reflect.DeepEqual(models, []string{"default-model", "explicit-model"}) {
		t.Fatalf("unexpected lifecycle models: %#v", models)
	}
}

func TestModelLifecycleRejectsUnsuccessfulResponse(t *testing.T) {
	provider := newTestProvider(t, "model", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelLifecycleResponse{})
	})

	err := provider.LoadModel(
		context.Background(),
		pkgProvider.ModelLoadRequest{},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}
