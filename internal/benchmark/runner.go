package benchmark

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const defaultCleanupTimeout = 30 * time.Second

var errScenarioPanic = errors.New("benchmark scenario panicked")

type RunnerOptions struct {
	Warmup         int
	Runs           int
	Timeout        time.Duration
	CleanupTimeout time.Duration
	Command        string
	MaestroVersion string
	MaestroCommit  string
}

type Runner struct {
	options RunnerOptions
	now     func() time.Time
	runID   func() string
}

func NewRunner(options RunnerOptions) (*Runner, error) {
	return newRunner(options, time.Now, randomRunID)
}

func newRunner(
	options RunnerOptions,
	now func() time.Time,
	runID func() string,
) (*Runner, error) {
	if options.Warmup < 0 {
		return nil, errors.New("benchmark warmup cannot be negative")
	}
	if options.Runs < 1 {
		return nil, errors.New("benchmark runs must be at least one")
	}
	if options.Timeout < 0 {
		return nil, errors.New("benchmark timeout cannot be negative")
	}
	if options.CleanupTimeout < 0 {
		return nil, errors.New("benchmark cleanup timeout cannot be negative")
	}
	if options.CleanupTimeout == 0 {
		options.CleanupTimeout = defaultCleanupTimeout
	}
	if now == nil || runID == nil {
		return nil, errors.New("benchmark runner dependencies are nil")
	}

	return &Runner{options: options, now: now, runID: runID}, nil
}

func (r *Runner) Run(
	ctx context.Context,
	manifest pkgBenchmark.Manifest,
	configuration pkgBenchmark.ConfigurationProfile,
	scenarios ...pkgBenchmark.Scenario,
) (pkgBenchmark.Report, error) {
	if ctx == nil {
		return pkgBenchmark.Report{}, errors.New("benchmark context is nil")
	}
	if err := manifest.Validate(); err != nil {
		return pkgBenchmark.Report{}, err
	}

	registry, err := scenarioRegistry(scenarios)
	if err != nil {
		return pkgBenchmark.Report{}, err
	}

	startedAt := r.now()
	configuration.Execution.Warmup = r.options.Warmup
	configuration.Execution.Runs = r.options.Runs
	configuration.Execution.TimeoutMS = durationMilliseconds(r.options.Timeout)
	configuration.Execution.CleanupTimeoutMS = durationMilliseconds(
		r.options.CleanupTimeout,
	)
	report := pkgBenchmark.Report{
		SchemaVersion: pkgBenchmark.ReportSchemaVersion,
		RunID:         r.runID(),
		CreatedAt:     startedAt,
		Metadata: pkgBenchmark.RunMetadata{
			Command: r.options.Command, MaestroVersion: r.options.MaestroVersion,
			MaestroCommit:   r.options.MaestroCommit,
			ManifestVersion: manifest.Version, ManifestOwner: manifest.Owner,
			SourceMilestone: manifest.SourceMilestone,
		},
		Configuration: configuration,
		Scenarios:     make([]pkgBenchmark.ScenarioReport, 0, len(manifest.Scenarios)),
	}

	for _, definition := range manifest.Scenarios {
		scenario, exists := registry[definition.ID]
		if !exists {
			report.Scenarios = append(report.Scenarios, missingScenarioReport(
				definition,
				startedAt,
			))
			continue
		}
		if scenario.Definition() != definition {
			return pkgBenchmark.Report{}, fmt.Errorf(
				"benchmark scenario %q definition differs from manifest",
				definition.ID,
			)
		}

		report.Scenarios = append(report.Scenarios, r.runScenario(ctx, scenario))
	}

	completedAt := r.now()
	if completedAt.Before(startedAt) {
		completedAt = startedAt
	}
	report.CompletedAt = completedAt
	report.DurationMS = durationMilliseconds(completedAt.Sub(startedAt))
	if err := report.Validate(); err != nil {
		return pkgBenchmark.Report{}, err
	}

	return report, nil
}

func scenarioRegistry(
	scenarios []pkgBenchmark.Scenario,
) (map[string]pkgBenchmark.Scenario, error) {
	registry := make(map[string]pkgBenchmark.Scenario, len(scenarios))
	for _, scenario := range scenarios {
		if nilScenario(scenario) {
			return nil, errors.New("benchmark scenario is nil")
		}
		definition := scenario.Definition()
		if err := definition.Validate(); err != nil {
			return nil, err
		}
		if _, exists := registry[definition.ID]; exists {
			return nil, fmt.Errorf("benchmark scenario %q is duplicated", definition.ID)
		}
		registry[definition.ID] = scenario
	}

	return registry, nil
}

