package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/controlledmutation"
	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func runWorkspace(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	if len(arguments) == 0 || arguments[0] == "--help" || arguments[0] == "-h" || arguments[0] == "help" {
		fmt.Fprintln(stdout, "usage: maestro workspace replace --file <path> --lines <start:end> [--config path] <instruction>")
		return 0
	}
	if arguments[0] != "replace" {
		fmt.Fprintln(stderr, "workspace failed: invalid_request")
		return 2
	}
	return runWorkspaceReplace(arguments[1:], stdin, stdout, stderr, dependencies)
}

func runWorkspaceReplace(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	return runControlledMutation("maestro workspace replace", arguments, false, stdin, stdout, stderr, dependencies)
}

func runMutate(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	preview := false
	filtered := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		if argument == "--preview" {
			if preview {
				fmt.Fprintln(stderr, "mutation failed: invalid_request")
				return 2
			}
			preview = true
			continue
		}
		filtered = append(filtered, argument)
	}
	return runControlledMutation("maestro mutate", filtered, preview, stdin, stdout, stderr, dependencies)
}

func runControlledMutation(name string, arguments []string, previewOnly bool, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	if duplicateMutationFlag(arguments) {
		fmt.Fprintln(stderr, "mutation failed: invalid_request")
		return 2
	}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	usage := func() {
		if name == "maestro mutate" {
			fmt.Fprintln(stdout, "usage: maestro mutate [--preview] --file <path> --lines <start:end> [--config path] <instruction>")
			return
		}
		fmt.Fprintln(stdout, "usage: maestro workspace replace --file <path> --lines <start:end> [--config path] <instruction>")
	}
	configPath := flags.String("config", "", "path to Maestro configuration")
	logical := flags.String("file", "", "single logical workspace file")
	lines := flags.String("lines", "", "inclusive start:end line range")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			usage()
			return 0
		}
		fmt.Fprintln(stderr, "mutation failed: invalid_request")
		return 2
	}
	if !previewOnly && (dependencies.isTerminal == nil || !dependencies.isTerminal(stdin)) {
		fmt.Fprintln(stderr, "mutation failed: tty_required")
		return 3
	}
	start, end, ok := parseLineRange(*lines)
	instruction := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if !ok || *logical == "" || instruction == "" || len(instruction) > maxInstructionBytes {
		fmt.Fprintln(stderr, "mutation failed: invalid_request")
		return 2
	}
	config, err := resolveAndLoadMutation(*configPath, dependencies)
	if err != nil {
		fmt.Fprintln(stderr, "mutation failed: invalid_request")
		renderConfigurationDiagnostic(stderr, err)
		return 2
	}
	service, err := controlledmutation.Build(config, controlledMutationDependencies(dependencies))
	if err != nil {
		fmt.Fprintf(stderr, "mutation failed: %s\n", mutationFailureCode(context.Background(), err))
		return mutationExitCode(context.Background(), err)
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	var approver pkgTool.Approver
	var previewer *application.PreviewApprover
	if previewOnly {
		previewer = application.NewPreviewApprover(stderr)
		approver = previewer
	} else {
		approver = application.NewTerminalApprover(bufio.NewReader(stdin), stderr, true)
	}
	result, err := service.Execute(ctx, controlledmutation.Request{File: *logical, StartLine: start, EndLine: end, Instruction: instruction, Approver: approver})
	if err != nil {
		if previewOnly && errors.Is(err, controlledmutation.ErrApprovalRejected) && previewer.Previewed() {
			fmt.Fprintln(stdout, "mode\tcontrolled_mutation")
			fmt.Fprintln(stdout, "terminal\tpreviewed")
			fmt.Fprintf(stdout, "model\t%s\n", result.Model)
			fmt.Fprintf(stdout, "file\t%s\n", *logical)
			fmt.Fprintf(stdout, "lines\t%d:%d\n", start, end)
			fmt.Fprintln(stdout, "effect\tunchanged")
			fmt.Fprintln(stdout, "durable\tfalse")
			return 0
		}
		fmt.Fprintf(stderr, "mutation failed: %s\n", mutationFailureCode(ctx, err))
		return mutationExitCode(ctx, err)
	}
	fmt.Fprintln(stdout, "mode\tcontrolled_mutation")
	fmt.Fprintf(stdout, "terminal\t%s\n", result.Terminal)
	fmt.Fprintf(stdout, "model\t%s\n", result.Model)
	fmt.Fprintf(stdout, "file\t%s\n", *logical)
	fmt.Fprintf(stdout, "lines\t%d:%d\n", start, end)
	fmt.Fprintf(stdout, "effect\t%s\n", result.Effect)
	fmt.Fprintf(stdout, "durable\t%t\n", result.Durable)
	return 0
}

