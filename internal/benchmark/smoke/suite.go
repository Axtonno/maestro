package smoke

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type Suite struct {
	config  Config
	runtime pkgProvider.Runtime
}

func NewScenarios(
	manifest pkgBenchmark.Manifest,
	config Config,
	runtime pkgProvider.Runtime,
) ([]pkgBenchmark.Scenario, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	if nilProviderRuntime(runtime) {
		return nil, errors.New("smoke provider runtime is nil")
	}
	suite := &Suite{config: config, runtime: runtime}
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

func nilProviderRuntime(runtime pkgProvider.Runtime) bool {
	if runtime == nil {
		return true
	}
	value := reflect.ValueOf(runtime)

	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (s *Suite) newScenario(
	definition pkgBenchmark.ScenarioDefinition,
) (pkgBenchmark.Scenario, error) {
	switch definition.ID {
	case "capability-instance":
		return s.capabilityInstance(definition), nil
	case "catalog-list-discover":
		return s.catalogListDiscover(definition), nil
	case "completion-simple":
		return s.completionSimple(definition), nil
	case "stream-terminal-close":
		return s.streamTerminalClose(definition), nil
	case "stream-cancel-deadline":
		return s.streamCancelDeadline(definition), nil
	case "embedding":
		return s.embedding(definition), nil
	case "lifecycle-load-unload":
		return s.lifecycleLoadUnload(definition), nil
	case "acquisition-pull-remove":
		if definition.MutationGuard != "MAESTRO_ALLOW_CATALOG_MUTATION=true" {
			return nil, fmt.Errorf(
				"smoke scenario %q has unsupported mutation guard %q",
				definition.ID,
				definition.MutationGuard,
			)
		}
		return s.acquisitionPullRemove(definition), nil
	case "structured-json":
		return s.structuredJSON(definition), nil
	case "structured-json-schema":
		return s.structuredJSONSchema(definition), nil
	case "tool-call-result":
		return s.toolCallResult(definition), nil
	case "tool-call-stream":
		return s.toolCallStream(definition), nil
	case "resilience-controlled-error":
		return s.resilienceControlledError(definition), nil
	case "observability-redaction":
		return s.observabilityRedaction(definition), nil
	default:
		return nil, fmt.Errorf("smoke scenario %q is not implemented", definition.ID)
	}
}

func (s *Suite) requireInstance(
	ctx context.Context,
	capabilities ...pkgProvider.Capability,
) (*pkgBenchmark.IterationResult, error) {
	if result := s.providerConfigured(); result != nil {
		return result, nil
	}

	return s.requireCapabilities(
		ctx,
		pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance},
		capabilities...,
	)
}

func (s *Suite) requireModel(
	ctx context.Context,
	role string,
	capabilities ...pkgProvider.Capability,
) (string, *pkgBenchmark.IterationResult, error) {
	if result := s.providerConfigured(); result != nil {
		return "", result, nil
	}
	model := strings.TrimSpace(s.config.Models[role])
	if model == "" {
		result := skipped("model_not_configured")
		return "", &result, nil
	}
	result, err := s.requireCapabilities(
		ctx,
		pkgProvider.CapabilityRequest{
			Target: pkgProvider.CapabilityTargetModel,
			Model:  model,
		},
		capabilities...,
	)

	return model, result, err
}

func (s *Suite) requireCapabilities(
	ctx context.Context,
	request pkgProvider.CapabilityRequest,
	capabilities ...pkgProvider.Capability,
) (*pkgBenchmark.IterationResult, error) {
	report, err := s.runtime.Capabilities(ctx, s.config.ProviderID, request)
	if err != nil {
		return nil, err
	}
	for _, required := range capabilities {
		var descriptor *pkgProvider.CapabilityDescriptor
		for index := range report.Capabilities {
			if report.Capabilities[index].Capability == required {
				descriptor = &report.Capabilities[index]
				break
			}
		}
		if descriptor == nil {
			return nil, fmt.Errorf("capability report omitted %q", required)
		}
		if descriptor.Support == pkgProvider.CapabilityUnsupported {
			result := unsupported("capability_not_supported")
			return &result, nil
		}
		if descriptor.Availability == pkgProvider.CapabilityAvailabilityUnavailable {
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

func passed(measurements ...pkgBenchmark.Measurement) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultPassed, Measurements: measurements,
	}
}

func failed(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultFailed,
		Error: &pkgBenchmark.ErrorRecord{Kind: "smoke_gate", Code: code},
	}
}

func skipped(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultSkipped, ReasonCode: code,
	}
}

func unsupported(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultUnsupported, ReasonCode: code,
	}
}

func countMeasurement(name string, value int) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Unit: "count", Value: float64(value)}
}