func nilScenario(scenario pkgBenchmark.Scenario) bool {
	if scenario == nil {
		return true
	}
	value := reflect.ValueOf(scenario)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (r *Runner) runScenario(
	ctx context.Context,
	scenario pkgBenchmark.Scenario,
) pkgBenchmark.ScenarioReport {
	totalIterations := r.options.Warmup + r.options.Runs
	samples := make([]pkgBenchmark.Sample, 0, totalIterations)
	state := pkgBenchmark.ResultPassed
	for position := 0; position < totalIterations; position++ {
		iteration := pkgBenchmark.Iteration{
			Warmup: position < r.options.Warmup,
		}
		if iteration.Warmup {
			iteration.Index = position + 1
		} else {
			iteration.Index = position - r.options.Warmup + 1
		}

		sample := r.runIteration(ctx, scenario, iteration)
		samples = append(samples, sample)
		if sample.State != pkgBenchmark.ResultPassed || sample.CleanupError != nil {
			state = sample.State
			if sample.CleanupError != nil {
				state = pkgBenchmark.ResultFailed
			}
			break
		}
	}

	return pkgBenchmark.ScenarioReport{
		Scenario:   scenario.Definition(),
		State:      state,
		Samples:    samples,
		Aggregates: aggregateSamples(samples),
	}
}

func (r *Runner) runIteration(
	ctx context.Context,
	scenario pkgBenchmark.Scenario,
	iteration pkgBenchmark.Iteration,
) pkgBenchmark.Sample {
	iterationContext := ctx
	cancel := func() {}
	if r.options.Timeout > 0 {
		iterationContext, cancel = context.WithTimeout(ctx, r.options.Timeout)
	}

	startedAt := r.now()
	result, runError, panicked := callScenario(iterationContext, scenario, iteration)
	operationCompletedAt := r.now()
	if operationCompletedAt.Before(startedAt) {
		operationCompletedAt = startedAt
	}
	contextError := iterationContext.Err()
	cancel()

	if panicked {
		result = failedResult("runner", "scenario_panic")
	} else if runError != nil {
		result = failedResultForError(runError, "scenario_failed")
	} else if contextError != nil {
		result = failedResultForError(contextError, "scenario_context")
	} else if err := result.Validate(); err != nil {
		result = failedResult("runner", "invalid_scenario_result")
	}

	cleanupError, cleanupPanicked := r.cleanup(ctx, scenario, iteration)
	var classifiedCleanupError *pkgBenchmark.ErrorRecord
	if cleanupPanicked {
		classifiedCleanupError = classifyError(errScenarioPanic, "runner", "cleanup_panic")
	} else if cleanupError != nil {
		classifiedCleanupError = classifyError(cleanupError, "cleanup", "cleanup_failed")
	}

	return pkgBenchmark.Sample{
		Iteration: iteration, StartedAt: startedAt,
		DurationMS: durationMilliseconds(operationCompletedAt.Sub(startedAt)),
		State:      result.State, ReasonCode: result.ReasonCode,
		Measurements: append([]pkgBenchmark.Measurement(nil), result.Measurements...),
		Error:        result.Error, CleanupError: classifiedCleanupError,
		Evaluation: cloneEvaluation(result.Evaluation),
	}
}

func cloneEvaluation(
	evaluation *pkgBenchmark.QualityEvaluation,
) *pkgBenchmark.QualityEvaluation {
	if evaluation == nil {
		return nil
	}
	cloned := *evaluation
	return &cloned
}

func callScenario(
	ctx context.Context,
	scenario pkgBenchmark.Scenario,
	iteration pkgBenchmark.Iteration,
) (result pkgBenchmark.IterationResult, runError error, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()

	result, runError = scenario.Run(ctx, iteration)

	return result, runError, false
}

func (r *Runner) cleanup(
	ctx context.Context,
	scenario pkgBenchmark.Scenario,
	iteration pkgBenchmark.Iteration,
) (cleanupError error, panicked bool) {
	cleanupContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		r.options.CleanupTimeout,
	)
	defer cancel()
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()

	return scenario.Cleanup(cleanupContext, iteration), false
}

func failedResult(kind string, code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultFailed,
		Error: &pkgBenchmark.ErrorRecord{Kind: kind, Code: code},
	}
}

func failedResultForError(err error, code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{
		State: pkgBenchmark.ResultFailed,
		Error: classifyError(err, "scenario", code),
	}
}

func classifyError(err error, kind string, code string) *pkgBenchmark.ErrorRecord {
	record := &pkgBenchmark.ErrorRecord{Kind: kind, Code: code}
	if errors.Is(err, context.Canceled) {
		record.Kind = "context"
		record.Code = "canceled"
		return record
	}
	if errors.Is(err, context.DeadlineExceeded) {
		record.Kind = "context"
		record.Code = "deadline_exceeded"
		return record
	}

	var providerError *pkgProvider.ProviderError
	if errors.As(err, &providerError) {
		record.Kind = "provider"
		record.Code = string(providerError.Kind)
		record.Operation = string(providerError.Operation)
		record.Provider = string(providerError.Provider)
		record.Model = providerError.Model
		record.StatusCode = providerError.StatusCode
		record.Retryable = providerError.Retryable
	}

	return record
}

func missingScenarioReport(
	definition pkgBenchmark.ScenarioDefinition,
	startedAt time.Time,
) pkgBenchmark.ScenarioReport {
	return pkgBenchmark.ScenarioReport{
		Scenario: definition,
		State:    pkgBenchmark.ResultSkipped,
		Samples: []pkgBenchmark.Sample{{
			Iteration:  pkgBenchmark.Iteration{Index: 1},
			StartedAt:  startedAt,
			State:      pkgBenchmark.ResultSkipped,
			ReasonCode: "scenario_not_registered",
		}},
	}
}

func durationMilliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func randomRunID() string {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err == nil {
		return hex.EncodeToString(random)
	}

	return fmt.Sprintf("local-%d", time.Now().UnixNano())
}
