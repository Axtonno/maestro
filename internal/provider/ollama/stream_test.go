package ollama

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestStreamMapsChunksAndFinishesWithEOF(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"{\"model\":\"gemma4\",\"message\":{\"role\":\"assistant\",\"content\":\"Hel\"},\"done\":false}\n" +
			"{\"model\":\"gemma4\",\"message\":{\"role\":\"assistant\",\"content\":\"lo\"},\"done\":false}\n" +
			"{\"model\":\"gemma4\",\"message\":{\"role\":\"assistant\",\"content\":\"\"},\"done\":true,\"done_reason\":\"stop\",\"prompt_eval_count\":8,\"eval_count\":2}\n",
	)}
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/chat" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://ollama.test", "gemma4", client)
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
	if err != nil {
		t.Fatalf("receive final chunk: %v", err)
	}
	if final.FinishReason != "stop" ||
		final.Usage.InputTokens != 8 ||
		final.Usage.OutputTokens != 2 {
		t.Fatalf("unexpected final chunk: %#v", final)
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
		t.Fatalf("expected idempotent close, got %d closes", body.closes())
	}
}

func TestStreamNormalizesToolCallFinishReasonAcrossChunks(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		expectedChunkCount   int
		expectedFinishReason string
		expectedToolCalls    int
		expectedUsage        pkgProvider.Usage
	}{
		{
			name: "tool call followed by stop",
			body: `{"model":"llama","message":{"role":"assistant","tool_calls":[{"id":"call-1","function":{"index":0,"name":"weather","arguments":{"city":"Rome"}}}]},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":""},"done":true,"done_reason":"stop","prompt_eval_count":8,"eval_count":2}` + "\n",
			expectedChunkCount:   2,
			expectedFinishReason: pkgProvider.FinishReasonToolCalls,
			expectedToolCalls:    1,
			expectedUsage: pkgProvider.Usage{
				InputTokens: 8, OutputTokens: 2,
			},
		},
		{
			name: "text followed by stop",
			body: `{"model":"llama","message":{"role":"assistant","content":"hello"},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}` + "\n",
			expectedChunkCount:   2,
			expectedFinishReason: pkgProvider.FinishReasonStop,
		},
		{
			name: "tool call followed by tool calls",
			body: `{"model":"llama","message":{"role":"assistant","tool_calls":[{"id":"call-1","function":{"index":0,"name":"weather","arguments":{"city":"Rome"}}}]},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":""},"done":true,"done_reason":"tool_calls"}` + "\n",
			expectedChunkCount:   2,
			expectedFinishReason: pkgProvider.FinishReasonToolCalls,
			expectedToolCalls:    1,
		},
		{
			name: "tool call followed by length",
			body: `{"model":"llama","message":{"role":"assistant","tool_calls":[{"id":"call-1","function":{"index":0,"name":"weather","arguments":{"city":"Rome"}}}]},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":""},"done":true,"done_reason":"length"}` + "\n",
			expectedChunkCount:   2,
			expectedFinishReason: pkgProvider.FinishReasonLength,
			expectedToolCalls:    1,
		},
		{
			name: "multiple text chunks followed by stop",
			body: `{"model":"llama","message":{"role":"assistant","content":"hel"},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":"lo"},"done":false}` + "\n" +
				`{"model":"llama","message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}` + "\n",
			expectedChunkCount:   3,
			expectedFinishReason: pkgProvider.FinishReasonStop,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &trackingBody{reader: strings.NewReader(test.body)}
			provider := newProviderWithBody(t, body)
			active, err := provider.Stream(
				context.Background(),
				pkgProvider.CompletionRequest{},
			)
			if err != nil {
				t.Fatalf("open stream: %v", err)
			}

			var chunks []pkgProvider.StreamChunk
			for {
				chunk, receiveError := active.Recv()
				if errors.Is(receiveError, io.EOF) {
					break
				}
				if receiveError != nil {
					t.Fatalf("receive chunk: %v", receiveError)
				}
				chunks = append(chunks, chunk)
			}

			if len(chunks) != test.expectedChunkCount {
				t.Fatalf("unexpected chunks: %#v", chunks)
			}
			terminal := chunks[len(chunks)-1]
			if terminal.FinishReason != test.expectedFinishReason {
				t.Fatalf("unexpected terminal chunk: %#v", terminal)
			}
			if terminal.Usage != test.expectedUsage {
				t.Fatalf("terminal metadata was not preserved: %#v", terminal)
			}

			toolCallCount := 0
			for index, chunk := range chunks {
				toolCallCount += len(chunk.ToolCalls)
				if index < len(chunks)-1 && chunk.FinishReason != "" {
					t.Fatalf("non-terminal chunk was transformed: %#v", chunk)
				}
			}
			if toolCallCount != test.expectedToolCalls {
				t.Fatalf("unexpected translated tool calls: %#v", chunks)
			}
		})
	}
}

func TestStreamCloseIsIdempotentAndPreventsReads(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"{\"message\":{\"content\":\"waiting\"},\"done\":false}\n",
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

func TestStreamRejectsTruncatedMalformedAndErrorChunks(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		target error
	}{
		{
			name:   "truncated",
			body:   "{\"message\":{\"content\":\"partial\"},\"done\":false}\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name:   "malformed",
			body:   "{not-json}\n",
			target: pkgProvider.ErrInvalidResponse,
		},
		{
			name: "mid-stream API error",
			body: "{\"message\":{\"content\":\"partial\"},\"done\":false}\n" +
				"{\"error\":\"model runner failed\"}\n",
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
			var classified *pkgProvider.ProviderError
			if !errors.As(receiveError, &classified) ||
				classified.Operation != pkgProvider.OperationStreaming ||
				classified.Model != "gemma4" {
				t.Fatalf("expected classified streaming error, got %v", receiveError)
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
		"{\"message\":{\"content\":\"unused\"},\"done\":false}\n",
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
	} else {
		var classified *pkgProvider.ProviderError
		if !errors.As(err, &classified) ||
			classified.Kind != pkgProvider.ErrorKindCanceled {
			t.Fatalf("expected classified cancellation, got %v", err)
		}
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("close canceled stream: %v", err)
	}
}

func TestStreamHandlesHTTPError(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"{\"error\":\"model not found\"}",
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
	provider, err := New("http://ollama.test", "gemma4", client)
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
	provider, err := New("http://ollama.test", "gemma4", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	return provider
}