func controlledMutationDependencies(dependencies commandDependencies) controlledmutation.Dependencies {
	configured := dependencies.application
	result := controlledmutation.Dependencies{Getenv: configured.Getenv}
	result.AfterPreview = dependencies.mutationAfterPreview
	if configured.ProviderFactory != nil {
		result.ProviderFactory = func(config productconfig.Config, secret string) (pkgProvider.Provider, error) {
			return configured.ProviderFactory(config, secret)
		}
	}
	return result
}

func parseLineRange(value string) (int, int, bool) {
	if strings.TrimSpace(value) != value || strings.Count(value, ":") != 1 {
		return 0, 0, false
	}
	parts := strings.Split(value, ":")
	start, errStart := strconv.Atoi(parts[0])
	end, errEnd := strconv.Atoi(parts[1])
	return start, end, errStart == nil && errEnd == nil && start > 0 && end >= start
}

func duplicateMutationFlag(arguments []string) bool {
	seen := map[string]bool{}
	for _, argument := range arguments {
		for _, name := range []string{"--config", "--file", "--lines"} {
			if argument == name || strings.HasPrefix(argument, name+"=") {
				if seen[name] {
					return true
				}
				seen[name] = true
			}
		}
	}
	return false
}

func mutationFailureCode(ctx context.Context, err error) string {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(ctx.Err(), context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(ctx.Err(), context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, controlledmutation.ErrTTYRequired):
		return "tty_required"
	case errors.Is(err, controlledmutation.ErrApprovalRejected), errors.Is(err, pkgTool.ErrPermissionDenied):
		return "approval_rejected"
	case errors.Is(err, controlledmutation.ErrInsufficientInfo), errors.Is(err, mutation.ErrInsufficientInformation):
		return "insufficient_information"
	case errors.Is(err, controlledmutation.ErrStaleSource), errors.Is(err, mutation.ErrStaleSource):
		return "stale_source"
	case errors.Is(err, controlledmutation.ErrModelIdentity):
		return "model_identity_mismatch"
	case errors.Is(err, controlledmutation.ErrProviderUnavailable):
		return "provider_unavailable"
	case errors.Is(err, controlledmutation.ErrCapabilityUnsupported):
		return "capability_unsupported"
	case errors.Is(err, controlledmutation.ErrResponseInvalid):
		return "response_invalid"
	case errors.Is(err, mutation.ErrSelectionOutOfBounds):
		return "selection_out_of_bounds"
	case errors.Is(err, mutation.ErrSensitiveTarget), errors.Is(err, pkgContext.ErrInvalidPath), errors.Is(err, pkgContext.ErrInvalidWorkspace):
		return "file_not_allowed"
	case errors.Is(err, controlledmutation.ErrInvalidRequest), errors.Is(err, productconfig.ErrInvalid):
		return "invalid_request"
	default:
		return "execution_failed"
	}
}

func mutationExitCode(ctx context.Context, err error) int {
	code := mutationFailureCode(ctx, err)
	switch code {
	case "canceled":
		return 130
	case "invalid_request", "selection_out_of_bounds", "file_not_allowed":
		return 2
	case "tty_required", "approval_rejected", "insufficient_information", "stale_source":
		return 3
	case "provider_unavailable", "capability_unsupported", "model_identity_mismatch":
		return 4
	default:
		return 1
	}
}
