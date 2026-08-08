package llamacpp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const modelPullCleanupTimeout = 5 * time.Second

var _ pkgProvider.ModelPullStream = (*modelPullStream)(nil)

type modelPullStream struct {
	provider *Provider
	context  context.Context
	model    string
	body     io.ReadCloser
	reader   *bufio.Reader

	recvMu  sync.Mutex
	stateMu sync.Mutex
	closed  bool
	done    bool
}

func (p *Provider) PullModel(
	ctx context.Context,
	request pkgProvider.ModelPullRequest,
) (streamValue pkgProvider.ModelPullStream, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationModelPull,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return nil, err
	}
	operationModel = model

	if err := p.startModelPull(ctx, model); err != nil {
		return nil, err
	}

	events, err := p.request(
		ctx,
		http.MethodGet,
		"/models/sse",
		nil,
		"text/event-stream",
	)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("subscribe to llama.cpp model events: %w", err),
			p.cancelModelPull(model),
		)
	}
	if err := requireSuccess(events); err != nil {
		events.Body.Close()

		return nil, errors.Join(
			fmt.Errorf("subscribe to llama.cpp model events: %w", err),
			p.cancelModelPull(model),
		)
	}

	return &modelPullStream{
		provider: p,
		context:  ctx,
		model:    model,
		body:     events.Body,
		reader:   bufio.NewReader(events.Body),
	}, nil
}

func (p *Provider) RemoveModel(
	ctx context.Context,
	request pkgProvider.ModelRemoveRequest,
) (operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationModelRemove,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return err
	}
	operationModel = model

	response := modelLifecycleResponse{}
	path := "/models?model=" + url.QueryEscape(model)
	if err := p.doJSON(
		ctx,
		http.MethodDelete,
		path,
		nil,
		&response,
	); err != nil {
		return fmt.Errorf("remove llama.cpp model: %w", err)
	}

	if hasLlamaCPPAPIError(response.Error) {
		return newLlamaCPPAPIError(0, response.Error)
	}
	if !response.Success {
		return fmt.Errorf(
			"remove llama.cpp model: operation was not successful: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}

func (p *Provider) startModelPull(ctx context.Context, model string) error {
	response := modelLifecycleResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/models",
		modelLifecycleRequest{Model: model},
		&response,
	); err != nil {
		return fmt.Errorf("start llama.cpp model pull: %w", err)
	}

	if hasLlamaCPPAPIError(response.Error) {
		return newLlamaCPPAPIError(0, response.Error)
	}
	if !response.Success {
		return fmt.Errorf(
			"start llama.cpp model pull: operation was not successful: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}

func (s *modelPullStream) Recv() (
	progress pkgProvider.ModelPullProgress,
	receiveError error,
) {
	defer func() {
		receiveError = classifyLlamaCPPError(
			pkgProvider.OperationModelPull,
			s.model,
			receiveError,
		)
	}()

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if err := s.currentState(); err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			_ = s.finish(true)
		}

		return pkgProvider.ModelPullProgress{}, err
	}

	for {
		data, err := readSSEData(s.reader)
		if err != nil {
			if contextError := s.context.Err(); contextError != nil {
				_ = s.finish(true)

				return pkgProvider.ModelPullProgress{}, contextError
			}

			_ = s.finish(true)
			if errors.Is(err, io.EOF) {
				return pkgProvider.ModelPullProgress{}, fmt.Errorf(
					"receive llama.cpp model pull: stream ended before completion: %w",
					pkgProvider.ErrInvalidResponse,
				)
			}

			return pkgProvider.ModelPullProgress{}, fmt.Errorf(
				"receive llama.cpp model pull: read event: %v: %w",
				err,
				pkgProvider.ErrInvalidResponse,
			)
		}

		if data == "" {
			continue
		}

		event := modelEvent{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			_ = s.finish(true)

			return pkgProvider.ModelPullProgress{}, fmt.Errorf(
				"receive llama.cpp model pull: decode event: %v: %w",
				err,
				pkgProvider.ErrInvalidResponse,
			)
		}

		if event.Model != s.model {
			continue
		}

		switch event.Event {
		case "download_progress":
			completed, total, err := aggregateDownloadProgress(event.Data)
			if err != nil {
				_ = s.finish(true)

				return pkgProvider.ModelPullProgress{}, err
			}

			return pkgProvider.ModelPullProgress{
				Model:          s.model,
				Stage:          pkgProvider.ModelPullStageDownloading,
				Detail:         event.Event,
				TotalBytes:     total,
				CompletedBytes: completed,
			}, nil
		case "download_finished":
			if _, err := s.provider.DiscoverModels(s.context); err != nil {
				_ = s.finish(false)

				return pkgProvider.ModelPullProgress{}, fmt.Errorf(
					"refresh llama.cpp models after pull: %w",
					err,
				)
			}

			_ = s.finish(false)

			return pkgProvider.ModelPullProgress{
				Model:  s.model,
				Stage:  pkgProvider.ModelPullStageCompleted,
				Detail: event.Event,
			}, nil
		case "download_failed":
			_ = s.finish(false)

			remote := newLlamaCPPAPIError(0, event.Data)
			if remote.message == "" {
				remote.message = "model download failed"
			}

			return pkgProvider.ModelPullProgress{}, remote
		default:
			continue
		}
	}
}

func (s *modelPullStream) Close() (closeError error) {
	defer func() {
		closeError = classifyLlamaCPPError(
			pkgProvider.OperationModelPull,
			s.model,
			closeError,
		)
	}()

	return s.finish(true)
}

func (s *modelPullStream) currentState() error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	if s.done {
		return io.EOF
	}

	if s.closed {
		return fmt.Errorf(
			"receive llama.cpp model pull after close: %w",
			pkgProvider.ErrInvalidStream,
		)
	}

	if err := s.context.Err(); err != nil {
		return err
	}

	return nil
}

func (s *modelPullStream) finish(cancelRemote bool) error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}

	s.closed = true
	if !cancelRemote {
		s.done = true
	}
	s.stateMu.Unlock()

	bodyError := s.body.Close()
	if !cancelRemote {
		return bodyError
	}

	cancelError := s.provider.cancelModelPull(s.model)

	return errors.Join(bodyError, cancelError)
}

func (p *Provider) cancelModelPull(model string) error {
	cleanupContext, cancel := context.WithTimeout(
		context.Background(),
		modelPullCleanupTimeout,
	)
	defer cancel()

	return p.changeModelLifecycle(
		cleanupContext,
		model,
		"/models/unload",
		"cancel pull",
	)
}

func aggregateDownloadProgress(
	data json.RawMessage,
) (completed int64, total int64, err error) {
	files := map[string]modelDownloadProgress{}
	if err := json.Unmarshal(data, &files); err != nil || len(files) == 0 {
		return 0, 0, fmt.Errorf(
			"receive llama.cpp model pull: invalid download progress: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	for _, progress := range files {
		if progress.Done < 0 || progress.Total < 0 ||
			(progress.Total > 0 && progress.Done > progress.Total) ||
			completed > math.MaxInt64-progress.Done ||
			total > math.MaxInt64-progress.Total {
			return 0, 0, fmt.Errorf(
				"receive llama.cpp model pull: invalid download progress: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}

		completed += progress.Done
		total += progress.Total
	}

	return completed, total, nil
}
