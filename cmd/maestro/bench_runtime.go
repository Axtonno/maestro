package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/runtimebench"
	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func runBenchRuntime(kind runtimebench.Kind, arguments []string, stdout io.Writer, stderr io.Writer) int {
	command := "maestro bench " + string(kind)
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "docs/runtime-benchmark-manifest.yaml", "path to the runtime benchmark manifest")
	providerID := flags.String("provider", smoke.ProviderOllama, "provider to benchmark")
	outputPath := flags.String("output", "-", "JSON report path, or - for stdout")
	markdownPath := flags.String("markdown", "", "optional Markdown report path")
	warmup := flags.Int("warmup", 1, "number of warmup iterations")
	runs := flags.Int("runs", 5, "number of measured iterations")
	timeout := flags.Duration("timeout", 5*time.Minute, "timeout per scenario iteration")
	cleanupTimeout := flags.Duration("cleanup-timeout", 30*time.Second, "timeout per scenario cleanup")
	providerPID := flags.Int("provider-pid", 0, "Linux provider PID to sample; zero samples Maestro")
	sampleInterval := flags.Duration("resource-sample-interval", 50*time.Millisecond, "Linux process resource sample interval")
	failOnFailure := flags.Bool("fail-on-failure", false, "return a non-zero status when a runtime scenario fails")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "%s does not accept positional arguments\n", command)
		return 2
	}
	if *timeout <= 0 || *cleanupTimeout <= 0 || *sampleInterval <= 0 || *providerPID < 0 {
		fmt.Fprintf(stderr, "%s timeouts, sample interval and PID must be valid\n", command)
		return 2
	}
	if err := validateBenchmarkReportPaths(*outputPath, *markdownPath); err != nil {
		fmt.Fprintf(stderr, "%s output: %v\n", command, err)
		return 2
	}
	hardware, err := internalBenchmark.CollectHardwareProfile()
	if err != nil {
		fmt.Fprintf(stderr, "collect benchmark hardware profile: %v\n", err)
		return 1
	}
	version, commit := internalBenchmark.BuildMetadata()

	manifest, err := internalBenchmark.LoadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "load runtime benchmark manifest: %v\n", err)
		return 1
	}
	manifest, err = runtimebench.SelectManifest(manifest, kind)
	if err != nil {
		fmt.Fprintf(stderr, "select runtime benchmark scenarios: %v\n", err)
		return 1
	}
	config, err := smoke.ConfigFromEnvironment(manifest, *providerID, *timeout)
	if err != nil {
		fmt.Fprintf(stderr, "configure runtime benchmark: %v\n", err)
		return 1
	}
	maestroRuntime, err := smoke.NewRuntime(config)
	if err != nil {
		fmt.Fprintf(stderr, "construct runtime benchmark: %v\n", err)
		return 1
	}
	sampler, err := runtimebench.NewProcessSampler(*providerPID, *sampleInterval)
	if err != nil {
		fmt.Fprintf(stderr, "configure runtime resource sampler: %v\n", err)
		return 1
	}
	scenarios, err := runtimebench.NewScenarios(manifest, config, maestroRuntime.Providers(), sampler)
	if err != nil {
		fmt.Fprintf(stderr, "construct runtime benchmark scenarios: %v\n", err)
		return 1
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Warmup: *warmup, Runs: *runs, Timeout: *timeout,
		CleanupTimeout: *cleanupTimeout, Command: "bench " + string(kind),
		MaestroVersion: version, MaestroCommit: commit,
	})
	if err != nil {
		fmt.Fprintf(stderr, "configure runtime benchmark runner: %v\n", err)
		return 2
	}
	profile := config.ConfigurationProfile()
	profile.Hardware = hardware
	profile.Generation.MaxTokens = 128
	if model := config.Models["chat"]; model != "" {
		profile.Model = pkgBenchmark.ModelProfile{ID: model}
	}

	runContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	report, runError := runner.Run(runContext, manifest, profile, scenarios...)
	stopSignals()
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), *cleanupTimeout)
	shutdownError := smoke.Shutdown(shutdownContext, maestroRuntime)
	cancelShutdown()
	if runError != nil {
		fmt.Fprintf(stderr, "execute runtime benchmark: %v\n", runError)
		return 1
	}
	if shutdownError != nil {
		fmt.Fprintf(stderr, "shutdown runtime benchmark: %v\n", shutdownError)
		return 1
	}
	err = writeBenchmarkReports(*outputPath, *markdownPath, stdout, report)
	if err != nil {
		fmt.Fprintf(stderr, "write runtime benchmark report: %v\n", err)
		return 1
	}
	if *failOnFailure && reportHasFailure(report) {
		return 1
	}
	return 0
}
