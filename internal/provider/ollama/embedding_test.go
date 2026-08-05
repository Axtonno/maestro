package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestEmbedMapsMultipleInputsAndResponse(t *testing.T) {
	provider := newTestProvider(t, "embeddinggemma", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/embed" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}

		decoded := embedRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode embed request: %v", err)
		}
		if decoded.Model != "embeddinggemma" || !reflect.DeepEqual(
			decoded.Input,
			[]string{"first", "second"},
		) {
			t.Errorf("unexpected embed request: %#v", decoded)
		}

		writeJSON(t, writer, embedResponse{
			Model: "embeddinggemma",
			Embeddings: [][]float32{
				{0.1, 0.2},
				{0.3, 0.4},
			},
			PromptEvalCount: 4,
		})
	})

	response, err := provider.Embed(
		context.Background(),
		pkgProvider.EmbeddingRequest{
			Inputs: []string{"first", "second"},
		},
	)
	if err != nil {
		t.Fatalf("embed: %v", err)
	}

	if response.Model != "embeddinggemma" ||
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
	provider := newTestProvider(t, "embeddinggemma", func(
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
		name       string
		embeddings [][]float32
	}{
		{name: "missing embedding"},
		{name: "wrong cardinality", embeddings: [][]float32{{1}, {2}}},
		{name: "empty vector", embeddings: [][]float32{{}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "embeddinggemma", func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writeJSON(t, writer, embedResponse{
					Embeddings: test.embeddings,
				})
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

func TestEmbedHandlesMalformedAndPayloadErrors(t *testing.T) {
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
				writeJSON(t, writer, embedResponse{Error: "embedding failed"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "embeddinggemma", test.handler)
			_, err := provider.Embed(
				context.Background(),
				pkgProvider.EmbeddingRequest{Inputs: []string{"input"}},
			)
			if err == nil {
				t.Fatal("expected embedding error")
			}
			if test.target != nil && !errors.Is(err, test.target) {
				t.Fatalf("expected %v, got %v", test.target, err)
			}
		})
	}
}
