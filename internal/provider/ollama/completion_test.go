package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestCompleteMapsRequestAndResponse(t *testing.T) {
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/chat" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content type %q", request.Header.Get("Content-Type"))
		}

		decoded := chatRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if decoded.Model != "gemma4" || decoded.Stream {
			t.Errorf("unexpected chat request: %#v", decoded)
		}
		if len(decoded.Messages) != 2 ||
			decoded.Messages[0].Role != "system" ||
			decoded.Messages[1].Content != "Hello" {
			t.Errorf("unexpected messages: %#v", decoded.Messages)
		}

		writeJSON(t, writer, chatResponse{
			Model: "gemma4",
			Message: chatMessage{
				Role:    "assistant",
				Content: "Hi",
			},
			Done:            true,
			DoneReason:      "stop",
			PromptEvalCount: 12,
			EvalCount:       3,
		})
	})

	response, err := provider.Complete(
		context.Background(),
		pkgProvider.CompletionRequest{
			Model: "gemma4",
			Messages: []pkgProvider.Message{
				{Role: pkgProvider.RoleSystem, Content: "Be concise"},
				{Role: pkgProvider.RoleUser, Content: "Hello"},
			},
		},
	)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	if response.Model != "gemma4" ||
		response.Message.Role != pkgProvider.RoleAssistant ||
		response.Message.Content != "Hi" ||
		response.FinishReason != "stop" ||
		response.Usage.InputTokens != 12 ||
		response.Usage.OutputTokens != 3 {
		t.Fatalf("unexpected completion response: %#v", response)
	}
}

func TestCompleteUsesDefaultModel(t *testing.T) {
	provider := newTestProvider(t, "qwen3", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		_ = json.NewDecoder(request.Body).Decode(&decoded)
		if decoded.Model != "qwen3" {
			t.Errorf("expected default model, got %q", decoded.Model)
		}

		writeJSON(t, writer, chatResponse{Done: true})
	})

	if _, err := provider.Complete(
		context.Background(),
		pkgProvider.CompletionRequest{},
	); err != nil {
		t.Fatalf("complete with default model: %v", err)
	}
}

func TestCompleteHandlesInvalidResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  error
	}{
		{
			name: "HTTP API error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNotFound)
				writeJSON(t, writer, errorResponse{Error: "model not found"})
			},
		},
		{
			name: "malformed JSON",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				_, _ = writer.Write([]byte("{"))
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "non-final response",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, chatResponse{Done: false})
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "payload error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, chatResponse{Error: "generation failed"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "gemma4", test.handler)
			_, err := provider.Complete(
				context.Background(),
				pkgProvider.CompletionRequest{},
			)
			if err == nil {
				t.Fatal("expected completion error")
			}
			if test.target != nil && !errors.Is(err, test.target) {
				t.Fatalf("expected %v, got %v", test.target, err)
			}
		})
	}
}

func TestCompletePreservesContextCancellation(t *testing.T) {
	provider := newTestProvider(t, "gemma4", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, chatResponse{Done: true})
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Complete(ctx, pkgProvider.CompletionRequest{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
