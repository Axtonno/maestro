package ollama

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestCompleteClassifiesOllamaHTTPFailures(t *testing.T) {
	tests := []struct {
		status    int
		kind      pkgProvider.ErrorKind
		sentinel  error
		retryable bool
	}{
		{400, pkgProvider.ErrorKindInvalidRequest, pkgProvider.ErrInvalidRequest, false},
		{401, pkgProvider.ErrorKindAuthentication, pkgProvider.ErrAuthentication, false},
		{404, pkgProvider.ErrorKindModelNotFound, pkgProvider.ErrModelNotFound, false},
		{429, pkgProvider.ErrorKindRateLimited, pkgProvider.ErrRateLimited, true},
		{500, pkgProvider.ErrorKindInternal, pkgProvider.ErrProviderInternal, true},
		{502, pkgProvider.ErrorKindUnavailable, pkgProvider.ErrProviderUnavailable, true},
		{507, pkgProvider.ErrorKindCapacityExhausted, pkgProvider.ErrCapacityExhausted, true},
	}

	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			provider := newTestProvider(t, "gemma4", func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				writeJSON(t, writer, map[string]string{"error": " remote\n failure "})
			})

			_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
			assertOllamaProviderError(t, err, test.kind, test.status, test.retryable)
			if !errors.Is(err, test.sentinel) {
				t.Fatalf("expected %v, got %v", test.sentinel, err)
			}
		})
	}
}

func TestCompleteClassifiesOllamaTransportAndContextFailures(t *testing.T) {
	transportCause := errors.New("connection refused")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportCause
	})}
	provider, err := New("http://ollama.test", "gemma4", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	_, err = provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertOllamaProviderError(t, err, pkgProvider.ErrorKindUnavailable, 0, true)
	if !errors.Is(err, transportCause) {
		t.Fatalf("transport cause was not preserved: %v", err)
	}

	timeoutClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, testTimeoutError{}
	})}
	timeoutProvider, err := New("http://ollama.test", "gemma4", timeoutClient)
	if err != nil {
		t.Fatalf("create timeout provider: %v", err)
	}
	_, err = timeoutProvider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertOllamaProviderError(t, err, pkgProvider.ErrorKindTransient, 0, true)
	if !errors.Is(err, pkgProvider.ErrTransient) {
		t.Fatalf("expected ErrTransient, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Complete(ctx, pkgProvider.CompletionRequest{})
	assertOllamaProviderError(t, err, pkgProvider.ErrorKindCanceled, 0, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context identity was not preserved: %v", err)
	}

	err = classifyOllamaError(
		pkgProvider.OperationCompletion,
		"gemma4",
		context.DeadlineExceeded,
	)
	assertOllamaProviderError(t, err, pkgProvider.ErrorKindDeadlineExceeded, 0, false)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline identity was not preserved: %v", err)
	}
}

type testTimeoutError struct{}

func (testTimeoutError) Error() string   { return "request timed out" }
func (testTimeoutError) Timeout() bool   { return true }
func (testTimeoutError) Temporary() bool { return true }

func TestCompleteClassifiesMalformedOllamaResponse(t *testing.T) {
	provider := newTestProvider(t, "gemma4", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte("{"))
	})

	_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertOllamaProviderError(t, err, pkgProvider.ErrorKindInvalidResponse, 0, false)
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func TestOllamaRemoteMessageIsBounded(t *testing.T) {
	provider := newTestProvider(t, "gemma4", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		writeJSON(t, writer, map[string]string{"error": strings.Repeat("x\n", 1000)})
	})

	_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	var classified *pkgProvider.ProviderError
	if !errors.As(err, &classified) {
		t.Fatalf("expected ProviderError, got %v", err)
	}
	if len(classified.Message) > 512 || strings.Contains(classified.Message, "\n") {
		t.Fatalf("unbounded remote message: %q", classified.Message)
	}
}

func assertOllamaProviderError(
	t *testing.T,
	err error,
	kind pkgProvider.ErrorKind,
	status int,
	retryable bool,
) {
	t.Helper()
	var classified *pkgProvider.ProviderError
	if !errors.As(err, &classified) {
		t.Fatalf("expected ProviderError, got %T: %v", err, err)
	}
	if classified.Kind != kind || classified.Operation != pkgProvider.OperationCompletion ||
		classified.Provider != providerID || classified.Model != "gemma4" ||
		classified.StatusCode != status || classified.Retryable != retryable {
		t.Fatalf("unexpected classification: %#v", classified)
	}
}
