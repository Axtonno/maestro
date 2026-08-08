package provider

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidProvider        = errors.New("invalid provider")
	ErrAlreadyRegistered      = errors.New("provider already registered")
	ErrNotFound               = errors.New("provider not found")
	ErrDefaultNotConfigured   = errors.New("default provider not configured")
	ErrUnsupportedCapability  = errors.New("provider capability not supported")
	ErrInvalidStream          = errors.New("invalid provider stream")
	ErrInvalidRequest         = errors.New("invalid provider request")
	ErrInvalidResponse        = errors.New("invalid provider response")
	ErrInvalidResidencyPolicy = errors.New("invalid model residency policy")
	ErrResidencyShuttingDown  = errors.New("model residency is shutting down")
	ErrAuthentication         = errors.New("provider authentication failed")
	ErrModelNotFound          = errors.New("provider model not found")
	ErrCapabilityNotFound     = errors.New("provider capability not found")
	ErrProviderUnavailable    = errors.New("provider unavailable")
	ErrCapacityExhausted      = errors.New("provider capacity exhausted")
	ErrRateLimited            = errors.New("provider rate limited")
	ErrTransient              = errors.New("transient provider error")
	ErrProviderInternal       = errors.New("provider internal error")
)

const maxProviderErrorDetailBytes = 512

type ErrorKind string

const (
	ErrorKindInvalidRequest     ErrorKind = "invalid_request"
	ErrorKindAuthentication     ErrorKind = "authentication"
	ErrorKindModelNotFound      ErrorKind = "model_not_found"
	ErrorKindCapabilityNotFound ErrorKind = "capability_not_found"
	ErrorKindUnavailable        ErrorKind = "unavailable"
	ErrorKindCapacityExhausted  ErrorKind = "capacity_exhausted"
	ErrorKindRateLimited        ErrorKind = "rate_limited"
	ErrorKindTransient          ErrorKind = "transient"
	ErrorKindInvalidResponse    ErrorKind = "invalid_response"
	ErrorKindInternal           ErrorKind = "internal"
	ErrorKindCanceled           ErrorKind = "canceled"
	ErrorKindDeadlineExceeded   ErrorKind = "deadline_exceeded"
)

type Operation string

const (
	OperationCompletion              Operation = "completion"
	OperationStreaming               Operation = "streaming"
	OperationEmbedding               Operation = "embedding"
	OperationModelListing            Operation = "model_listing"
	OperationModelDiscovery          Operation = "model_discovery"
	OperationModelLoad               Operation = "model_load"
	OperationModelUnload             Operation = "model_unload"
	OperationModelPull               Operation = "model_pull"
	OperationModelRemove             Operation = "model_remove"
	OperationCapabilityIntrospection Operation = "capability_introspection"
	OperationResidencyPolicy         Operation = "residency_policy"
	OperationProviderShutdown        Operation = "provider_shutdown"
)

type ProviderErrorDetails struct {
	Kind       ErrorKind
	Operation  Operation
	Provider   ID
	Model      string
	StatusCode int
	Retryable  bool
	RemoteType string
	RemoteCode string
	Message    string
}

// ProviderError is the stable provider-neutral error envelope. Cause remains
// available through errors.Is/errors.As and Unwrap.
type ProviderError struct {
	Kind       ErrorKind
	Operation  Operation
	Provider   ID
	Model      string
	StatusCode int
	Retryable  bool
	RemoteType string
	RemoteCode string
	Message    string

	cause error
}

func NewProviderError(
	details ProviderErrorDetails,
	cause error,
) *ProviderError {
	return &ProviderError{
		Kind:       details.Kind,
		Operation:  details.Operation,
		Provider:   details.Provider,
		Model:      sanitizeProviderErrorDetail(details.Model),
		StatusCode: details.StatusCode,
		Retryable:  details.Retryable,
		RemoteType: sanitizeProviderErrorDetail(details.RemoteType),
		RemoteCode: sanitizeProviderErrorDetail(details.RemoteCode),
		Message:    sanitizeProviderErrorDetail(details.Message),
		cause:      cause,
	}
}

func (e *ProviderError) Error() string {
	if e == nil {
		return "<nil>"
	}

	message := fmt.Sprintf(
		"provider %q operation %q failed (%s)",
		e.Provider,
		e.Operation,
		e.Kind,
	)
	if e.Model != "" {
		message += fmt.Sprintf(" for model %q", e.Model)
	}
	if e.StatusCode != 0 {
		message += fmt.Sprintf(" with status %d", e.StatusCode)
	}
	if e.Message != "" {
		message += ": " + e.Message
	}

	return message
}

func (e *ProviderError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

func (e *ProviderError) Is(target error) bool {
	if e == nil {
		return false
	}

	switch e.Kind {
	case ErrorKindInvalidRequest:
		return target == ErrInvalidRequest
	case ErrorKindAuthentication:
		return target == ErrAuthentication
	case ErrorKindModelNotFound:
		return target == ErrModelNotFound
	case ErrorKindCapabilityNotFound:
		return target == ErrCapabilityNotFound ||
			target == ErrUnsupportedCapability
	case ErrorKindUnavailable:
		return target == ErrProviderUnavailable
	case ErrorKindCapacityExhausted:
		return target == ErrCapacityExhausted
	case ErrorKindRateLimited:
		return target == ErrRateLimited
	case ErrorKindTransient:
		return target == ErrTransient
	case ErrorKindInvalidResponse:
		return target == ErrInvalidResponse
	case ErrorKindInternal:
		return target == ErrProviderInternal
	default:
		return false
	}
}

func sanitizeProviderErrorDetail(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= maxProviderErrorDetailBytes {
		return value
	}

	value = value[:maxProviderErrorDetailBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}

	return value
}
