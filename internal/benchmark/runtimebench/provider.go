package runtimebench

import (
	"context"
	"errors"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (s *Suite) providerCapabilityLatency(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			if result := s.providerConfigured(); result != nil {
				return *result, nil
			}
			started := time.Now()
			var report pkgProvider.CapabilityReport
			resources, err := s.capture(func() error {
				var operationError error
				report, operationError = s.runtime.Capabilities(ctx, s.config.ProviderID, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance})
				return operationError
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if report.Provider != s.config.ProviderID || len(report.Capabilities) != len(pkgProvider.KnownCapabilities()) {
				return failed("invalid_capability_report"), nil
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("capability_latency_ms", time.Since(started)),
				countMeasurement("capability_count", len(report.Capabilities)),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
	}
}

func (s *Suite) providerCatalogLatency(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			preflight, err := s.requireInstance(ctx, pkgProvider.CapabilityModelListing, pkgProvider.CapabilityModelDiscovery)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			var listed []pkgProvider.Model
			var discovered []pkgProvider.ModelInfo
			var listElapsed, discoveryElapsed time.Duration
			resources, err := s.capture(func() error {
				started := time.Now()
				var operationError error
				listed, operationError = s.runtime.Models(ctx, s.config.ProviderID)
				listElapsed = time.Since(started)
				if operationError != nil {
					return operationError
				}
				started = time.Now()
				discovered, operationError = s.runtime.DiscoverModels(ctx, s.config.ProviderID)
				discoveryElapsed = time.Since(started)
				return operationError
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("model_listing_latency_ms", listElapsed),
				elapsedMeasurement("model_discovery_latency_ms", discoveryElapsed),
				countMeasurement("listed_model_count", len(listed)),
				countMeasurement("discovered_model_count", len(discovered)),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
	}
}

func (s *Suite) providerRetryControlled(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var policyModel string
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "chat", pkgProvider.CapabilityCompletion)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			if s.faults == nil {
				return unsupported("controlled_fault_provider_unavailable"), nil
			}
			policyModel = model
			s.faults.reset(faultFirst)
			if err := s.runtime.SetResiliencePolicy(ctx, faultProviderID, retryPolicy(model)); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			started := time.Now()
			resources, err := s.capture(func() error {
				_, operationError := s.runtime.Complete(ctx, faultProviderID, completionRequest(model))
				return operationError
			})
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if s.faults.callCount() != 2 {
				return failed("controlled_retry_count_invalid"), nil
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("retry_recovery_latency_ms", time.Since(started)),
				countMeasurement("attempt_count", 2),
				countMeasurement("retry_count", 1),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if s.faults != nil {
				s.faults.reset(faultNone)
			}
			if policyModel == "" {
				return nil
			}
			model := policyModel
			policyModel = ""
			return s.runtime.SetResiliencePolicy(ctx, faultProviderID, baselinePolicy(model))
		},
	}
}

func (s *Suite) providerCircuitBreaker(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Scenario {
	var policyModel string
	var nextProbeAt time.Time
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, "chat", pkgProvider.CapabilityCompletion)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			if s.faults == nil {
				return unsupported("controlled_fault_provider_unavailable"), nil
			}
			policyModel = model
			s.faults.reset(faultAlways)
			if err := s.runtime.SetResiliencePolicy(ctx, faultProviderID, circuitPolicy(model)); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			started := time.Now()
			resources, captureError := s.capture(func() error {
				_, firstError := s.runtime.Complete(ctx, faultProviderID, completionRequest(model))
				if firstError == nil || !errors.Is(firstError, pkgProvider.ErrTransient) {
					return errors.New("controlled circuit fault was not returned")
				}
				snapshot, exists, stateError := s.runtime.CircuitState(faultProviderID, pkgProvider.OperationCompletion, model)
				if stateError != nil {
					return stateError
				}
				if !exists || snapshot.State != pkgProvider.CircuitStateOpen {
					return errors.New("controlled circuit did not open")
				}
				nextProbeAt = snapshot.NextProbeAt
				callsBeforeBlock := s.faults.callCount()
				_, blockedError := s.runtime.Complete(ctx, faultProviderID, completionRequest(model))
				if !errors.Is(blockedError, pkgProvider.ErrCircuitOpen) || s.faults.callCount() != callsBeforeBlock {
					return errors.New("open circuit did not block the provider call")
				}
				return nil
			})
			if captureError != nil {
				return failed("controlled_circuit_sequence_invalid"), nil
			}
			measurements := []pkgBenchmark.Measurement{
				elapsedMeasurement("circuit_open_latency_ms", time.Since(started)),
				countMeasurement("blocked_call_count", 1),
			}
			measurements = append(measurements, resources...)
			return passed(measurements...), nil
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			if s.faults != nil {
				s.faults.reset(faultNone)
			}
			if policyModel == "" {
				return nil
			}
			model := policyModel
			policyModel = ""
			if err := s.runtime.SetResiliencePolicy(ctx, faultProviderID, baselinePolicy(model)); err != nil {
				return err
			}
			wait := time.Until(nextProbeAt)
			nextProbeAt = time.Time{}
			if wait <= 0 {
				return nil
			}
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

func retryPolicy(model string) pkgProvider.ResiliencePolicy {
	return pkgProvider.ResiliencePolicy{
		Operation: pkgProvider.OperationCompletion, Model: model, MaxAttempts: 2,
		InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond, BackoffMultiplier: 1,
	}
}

func circuitPolicy(model string) pkgProvider.ResiliencePolicy {
	return pkgProvider.ResiliencePolicy{
		Operation: pkgProvider.OperationCompletion, Model: model, MaxAttempts: 1,
		CircuitBreaker: pkgProvider.CircuitBreakerPolicy{
			FailureThreshold: 1, OpenDuration: 5 * time.Millisecond, HalfOpenMaxAttempts: 1,
		},
	}
}

func baselinePolicy(model string) pkgProvider.ResiliencePolicy {
	return pkgProvider.ResiliencePolicy{Operation: pkgProvider.OperationCompletion, Model: model, MaxAttempts: 1}
}
