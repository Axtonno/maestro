//go:build integration

package llamacpp_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	"github.com/antonio-cafeo/maestro/pkg/provider/llamacpp"
)

func TestLlamaCPPIntegration(t *testing.T) {
	baseURL := os.Getenv("MAESTRO_LLAMACPP_BASE_URL")
	if baseURL == "" {
		t.Skip("MAESTRO_LLAMACPP_BASE_URL is not configured")
	}

	provider, err := llamacpp.New(llamacpp.Config{
		BaseURL: baseURL,
		Timeout: 2 * time.Minute,
		APIKey:  os.Getenv("MAESTRO_LLAMACPP_API_KEY"),
	})
	if err != nil {
		t.Fatalf("create llama.cpp provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("models", func(t *testing.T) {
		if _, err := provider.Models(ctx); err != nil {
			t.Fatalf("list models: %v", err)
		}
	})

	chatModel := os.Getenv("MAESTRO_LLAMACPP_CHAT_MODEL")
	if chatModel == "" {
		t.Log("MAESTRO_LLAMACPP_CHAT_MODEL is not configured; skipping chat tests")
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

		t.Run("stream cancellation", func(t *testing.T) {
			streamContext, cancelStream := context.WithCancel(context.Background())
			stream, err := provider.Stream(
				streamContext,
				pkgProvider.CompletionRequest{
					Model: chatModel,
					Messages: []pkgProvider.Message{{
						Role:    pkgProvider.RoleUser,
						Content: "Write a detailed explanation of local AI runtimes.",
					}},
				},
			)
			if err != nil {
				cancelStream()
				t.Fatalf("open cancellable stream: %v", err)
			}

			cancelStream()
			if _, err := stream.Recv(); !errors.Is(err, context.Canceled) {
				_ = stream.Close()
				t.Fatalf("expected context cancellation, got %v", err)
			}
			if err := stream.Close(); err != nil {
				t.Fatalf("close canceled stream: %v", err)
			}
		})
	}

	embedModel := os.Getenv("MAESTRO_LLAMACPP_EMBED_MODEL")
	if embedModel == "" {
		t.Log("MAESTRO_LLAMACPP_EMBED_MODEL is not configured; skipping embed test")
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
