package llamacpp

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

func TestNewBuildsLlamaCPPProviderWithDefaults(t *testing.T) {
	provider, err := New(Config{})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if provider.ID() != ID {
		t.Fatalf("expected llama.cpp ID, got %q", provider.ID())
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
		{name: "negative timeout", config: Config{Timeout: -time.Second}},
		{name: "base URL with v1 path", config: Config{
			BaseURL: "http://localhost:8080/v1",
		}},
		{name: "invalid scheme", config: Config{
			BaseURL: "ftp://localhost:8080",
		}},
		{name: "invalid default model", config: Config{
			DefaultModel: " local",
		}},
		{name: "invalid API key", config: Config{APIKey: " secret"}},
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

func TestNewUsesInjectedHTTPClientAndAPIKey(t *testing.T) {
	called := false
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		called = true
		if request.URL.String() != "http://llamacpp.test/v1/models" {
			t.Errorf("unexpected URL %q", request.URL.String())
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected authorization header")
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
			Request:    request,
		}, nil
	})}
	provider, err := New(Config{
		BaseURL:    "http://llamacpp.test",
		Timeout:    time.Nanosecond,
		APIKey:     "secret",
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
