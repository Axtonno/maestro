package llamacpp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

var _ pkgProvider.Stream = (*stream)(nil)

type stream struct {
	context context.Context
	model   string
	body    io.ReadCloser
	reader  *bufio.Reader

	recvMu   sync.Mutex
	stateMu  sync.Mutex
	closed   bool
	done     bool
	terminal bool
}

func (p *Provider) Stream(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (streamValue pkgProvider.Stream, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationStreaming,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return nil, err
	}
	operationModel = model

	response, err := p.request(
		ctx,
		http.MethodPost,
		"/v1/chat/completions",
		newChatRequest(model, request.Messages, true),
		"text/event-stream",
	)
	if err != nil {
		return nil, fmt.Errorf("stream with llama.cpp: %w", err)
	}

	if err := requireSuccess(response); err != nil {
		response.Body.Close()

		return nil, fmt.Errorf("stream with llama.cpp: %w", err)
	}

	return &stream{
		context: ctx,
		model:   model,
		body:    response.Body,
		reader:  bufio.NewReader(response.Body),
	}, nil
}

func (s *stream) Recv() (
	chunk pkgProvider.StreamChunk,
	receiveError error,
) {
	defer func() {
		receiveError = classifyLlamaCPPError(
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

	for {
		data, err := readSSEData(s.reader)
		if err != nil {
			if contextError := s.context.Err(); contextError != nil {
				s.finish()

				return pkgProvider.StreamChunk{}, contextError
			}

			s.finish()
			if errors.Is(err, io.EOF) {
				return pkgProvider.StreamChunk{}, fmt.Errorf(
					"receive llama.cpp stream: stream ended before [DONE]: %w",
					pkgProvider.ErrInvalidResponse,
				)
			}

			return pkgProvider.StreamChunk{}, fmt.Errorf(
				"receive llama.cpp stream: read event: %v: %w",
				err,
				pkgProvider.ErrInvalidResponse,
			)
		}

		if data == "" {
			continue
		}

		if strings.TrimSpace(data) == "[DONE]" {
			if !s.terminal {
				s.finish()

				return pkgProvider.StreamChunk{}, fmt.Errorf(
					"receive llama.cpp stream: [DONE] before final choice: %w",
					pkgProvider.ErrInvalidResponse,
				)
			}

			s.finish()

			return pkgProvider.StreamChunk{}, io.EOF
		}

		response := chatResponse{}
		if err := json.Unmarshal([]byte(data), &response); err != nil {
			s.finish()

			return pkgProvider.StreamChunk{}, fmt.Errorf(
				"receive llama.cpp stream: decode event: %v: %w",
				err,
				pkgProvider.ErrInvalidResponse,
			)
		}

		if hasLlamaCPPAPIError(response.Error) {
			s.finish()

			return pkgProvider.StreamChunk{}, newLlamaCPPAPIError(0, response.Error)
		}

		if len(response.Choices) == 0 {
			if response.Usage == nil || !s.terminal {
				s.finish()

				return pkgProvider.StreamChunk{}, fmt.Errorf(
					"receive llama.cpp stream: invalid choice-less event: %w",
					pkgProvider.ErrInvalidResponse,
				)
			}

			return pkgProvider.StreamChunk{
				Model: response.Model,
				Usage: providerUsage(response.Usage),
			}, nil
		}

		if len(response.Choices) != 1 || s.terminal {
			s.finish()

			return pkgProvider.StreamChunk{}, fmt.Errorf(
				"receive llama.cpp stream: invalid choice count or ordering: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}

		choice := response.Choices[0]
		if choice.Index != 0 {
			s.finish()

			return pkgProvider.StreamChunk{}, fmt.Errorf(
				"receive llama.cpp stream: invalid choice index %d: %w",
				choice.Index,
				pkgProvider.ErrInvalidResponse,
			)
		}

		finishReason := ""
		if choice.FinishReason != nil {
			finishReason = *choice.FinishReason
			s.terminal = true
		}

		return pkgProvider.StreamChunk{
			Model:        response.Model,
			Content:      choice.Delta.Content,
			FinishReason: finishReason,
			Usage:        providerUsage(response.Usage),
		}, nil
	}
}

func (s *stream) Close() (closeError error) {
	defer func() {
		closeError = classifyLlamaCPPError(
			pkgProvider.OperationStreaming,
			s.model,
			closeError,
		)
	}()

	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}
	s.closed = true
	s.stateMu.Unlock()

	return s.body.Close()
}

func (s *stream) currentState() error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.done {
		return io.EOF
	}

	if s.closed {
		return fmt.Errorf(
			"receive llama.cpp stream after close: %w",
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

func readSSEData(reader *bufio.Reader) (string, error) {
	dataLines := make([]string, 0, 1)

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			switch {
			case line == "":
				if len(dataLines) > 0 {
					return strings.Join(dataLines, "\n"), nil
				}
			case strings.HasPrefix(line, ":"):
				// SSE comments are ignored.
			default:
				field, value, found := strings.Cut(line, ":")
				if !found {
					field = line
					value = ""
				}
				if strings.HasPrefix(value, " ") {
					value = value[1:]
				}
				if field == "data" {
					dataLines = append(dataLines, value)
				}
			}
		}

		if err != nil {
			if len(dataLines) > 0 {
				return strings.Join(dataLines, "\n"), nil
			}

			return "", err
		}
	}
}
