package llamacpp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const maxErrorBodySize = 64 << 10
const maxRemoteErrorDetailBytes = 512

type apiError struct {
	statusCode int
	message    string
	errorType  string
	errorCode  string
}

func (e *apiError) Error() string {
	message := boundedRemoteErrorDetail(e.message)
	if e.statusCode == 0 {
		return fmt.Sprintf("llama.cpp API error: %s", message)
	}

	return fmt.Sprintf(
		"llama.cpp API error (status %d): %s",
		e.statusCode,
		message,
	)
}

func (p *Provider) doJSON(
	ctx context.Context,
	method string,
	path string,
	payload any,
	target any,
) error {
	response, err := p.request(ctx, method, path, payload, "application/json")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := requireSuccess(response); err != nil {
		return err
	}

	if err := decodeSingleJSON(response.Body, target); err != nil {
		return fmt.Errorf(
			"decode llama.cpp response: %w: %w",
			err,
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}

func (p *Provider) request(
	ctx context.Context,
	method string,
	path string,
	payload any,
	accept string,
) (*http.Response, error) {
	if ctx == nil {
		return nil, fmt.Errorf(
			"create llama.cpp request: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode llama.cpp request: %w", err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		p.endpoint(path),
		body,
	)
	if err != nil {
		return nil, fmt.Errorf("create llama.cpp request: %w", err)
	}

	request.Header.Set("Accept", accept)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if p.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	response, err := p.client.Do(request)
	if err != nil {
		if contextError := ctx.Err(); contextError != nil {
			return nil, contextError
		}

		return nil, fmt.Errorf(
			"send llama.cpp request: %w",
			&transportError{cause: err},
		)
	}

	return response, nil
}

func requireSuccess(response *http.Response) error {
	if response.StatusCode >= http.StatusOK &&
		response.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	return decodeAPIError(response.StatusCode, response.Body)
}

func decodeAPIError(statusCode int, body io.Reader) error {
	limited := io.LimitReader(body, maxErrorBodySize+1)
	encoded, err := io.ReadAll(limited)
	if err != nil {
		return &apiError{
			statusCode: statusCode,
			message:    http.StatusText(statusCode),
		}
	}

	if len(encoded) > maxErrorBodySize {
		encoded = encoded[:maxErrorBodySize]
	}

	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(encoded, &envelope) == nil && len(envelope.Error) > 0 {
		decoded := newLlamaCPPAPIError(statusCode, envelope.Error)
		if decoded.message == "" {
			decoded.message = http.StatusText(statusCode)
		}

		return decoded
	}

	message := strings.TrimSpace(string(encoded))
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &apiError{statusCode: statusCode, message: message}
}

func errorMessage(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var detail errorDetail
	if json.Unmarshal(raw, &detail) == nil {
		if message := strings.TrimSpace(detail.Message); message != "" {
			return message
		}
	}

	var message string
	if json.Unmarshal(raw, &message) == nil {
		return strings.TrimSpace(message)
	}

	return ""
}

func errorMessageFromEnvelope(encoded []byte) string {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(encoded, &envelope) != nil {
		return ""
	}

	return errorMessage(envelope.Error)
}

func decodeSingleJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing any
	err := decoder.Decode(&trailing)
	if !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("response contains multiple JSON values")
		}

		return fmt.Errorf("decode trailing response data: %w", err)
	}

	return nil
}
