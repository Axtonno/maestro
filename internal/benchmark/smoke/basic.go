package smoke

import (
	"context"
	"math"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) capabilityInstance(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			if result := s.providerConfigured(); result != nil {
				return *result, nil
			}
			report, err := s.runtime.Capabilities(
				ctx,
				s.config.ProviderID,
				pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if report.Provider != s.config.ProviderID ||
				len(report.Capabilities) != len(pkgProvider.KnownCapabilities()) {
				return failed("invalid_capability_report"), nil
			}

			return passed(countMeasurement(
				"capability_count",
				len(report.Capabilities),
			)), nil
		},
	}
}

func (s *Suite) catalogListDiscover(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			preflight, err := s.requireInstance(
				ctx,
				pkgProvider.CapabilityModelListing,
				pkgProvider.CapabilityModelDiscovery,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			models, err := s.runtime.Models(ctx, s.config.ProviderID)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			infos, err := s.runtime.DiscoverModels(ctx, s.config.ProviderID)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			for _, model := range models {
				if strings.TrimSpace(model.ID) == "" {
					return failed("catalog_contains_empty_model_id"), nil
				}
			}
			for _, info := range infos {
				if strings.TrimSpace(info.Model.ID) == "" {
					return failed("discovery_contains_empty_model_id"), nil
				}
			}

			return passed(
				countMeasurement("listed_model_count", len(models)),
				countMeasurement("discovered_model_count", len(infos)),
			), nil
		},
	}
}

func (s *Suite) completionSimple(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityCompletion,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			response, err := s.runtime.Complete(
				ctx,
				s.config.ProviderID,
				simpleCompletionRequest(model),
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if strings.TrimSpace(response.Message.Content) == "" {
				return failed("completion_content_empty"), nil
			}

			return passed(
				countMeasurement("input_tokens", response.Usage.InputTokens),
				countMeasurement("output_tokens", response.Usage.OutputTokens),
			), nil
		},
	}
}

func (s *Suite) embedding(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"embedding",
				pkgProvider.CapabilityEmbedding,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			response, err := s.runtime.Embed(
				ctx,
				s.config.ProviderID,
				pkgProvider.EmbeddingRequest{
					Model: model, Inputs: []string{"Maestro smoke benchmark"},
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if len(response.Embeddings) != 1 || len(response.Embeddings[0]) == 0 {
				return failed("embedding_shape_invalid"), nil
			}
			for _, value := range response.Embeddings[0] {
				if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
					return failed("embedding_value_not_finite"), nil
				}
			}

			return passed(
				countMeasurement("embedding_dimension", len(response.Embeddings[0])),
				countMeasurement("input_tokens", response.Usage.InputTokens),
			), nil
		},
	}
}

func simpleCompletionRequest(model string) pkgProvider.CompletionRequest {
	return pkgProvider.CompletionRequest{
		Model: model,
		Messages: []pkgProvider.Message{{
			Role:    pkgProvider.RoleUser,
			Content: "Reply with one short sentence containing the word Maestro.",
		}},
		Options: pkgProvider.GenerationOptions{MaxTokens: 64},
	}
}

func resultOrZero(
	result *pkgBenchmark.IterationResult,
) pkgBenchmark.IterationResult {
	if result == nil {
		return pkgBenchmark.IterationResult{}
	}

	return *result
}
