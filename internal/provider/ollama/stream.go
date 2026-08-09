package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

var _ pkgProvider.Stream = (*stream)(nil)

type stream struct {
	context      context.Context
	model        string
	body         io.ReadCloser
	decoder      *json.Decoder
	output       *pkgProvider.StructuredOutput
	content      bytes.Buffer
	toolCallSeen bool

	recvMu  sync.Mutex
	stateMu sync.Mutex
	closed  bool
	done    bool
}

func (p *Provider) Stream(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (result pkgProvider.Stream, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationStreaming,
			operationModel,
			operationError,
		)
	}()
	if err := validateOllamaCompletionRequest(request); err != nil {
		return nil, err
	}

	model, err := p.model(request.Model)
	if err != nil {
		return nil, err
	}
	operationModel = model

	response, err := p.request(
		ctx,
		http.MethodPost,
		"/api/chat",
		newChatRequest(model, request, true),
	)
	if err != nil {
		return nil, fmt.Errorf("stream with Ollama: %w", err)
	}

	if err := requireSuccess(response); err != nil {
		response.Body.Close()

		return nil, fmt.Errorf("stream with Ollama: %w", err)
	}

	return &stream{
		context: ctx,
		model:   model,
		body:    response.Body,
		decoder: json.NewDecoder(response.Body),
		output:  request.Output,
	}, nil
}

func (s *stream) Recv() (
	chunk pkgProvider.StreamChunk,
	receiveError error,
) {
	defer func() {
		receiveError = classifyOllamaError(
			pkgProvider.OperationStreaming,
			s.model,
			receiveError,
		)
	}()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if err := s.currentState(); err != nil {
		return pkgProvider.StreamChunk{}, err
	}

	response := chatResponse{}
	if err := s.decoder.Decode(&response); err != nil {
		if contextError := s.context.Err(); contextError != nil {
			s.finish()

			return pkgProvider.StreamChunk{}, contextError
		}

		s.finish()

		if errors.Is(err, io.EOF) {
			return pkgProvider.StreamChunk{}, fmt.Errorf(
				"receive Ollama stream: stream ended before a final chunk: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}

		return pkgProvider.StreamChunk{}, fmt.Errorf(
			"receive Ollama stream: decode chunk: %v: %w",
			err,
			pkgProvider.ErrInvalidResponse,
		)
	}

	if response.Error != "" {
		s.finish()

		return pkgProvider.StreamChunk{}, &apiError{
			message: response.Error,
		}
	}
	toolCalls, err := ollamaToolCallDeltas(response.Message.ToolCalls)
	if err != nil {
		s.finish()

		return pkgProvider.StreamChunk{}, err
	}
	if len(toolCalls) > 0 {
		s.toolCallSeen = true
	}
	if s.output != nil {
		_, _ = s.content.WriteString(response.Message.Content)
		if response.Done {
			if err := validateStructuredContent(s.output, s.content.Bytes()); err != nil {
				s.finish()

				return pkgProvider.StreamChunk{}, err
			}
		}
	}

	chunk = pkgProvider.StreamChunk{
		Model:     response.Model,
		Content:   response.Message.Content,
		ToolCalls: toolCalls,
		FinishReason: normalizeOllamaFinishReason(
			response.Done,
			response.DoneReason,
			s.toolCallSeen,
		),
		Usage: pkgProvider.Usage{
			InputTokens:  response.PromptEvalCount,
			OutputTokens: response.EvalCount,
		},
	}

	if response.Done {
		s.finish()
	}

	return chunk, nil
}

func (s *stream) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}
	s.closed = true
	s.stateMu.Unlock()

	return classifyOllamaError(
		pkgProvider.OperationStreaming,
		s.model,
		s.body.Close(),
	)
}

func (s *stream) currentState() error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.done {
		return io.EOF
	}

	if s.closed {
		return fmt.Errorf(
			"receive Ollama stream after close: %w",
			pkgProvider.ErrInvalidStream,
		)
	}

	if err := s.context.Err(); err != nil {
		return err
	}

	return nil
}

func (s *stream) finish() {
	s.stateMu.Lock()
	if s.done {
		s.stateMu.Unlock()

		return
	}

	s.done = true
	alreadyClosed := s.closed
	s.closed = true
	s.stateMu.Unlock()

	if !alreadyClosed {
		_ = s.body.Close()
	}
}
