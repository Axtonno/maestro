package provider

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type resilienceStream struct {
	context           context.Context
	opener            func() (pkgProvider.Stream, error)
	execution         *resilienceExecution
	observation       *operationObservation
	finishObservation bool

	stateMu   sync.Mutex
	receiving atomic.Bool
	stream    pkgProvider.Stream
	attempts  uint
	delivered bool
	closed    bool
}

func openStreamWithResilience(
	ctx context.Context,
	execution *resilienceExecution,
	observation *operationObservation,
	finishObservation bool,
	opener func() (pkgProvider.Stream, error),
) (pkgProvider.Stream, error) {
	if execution == nil && observation == nil {
		return opener()
	}

	stream := &resilienceStream{
		context: ctx, opener: opener, execution: execution,
		observation: observation, finishObservation: finishObservation,
	}
	if err := stream.open(); err != nil {
		execution.finish(err)

		return nil, err
	}

	return stream, nil
}

func (s *resilienceStream) open() error {
	for {
		s.attempts++
		s.observation.attempt(s.attempts)
		stream, err := s.opener()
		if err == nil && nilStream(stream) {
			err = pkgProvider.ErrInvalidStream
		}
		if err == nil {
			if !s.setStream(stream) {
				_ = stream.Close()

				return pkgProvider.ErrInvalidStream
			}

			return nil
		}
		if stream != nil && !nilStream(stream) {
			_ = stream.Close()
		}
		if s.execution == nil || !s.execution.canRetry(s.attempts, err) {
			return err
		}
		if waitError := s.execution.waitBeforeRetry(
			s.context,
			s.attempts,
			err,
		); waitError != nil {
			if errors.Is(waitError, errRetryBudgetExhausted) {
				return err
			}

			return waitError
		}
	}
}

func (s *resilienceStream) Recv() (pkgProvider.StreamChunk, error) {
	if !s.receiving.CompareAndSwap(false, true) {
		return pkgProvider.StreamChunk{}, pkgProvider.ErrInvalidStream
	}
	defer s.receiving.Store(false)

	stream, closed := s.currentStream()
	if closed {
		return pkgProvider.StreamChunk{}, pkgProvider.ErrInvalidStream
	}

	for {
		chunk, err := stream.Recv()
		if err == nil {
			s.delivered = true
			s.execution.finish(nil)
			s.observation.recordUsage(chunk.Usage)

			return chunk, nil
		}
		if errors.Is(err, io.EOF) {
			s.execution.finish(nil)
			if s.finishObservation {
				s.observation.finish(nil)
			}

			return chunk, err
		}
		if s.delivered || s.execution == nil ||
			!s.execution.canRetry(s.attempts, err) {
			s.execution.finish(err)
			if s.finishObservation {
				s.observation.finish(err)
			}

			return chunk, err
		}

		_ = stream.Close()
		if waitError := s.execution.waitBeforeRetry(
			s.context,
			s.attempts,
			err,
		); waitError != nil {
			if !errors.Is(waitError, errRetryBudgetExhausted) {
				err = waitError
			}
			s.execution.finish(err)
			if s.finishObservation {
				s.observation.finish(err)
			}

			return pkgProvider.StreamChunk{}, err
		}
		if openError := s.openAfterWait(); openError != nil {
			s.execution.finish(openError)
			if s.finishObservation {
				s.observation.finish(openError)
			}

			return pkgProvider.StreamChunk{}, openError
		}
		stream, closed = s.currentStream()
		if closed {
			err = pkgProvider.ErrInvalidStream
			s.execution.finish(err)
			if s.finishObservation {
				s.observation.finishClosed(nil)
			}

			return pkgProvider.StreamChunk{}, err
		}
	}
}

func (s *resilienceStream) openAfterWait() error {
	for {
		s.attempts++
		s.observation.attempt(s.attempts)
		stream, err := s.opener()
		if err == nil && nilStream(stream) {
			err = pkgProvider.ErrInvalidStream
		}
		if err == nil {
			if !s.setStream(stream) {
				_ = stream.Close()

				return pkgProvider.ErrInvalidStream
			}

			return nil
		}
		if stream != nil && !nilStream(stream) {
			_ = stream.Close()
		}
		if !s.execution.canRetry(s.attempts, err) {
			return err
		}
		if waitError := s.execution.waitBeforeRetry(
			s.context,
			s.attempts,
			err,
		); waitError != nil {
			if errors.Is(waitError, errRetryBudgetExhausted) {
				return err
			}

			return waitError
		}
	}
}

func (s *resilienceStream) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}
	s.closed = true
	stream := s.stream
	s.stateMu.Unlock()

	err := stream.Close()
	s.execution.finish(err)
	if s.finishObservation {
		s.observation.finishClosed(err)
	}

	return err
}

func (s *resilienceStream) currentStream() (pkgProvider.Stream, bool) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	return s.stream, s.closed
}

func (s *resilienceStream) setStream(stream pkgProvider.Stream) bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.closed {
		return false
	}

	s.stream = stream

	return true
}

type resilienceModelPullStream struct {
	stream      pkgProvider.ModelPullStream
	execution   *resilienceExecution
	observation *operationObservation

	stateMu   sync.Mutex
	receiving atomic.Bool
	closed    bool
}

func (s *resilienceModelPullStream) Recv() (
	pkgProvider.ModelPullProgress,
	error,
) {
	if !s.receiving.CompareAndSwap(false, true) {
		return pkgProvider.ModelPullProgress{}, pkgProvider.ErrInvalidStream
	}
	defer s.receiving.Store(false)

	s.stateMu.Lock()
	closed := s.closed
	s.stateMu.Unlock()
	if closed {
		return pkgProvider.ModelPullProgress{}, pkgProvider.ErrInvalidStream
	}

	progress, err := s.stream.Recv()
	s.observation.recordProgress(progress.CompletedBytes, progress.TotalBytes)
	if err != nil {
		s.execution.finish(err)
		s.observation.finish(err)
	} else if progress.Stage == pkgProvider.ModelPullStageCompleted {
		s.execution.finish(nil)
		s.observation.finish(nil)
	}

	return progress, err
}

func (s *resilienceModelPullStream) Close() error {
	s.stateMu.Lock()
	if s.closed {
		s.stateMu.Unlock()

		return nil
	}
	s.closed = true
	s.stateMu.Unlock()

	err := s.stream.Close()
	s.execution.finish(err)
	s.observation.finishClosed(err)

	return err
}
