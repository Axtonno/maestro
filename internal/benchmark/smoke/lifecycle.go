package smoke

import (
	"context"
	"errors"
	"io"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) lifecycleLoadUnload(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var cleanupModel string

	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"lifecycle",
				pkgProvider.CapabilityModelLoad,
				pkgProvider.CapabilityModelUnload,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			cleanupModel = model
			if err := s.runtime.LoadModel(
				ctx,
				s.config.ProviderID,
				pkgProvider.ModelLoadRequest{Model: model},
			); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}

			return passed(), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if cleanupModel == "" {
				return nil
			}
			model := cleanupModel
			cleanupModel = ""

			return s.runtime.UnloadModel(
				ctx,
				s.config.ProviderID,
				pkgProvider.ModelUnloadRequest{Model: model},
			)
		},
	}
}

func (s *Suite) acquisitionPullRemove(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var active pkgProvider.ModelPullStream
	var cleanupModel string

	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
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
			preflight, err := s.requireInstance(
				ctx,
				pkgProvider.CapabilityModelDiscovery,
				pkgProvider.CapabilityModelPull,
				pkgProvider.CapabilityModelRemove,
			)
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

			active, err = s.runtime.PullModel(
				ctx,
				s.config.ProviderID,
				pkgProvider.ModelPullRequest{Model: model},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			cleanupModel = model
			progressCount := 0
			completed := false
			for {
				progress, receiveError := active.Recv()
				if receiveError != nil {
					return pkgBenchmark.IterationResult{}, receiveError
				}
				progressCount++
				if progress.Stage == pkgProvider.ModelPullStageCompleted {
					completed = true
					break
				}
			}
			if !completed {
				return failed("model_pull_terminal_missing"), nil
			}
			if _, err := active.Recv(); !errors.Is(err, io.EOF) {
				if err == nil {
					return failed("model_pull_continued_after_terminal"), nil
				}
				return pkgBenchmark.IterationResult{}, err
			}

			return passed(countMeasurement("pull_progress_count", progressCount)), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			var cleanupError error
			if active != nil {
				cleanupError = active.Close()
				active = nil
			}
			if cleanupModel == "" {
				return cleanupError
			}
			model := cleanupModel
			cleanupModel = ""
			removeError := s.runtime.RemoveModel(
				ctx,
				s.config.ProviderID,
				pkgProvider.ModelRemoveRequest{Model: model},
			)
			if errors.Is(removeError, pkgProvider.ErrModelNotFound) {
				removeError = nil
			}

			return errors.Join(cleanupError, removeError)
		},
	}
}
