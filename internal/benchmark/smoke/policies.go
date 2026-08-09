package smoke

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) resilienceControlledError(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var policyModel string

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
			policyModel = model
			if err := s.runtime.SetResiliencePolicy(
				ctx,
				s.config.ProviderID,
				pkgProvider.ResiliencePolicy{
					Operation: pkgProvider.OperationCompletion,
					Model:     model, MaxAttempts: 2,
					InitialBackoff:    time.Millisecond,
					MaxBackoff:        time.Millisecond,
					BackoffMultiplier: 1,
				},
			); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			operationContext, cancel := context.WithCancel(ctx)
			cancel()
			_, err = s.runtime.Complete(
				operationContext,
				s.config.ProviderID,
				simpleCompletionRequest(model),
			)
			if !errors.Is(err, context.Canceled) {
				if err == nil {
					return failed("controlled_error_not_returned"), nil
				}
				return pkgBenchmark.IterationResult{}, err
			}

			return passed(), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if policyModel == "" {
				return nil
			}
			model := policyModel
			policyModel = ""

			return s.runtime.SetResiliencePolicy(
				ctx,
				s.config.ProviderID,
				pkgProvider.ResiliencePolicy{
					Operation: pkgProvider.OperationCompletion,
					Model:     model, MaxAttempts: 1,
				},
			)
		},
	}
}

type recordingObserver struct {
	mu     sync.Mutex
	events []pkgProvider.ProviderEvent
}

func (o *recordingObserver) ObserveProviderEvent(event pkgProvider.ProviderEvent) error {
	o.mu.Lock()
	o.events = append(o.events, event)
	o.mu.Unlock()

	return nil
}

func (o *recordingObserver) snapshot() []pkgProvider.ProviderEvent {
	o.mu.Lock()
	defer o.mu.Unlock()

	return append([]pkgProvider.ProviderEvent(nil), o.events...)
}

func (s *Suite) observabilityRedaction(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	const sensitivePrompt = "maestro-smoke-sensitive-prompt"
	var observerInstalled bool

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
			observer := &recordingObserver{}
			s.runtime.SetObserver(observer)
			observerInstalled = true
			_, err = s.runtime.Complete(
				ctx,
				s.config.ProviderID,
				pkgProvider.CompletionRequest{
					Model: model,
					Messages: []pkgProvider.Message{{
						Role: pkgProvider.RoleUser, Content: sensitivePrompt,
					}},
					Options: pkgProvider.GenerationOptions{MaxTokens: 32},
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			events := observer.snapshot()
			if len(events) < 3 ||
				events[0].Kind != pkgProvider.ProviderEventOperationStarted ||
				events[len(events)-1].Kind != pkgProvider.ProviderEventOperationCompleted {
				return failed("observability_event_sequence_invalid"), nil
			}
			operationID := events[0].OperationID
			for _, event := range events {
				if operationID == 0 || event.OperationID != operationID ||
					event.Provider != s.config.ProviderID ||
					event.Operation != pkgProvider.OperationCompletion {
					return failed("observability_event_correlation_invalid"), nil
				}
			}
			if strings.Contains(fmt.Sprintf("%#v", events), sensitivePrompt) {
				return failed("observability_content_leaked"), nil
			}

			return passed(countMeasurement("provider_event_count", len(events))), nil
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			if observerInstalled {
				s.runtime.SetObserver(nil)
				observerInstalled = false
			}

			return nil
		},
	}
}
