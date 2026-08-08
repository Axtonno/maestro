package provider

import (
	"context"
	"errors"
	"io"
	"sync"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type resilienceStream struct {
	context   context.Context
	opener    func() (pkgProvider.Stream, error)
	execution *resilienceExecution

	mu        sync.Mutex
	stream    pkgProvider.Stream
	attempts  uint
	delivered bool
	closed    bool
}

func openStreamWithResilience(
	ctx context.Context,
	execution *resilienceExecution,
	opener func() (pkgProvider.Stream, error),
) (pkgProvider.Stream, error) {
	if execution == nil {
		return opener()
	}

	stream := &resilienceStream{
		context: ctx, opener: opener, execution: execution,
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
		stream, err := s.opener()
		if err == nil && nilStream(stream) {
			err = pkgProvider.ErrInvalidStream
		}
		if err == nil {
			s.stream = stream

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
		); waitError != nil {
			if errors.Is(waitError, errRetryBudgetExhausted) {
				return err
			}

			return waitError
		}
	}
}

func (s *resilienceStream) Recv() (pkgProvider.StreamChunk, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return pkgProvider.StreamChunk{}, pkgProvider.ErrInvalidStream
	}

	for {
		chunk, err := s.stream.Recv()
		if err == nil {
			s.delivered = true
			s.execution.finish(nil)

			return chunk, nil
		}
		if errors.Is(err, io.EOF) {
			s.execution.finish(nil)

			return chunk, err
		}
		if s.delivered || !s.execution.canRetry(s.attempts, err) {
			s.execution.finish(err)

			return chunk, err
		}

		_ = s.stream.Close()
		if waitError := s.execution.waitBeforeRetry(
			s.context,
			s.attempts,
		); waitError != nil {
			if !errors.Is(waitError, errRetryBudgetExhausted) {
				err = waitError
			}
			s.execution.finish(err)

			return pkgProvider.StreamChunk{}, err
		}
		if openError := s.openAfterWait(); openError != nil {
			s.execution.finish(openError)

			return pkgProvider.StreamChunk{}, openError
		}
	}
}

func (s *resilienceStream) openAfterWait() error {
	for {
		s.attempts++
		stream, err := s.opener()
		if err == nil && nilStream(stream) {
			err = pkgProvider.ErrInvalidStream
		}
		if err == nil {
			s.stream = stream

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
		); waitError != nil {
			if errors.Is(waitError, errRetryBudgetExhausted) {
				return err
			}

			return waitError
		}
	}
}

func (s *resilienceStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	err := s.stream.Close()
	s.execution.finish(err)

	return err
}

type resilienceModelPullStream struct {
	stream    pkgProvider.ModelPullStream
	execution *resilienceExecution

	mu     sync.Mutex
	closed bool
}

func (s *resilienceModelPullStream) Recv() (
	pkgProvider.ModelPullProgress,
	error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return pkgProvider.ModelPullProgress{}, pkgProvider.ErrInvalidStream
	}

	progress, err := s.stream.Recv()
	if err != nil {
		s.execution.finish(err)
	} else if progress.Stage == pkgProvider.ModelPullStageCompleted {
		s.execution.finish(nil)
	}

	return progress, err
}

func (s *resilienceModelPullStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	err := s.stream.Close()
	s.execution.finish(err)

	return err
}
