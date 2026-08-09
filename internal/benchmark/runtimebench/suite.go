package runtimebench

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type Kind string

const (
	KindProvider Kind = "provider"
	KindModel    Kind = "model"

	faultProviderID pkgProvider.ID = "benchmark-controlled-fault"
)

type Suite struct {
	config  smoke.Config
	runtime pkgProvider.Runtime
	sampler Sampler
	faults  *faultCompleter
}

func SelectManifest(manifest pkgBenchmark.Manifest, kind Kind) (pkgBenchmark.Manifest, error) {
	if err := manifest.Validate(); err != nil {
		return pkgBenchmark.Manifest{}, err
	}
	prefix := string(kind) + "-"
	if kind != KindProvider && kind != KindModel {
		return pkgBenchmark.Manifest{}, fmt.Errorf("unknown runtime benchmark kind %q", kind)
	}
	selected := manifest
	selected.Scenarios = nil
	for _, definition := range manifest.Scenarios {
		if strings.HasPrefix(definition.ID, prefix) {
			selected.Scenarios = append(selected.Scenarios, definition)
		}
	}
	if len(selected.Scenarios) == 0 {
		return pkgBenchmark.Manifest{}, fmt.Errorf("runtime benchmark manifest has no %s scenarios", kind)
	}
	if err := selected.Validate(); err != nil {
		return pkgBenchmark.Manifest{}, err
	}

	return selected, nil
}

func NewScenarios(
	manifest pkgBenchmark.Manifest,
	config smoke.Config,
	runtime pkgProvider.Runtime,
	sampler Sampler,
) ([]pkgBenchmark.Scenario, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	if nilRuntime(runtime) {
		return nil, errors.New("runtime benchmark provider runtime is nil")
	}
	if sampler == nil {
		sampler = NoopSampler()
	}
	suite := &Suite{config: config, runtime: runtime, sampler: sampler}
	if config.Enabled {
		configured, err := runtime.Resolve(config.ProviderID)
		if err != nil {
			return nil, err
		}
		if completer, ok := configured.(pkgProvider.Completer); ok {
			suite.faults = &faultCompleter{delegate: completer}
			if err := runtime.Register(suite.faults); err != nil {
				return nil, fmt.Errorf("register controlled fault provider: %w", err)
			}
		}
	}

	scenarios := make([]pkgBenchmark.Scenario, 0, len(manifest.Scenarios))
	for _, definition := range manifest.Scenarios {
		scenario, err := suite.newScenario(definition)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios, nil
}

func nilRuntime(runtime pkgProvider.Runtime) bool {
	if runtime == nil {
		return true
	}
	value := reflect.ValueOf(runtime)

	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (s *Suite) newScenario(definition pkgBenchmark.ScenarioDefinition) (pkgBenchmark.Scenario, error) {
	switch definition.ID {
	case "provider-capability-latency":
		return s.providerCapabilityLatency(definition), nil
	case "provider-catalog-latency":
		return s.providerCatalogLatency(definition), nil
	case "provider-retry-controlled":
		return s.providerRetryControlled(definition), nil
	case "provider-circuit-breaker":
		return s.providerCircuitBreaker(definition), nil
	case "model-completion-latency":
		return s.modelCompletionLatency(definition), nil
	case "model-stream-performance":
		return s.modelStreamPerformance(definition), nil
	case "model-stream-cancellation":
		return s.modelStreamCancellation(definition), nil
	case "model-embedding-performance":
		return s.modelEmbeddingPerformance(definition), nil
	case "model-lifecycle-load-unload":
		return s.modelLifecycleLoadUnload(definition), nil
	case "model-cold-warm":
		return s.modelColdWarm(definition), nil
	case "model-pull-cancellation":
		if definition.MutationGuard != "MAESTRO_ALLOW_CATALOG_MUTATION=true" {
			return nil, fmt.Errorf("runtime benchmark scenario %q has unsupported mutation guard %q", definition.ID, definition.MutationGuard)
		}
		return s.modelPullCancellation(definition), nil
	default:
		return nil, fmt.Errorf("runtime benchmark scenario %q is not implemented", definition.ID)
	}
}

func (s *Suite) requireInstance(ctx context.Context, capabilities ...pkgProvider.Capability) (*pkgBenchmark.IterationResult, error) {
	if result := s.providerConfigured(); result != nil {
		return result, nil
	}
	return s.requireCapabilities(ctx, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance}, capabilities...)
}

func (s *Suite) requireModel(ctx context.Context, role string, capabilities ...pkgProvider.Capability) (string, *pkgBenchmark.IterationResult, error) {
	if result := s.providerConfigured(); result != nil {
		return "", result, nil
	}
	model := strings.TrimSpace(s.config.Models[role])
	if model == "" {
		result := skipped("model_not_configured")
		return "", &result, nil
	}
	result, err := s.requireCapabilities(ctx, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetModel, Model: model}, capabilities...)
	return model, result, err
}

