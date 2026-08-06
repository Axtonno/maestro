package ollama

import (
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

var _ pkgProvider.ModelPullStream = (*modelPullStream)(nil)

type modelPullStream struct {
	context context.Context
	model   string
	body    io.ReadCloser
	decoder *json.Decoder

	recvMu  sync.Mutex
	stateMu sync.Mutex
	closed  bool
	done    bool
}

func (p *Provider) PullModel(
	ctx context.Context,
	request pkgProvider.ModelPullRequest,
) (pkgProvider.ModelPullStream, error) {
	model, err := p.model(request.Model)
	if err != nil {
		return nil, err
	}

	response, err := p.request(
		ctx,
		http.MethodPost,
		"/api/pull",
		modelPullRequest{Model: model, Stream: true},
	)
	if err != nil {
		return nil, fmt.Errorf("pull Ollama model: %w", err)
	}

	if err := requireSuccess(response); err != nil {
		response.Body.Close()

		return nil, fmt.Errorf("pull Ollama model: %w", err)
	}

	return &modelPullStream{
		context: ctx,
		model:   model,
		body:    response.Body,
		decoder: json.NewDecoder(response.Body),
	}, nil
}

func (p *Provider) RemoveModel(
	ctx context.Context,
	request pkgProvider.ModelRemoveRequest,
) error {
	model, err := p.model(request.Model)
	if err != nil {
		return err
	}

	response, err := p.request(
		ctx,
		http.MethodDelete,
		"/api/delete",
		modelRemoveRequest{Model: model},
	)
	if err != nil {
		return fmt.Errorf("remove Ollama model: %w", err)
	}
	defer response.Body.Close()

	if err := requireSuccess(response); err != nil {
		return fmt.Errorf("remove Ollama model: %w", err)
	}

	return nil
}

func (s *modelPullStream) Recv() (pkgProvider.ModelPullProgress, error) {
	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if err := s.currentState(); err != nil {
		return pkgProvider.ModelPullProgress{}, err
	}

	response := modelPullResponse{}
	if err := s.decoder.Decode(&response); err != nil {
		if contextError := s.context.Err(); contextError != nil {
			s.finish()

			return pkgProvider.ModelPullProgress{}, contextError
		}

		s.finish()
		if errors.Is(err, io.EOF) {
			return pkgProvider.ModelPullProgress{}, fmt.Errorf(
				"receive Ollama model pull: stream ended before completion: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}

		return pkgProvider.ModelPullProgress{}, fmt.Errorf(
			"receive Ollama model pull: decode progress: %v: %w",
			err,
			pkgProvider.ErrInvalidResponse,
		)
	}

	if response.Error != "" {
		s.finish()

		return pkgProvider.ModelPullProgress{}, &apiError{message: response.Error}
	}

	stage := ollamaPullStage(response.Status)
	if strings.TrimSpace(response.Status) == "" || response.Total < 0 ||
		response.Completed < 0 ||
		(response.Total > 0 && response.Completed > response.Total) {
		s.finish()

		return pkgProvider.ModelPullProgress{}, fmt.Errorf(
			"receive Ollama model pull: invalid progress: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	progress := pkgProvider.ModelPullProgress{
		Model:          s.model,
		Stage:          stage,
		Detail:         response.Status,
		Digest:         response.Digest,
		TotalBytes:     response.Total,
		CompletedBytes: response.Completed,
	}

	if stage == pkgProvider.ModelPullStageCompleted {
		s.finish()
	}

	return progress, nil
}

func (s *modelPullStream) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}
	s.closed = true
	s.stateMu.Unlock()

	return s.body.Close()
}

func (s *modelPullStream) currentState() error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.done {
		return io.EOF
	}

	if s.closed {
		return fmt.Errorf(
			"receive Ollama model pull after close: %w",
			pkgProvider.ErrInvalidStream,
		)
	}

	if err := s.context.Err(); err != nil {
		return err
	}

	return nil
}

func (s *modelPullStream) finish() {
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

func ollamaPullStage(status string) pkgProvider.ModelPullStage {
	normalized := strings.ToLower(strings.TrimSpace(status))

	switch {
	case normalized == "success":
		return pkgProvider.ModelPullStageCompleted
	case strings.HasPrefix(normalized, "pulling manifest"):
		return pkgProvider.ModelPullStageResolving
	case strings.HasPrefix(normalized, "pulling "):
		return pkgProvider.ModelPullStageDownloading
	case strings.HasPrefix(normalized, "verifying"):
		return pkgProvider.ModelPullStageVerifying
	case strings.HasPrefix(normalized, "writing manifest"),
		strings.HasPrefix(normalized, "removing any unused layers"):
		return pkgProvider.ModelPullStageFinalizing
	default:
		return pkgProvider.ModelPullStageUnknown
	}
}
