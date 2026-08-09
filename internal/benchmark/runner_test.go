package benchmark

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestRunnerExecutesWarmupRunsCleanupAndAggregates(t *testing.T) {
	definition := testScenarioDefinition()
	values := []float64{999, 10, 20, 30}
	runCount := 0
	cleanupCount := 0
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			value := values[runCount]
			runCount++
			return pkgBenchmark.IterationResult{
				State: pkgBenchmark.ResultPassed,
				Measurements: []pkgBenchmark.Measurement{{
					Name: "total_latency_ms", Unit: "ms", Value: value,
				}},
			}, nil
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			cleanupCount++
			return nil
		},
	}
	runner := deterministicRunner(t, RunnerOptions{Warmup: 1, Runs: 3})

	report, err := runner.Run(
		context.Background(),
		testManifest(definition),
		pkgBenchmark.ConfigurationProfile{},
		scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	if runCount != 4 || cleanupCount != 4 {
		t.Fatalf("runs=%d cleanup=%d, expected 4 each", runCount, cleanupCount)
	}
	if report.RunID != "deterministic-run" || len(report.Scenarios) != 1 ||
		len(report.Scenarios[0].Samples) != 4 {
		t.Fatalf("unexpected report: %#v", report)
	}
	aggregate := report.Scenarios[0].Aggregates[0]
	if aggregate.Count != 3 || aggregate.Min != 10 || aggregate.Median != 20 ||
		aggregate.P95 != nil || aggregate.Max != 30 {
		t.Fatalf("warmup leaked into aggregate: %#v", aggregate)
	}
	if report.Configuration.Execution.Warmup != 1 ||
		report.Configuration.Execution.Runs != 3 ||
		report.Configuration.Execution.CleanupTimeoutMS != 30000 {
		t.Fatalf("unexpected execution profile: %#v", report.Configuration.Execution)
	}
}

func TestRunnerAlwaysCleansUpAfterPanicWithIndependentContext(t *testing.T) {
	definition := testScenarioDefinition()
	cleanupCalled := false
	cleanupWasCanceled := false
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			panic("sensitive panic payload")
		},
		CleanupFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) error {
			cleanupCalled = true
			cleanupWasCanceled = ctx.Err() != nil
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})

	report, err := runner.Run(
		ctx,
		testManifest(definition),
		pkgBenchmark.ConfigurationProfile{},
		scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	sample := report.Scenarios[0].Samples[0]
	if !cleanupCalled || cleanupWasCanceled {
		t.Fatalf("cleanup called=%t canceled=%t", cleanupCalled, cleanupWasCanceled)
	}
	if sample.State != pkgBenchmark.ResultFailed || sample.Error == nil ||
		sample.Error.Code != "scenario_panic" {
		t.Fatalf("panic was not classified: %#v", sample)
	}
}

func TestRunnerClassifiesProviderCleanupErrorsWithoutMessages(t *testing.T) {
	definition := testScenarioDefinition()
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultPassed}, nil
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			return pkgProvider.NewProviderError(pkgProvider.ProviderErrorDetails{
				Kind: pkgProvider.ErrorKindUnavailable, Operation: pkgProvider.OperationModelUnload,
				Provider: "ollama", Model: "/home/user/private/model.gguf",
				StatusCode: 503, Retryable: true, Message: "secret remote payload",
			}, errors.New("secret cause"))
		},
	}
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), testManifest(definition),
		pkgBenchmark.ConfigurationProfile{}, scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	record := report.Scenarios[0].Samples[0].CleanupError
	if record == nil || record.Kind != "provider" || record.Code != "unavailable" ||
		record.Operation != "model_unload" || record.StatusCode != 503 || !record.Retryable {
		t.Fatalf("unexpected cleanup classification: %#v", record)
	}
}

