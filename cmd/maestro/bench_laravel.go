package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	maestro "github.com/antonio-cafeo/maestro"
	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/developer"
	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	"github.com/antonio-cafeo/maestro/pkg/plugin/laravel"
)

func runBenchLaravel(arguments []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("maestro bench laravel", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "docs/developer-benchmark-manifest.yaml", "path to the developer benchmark manifest")
	providerID := flags.String("provider", smoke.ProviderOllama, "provider to benchmark")
	datasetID := flags.String("dataset", developer.DatasetID+"@"+developer.DatasetVersion, "embedded dataset identity")
	outputPath := flags.String("output", "-", "JSON report path, or - for stdout")
	warmup := flags.Int("warmup", 0, "number of warmup iterations")
	runs := flags.Int("runs", 1, "number of measured iterations")
	timeout := flags.Duration("timeout", 5*time.Minute, "timeout per scenario iteration")
	cleanupTimeout := flags.Duration("cleanup-timeout", 30*time.Second, "timeout for runtime and scenario cleanup")
	failOnFailure := flags.Bool("fail-on-failure", false, "return non-zero when a technical scenario fails")
	minimumScore := flags.Int("minimum-score", -1, "optional quality gate from 0 to 3; -1 disables it")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro bench laravel does not accept positional arguments")
		return 2
	}
	if *timeout <= 0 || *cleanupTimeout <= 0 || *minimumScore < -1 || *minimumScore > 3 {
		fmt.Fprintln(stderr, "maestro bench laravel timeouts and minimum score are invalid")
		return 2
	}
	expectedDataset := developer.DatasetID + "@" + developer.DatasetVersion
	if *datasetID != expectedDataset {
		fmt.Fprintf(stderr, "unsupported developer dataset %q; expected %q\n", *datasetID, expectedDataset)
		return 2
	}

	manifest, err := internalBenchmark.LoadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "load developer benchmark manifest: %v\n", err)
		return 1
	}
	dataset, err := developer.LoadDataset()
	if err != nil {
		fmt.Fprintf(stderr, "load developer benchmark dataset: %v\n", err)
		return 1
	}
	workspace, cleanupWorkspace, err := dataset.Materialize()
	if err != nil {
		fmt.Fprintf(stderr, "materialize developer benchmark dataset: %v\n", err)
		return 1
	}
	defer func() { _ = cleanupWorkspace() }()

	config, err := smoke.ConfigFromEnvironment(manifest, *providerID, *timeout)
	if err != nil {
		fmt.Fprintf(stderr, "configure developer benchmark: %v\n", err)
		return 1
	}
	maestroRuntime, err := smoke.NewRuntime(config)
	if err != nil {
		fmt.Fprintf(stderr, "construct developer benchmark runtime: %v\n", err)
		return 1
	}
	runtimeActive := true
	defer func() {
		if !runtimeActive {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), *cleanupTimeout)
		defer cancel()
		_ = maestroRuntime.Stop(ctx)
		_ = smoke.Shutdown(ctx, maestroRuntime)
	}()
	if err := maestroRuntime.Plugins().RegisterLoader(
		laravel.ID,
		laravel.NewLoader(laravel.Config{Root: workspace}),
	); err != nil {
		fmt.Fprintf(stderr, "register Laravel benchmark plugin: %v\n", err)
		return 1
	}
	if _, err := maestroRuntime.Plugins().Load(context.Background(), laravel.ID); err != nil {
		fmt.Fprintf(stderr, "load Laravel benchmark plugin: %v\n", err)
		return 1
	}
	if err := maestroRuntime.Start(context.Background()); err != nil {
		fmt.Fprintf(stderr, "start Laravel benchmark plugin: %v\n", err)
		return 1
	}

	scenarios, err := developer.NewScenarios(manifest, dataset, config, maestroRuntime.Providers())
	if err != nil {
		fmt.Fprintf(stderr, "construct developer benchmark scenarios: %v\n", err)
		return 1
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Warmup: *warmup, Runs: *runs, Timeout: *timeout,
		CleanupTimeout: *cleanupTimeout, Command: "bench laravel",
	})
	if err != nil {
		fmt.Fprintf(stderr, "configure developer benchmark runner: %v\n", err)
		return 2
	}
	profile := config.ConfigurationProfile()
	profile.Hardware = pkgBenchmark.HardwareProfile{OS: runtime.GOOS, Architecture: runtime.GOARCH, LogicalCPUs: runtime.NumCPU()}
	profile.Dataset = pkgBenchmark.DatasetProfile{ID: dataset.ID, Version: dataset.Version}
	profile.Plugins = []pkgBenchmark.PluginProfile{{ID: string(laravel.ID), Version: laravel.Version}}
	profile.Generation.MaxTokens = 1024
	if model := config.Models["chat"]; model != "" {
		profile.Model = pkgBenchmark.ModelProfile{ID: model}
	}

	runContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	report, runError := runner.Run(runContext, manifest, profile, scenarios...)
	stopSignals()
	shutdownError := shutdownDeveloperBenchmark(*cleanupTimeout, maestroRuntime, cleanupWorkspace)
	runtimeActive = false
	if runError != nil {
		fmt.Fprintf(stderr, "execute developer benchmark: %v\n", runError)
		return 1
	}
	if shutdownError != nil {
		fmt.Fprintf(stderr, "shutdown developer benchmark: %v\n", shutdownError)
		return 1
	}
	if *outputPath == "-" {
		err = internalBenchmark.EncodeReportJSON(stdout, report)
	} else {
		err = internalBenchmark.WriteReportJSON(*outputPath, report)
	}
	if err != nil {
		fmt.Fprintf(stderr, "write developer benchmark report: %v\n", err)
		return 1
	}
	if *failOnFailure && reportHasFailure(report) {
		return 1
	}
	if *minimumScore >= 0 && reportBelowMinimumScore(report, *minimumScore) {
		return 1
	}
	return 0
}

func shutdownDeveloperBenchmark(timeout time.Duration, runtime maestro.Runtime, cleanup func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return errors.Join(runtime.Stop(ctx), smoke.Shutdown(ctx, runtime), cleanup())
}

func reportBelowMinimumScore(report pkgBenchmark.Report, minimum int) bool {
	for _, scenario := range report.Scenarios {
		for _, sample := range scenario.Samples {
			if sample.Iteration.Warmup {
				continue
			}
			if sample.Evaluation == nil || sample.Evaluation.Score < minimum {
				return true
			}
		}
	}
	return false
}