func (s *Suite) requireCapabilities(ctx context.Context, request pkgProvider.CapabilityRequest, capabilities ...pkgProvider.Capability) (*pkgBenchmark.IterationResult, error) {
	report, err := s.runtime.Capabilities(ctx, s.config.ProviderID, request)
	if err != nil {
		return nil, err
	}
	for _, required := range capabilities {
		var found *pkgProvider.CapabilityDescriptor
		for index := range report.Capabilities {
			if report.Capabilities[index].Capability == required {
				found = &report.Capabilities[index]
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("capability report omitted %q", required)
		}
		if found.Support == pkgProvider.CapabilityUnsupported {
			result := unsupported("capability_not_supported")
			return &result, nil
		}
		if found.Availability == pkgProvider.CapabilityAvailabilityUnavailable {
			result := skipped("capability_unavailable")
			return &result, nil
		}
	}
	return nil, nil
}

func (s *Suite) providerConfigured() *pkgBenchmark.IterationResult {
	if s.config.Enabled {
		return nil
	}
	result := skipped("provider_not_configured")
	return &result
}

func (s *Suite) capture(operation func() error) (measurements []pkgBenchmark.Measurement, operationError error) {
	session := s.sampler.Start()
	defer func() {
		measurements = session.Stop()
	}()
	operationError = operation()
	return measurements, operationError
}

func elapsedMeasurement(name string, elapsed time.Duration) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Value: float64(elapsed) / float64(time.Millisecond), Unit: "ms", Method: "monotonic_clock"}
}

func countMeasurement(name string, value int) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Value: float64(value), Unit: "count"}
}

func rateMeasurement(name string, value float64, unit, method string) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Value: value, Unit: unit, Method: method}
}

func passed(measurements ...pkgBenchmark.Measurement) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultPassed, Measurements: measurements}
}

func failed(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultFailed, Error: &pkgBenchmark.ErrorRecord{Kind: "runtime_gate", Code: code}}
}

func skipped(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultSkipped, ReasonCode: code}
}

func unsupported(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultUnsupported, ReasonCode: code}
}

func resultOrZero(result *pkgBenchmark.IterationResult) pkgBenchmark.IterationResult {
	if result == nil {
		return pkgBenchmark.IterationResult{}
	}
	return *result
}

func completionRequest(model string) pkgProvider.CompletionRequest {
	return pkgProvider.CompletionRequest{
		Model:    model,
		Messages: []pkgProvider.Message{{Role: pkgProvider.RoleUser, Content: "Reply with one concise sentence about local AI runtimes."}},
		Options:  pkgProvider.GenerationOptions{MaxTokens: 128},
	}
}

type faultMode uint8

const (
	faultNone faultMode = iota
	faultFirst
	faultAlways
)

type faultCompleter struct {
	delegate pkgProvider.Completer
	mu       sync.Mutex
	mode     faultMode
	calls    int
}

func (f *faultCompleter) ID() pkgProvider.ID { return faultProviderID }

func (f *faultCompleter) reset(mode faultMode) {
	f.mu.Lock()
	f.mode = mode
	f.calls = 0
	f.mu.Unlock()
}

func (f *faultCompleter) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *faultCompleter) Complete(ctx context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	f.mu.Lock()
	f.calls++
	call := f.calls
	mode := f.mode
	f.mu.Unlock()
	if mode == faultAlways || (mode == faultFirst && call == 1) {
		return pkgProvider.CompletionResponse{}, pkgProvider.NewProviderError(pkgProvider.ProviderErrorDetails{
			Kind: pkgProvider.ErrorKindTransient, Operation: pkgProvider.OperationCompletion,
			Provider: faultProviderID, Model: request.Model, Retryable: true,
			Message: "controlled benchmark fault",
		}, pkgProvider.ErrTransient)
	}
	return f.delegate.Complete(ctx, request)
}
