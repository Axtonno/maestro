package ollama

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

type apiError struct {
	statusCode int
	message    string
}

func (e *apiError) Error() string {
	message := boundedRemoteErrorDetail(e.message)
	if e.statusCode == 0 {
		return fmt.Sprintf("Ollama API error: %s", message)
	}

	return fmt.Sprintf(
		"Ollama API error (status %d): %s",
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
	response, err := p.request(ctx, method, path, payload)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := requireSuccess(response); err != nil {
		return err
	}

	if err := decodeSingleJSON(response.Body, target); err != nil {
		return fmt.Errorf(
			"decode Ollama response: %w: %w",
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
) (*http.Response, error) {
	if ctx == nil {
		return nil, fmt.Errorf(
			"create Ollama request: context is nil: %w",
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
			return nil, fmt.Errorf("encode Ollama request: %w", err)
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
		return nil, fmt.Errorf("create Ollama request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := p.client.Do(request)
	if err != nil {
		if contextError := ctx.Err(); contextError != nil {
			return nil, contextError
		}

		return nil, fmt.Errorf(
			"send Ollama request: %w",
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

	decoded := errorResponse{}
	message := ""
	if json.Unmarshal(encoded, &decoded) == nil {
		message = strings.TrimSpace(decoded.Error)
	}

	if message == "" {
		message = strings.TrimSpace(string(encoded))
	}

	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &apiError{
		statusCode: statusCode,
		message:    message,
	}
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
