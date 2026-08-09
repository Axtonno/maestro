package runtimebench

import (
	"context"
	"errors"
	"io"
	"math"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) modelCompletionLatency(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "chat", pkgProvider.CapabilityCompletion)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			var response pkgProvider.CompletionResponse
			started := time.Now()
			resources, err := s.capture(func() error {
				var operationError error
				response, operationError = s.runtime.Complete(ctx, s.config.ProviderID, completionRequest(model))
				return operationError
			})
			elapsed := time.Since(started)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("total_latency_ms", elapsed),
				countMeasurement("input_tokens", response.Usage.InputTokens),
				countMeasurement("output_tokens", response.Usage.OutputTokens),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
	}
}

func (s *Suite) modelStreamPerformance(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var active pkgProvider.Stream
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "chat", pkgProvider.CapabilityStreaming)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			started := time.Now()
			var firstContentAt time.Time
			var completedAt time.Time
			var chunks, outputTokens int
			var terminal bool
			resources, err := s.capture(func() error {
				var operationError error
				active, operationError = s.runtime.Stream(ctx, s.config.ProviderID, completionRequest(model))
				if operationError != nil {
					return operationError
				}
				for {
					chunk, receiveError := active.Recv()
					if errors.Is(receiveError, io.EOF) {
						completedAt = time.Now()
						return nil
					}
					if receiveError != nil {
						return receiveError
					}
					chunks++
					if firstContentAt.IsZero() && (chunk.Content != "" || len(chunk.ToolCalls) > 0) {
						firstContentAt = time.Now()
					}
					if chunk.Usage.OutputTokens > outputTokens {
						outputTokens = chunk.Usage.OutputTokens
					}
					if chunk.FinishReason != "" {
						terminal = true
					}
				}
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if firstContentAt.IsZero() || !terminal || chunks == 0 {
				return failed("stream_terminal_or_content_missing"), nil
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("time_to_first_token_ms", firstContentAt.Sub(started)),
				elapsedMeasurement("total_latency_ms", completedAt.Sub(started)),
				countMeasurement("stream_chunk_count", chunks),
				countMeasurement("stream_completed", 1),
			}
			generationDuration := completedAt.Sub(firstContentAt)
			if outputTokens > 0 && generationDuration > 0 {
				measurements = append(measurements, rateMeasurement(
					"tokens_per_second",
					float64(outputTokens)/generationDuration.Seconds(),
					"tokens/s",
					"provider_reported_usage",
				))
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
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

func (s *Suite) modelStreamCancellation(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var active pkgProvider.Stream
	var cancel context.CancelFunc
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "chat", pkgProvider.CapabilityStreaming)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			streamContext, cancelStream := context.WithCancel(ctx)
			cancel = cancelStream
			request := completionRequest(model)
			request.Options.MaxTokens = 1024
			request.Messages[0].Content = "Write a long technical explanation of local AI runtime architecture."
			active, err = s.runtime.Stream(streamContext, s.config.ProviderID, request)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			cancelStarted := time.Now()
			cancelStream()
			session := s.sampler.Start()
			chunks := 0
			for {
				_, receiveError := active.Recv()
				switch {
				case errors.Is(receiveError, context.Canceled), errors.Is(receiveError, context.DeadlineExceeded):
					measurements := []pkgBenchmark.Measurement{
						elapsedMeasurement("cancel_latency_ms", time.Since(cancelStarted)),
						countMeasurement("chunks_before_cancellation", chunks),
					}
					measurements = append(measurements, session.Stop()...)
					return passed(measurements...), nil
				case errors.Is(receiveError, io.EOF):
					session.Stop()
					return failed("stream_completed_before_cancellation"), nil
				case receiveError != nil:
					session.Stop()
					return pkgBenchmark.IterationResult{}, receiveError
				default:
					chunks++
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

func (s *Suite) modelEmbeddingPerformance(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	const batchSize = 8
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "embedding", pkgProvider.CapabilityEmbedding)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			inputs := make([]string, batchSize)
			for index := range inputs {
				inputs[index] = "Maestro runtime benchmark embedding fixture"
			}
			var response pkgProvider.EmbeddingResponse
			started := time.Now()
			resources, err := s.capture(func() error {
				var operationError error
				response, operationError = s.runtime.Embed(ctx, s.config.ProviderID, pkgProvider.EmbeddingRequest{Model: model, Inputs: inputs})
				return operationError
			})
			elapsed := time.Since(started)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if len(response.Embeddings) != batchSize || len(response.Embeddings[0]) == 0 {
				return failed("embedding_shape_invalid"), nil
			}
			for _, embedding := range response.Embeddings {
				if len(embedding) != len(response.Embeddings[0]) {
					return failed("embedding_dimension_inconsistent"), nil
				}
				for _, value := range embedding {
					if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
						return failed("embedding_value_not_finite"), nil
					}
				}
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("embedding_latency_ms", elapsed),
				rateMeasurement("embeddings_per_second", batchSize/elapsed.Seconds(), "embeddings/s", "fixed_batch_size"),
				countMeasurement("embedding_batch_size", batchSize),
				countMeasurement("embedding_dimension", len(response.Embeddings[0])),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
	}
}

func (s *Suite) modelLifecycleLoadUnload(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var cleanupModel string
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "lifecycle", pkgProvider.CapabilityModelLoad, pkgProvider.CapabilityModelUnload)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			cleanupModel = model
			var loadElapsed, unloadElapsed time.Duration
			resources, err := s.capture(func() error {
				started := time.Now()
				if operationError := s.runtime.LoadModel(ctx, s.config.ProviderID, pkgProvider.ModelLoadRequest{Model: model}); operationError != nil {
					return operationError
				}
				loadElapsed = time.Since(started)
				started = time.Now()
				if operationError := s.runtime.UnloadModel(ctx, s.config.ProviderID, pkgProvider.ModelUnloadRequest{Model: model}); operationError != nil {
					return operationError
				}
				unloadElapsed = time.Since(started)
				cleanupModel = ""
				return nil
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("load_latency_ms", loadElapsed),
				elapsedMeasurement("unload_latency_ms", unloadElapsed),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if cleanupModel == "" {
				return nil
			}
			model := cleanupModel
			cleanupModel = ""
			return s.runtime.UnloadModel(ctx, s.config.ProviderID, pkgProvider.ModelUnloadRequest{Model: model})
		},
	}
}

func (s *Suite) modelColdWarm(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var cleanupModel string
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "lifecycle", pkgProvider.CapabilityCompletion, pkgProvider.CapabilityModelUnload)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			cleanupModel = model
			if err := s.runtime.UnloadModel(ctx, s.config.ProviderID, pkgProvider.ModelUnloadRequest{Model: model}); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			var coldElapsed, warmElapsed time.Duration
			resources, err := s.capture(func() error {
				started := time.Now()
				if _, operationError := s.runtime.Complete(ctx, s.config.ProviderID, completionRequest(model)); operationError != nil {
					return operationError
				}
				coldElapsed = time.Since(started)
				started = time.Now()
				if _, operationError := s.runtime.Complete(ctx, s.config.ProviderID, completionRequest(model)); operationError != nil {
					return operationError
				}
				warmElapsed = time.Since(started)
				return nil
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("cold_latency_ms", coldElapsed),
				elapsedMeasurement("warm_latency_ms", warmElapsed),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if cleanupModel == "" {
				return nil
			}
			model := cleanupModel
			cleanupModel = ""
			return s.runtime.UnloadModel(ctx, s.config.ProviderID, pkgProvider.ModelUnloadRequest{Model: model})
		},
	}
}

func (s *Suite) modelPullCancellation(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var active pkgProvider.ModelPullStream
	var cancel context.CancelFunc
	var cleanupModel string
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			if result := s.providerConfigured(); result != nil {
				return *result, nil
			}
			if !s.config.AllowCatalogMutation {
				return skipped("catalog_mutation_not_allowed"), nil
			}
			model := s.config.Models["acquisition_fixture"]
			if model == "" {
				return skipped("model_not_configured"), nil
			}
			preflight, err := s.requireInstance(ctx, pkgProvider.CapabilityModelDiscovery, pkgProvider.CapabilityModelPull, pkgProvider.CapabilityModelRemove)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			models, err := s.runtime.DiscoverModels(ctx, s.config.ProviderID)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			for _, existing := range models {
				if existing.Model.ID == model {
					return skipped("acquisition_fixture_already_present"), nil
				}
			}
			pullContext, cancelPull := context.WithCancel(ctx)
			cancel = cancelPull
			active, err = s.runtime.PullModel(pullContext, s.config.ProviderID, pkgProvider.ModelPullRequest{Model: model})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			cleanupModel = model
			first, err := active.Recv()
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if first.Stage == pkgProvider.ModelPullStageCompleted {
				return skipped("pull_completed_before_cancellation"), nil
			}
			cancelStarted := time.Now()
			cancelPull()
			session := s.sampler.Start()
			for {
				progress, receiveError := active.Recv()
				switch {
				case errors.Is(receiveError, context.Canceled), errors.Is(receiveError, context.DeadlineExceeded):
					measurements := []pkgBenchmark.Measurement{elapsedMeasurement("cancel_latency_ms", time.Since(cancelStarted))}
					measurements = append(measurements, session.Stop()...)
					return passed(measurements...), nil
				case receiveError != nil:
					session.Stop()
					return pkgBenchmark.IterationResult{}, receiveError
				case progress.Stage == pkgProvider.ModelPullStageCompleted:
					session.Stop()
					return failed("pull_completed_before_cancellation"), nil
				}
			}
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if cancel != nil {
				cancel()
				cancel = nil
			}
			var closeError error
			if active != nil {
				closeError = active.Close()
				active = nil
			}
			if cleanupModel == "" {
				return closeError
			}
			model := cleanupModel
			cleanupModel = ""
			removeError := s.runtime.RemoveModel(ctx, s.config.ProviderID, pkgProvider.ModelRemoveRequest{Model: model})
			if errors.Is(removeError, pkgProvider.ErrModelNotFound) {
				removeError = nil
			}
			return errors.Join(closeError, removeError)
		},
	}
}
