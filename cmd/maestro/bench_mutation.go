package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/application"
	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/mutation"
	"github.com/antonio-cafeo/maestro/internal/buildinfo"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func runBenchMutation(
	arguments []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	dependencies commandDependencies,
) int {
	flags := flag.NewFlagSet("maestro bench mutation", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profilePath := flags.String(
		"profile",
		"docs/mutation-qualification-profile.yaml",
		"path to the mutation qualification profile",
	)
	mode := flags.String("mode", "validate", "execution mode: validate, deterministic, preflight, gate-a, gate-b, or gate-c")
	outputPath := flags.String("output", "-", "deterministic JSON report path, or - for stdout")
	markdownPath := flags.String("markdown", "", "optional deterministic Markdown report path")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro bench mutation does not accept positional arguments")
		return 2
	}
	profile, err := mutation.LoadProfile(*profilePath)
	if err != nil {
		fmt.Fprintf(stderr, "load mutation qualification profile: %v\n", err)
		return 1
	}
	fixture, err := mutation.MaterializeFixture(context.Background(), profile)
	if err != nil {
		fmt.Fprintf(stderr, "validate mutation qualification fixture: %v\n", err)
		return 1
	}
	if err := fixture.Cleanup(); err != nil {
		fmt.Fprintln(stderr, "validate mutation qualification fixture: cleanup_failed")
		return 1
	}
	if *mode == "deterministic" {
		return runDeterministicMutationBenchmark(profile, *outputPath, *markdownPath, stdout, stderr)
	}
	if *mode == "preflight" {
		if *markdownPath != "" || *outputPath != "-" {
			fmt.Fprintln(stderr, "mutation qualification preflight does not write a gate report")
			return 2
		}
		return runMutationPreflight(profile, stdout, stderr, dependencies)
	}
	if *mode == "gate-a" || *mode == "gate-b" || *mode == "gate-c" {
		return runLiveMutationBenchmark(
			profile, *mode, *outputPath, *markdownPath, stdin, stdout, stderr, dependencies,
		)
	}
	if *mode != "validate" {
		fmt.Fprintf(stderr, "mutation qualification mode %q is invalid\n", *mode)
		return 2
	}
	if *markdownPath != "" || *outputPath != "-" {
		fmt.Fprintln(stderr, "mutation qualification report paths require --mode deterministic")
		return 2
	}
	fmt.Fprintf(stdout, "mutation qualification profile valid: version=%d gates=%d scenarios=%d\n",
		profile.Version, len(profile.Gates), len(profile.MutationMatrix))
	fmt.Fprintf(stdout, "profile_sha256\t%s\n", profile.Digest())
	fmt.Fprintf(stdout, "candidate\t%s/%s\t%s\n", profile.Target.Platform, profile.Target.Provider, profile.Target.Model)
	return 0
}

func runMutationPreflight(
	profile mutation.Profile,
	stdout io.Writer,
	stderr io.Writer,
	dependencies commandDependencies,
) int {
	hardware, err := internalBenchmark.CollectHardwareProfile()
	if err != nil {
		fmt.Fprintln(stderr, "fail\thardware\tprofile_unavailable")
		return 1
	}
	failed := false
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		fmt.Fprintln(stdout, "fail\tplatform\ttarget_mismatch")
		failed = true
	} else {
		fmt.Fprintln(stdout, "pass\tplatform\tlinux_amd64")
	}
	lower := profile.Target.HardwareLowerBound
	swapMB := linuxSwapMB()
	if hardware.LogicalCPUs < lower.LogicalCPUs || hardware.MemoryMB < lower.RAMGiB*1024 ||
		swapMB+1 < lower.SwapGiB*1024 || !strings.Contains(hardware.CPU, lower.CPU) {
		fmt.Fprintln(stdout, "fail\thardware\tlower_bound_not_met")
		failed = true
	} else {
		fmt.Fprintf(stdout, "pass\thardware\tlogical_cpus=%d memory_mb=%d swap_mb=%d\n",
			hardware.LogicalCPUs, hardware.MemoryMB, swapMB)
	}
	fixture, err := mutation.MaterializeFixture(context.Background(), profile)
	if err != nil {
		fmt.Fprintln(stdout, "fail\tfixture\tinvalid")
		return 1
	}
	defer fixture.Cleanup()
	config, err := loadMutationPreflightConfig(profile, fixture.Root)
	if err != nil {
		fmt.Fprintln(stdout, "fail\tconfiguration\tinvalid")
		return 1
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	checks := application.Doctor(ctx, config, dependencies.application)
	for _, check := range checks {
		status := string(check.Status)
		fmt.Fprintf(stdout, "%s\tdoctor_%s\t%s\n", status, check.Name, check.Detail)
		failed = failed || check.Status == application.CheckFail
	}
	if failed {
		return 1
	}
	return 0
}

func loadMutationPreflightConfig(profile mutation.Profile, root string) (productconfig.Config, error) {
	name, err := profile.Resolve(profile.Configuration.ProductProfile)
	if err != nil {
		return productconfig.Config{}, err
	}
	config, err := productconfig.Load(name)
	if err != nil {
		return productconfig.Config{}, err
	}
	config.Workspace.Root = root
	return config, config.ValidateExecutionProfile()
}

func linuxSwapMB() int64 {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "SwapTotal:" {
			continue
		}
		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err == nil && value >= 0 {
			return value / 1024
		}
	}
	return 0
}

