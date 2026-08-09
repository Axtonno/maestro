package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func runBenchSmoke(arguments []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("maestro bench smoke", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String(
		"manifest",
		"docs/provider-smoke-benchmark-manifest.yaml",
		"path to the benchmark manifest",
	)
	providerID := flags.String("provider", smoke.ProviderOllama, "provider to benchmark")
	outputPath := flags.String("output", "-", "JSON report path, or - for stdout")
	warmup := flags.Int("warmup", 0, "number of warmup iterations")
	runs := flags.Int("runs", 1, "number of measured iterations")
	timeout := flags.Duration("timeout", 2*time.Minute, "timeout per scenario iteration")
	cleanupTimeout := flags.Duration(
		"cleanup-timeout",
		30*time.Second,
		"timeout per scenario cleanup",
	)
	failOnFailure := flags.Bool(
		"fail-on-failure",
		false,
		"return a non-zero status when a smoke scenario fails",
	)
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro bench smoke does not accept positional arguments")
		return 2
	}
	if *timeout <= 0 || *cleanupTimeout <= 0 {
		fmt.Fprintln(stderr, "maestro bench smoke timeouts must be positive")
		return 2
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Warmup: *warmup, Runs: *runs, Timeout: *timeout,
		CleanupTimeout: *cleanupTimeout, Command: "bench smoke",
	})
	if err != nil {
		fmt.Fprintf(stderr, "configure smoke runner: %v\n", err)
		return 2
	}

	manifest, err := internalBenchmark.LoadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "load smoke manifest: %v\n", err)
		return 1
	}
	config, err := smoke.ConfigFromEnvironment(
		manifest,
		*providerID,
		*timeout,
	)
	if err != nil {
		fmt.Fprintf(stderr, "configure smoke benchmark: %v\n", err)
		return 1
	}
	maestroRuntime, err := smoke.NewRuntime(config)
	if err != nil {
		fmt.Fprintf(stderr, "construct smoke runtime: %v\n", err)
		return 1
	}
	scenarios, err := smoke.NewScenarios(
		manifest,
		config,
		maestroRuntime.Providers(),
	)
	if err != nil {
		fmt.Fprintf(stderr, "construct smoke scenarios: %v\n", err)
		return 1
	}
	profile := config.ConfigurationProfile()
	profile.Hardware = pkgBenchmark.HardwareProfile{
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
		LogicalCPUs: runtime.NumCPU(),
	}
	runContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	report, runError := runner.Run(
		runContext,
		manifest,
		profile,
		scenarios...,
	)
	stopSignals()
	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		*cleanupTimeout,
	)
	shutdownError := smoke.Shutdown(shutdownContext, maestroRuntime)
	cancelShutdown()
	if runError != nil {
		fmt.Fprintf(stderr, "execute smoke benchmark: %v\n", runError)
		return 1
	}
	if shutdownError != nil {
		fmt.Fprintf(stderr, "shutdown smoke runtime: %v\n", shutdownError)
		return 1
	}
	if *outputPath == "-" {
		err = internalBenchmark.EncodeReportJSON(stdout, report)
	} else {
		err = internalBenchmark.WriteReportJSON(*outputPath, report)
	}
	if err != nil {
		fmt.Fprintf(stderr, "write smoke report: %v\n", err)
		return 1
	}
	if *failOnFailure && reportHasFailure(report) {
		return 1
	}

	return 0
}

func reportHasFailure(report pkgBenchmark.Report) bool {
	for _, scenario := range report.Scenarios {
		if scenario.State == pkgBenchmark.ResultFailed {
			return true
		}
	}

	return false
}
