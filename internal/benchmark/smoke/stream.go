package smoke

import (
	"context"
	"errors"
	"io"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) streamTerminalClose(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var active pkgProvider.Stream

	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityStreaming,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			active, err = s.runtime.Stream(
				ctx,
				s.config.ProviderID,
				simpleCompletionRequest(model),
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}

			chunkCount := 0
			terminalSeen := false
			for {
				chunk, receiveError := active.Recv()
				if errors.Is(receiveError, io.EOF) {
					break
				}
				if receiveError != nil {
					return pkgBenchmark.IterationResult{}, receiveError
				}
				chunkCount++
				if chunk.FinishReason != "" {
					terminalSeen = true
				}
			}
			if chunkCount == 0 {
				return failed("stream_has_no_chunks"), nil
			}
			if !terminalSeen {
				return failed("stream_terminal_missing"), nil
			}

			return passed(countMeasurement("stream_chunk_count", chunkCount)), nil
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			if active == nil {
				return nil
			}
			err := active.Close()
			active = nil

			return err
		},
	}
}

func (s *Suite) streamCancelDeadline(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var active pkgProvider.Stream
	var cancel context.CancelFunc

	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityStreaming,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			streamContext, cancelStream := context.WithCancel(ctx)
			cancel = cancelStream
			request := simpleCompletionRequest(model)
			request.Options.MaxTokens = 512
			request.Messages[0].Content = "Write a detailed explanation of local AI runtimes."
			active, err = s.runtime.Stream(
				streamContext,
				s.config.ProviderID,
				request,
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			cancelStream()

			chunksBeforeCancel := 0
			for {
				_, receiveError := active.Recv()
				switch {
				case errors.Is(receiveError, context.Canceled),
					errors.Is(receiveError, context.DeadlineExceeded):
					return passed(countMeasurement(
						"chunks_before_cancellation",
						chunksBeforeCancel,
					)), nil
				case errors.Is(receiveError, io.EOF):
					return failed("stream_completed_before_cancellation"), nil
				case receiveError != nil:
					return pkgBenchmark.IterationResult{}, receiveError
				default:
					chunksBeforeCancel++
				}
			}
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			if cancel != nil {
				cancel()
				cancel = nil
			}
			if active == nil {
				return nil
			}
			err := active.Close()
			active = nil

			return err
		},
	}
}