func runLiveMutationBenchmark(
	profile mutation.Profile,
	mode string,
	outputPath string,
	markdownPath string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	dependencies commandDependencies,
) int {
	if err := validateBenchmarkReportPaths(outputPath, markdownPath); err != nil {
		fmt.Fprintf(stderr, "mutation qualification output: %v\n", err)
		return 2
	}
	gate := mutation.GateA
	if mode == "gate-b" {
		gate = mutation.GateB
	} else if mode == "gate-c" {
		gate = mutation.GateC
	}
	interactive := dependencies.isTerminal != nil && dependencies.isTerminal(stdin)
	if gate == mutation.GateC && !interactive {
		fmt.Fprintln(stderr, "mutation qualification Gate C requires an interactive TTY")
		return 3
	}
	hardware, err := internalBenchmark.CollectHardwareProfile()
	if err != nil {
		fmt.Fprintln(stderr, "mutation qualification hardware: unavailable")
		return 1
	}
	build := mutationBuildIdentity()
	if executable, executableError := os.Executable(); executableError == nil {
		build.Digest, _ = fileSHA256(executable)
	}
	options := mutation.LiveOptions{
		Profile: profile, Dependencies: dependencies.application,
		Hardware: hardware, Build: build,
	}
	if gate == mutation.GateC {
		input := bufio.NewReader(stdin)
		options.Approver = func(int) pkgTool.Approver {
			return application.NewTerminalApprover(input, stderr, true)
		}
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	report, err := mutation.RunLiveGate(ctx, gate, options)
	if err != nil {
		fmt.Fprintln(stderr, "mutation qualification live gate: execution_failed")
		return 1
	}
	if err := writeMutationQualificationReport(outputPath, markdownPath, stdout, report); err != nil {
		fmt.Fprintln(stderr, "mutation qualification live report: write_failed")
		return 1
	}
	if report.State != "passed" {
		return 1
	}
	return 0
}

func writeMutationQualificationReport(
	outputPath string,
	markdownPath string,
	stdout io.Writer,
	report mutation.Report,
) error {
	if outputPath == "-" {
		if err := mutation.EncodeReportJSON(stdout, report); err != nil {
			return err
		}
	} else if err := mutation.WriteReport(outputPath, func(writer io.Writer) error {
		return mutation.EncodeReportJSON(writer, report)
	}); err != nil {
		return err
	}
	if markdownPath != "" {
		return mutation.WriteReport(markdownPath, func(writer io.Writer) error {
			return mutation.EncodeReportMarkdown(writer, report)
		})
	}
	return nil
}

func runDeterministicMutationBenchmark(
	profile mutation.Profile,
	outputPath string,
	markdownPath string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if err := validateBenchmarkReportPaths(outputPath, markdownPath); err != nil {
		fmt.Fprintf(stderr, "mutation qualification output: %v\n", err)
		return 2
	}
	repository, err := profile.Resolve("..")
	if err != nil {
		fmt.Fprintln(stderr, "mutation qualification suite: repository_unavailable")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), profile.Protocol.RunDeadline.Duration)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-count=1",
		"./internal/tool", "./internal/application", "./internal/agent", "./internal/benchmark/mutation")
	command.Dir = repository
	command.Env = os.Environ()
	if os.Getenv("GOCACHE") == "" {
		command.Env = append(command.Env, "GOCACHE=/tmp/maestro-go-build")
	}
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		fmt.Fprintln(stderr, "mutation qualification deterministic suite: failed")
		return 1
	}
	hardware, err := internalBenchmark.CollectHardwareProfile()
	if err != nil {
		fmt.Fprintln(stderr, "mutation qualification hardware: unavailable")
		return 1
	}
	build := mutationBuildIdentity()
	if executable, executableError := os.Executable(); executableError == nil {
		build.Digest, _ = fileSHA256(executable)
	}
	report, err := mutation.NewDeterministicReport(ctx, profile, hardware, build, time.Now().UTC())
	if err != nil {
		fmt.Fprintln(stderr, "mutation qualification deterministic report: invalid")
		return 1
	}
	if outputPath == "-" {
		if err := mutation.EncodeReportJSON(stdout, report); err != nil {
			fmt.Fprintln(stderr, "mutation qualification deterministic report: write_failed")
			return 1
		}
	} else if err := mutation.WriteReport(outputPath, func(writer io.Writer) error {
		return mutation.EncodeReportJSON(writer, report)
	}); err != nil {
		fmt.Fprintln(stderr, "mutation qualification deterministic report: write_failed")
		return 1
	}
	if markdownPath != "" {
		if err := mutation.WriteReport(markdownPath, func(writer io.Writer) error {
			return mutation.EncodeReportMarkdown(writer, report)
		}); err != nil {
			fmt.Fprintln(stderr, "mutation qualification deterministic Markdown: write_failed")
			return 1
		}
	}
	return 0
}

func fileSHA256(name string) (string, error) {
	file, err := os.Open(filepath.Clean(name))
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func mutationBuildIdentity() mutation.ReportBuild {
	current := buildinfo.Current()
	build := mutation.ReportBuild{Version: current.Version, Commit: current.Commit}
	if current.Version == "devel" || current.Commit == "unknown" {
		version, commit := internalBenchmark.BuildMetadata()
		if version != "" {
			build.Version = version
		}
		if commit != "" {
			build.Commit = commit
		}
	}
	return build
}
