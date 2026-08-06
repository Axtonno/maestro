package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestStreamMapsSSEChunksUsageAndCompletion(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		": keep-alive\r\n\r\n" +
			"data: {\"model\":\"local\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"Hel\"},\"finish_reason\":null}]}\r\n\r\n" +
			"data: {\"model\":\"local\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"lo\"},\"finish_reason\":null}]}\n\n" +
			"data: {\"model\":\"local\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n" +
			"data: {\"model\":\"local\",\"choices\":[],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":2}}\n\n" +
			"data: [DONE]\n\n",
	)}
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		if request.Method != http.MethodPost ||
			request.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("unexpected accept header %q", request.Header.Get("Accept"))
		}

		decoded := chatRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if !decoded.Stream || decoded.N != 1 ||
			decoded.StreamOptions == nil ||
			!decoded.StreamOptions.IncludeUsage {
			t.Errorf("unexpected stream request: %#v", decoded)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://llamacpp.test", "local", "", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	stream, err := provider.Stream(
		context.Background(),
		pkgProvider.CompletionRequest{},
	)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}

	first, err := stream.Recv()
	if err != nil || first.Content != "Hel" {
		t.Fatalf("unexpected first chunk %#v, %v", first, err)
	}
	second, err := stream.Recv()
	if err != nil || second.Content != "lo" {
		t.Fatalf("unexpected second chunk %#v, %v", second, err)
	}
	final, err := stream.Recv()
	if err != nil || final.FinishReason != "stop" {
		t.Fatalf("unexpected final choice %#v, %v", final, err)
	}
	usage, err := stream.Recv()
	if err != nil || usage.Usage.InputTokens != 8 ||
		usage.Usage.OutputTokens != 2 {
		t.Fatalf("unexpected usage chunk %#v, %v", usage, err)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
	if body.closes() != 1 {
		t.Fatalf("expected body closed once, got %d", body.closes())
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close completed stream: %v", err)
	}
	if body.closes() != 1 {
		t.Fatalf("expected idempotent close, got %d", body.closes())
	}
}

func TestStreamCloseIsIdempotentAndPreventsReads(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"waiting\"},\"finish_reason\":null}]}\n\n",
	)}
	provider := newProviderWithBody(t, body)
	stream, err := provider.Stream(
		context.Background(),
		pkgProvider.CompletionRequest{},
	)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close stream again: %v", err)
	}
	if _, err := stream.Recv(); !errors.Is(err, pkgProvider.ErrInvalidStream) {
		t.Fatalf("expected ErrInvalidStream, got %v", err)
	}
	if body.closes() != 1 {
		t.Fatalf("expected one body close, got %d", body.closes())
	}
}

func TestStreamRejectsInvalidEventSequences(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		target error
	}{
		{
			name:   "truncated before done",
			body:   "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"partial\"},\"finish_reason\":null}]}\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name:   "malformed JSON",
			body:   "data: {not-json}\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "API error event",
			body: "data: {\"error\":{\"message\":\"generation failed\"}}\n\n",
		},
		{
			name:   "done before finish reason",
			body:   "data: [DONE]\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name:   "choice-less event before finish",
			body:   "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":1}}\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name:   "invalid choice index",
			body:   "data: {\"choices\":[{\"index\":1,\"delta\":{},\"finish_reason\":null}]}\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "choice after finish",
			body: "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n" +
				"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"late\"},\"finish_reason\":null}]}\n\n",
			target: pkgProvider.ErrInvalidResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &trackingBody{reader: strings.NewReader(test.body)}
			provider := newProviderWithBody(t, body)
			stream, err := provider.Stream(
				context.Background(),
				pkgProvider.CompletionRequest{},
			)
			if err != nil {
				t.Fatalf("open stream: %v", err)
			}

			var receiveError error
			for receiveError == nil {
				_, receiveError = stream.Recv()
			}
			if errors.Is(receiveError, io.EOF) {
				t.Fatalf("expected stream failure, got EOF")
			}
			if test.target != nil && !errors.Is(receiveError, test.target) {
				t.Fatalf("expected %v, got %v", test.target, receiveError)
			}
			if body.closes() != 1 {
				t.Fatalf("expected body close after failure, got %d", body.closes())
			}
		})
	}
}

func TestStreamPreservesCancellation(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":null}]}\n\n",
	)}
	provider := newProviderWithBody(t, body)
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := provider.Stream(ctx, pkgProvider.CompletionRequest{})
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	cancel()

	if _, err := stream.Recv(); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close canceled stream: %v", err)
	}
}

func TestStreamHandlesHTTPError(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		`{"error":{"message":"model not found"}}`,
	)}
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://llamacpp.test", "local", "", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if _, err := provider.Stream(
		context.Background(),
		pkgProvider.CompletionRequest{},
	); err == nil {
		t.Fatal("expected HTTP error")
	}
	if body.closes() != 1 {
		t.Fatalf("expected error body close, got %d", body.closes())
	}
}

func newProviderWithBody(t *testing.T, body io.ReadCloser) *Provider {
	t.Helper()

	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://llamacpp.test", "local", "", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	return provider
}
