package ollama

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const maxRemoteErrorDetailBytes = 512

type transportError struct {
	cause error
}

func (e *transportError) Error() string {
	return fmt.Sprintf("Ollama transport error: %v", e.cause)
}

func (e *transportError) Unwrap() error {
	return e.cause
}

func classifyOllamaError(
	operation pkgProvider.Operation,
	model string,
	err error,
) error {
	if err == nil || errors.Is(err, io.EOF) {
		return err
	}

	var classified *pkgProvider.ProviderError
	if errors.As(err, &classified) {
		return err
	}

	details := pkgProvider.ProviderErrorDetails{
		Operation: operation,
		Provider:  providerID,
		Model:     model,
	}

	switch {
	case errors.Is(err, context.Canceled):
		details.Kind = pkgProvider.ErrorKindCanceled
	case errors.Is(err, context.DeadlineExceeded):
		details.Kind = pkgProvider.ErrorKindDeadlineExceeded
	case errors.Is(err, pkgProvider.ErrInvalidRequest):
		details.Kind = pkgProvider.ErrorKindInvalidRequest
	case errors.Is(err, pkgProvider.ErrInvalidStream):
		details.Kind = pkgProvider.ErrorKindInvalidRequest
	case errors.Is(err, pkgProvider.ErrUnsupportedCapability):
		details.Kind = pkgProvider.ErrorKindCapabilityNotFound
	case errors.Is(err, pkgProvider.ErrInvalidResponse):
		details.Kind = pkgProvider.ErrorKindInvalidResponse
	default:
		var remote *apiError
		var transport *transportError
		switch {
		case errors.As(err, &remote):
			details.StatusCode = remote.statusCode
			details.Message = remote.message
			details.Kind, details.Retryable = classifyOllamaStatus(
				operation,
				model,
				remote.statusCode,
			)
		case errors.As(err, &transport):
			var networkError net.Error
			if errors.As(transport.cause, &networkError) && networkError.Timeout() {
				details.Kind = pkgProvider.ErrorKindTransient
			} else {
				details.Kind = pkgProvider.ErrorKindUnavailable
			}
			details.Retryable = true
			details.Message = "transport request failed"
		default:
			details.Kind = pkgProvider.ErrorKindInternal
		}
	}

	return pkgProvider.NewProviderError(details, err)
}

func classifyOllamaStatus(
	operation pkgProvider.Operation,
	model string,
	statusCode int,
) (pkgProvider.ErrorKind, bool) {
	switch statusCode {
	case 0:
		return pkgProvider.ErrorKindInternal, false
	case http.StatusBadRequest,
		http.StatusRequestEntityTooLarge,
		http.StatusUnprocessableEntity:
		return pkgProvider.ErrorKindInvalidRequest, false
	case http.StatusUnauthorized, http.StatusForbidden:
		return pkgProvider.ErrorKindAuthentication, false
	case http.StatusNotFound:
		if model != "" && operation != pkgProvider.OperationModelListing &&
			operation != pkgProvider.OperationModelDiscovery {
			return pkgProvider.ErrorKindModelNotFound, false
		}

		return pkgProvider.ErrorKindCapabilityNotFound, false
	case http.StatusRequestTimeout:
		return pkgProvider.ErrorKindTransient, true
	case http.StatusConflict:
		return pkgProvider.ErrorKindUnavailable, true
	case http.StatusTooManyRequests:
		return pkgProvider.ErrorKindRateLimited, true
	case http.StatusNotImplemented:
		return pkgProvider.ErrorKindCapabilityNotFound, false
	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return pkgProvider.ErrorKindUnavailable, true
	case http.StatusInsufficientStorage:
		return pkgProvider.ErrorKindCapacityExhausted, true
	default:
		if statusCode >= 500 {
			return pkgProvider.ErrorKindInternal, true
		}
		if statusCode >= 400 {
			return pkgProvider.ErrorKindInvalidRequest, false
		}

		return pkgProvider.ErrorKindInternal, false
	}
}

func boundedRemoteErrorDetail(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= maxRemoteErrorDetailBytes {
		return value
	}

	value = value[:maxRemoteErrorDetailBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}

	return value
}
