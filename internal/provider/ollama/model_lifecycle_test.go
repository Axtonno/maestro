package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestModelLifecycleMapsLoadAndUnloadRequests(t *testing.T) {
	requests := make([]modelLifecycleRequest, 0, 2)
	provider := newTestProvider(t, "default-model", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/generate" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}

		decoded := modelLifecycleRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode lifecycle request: %v", err)
		}
		requests = append(requests, decoded)
		writeJSON(t, writer, modelLifecycleResponse{Model: decoded.Model, Done: true})
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

	if len(requests) != 2 {
		t.Fatalf("expected two lifecycle requests, got %d", len(requests))
	}
	if requests[0].Model != "default-model" || requests[0].Stream ||
		requests[0].KeepAlive != -1 {
		t.Fatalf("unexpected load request: %#v", requests[0])
	}
	if requests[1].Model != "explicit-model" || requests[1].Stream ||
		requests[1].KeepAlive != 0 {
		t.Fatalf("unexpected unload request: %#v", requests[1])
	}
}

func TestModelLifecycleRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  error
	}{
		{
			name: "non-final response",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, modelLifecycleResponse{})
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "payload error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, modelLifecycleResponse{Error: "load failed"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "model", test.handler)
			err := provider.LoadModel(
				context.Background(),
				pkgProvider.ModelLoadRequest{},
			)
			if err == nil {
				t.Fatal("expected lifecycle error")
			}
			if test.target != nil && !errors.Is(err, test.target) {
				t.Fatalf("expected %v, got %v", test.target, err)
			}
		})
	}
}
