//go:build integration

package ollama_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	"github.com/antonio-cafeo/maestro/pkg/provider/ollama"
)

func TestOllamaIntegration(t *testing.T) {
	baseURL := os.Getenv("MAESTRO_OLLAMA_BASE_URL")
	if baseURL == "" {
		t.Skip("MAESTRO_OLLAMA_BASE_URL is not configured")
	}

	provider, err := ollama.New(ollama.Config{
		BaseURL: baseURL,
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("create Ollama provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("models", func(t *testing.T) {
		if _, err := provider.Models(ctx); err != nil {
			t.Fatalf("list models: %v", err)
		}
	})

	chatModel := os.Getenv("MAESTRO_OLLAMA_CHAT_MODEL")
	if chatModel == "" {
		t.Log("MAESTRO_OLLAMA_CHAT_MODEL is not configured; skipping chat tests")
	} else {
		t.Run("completion", func(t *testing.T) {
			_, err := provider.Complete(ctx, pkgProvider.CompletionRequest{
				Model: chatModel,
				Messages: []pkgProvider.Message{{
					Role:    pkgProvider.RoleUser,
					Content: "Reply with the word Maestro.",
				}},
			})
			if err != nil {
				t.Fatalf("complete: %v", err)
			}
		})

		t.Run("stream", func(t *testing.T) {
			stream, err := provider.Stream(ctx, pkgProvider.CompletionRequest{
				Model: chatModel,
				Messages: []pkgProvider.Message{{
					Role:    pkgProvider.RoleUser,
					Content: "Reply with the word Maestro.",
				}},
			})
			if err != nil {
				t.Fatalf("open stream: %v", err)
			}
			defer stream.Close()

			for {
				_, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("receive stream: %v", err)
				}
			}
		})
	}

	embedModel := os.Getenv("MAESTRO_OLLAMA_EMBED_MODEL")
	if embedModel == "" {
		t.Log("MAESTRO_OLLAMA_EMBED_MODEL is not configured; skipping embed test")
	} else {
		t.Run("embedding", func(t *testing.T) {
			_, err := provider.Embed(ctx, pkgProvider.EmbeddingRequest{
				Model:  embedModel,
				Inputs: []string{"Maestro"},
			})
			if err != nil {
				t.Fatalf("embed: %v", err)
			}
		})
	}
}
