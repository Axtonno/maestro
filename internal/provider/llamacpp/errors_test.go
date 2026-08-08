package llamacpp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestCompleteClassifiesLlamaCPPStructuredFailures(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		remoteType string
		kind       pkgProvider.ErrorKind
		sentinel   error
		retryable  bool
	}{
		{"type overrides status", 400, "authentication_error", pkgProvider.ErrorKindAuthentication, pkgProvider.ErrAuthentication, false},
		{"unsupported", 400, "not_supported_error", pkgProvider.ErrorKindCapabilityNotFound, pkgProvider.ErrCapabilityNotFound, false},
		{"model absent", 404, "not_found_error", pkgProvider.ErrorKindModelNotFound, pkgProvider.ErrModelNotFound, false},
		{"rate limit", 429, "rate_limit_error", pkgProvider.ErrorKindRateLimited, pkgProvider.ErrRateLimited, true},
		{"unavailable", 503, "unavailable_error", pkgProvider.ErrorKindUnavailable, pkgProvider.ErrProviderUnavailable, true},
		{"capacity", 507, "capacity_error", pkgProvider.ErrorKindCapacityExhausted, pkgProvider.ErrCapacityExhausted, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "local", "", func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(test.status)
				writeJSON(t, writer, map[string]any{"error": map[string]any{
					"message": "remote failure", "type": test.remoteType, "code": "E_TEST",
				}})
			})

			_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
			classified := assertLlamaCPPProviderError(t, err, test.kind, test.status, test.retryable)
			if classified.RemoteType != test.remoteType || classified.RemoteCode != "E_TEST" {
				t.Fatalf("structured details were not preserved: %#v", classified)
			}
			if !errors.Is(err, test.sentinel) {
				t.Fatalf("expected %v, got %v", test.sentinel, err)
			}
		})
	}
}

func TestCompleteClassifiesLlamaCPPTransportContextAndMalformedFailures(t *testing.T) {
	transportCause := errors.New("connection refused")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportCause
	})}
	provider, err := New("http://llamacpp.test", "local", "", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	_, err = provider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertLlamaCPPProviderError(t, err, pkgProvider.ErrorKindUnavailable, 0, true)
	if !errors.Is(err, transportCause) {
		t.Fatalf("transport cause was not preserved: %v", err)
	}

	timeoutClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, testTimeoutError{}
	})}
	timeoutProvider, err := New("http://llamacpp.test", "local", "", timeoutClient)
	if err != nil {
		t.Fatalf("create timeout provider: %v", err)
	}
	_, err = timeoutProvider.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertLlamaCPPProviderError(t, err, pkgProvider.ErrorKindTransient, 0, true)
	if !errors.Is(err, pkgProvider.ErrTransient) {
		t.Fatalf("expected ErrTransient, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Complete(ctx, pkgProvider.CompletionRequest{})
	assertLlamaCPPProviderError(t, err, pkgProvider.ErrorKindCanceled, 0, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context identity was not preserved: %v", err)
	}

	err = classifyLlamaCPPError(
		pkgProvider.OperationCompletion,
		"local",
		context.DeadlineExceeded,
	)
	assertLlamaCPPProviderError(t, err, pkgProvider.ErrorKindDeadlineExceeded, 0, false)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline identity was not preserved: %v", err)
	}

	malformed := newTestProvider(t, "local", "", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte("{"))
	})
	_, err = malformed.Complete(context.Background(), pkgProvider.CompletionRequest{})
	assertLlamaCPPProviderError(t, err, pkgProvider.ErrorKindInvalidResponse, 0, false)
	if !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

type testTimeoutError struct{}

func (testTimeoutError) Error() string   { return "request timed out" }
func (testTimeoutError) Timeout() bool   { return true }
func (testTimeoutError) Temporary() bool { return true }

func assertLlamaCPPProviderError(
	t *testing.T,
	err error,
	kind pkgProvider.ErrorKind,
	status int,
	retryable bool,
) *pkgProvider.ProviderError {
	t.Helper()
	var classified *pkgProvider.ProviderError
	if !errors.As(err, &classified) {
		t.Fatalf("expected ProviderError, got %T: %v", err, err)
	}
	if classified.Kind != kind || classified.Operation != pkgProvider.OperationCompletion ||
		classified.Provider != providerID || classified.Model != "local" ||
		classified.StatusCode != status || classified.Retryable != retryable {
		t.Fatalf("unexpected classification: %#v", classified)
	}

	return classified
}
