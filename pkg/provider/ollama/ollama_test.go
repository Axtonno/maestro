package ollama

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestNewBuildsOllamaProviderWithDefaults(t *testing.T) {
	provider, err := New(Config{})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if provider.ID() != ID {
		t.Fatalf("expected ollama ID, got %q", provider.ID())
	}

	var _ pkgProvider.Completer = provider
	var _ pkgProvider.Streamer = provider
	var _ pkgProvider.Embedder = provider
	var _ pkgProvider.ModelLister = provider
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "negative timeout",
			config: Config{
				Timeout: -time.Second,
			},
		},
		{
			name: "base URL with API path",
			config: Config{
				BaseURL: "http://localhost:11434/api",
			},
		},
		{
			name: "invalid scheme",
			config: Config{
				BaseURL: "ftp://localhost:11434",
			},
		},
		{
			name: "invalid default model",
			config: Config{
				DefaultModel: " gemma4",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("expected ErrInvalidConfig, got %v", err)
			}
		})
	}
}

func TestNewUsesInjectedHTTPClient(t *testing.T) {
	called := false
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		called = true
		if request.URL.String() != "http://ollama.test/api/tags" {
			t.Errorf("unexpected URL %q", request.URL.String())
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"models":[]}`)),
			Request:    request,
		}, nil
	})}
	provider, err := New(Config{
		BaseURL:    "http://ollama.test",
		Timeout:    time.Nanosecond,
		HTTPClient: client,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if _, err := provider.Models(context.Background()); err != nil {
		t.Fatalf("list models: %v", err)
	}
	if !called {
		t.Fatal("injected HTTP client was not used")
	}
}
