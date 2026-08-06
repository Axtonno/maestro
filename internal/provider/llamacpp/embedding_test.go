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

func TestEmbedMapsAndOrdersMultipleInputs(t *testing.T) {
	provider := newTestProvider(t, "embed-model", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/embeddings" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		decoded := embeddingRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if decoded.Model != "embed-model" ||
			decoded.EncodingFormat != "float" ||
			!reflect.DeepEqual(decoded.Input, []string{"first", "second"}) {
			t.Errorf("unexpected embedding request: %#v", decoded)
		}

		writeJSON(t, writer, embeddingResponse{
			Model: "embed-model",
			Data: []embeddingData{
				{Index: 1, Embedding: []float32{0.3, 0.4}},
				{Index: 0, Embedding: []float32{0.1, 0.2}},
			},
			Usage: usage{PromptTokens: 4},
		})
	})

	response, err := provider.Embed(
		context.Background(),
		pkgProvider.EmbeddingRequest{Inputs: []string{"first", "second"}},
	)
	if err != nil {
		t.Fatalf("embed: %v", err)
	}

	if response.Model != "embed-model" ||
		response.Usage.InputTokens != 4 ||
		!reflect.DeepEqual(response.Embeddings, [][]float32{
			{0.1, 0.2},
			{0.3, 0.4},
		}) {
		t.Fatalf("unexpected embedding response: %#v", response)
	}
}

func TestEmbedRejectsEmptyInputBeforeRequest(t *testing.T) {
	called := false
	provider := newTestProvider(t, "embed-model", "", func(
		_ http.ResponseWriter,
		_ *http.Request,
	) {
		called = true
	})

	_, err := provider.Embed(
		context.Background(),
		pkgProvider.EmbeddingRequest{},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
	if called {
		t.Fatal("adapter made an HTTP request for empty input")
	}
}

func TestEmbedValidatesResponseShape(t *testing.T) {
	tests := []struct {
		name string
		data []embeddingData
	}{
		{name: "missing embedding"},
		{name: "wrong cardinality", data: []embeddingData{
			{Index: 0, Embedding: []float32{1}},
			{Index: 1, Embedding: []float32{2}},
		}},
		{name: "empty vector", data: []embeddingData{{Index: 0}}},
		{name: "index out of range", data: []embeddingData{{
			Index: 1, Embedding: []float32{1},
		}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "embed-model", "", func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writeJSON(t, writer, embeddingResponse{Data: test.data})
			})
			_, err := provider.Embed(
				context.Background(),
				pkgProvider.EmbeddingRequest{Inputs: []string{"input"}},
			)
			if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
				t.Fatalf("expected ErrInvalidResponse, got %v", err)
			}
		})
	}
}

func TestEmbedRejectsDuplicateIndexes(t *testing.T) {
	provider := newTestProvider(t, "embed-model", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, embeddingResponse{Data: []embeddingData{
			{Index: 0, Embedding: []float32{1}},
			{Index: 0, Embedding: []float32{2}},
		}})
	})

	_, err := provider.Embed(
		context.Background(),
		pkgProvider.EmbeddingRequest{Inputs: []string{"first", "second"}},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestEmbedHandlesPayloadError(t *testing.T) {
	provider := newTestProvider(t, "embed-model", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, map[string]any{
			"error": map[string]any{"message": "embedding failed"},
		})
	})

	_, err := provider.Embed(
		context.Background(),
		pkgProvider.EmbeddingRequest{Inputs: []string{"input"}},
	)
	if err == nil {
		t.Fatal("expected embedding error")
	}
}
