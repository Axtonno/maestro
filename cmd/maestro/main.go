package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/antonio-cafeo/maestro/internal/application"
	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/buildinfo"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout io.Writer, stderr io.Writer) int {
	return runWithIO(arguments, os.Stdin, stdout, stderr, defaultCommandDependencies())
}

type commandDependencies struct {
	application application.Dependencies
	buildInfo   func() buildinfo.Info
	context     func() (context.Context, context.CancelFunc)
	isTerminal  func(io.Reader) bool
}

func defaultCommandDependencies() commandDependencies {
	return commandDependencies{
		application: application.DefaultDependencies(),
		buildInfo:   buildinfo.Current,
		context: func() (context.Context, context.CancelFunc) {
			return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		},
		isTerminal: isTerminalReader,
	}
}

func isTerminalReader(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func runWithIO(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	if len(arguments) == 0 || arguments[0] == "--help" || arguments[0] == "-h" || (arguments[0] == "help" && len(arguments) == 1) {
		printRootUsage(stdout)
		return 0
	}
	if arguments[0] == "help" {
		return runWithIO([]string{arguments[1], "--help"}, stdin, stdout, stderr, dependencies)
	}
	switch arguments[0] {
	case "doctor":
		return runDoctor(arguments[1:], stdout, stderr, dependencies)
	case "models":
		return runModels(arguments[1:], stdout, stderr, dependencies)
	case "agents":
		return runAgents(arguments[1:], stdout, stderr, dependencies)
	case "run":
		return runAgent(arguments[1:], stdin, stdout, stderr, dependencies)
	case "version":
		return runVersion(arguments[1:], stdout, stderr, dependencies)
	case "bench":
		return runBench(arguments[1:], stdin, stdout, stderr, dependencies)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", arguments[0])
		printRootUsage(stderr)
		return 2
	}
}

func runBench(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
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
	case "laravel":
		return runBenchLaravel(arguments[1:], stdout, stderr)
	case "mutation":
		return runBenchMutation(arguments[1:], stdin, stdout, stderr, dependencies)
	case "render":
		return runBenchRender(arguments[1:], stdout, stderr)
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
	fmt.Fprintln(writer, "usage: maestro <command>")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "commands:")
	fmt.Fprintln(writer, "  doctor   validate configuration and operational prerequisites")
	fmt.Fprintln(writer, "  models   list models from the explicitly configured provider")
	fmt.Fprintln(writer, "  agents   list registered agents and capabilities")
	fmt.Fprintln(writer, "  run      execute the configured reference agent")
	fmt.Fprintln(writer, "  version  print build version and commit")
	fmt.Fprintln(writer, "  bench    run benchmark and evaluation commands")
}

func printBenchUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: maestro bench <command>")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "commands:")
	fmt.Fprintln(writer, "  laravel   execute the versioned developer benchmark")
	fmt.Fprintln(writer, "  model     benchmark model latency, streaming and lifecycle")
	fmt.Fprintln(writer, "  mutation  validate and execute the mutation qualification matrix")
	fmt.Fprintln(writer, "  provider  benchmark provider catalog and resilience")
	fmt.Fprintln(writer, "  render    derive a Markdown report from benchmark JSON")
	fmt.Fprintln(writer, "  smoke     execute the live provider smoke matrix")
	fmt.Fprintln(writer, "  validate  validate a versioned benchmark manifest")
}
