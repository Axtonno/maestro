package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type transportError struct {
	cause error
}

func (e *transportError) Error() string {
	return fmt.Sprintf("llama.cpp transport error: %v", e.cause)
}

func (e *transportError) Unwrap() error {
	return e.cause
}

func newLlamaCPPAPIError(
	statusCode int,
	raw json.RawMessage,
) *apiError {
	detail := errorDetail{}
	_ = json.Unmarshal(raw, &detail)

	message := strings.TrimSpace(detail.Message)
	if message == "" {
		var plain string
		if json.Unmarshal(raw, &plain) == nil {
			message = strings.TrimSpace(plain)
		}
	}

	code := ""
	if detail.Code != nil {
		code = fmt.Sprint(detail.Code)
	}

	return &apiError{
		statusCode: statusCode,
		message:    message,
		errorType:  strings.TrimSpace(detail.Type),
		errorCode:  strings.TrimSpace(code),
	}
}

func hasLlamaCPPAPIError(raw json.RawMessage) bool {
	value := strings.TrimSpace(string(raw))

	return value != "" && value != "null"
}

func classifyLlamaCPPError(
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
			details.RemoteType = remote.errorType
			details.RemoteCode = remote.errorCode
			details.Message = remote.message
			details.Kind, details.Retryable = classifyLlamaCPPRemote(
				operation,
				model,
				remote,
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

func classifyLlamaCPPRemote(
	operation pkgProvider.Operation,
	model string,
	remote *apiError,
) (pkgProvider.ErrorKind, bool) {
	switch remote.errorType {
	case "invalid_request_error", "exceed_context_size_error":
		return pkgProvider.ErrorKindInvalidRequest, false
	case "authentication_error", "permission_error":
		return pkgProvider.ErrorKindAuthentication, false
	case "not_found_error":
		return llamaCPPNotFoundKind(operation, model), false
	case "not_supported_error":
		return pkgProvider.ErrorKindCapabilityNotFound, false
	case "unavailable_error":
		return pkgProvider.ErrorKindUnavailable, true
	case "rate_limit_error":
		return pkgProvider.ErrorKindRateLimited, true
	case "capacity_error", "insufficient_quota":
		return pkgProvider.ErrorKindCapacityExhausted, true
	case "server_error":
		return pkgProvider.ErrorKindInternal, true
	}

	switch remote.statusCode {
	case 0:
		return pkgProvider.ErrorKindInternal, false
	case http.StatusBadRequest,
		http.StatusRequestEntityTooLarge,
		http.StatusUnprocessableEntity:
		return pkgProvider.ErrorKindInvalidRequest, false
	case http.StatusUnauthorized, http.StatusForbidden:
		return pkgProvider.ErrorKindAuthentication, false
	case http.StatusNotFound:
		return llamaCPPNotFoundKind(operation, model), false
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
		if remote.statusCode >= 500 {
			return pkgProvider.ErrorKindInternal, true
		}
		if remote.statusCode >= 400 {
			return pkgProvider.ErrorKindInvalidRequest, false
		}

		return pkgProvider.ErrorKindInternal, false
	}
}

func llamaCPPNotFoundKind(
	operation pkgProvider.Operation,
	model string,
) pkgProvider.ErrorKind {
	if model != "" && operation != pkgProvider.OperationModelListing &&
		operation != pkgProvider.OperationModelDiscovery {
		return pkgProvider.ErrorKindModelNotFound
	}

	return pkgProvider.ErrorKindCapabilityNotFound
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