func TestRunnerClassifiesProviderScenarioErrors(t *testing.T) {
	definition := testScenarioDefinition()
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{}, pkgProvider.NewProviderError(
				pkgProvider.ProviderErrorDetails{
					Kind:      pkgProvider.ErrorKindRateLimited,
					Operation: pkgProvider.OperationCompletion,
					Provider:  "ollama", Model: "qwen", StatusCode: 429,
					Retryable: true, Message: "sensitive remote error",
				},
				errors.New("sensitive cause"),
			)
		},
	}
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), testManifest(definition),
		pkgBenchmark.ConfigurationProfile{}, scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	record := report.Scenarios[0].Samples[0].Error
	if record == nil || record.Kind != "provider" || record.Code != "rate_limited" ||
		record.Operation != "completion" || record.Provider != "ollama" ||
		record.Model != "qwen" || record.StatusCode != 429 || !record.Retryable {
		t.Fatalf("unexpected provider classification: %#v", record)
	}
}

func TestRunnerClassifiesIterationDeadline(t *testing.T) {
	definition := testScenarioDefinition()
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			<-ctx.Done()
			return pkgBenchmark.IterationResult{}, ctx.Err()
		},
	}
	runner := deterministicRunner(t, RunnerOptions{
		Runs: 1, Timeout: time.Millisecond, CleanupTimeout: time.Second,
	})

	report, err := runner.Run(
		context.Background(), testManifest(definition),
		pkgBenchmark.ConfigurationProfile{}, scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	record := report.Scenarios[0].Samples[0].Error
	if record == nil || record.Kind != "context" ||
		record.Code != "deadline_exceeded" {
		t.Fatalf("unexpected deadline classification: %#v", record)
	}
}

func TestRunnerReportsUnregisteredScenariosAsSkipped(t *testing.T) {
	definition := testScenarioDefinition()
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), testManifest(definition),
		pkgBenchmark.ConfigurationProfile{},
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	scenario := report.Scenarios[0]
	if scenario.State != pkgBenchmark.ResultSkipped ||
		scenario.Samples[0].ReasonCode != "scenario_not_registered" {
		t.Fatalf("unexpected missing scenario report: %#v", scenario)
	}
}

func TestRunnerRejectsTypedNilScenarios(t *testing.T) {
	type pointerScenario struct {
		pkgBenchmark.ScenarioFuncs
	}
	var scenario *pointerScenario
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})

	_, err := runner.Run(
		context.Background(), testManifest(testScenarioDefinition()),
		pkgBenchmark.ConfigurationProfile{}, scenario,
	)
	if err == nil {
		t.Fatal("expected typed nil scenario rejection")
	}
}

func deterministicRunner(t *testing.T, options RunnerOptions) *Runner {
	t.Helper()
	current := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	now := func() time.Time {
		value := current
		current = current.Add(time.Millisecond)
		return value
	}
	runner, err := newRunner(options, now, func() string { return "deterministic-run" })
	if err != nil {
		t.Fatalf("construct runner: %v", err)
	}

	return runner
}

func testScenarioDefinition() pkgBenchmark.ScenarioDefinition {
	return pkgBenchmark.ScenarioDefinition{
		ID: "completion", Capability: "completion", ModelRole: "chat", Cleanup: "none",
	}
}

func testManifest(definition pkgBenchmark.ScenarioDefinition) pkgBenchmark.Manifest {
	return pkgBenchmark.Manifest{
		Version: pkgBenchmark.ManifestSchemaVersion,
		Owner:   "tests", SourceMilestone: "tests",
		ReportRedaction: pkgBenchmark.RedactionPolicy{Exclude: []string{
			"prompts", "responses", "tool_arguments", "tool_results",
			"credentials", "user_paths",
		}},
		Providers: map[string]pkgBenchmark.ProviderManifest{"test": {}},
		Scenarios: []pkgBenchmark.ScenarioDefinition{definition},
		ResultStates: []pkgBenchmark.ResultState{
			pkgBenchmark.ResultPassed, pkgBenchmark.ResultFailed,
			pkgBenchmark.ResultSkipped, pkgBenchmark.ResultUnsupported,
		},
	}
}
