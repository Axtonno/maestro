package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestCompleteMapsRequestAndResponse(t *testing.T) {
	provider := newTestProvider(t, "", "secret", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost ||
			request.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected authorization header")
		}

		decoded := chatRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if decoded.Model != "local-chat" || decoded.Stream || decoded.N != 1 ||
			decoded.StreamOptions != nil {
			t.Errorf("unexpected chat request: %#v", decoded)
		}
		if len(decoded.Messages) != 2 ||
			decoded.Messages[0].Role != "system" ||
			decoded.Messages[1].Content != "Hello" {
			t.Errorf("unexpected messages: %#v", decoded.Messages)
		}

		writeJSON(t, writer, chatResponse{
			Model: "local-chat",
			Choices: []chatChoice{{
				Index: 0,
				Message: chatMessage{
					Role:    "assistant",
					Content: "Hi",
				},
				FinishReason: stringPointer("stop"),
			}},
			Usage: &usage{PromptTokens: 12, CompletionTokens: 3},
		})
	})

	response, err := provider.Complete(
		context.Background(),
		pkgProvider.CompletionRequest{
			Model: "local-chat",
			Messages: []pkgProvider.Message{
				{Role: pkgProvider.RoleSystem, Content: "Be concise"},
				{Role: pkgProvider.RoleUser, Content: "Hello"},
			},
		},
	)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	if response.Model != "local-chat" ||
		response.Message.Role != pkgProvider.RoleAssistant ||
		response.Message.Content != "Hi" ||
		response.FinishReason != "stop" ||
		response.Usage.InputTokens != 12 ||
		response.Usage.OutputTokens != 3 {
		t.Fatalf("unexpected completion response: %#v", response)
	}
}

func TestCompleteUsesDefaultModel(t *testing.T) {
	provider := newTestProvider(t, "local-default", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		_ = json.NewDecoder(request.Body).Decode(&decoded)
		if decoded.Model != "local-default" {
			t.Errorf("expected default model, got %q", decoded.Model)
		}

		writeJSON(t, writer, chatResponse{Choices: []chatChoice{{
			FinishReason: stringPointer("stop"),
		}}})
	})

	if _, err := provider.Complete(
		context.Background(),
		pkgProvider.CompletionRequest{},
	); err != nil {
		t.Fatalf("complete with default model: %v", err)
	}
}

func TestCompleteRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  error
	}{
		{
			name: "HTTP API error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNotFound)
				writeJSON(t, writer, map[string]any{
					"error": map[string]any{"message": "model not found"},
				})
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
			name: "missing choice",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, chatResponse{})
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "non-final choice",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, chatResponse{Choices: []chatChoice{{}}})
			},
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "payload error",
			handler: func(writer http.ResponseWriter, _ *http.Request) {
				writeJSON(t, writer, map[string]any{
					"error": map[string]any{"message": "generation failed"},
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "local", "", test.handler)
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
	provider := newTestProvider(t, "local", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, chatResponse{})
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Complete(ctx, pkgProvider.CompletionRequest{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
