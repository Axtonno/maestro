package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestProviderErrorMatchesStableSentinels(t *testing.T) {
	tests := []struct {
		kind     ErrorKind
		sentinel error
	}{
		{ErrorKindInvalidRequest, ErrInvalidRequest},
		{ErrorKindAuthentication, ErrAuthentication},
		{ErrorKindModelNotFound, ErrModelNotFound},
		{ErrorKindCapabilityNotFound, ErrCapabilityNotFound},
		{ErrorKindCapabilityNotFound, ErrUnsupportedCapability},
		{ErrorKindUnavailable, ErrProviderUnavailable},
		{ErrorKindCapacityExhausted, ErrCapacityExhausted},
		{ErrorKindRateLimited, ErrRateLimited},
		{ErrorKindTransient, ErrTransient},
		{ErrorKindInvalidResponse, ErrInvalidResponse},
		{ErrorKindInternal, ErrProviderInternal},
	}

	for _, test := range tests {
		t.Run(string(test.kind)+"/"+test.sentinel.Error(), func(t *testing.T) {
			err := NewProviderError(ProviderErrorDetails{Kind: test.kind}, nil)
			if !errors.Is(err, test.sentinel) {
				t.Fatalf("expected %v to match %v", err, test.sentinel)
			}
		})
	}
}

func TestProviderErrorPreservesCauseAndMetadata(t *testing.T) {
	cause := errors.New("socket failure")
	err := NewProviderError(ProviderErrorDetails{
		Kind:       ErrorKindUnavailable,
		Operation:  OperationCompletion,
		Provider:   ID("ollama"),
		Model:      "gemma4",
		StatusCode: 503,
		Retryable:  true,
		RemoteType: " unavailable_error ",
		RemoteCode: " E_BUSY ",
		Message:    " service\n temporarily\t unavailable ",
	}, cause)

	if !errors.Is(err, cause) {
		t.Fatalf("expected cause to remain discoverable, got %v", err)
	}
	var classified *ProviderError
	if !errors.As(err, &classified) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if classified.Operation != OperationCompletion ||
		classified.Provider != ID("ollama") || classified.Model != "gemma4" ||
		classified.StatusCode != 503 || !classified.Retryable ||
		classified.RemoteType != "unavailable_error" ||
		classified.RemoteCode != "E_BUSY" ||
		classified.Message != "service temporarily unavailable" {
		t.Fatalf("unexpected metadata: %#v", classified)
	}
	if text := err.Error(); !strings.Contains(text, "ollama") ||
		!strings.Contains(text, "completion") ||
		!strings.Contains(text, "gemma4") ||
		!strings.Contains(text, "503") {
		t.Fatalf("error text lacks bounded context: %q", text)
	}
}

func TestProviderErrorPreservesContextIdentity(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		err := NewProviderError(ProviderErrorDetails{Kind: ErrorKindCanceled}, cause)
		if !errors.Is(err, cause) {
			t.Fatalf("expected context identity %v, got %v", cause, err)
		}
	}
}

func TestProviderErrorBoundsRemoteDetails(t *testing.T) {
	message := strings.Repeat("é\n", maxProviderErrorDetailBytes)
	err := NewProviderError(ProviderErrorDetails{Message: message}, nil)

	if len(err.Message) > maxProviderErrorDetailBytes {
		t.Fatalf("message has %d bytes", len(err.Message))
	}
	if !utf8.ValidString(err.Message) {
		t.Fatalf("message is not valid UTF-8: %q", err.Message)
	}
	if strings.ContainsAny(err.Message, "\r\n\t") {
		t.Fatalf("message contains control whitespace: %q", err.Message)
	}
}
