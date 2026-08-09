package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/doctor"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout io.Writer, stderr io.Writer) int {
	if len(arguments) > 0 && arguments[0] == "bench" {
		return runBench(arguments[1:], stdout, stderr)
	}
	if len(arguments) > 0 {
		fmt.Fprintf(stderr, "unknown command %q\n", arguments[0])
		printRootUsage(stderr)
		return 2
	}

	fmt.Fprintln(stdout, "Maestro AI Runtime")
	fmt.Fprintln(stdout)

	info := doctor.CollectSystemInfo()

	fmt.Fprintln(stdout, "System")
	fmt.Fprintln(stdout, " OS:", info.OS)
	fmt.Fprintln(stdout, " Arch:", info.Arch)
	fmt.Fprintln(stdout, " CPU:", info.CPU)
	fmt.Fprintln(stdout, " Home:", info.HomeDir)

	return 0
}

func runBench(arguments []string, stdout io.Writer, stderr io.Writer) int {
	if len(arguments) == 0 || arguments[0] == "help" || arguments[0] == "--help" ||
		arguments[0] == "-h" {
		printBenchUsage(stdout)
		return 0
	}
	switch arguments[0] {
	case "validate":
		return runBenchValidate(arguments[1:], stdout, stderr)
	case "smoke":
		return runBenchSmoke(arguments[1:], stdout, stderr)
	case "provider":
		return runBenchRuntime("provider", arguments[1:], stdout, stderr)
	case "model":
		return runBenchRuntime("model", arguments[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown bench command %q\n", arguments[0])
		printBenchUsage(stderr)
		return 2
	}
}

func runBenchValidate(arguments []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("maestro bench validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String(
		"manifest",
		"docs/provider-smoke-benchmark-manifest.yaml",
		"path to the benchmark manifest",
	)
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro bench validate does not accept positional arguments")
		return 2
	}

	manifest, err := internalBenchmark.LoadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "invalid benchmark manifest: %v\n", err)
		return 1
	}
	fmt.Fprintf(
		stdout,
		"benchmark manifest valid: version=%d providers=%d scenarios=%d\n",
		manifest.Version,
		len(manifest.Providers),
		len(manifest.Scenarios),
	)

	return 0
}

func printRootUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: maestro [bench]")
}

func printBenchUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: maestro bench <command>")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "commands:")
	fmt.Fprintln(writer, "  model     benchmark model latency, streaming and lifecycle")
	fmt.Fprintln(writer, "  provider  benchmark provider catalog and resilience")
	fmt.Fprintln(writer, "  smoke     execute the live provider smoke matrix")
	fmt.Fprintln(writer, "  validate  validate a versioned benchmark manifest")
}
