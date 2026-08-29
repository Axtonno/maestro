package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/buildinfo"
	"github.com/antonio-cafeo/maestro/internal/directchat"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
)

const maxInstructionBytes = 1 << 20

func runDoctor(arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	flags := flag.NewFlagSet("maestro doctor", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	usage := func() { fmt.Fprintln(stdout, "usage: maestro doctor [--config path] [--mode agent|chat]") }
	flags.Usage = func() {}
	configPath := flags.String("config", "", "path to Maestro configuration")
	mode := flags.String("mode", "agent", "execution mode to validate")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			usage()
			return 0
		}
		fmt.Fprintln(stderr, "doctor failed: invalid_request")
		return 2
	}
	if flags.NArg() != 0 || (*mode != "agent" && *mode != "chat") {
		fmt.Fprintln(stderr, "doctor failed: invalid_request")
		return 2
	}
	var config productconfig.Config
	var err error
	if *mode == "chat" {
		config, err = resolveAndLoadChat(*configPath, dependencies)
	} else {
		config, err = resolveAndLoad(*configPath, dependencies)
	}
	if err != nil {
		if *mode == "chat" {
			fmt.Fprintln(stderr, "doctor failed: invalid_request")
		} else {
			fmt.Fprintf(stderr, "configuration invalid: %v\n", err)
		}
		return 2
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	if *mode == "chat" {
		checks := directchat.Doctor(ctx, config, directChatDependencies(dependencies))
		failed := false
		for _, check := range checks {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", check.Status, check.Name, check.Detail)
			failed = failed || check.Status == directchat.CheckFail
		}
		if ctx.Err() != nil {
			return 130
		}
		if failed {
			return 1
		}
		return 0
	}
	checks := application.Doctor(ctx, config, dependencies.application)
	failed := false
	for _, check := range checks {
		fmt.Fprintf(stdout, "%s\t%s\t%s\n", check.Status, check.Name, check.Detail)
		failed = failed || check.Status == application.CheckFail
	}
	if ctx.Err() != nil {
		return 130
	}
	if failed {
		return 1
	}
	return 0
}

func runModels(arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	config, help, code := loadConfigForCommand("maestro models", arguments, stdout, stderr, dependencies, false)
	if code != 0 || help {
		return code
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	configured, err := application.Build(config, dependencies.application)
	if err != nil {
		fmt.Fprintf(stderr, "models configuration failed: %v\n", err)
		return 2
	}
	defer closeApplication(configured)
	models, err := configured.Models(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "models unavailable: %v\n", err)
		if ctx.Err() != nil {
			return 130
		}
		return 4
	}
	fmt.Fprintf(stdout, "provider\t%s\n", config.Provider.ID)
	for _, model := range models {
		fmt.Fprintln(stdout, model.ID)
	}
	return 0
}

func runAgents(arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	config, help, code := loadConfigForCommand("maestro agents", arguments, stdout, stderr, dependencies, false)
	if code != 0 || help {
		return code
	}
	found := false
	for _, descriptor := range application.AgentDescriptors() {
		capabilities := descriptor.Capabilities()
		values := make([]string, len(capabilities))
		for index, capability := range capabilities {
			values[index] = string(capability)
		}
		fmt.Fprintf(stdout, "%s\t%s\t%s\n", descriptor.ID(), descriptor.Version(), strings.Join(values, ","))
		found = found || string(descriptor.ID()) == config.Agent.ID
	}
	if !found {
		fmt.Fprintf(stderr, "configured agent %q is not registered\n", config.Agent.ID)
		return 1
	}
	return 0
}

func runAgent(name string, arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	config, positionals, help, code := loadRunConfig(name, arguments, stdout, stderr, dependencies)
	if code != 0 || help {
		return code
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	interactive := dependencies.isTerminal != nil && dependencies.isTerminal(stdin)
	input := bufio.NewReader(stdin)
	instruction := strings.TrimSpace(strings.Join(positionals, " "))
	if instruction == "" {
		encoded, err := readInstruction(ctx, input, interactive, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "read instruction: %v\n", err)
			if ctx.Err() != nil {
				return 130
			}
			return 2
		}
		instruction = strings.TrimSpace(encoded)
	}
	if instruction == "" {
		fmt.Fprintf(stderr, "%s requires an instruction argument or stdin\n", name)
		return 2
	}
	configured, err := application.Build(config, dependencies.application)
	if err != nil {
		fmt.Fprintf(stderr, "run configuration failed: %v\n", err)
		return 2
	}
	defer closeApplication(configured)
	renderer := application.NewProgressRenderer(stderr)
	if err := renderer.Subscribe(configured.Runtime().EventBus()); err != nil {
		fmt.Fprintln(stderr, "run setup failed: progress_unavailable")
		return 1
	}
	renderer.RenderLimits(config.AgentLimits())
	approver := application.NewTerminalApprover(input, stderr, interactive)
	result, err := configured.ExecuteWithOptions(ctx, instruction, application.ExecuteOptions{Approver: approver})
	if err != nil {
		fmt.Fprintf(stderr, "run failed: %s\n", runFailureCode(ctx, err))
		return exitCodeForRunError(ctx, err)
	}
	session := result.Session()
	counters := session.Counters()
	fmt.Fprintf(stdout, "run\t%s\n", session.Run())
	fmt.Fprintf(stdout, "terminal\t%s\n", session.Terminal())
	fmt.Fprintf(stdout, "model_turns\t%d\n", counters.ModelTurns)
	fmt.Fprintf(stdout, "tool_calls\t%d\n", counters.ToolCalls)
	fmt.Fprintf(stdout, "input_tokens\t%d\n", counters.InputTokens)
	fmt.Fprintf(stdout, "output_tokens\t%d\n", counters.OutputTokens)
	fmt.Fprintln(stdout, "result")
	fmt.Fprintln(stdout, result.Content())
	return 0
}

func readInstruction(ctx context.Context, input *bufio.Reader, interactive bool, stderr io.Writer) (string, error) {
	type result struct {
		line string
		err  error
	}
	if interactive {
		fmt.Fprint(stderr, "instruction: ")
	}
	completed := make(chan result, 1)
	go func() {
		if interactive {
			line, err := readBoundedLine(input, maxInstructionBytes)
			completed <- result{line: line, err: err}
			return
		}
		encoded, err := io.ReadAll(io.LimitReader(input, maxInstructionBytes+1))
		if err == nil && len(encoded) > maxInstructionBytes {
			err = errors.New("instruction exceeds 1048576 bytes")
		}
		completed <- result{line: string(encoded), err: err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case value := <-completed:
		return value.line, value.err
	}
}

func readBoundedLine(input *bufio.Reader, maximum int) (string, error) {
	var value strings.Builder
	for {
		fragment, err := input.ReadSlice('\n')
		if value.Len()+len(fragment) > maximum {
			return "", errors.New("instruction exceeds 1048576 bytes")
		}
		value.Write(fragment)
		switch {
		case err == nil:
			return value.String(), nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		default:
			return "", err
		}
	}
}

func runVersion(arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	flags := flag.NewFlagSet("maestro version", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintln(stdout, "usage: maestro version") }
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "maestro version does not accept positional arguments")
		return 2
	}
	current := buildinfo.Current()
	if dependencies.buildInfo != nil {
		current = dependencies.buildInfo()
	}
	fmt.Fprintf(stdout, "maestro %s\n", current.Version)
	fmt.Fprintf(stdout, "commit %s\n", current.Commit)
	if current.Dirty {
		fmt.Fprintln(stdout, "dirty true")
	}
	return 0
}

func loadConfigForCommand(name string, arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies, allowPositionals bool) (productconfig.Config, bool, int) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintf(stdout, "usage: %s [--config path]\n", name) }
	configPath := flags.String("config", "", "path to Maestro configuration")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return productconfig.Config{}, true, 0
		}
		return productconfig.Config{}, false, 2
	}
	if !allowPositionals && flags.NArg() != 0 {
		fmt.Fprintf(stderr, "%s does not accept positional arguments\n", name)
		return productconfig.Config{}, false, 2
	}
	config, err := resolveAndLoad(*configPath, dependencies)
	if err != nil {
		fmt.Fprintf(stderr, "configuration invalid: %v\n", err)
		return productconfig.Config{}, false, 2
	}
	return config, false, 0
}

func loadRunConfig(name string, arguments []string, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) (productconfig.Config, []string, bool, int) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprintf(stdout, "usage: %s [--config path] [instruction]\n", name) }
	configPath := flags.String("config", "", "path to Maestro configuration")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return productconfig.Config{}, nil, true, 0
		}
		return productconfig.Config{}, nil, false, 2
	}
	config, err := resolveAndLoad(*configPath, dependencies)
	if err != nil {
		fmt.Fprintf(stderr, "configuration invalid: %v\n", err)
		return productconfig.Config{}, nil, false, 2
	}
	return config, flags.Args(), false, 0
}

func resolveAndLoad(explicit string, dependencies commandDependencies) (productconfig.Config, error) {
	getenv := dependencies.application.Getenv
	path, err := productconfig.ResolvePath(explicit, getenv)
	if err != nil {
		return productconfig.Config{}, err
	}
	return productconfig.Load(path)
}

func resolveAndLoadChat(explicit string, dependencies commandDependencies) (productconfig.Config, error) {
	getenv := dependencies.application.Getenv
	path, err := productconfig.ResolvePath(explicit, getenv)
	if err != nil {
		return productconfig.Config{}, err
	}
	return productconfig.LoadChat(path)
}

func commandContext(dependencies commandDependencies) (context.Context, context.CancelFunc) {
	if dependencies.context != nil {
		return dependencies.context()
	}
	return context.WithCancel(context.Background())
}

func exitCodeForRunError(ctx context.Context, err error) int {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, pkgAgent.ErrRunCanceled) {
		return 130
	}
	if errors.Is(err, pkgAgent.ErrPermissionDenied) {
		return 3
	}
	if errors.Is(err, pkgAgent.ErrProviderFailed) || errors.Is(err, pkgAgent.ErrNotFound) {
		return 4
	}
	if errors.Is(err, pkgAgent.ErrInvalidRequest) {
		return 2
	}
	return 1
}

func runFailureCode(ctx context.Context, err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	switch exitCodeForRunError(ctx, err) {
	case 130:
		return "canceled"
	case 3:
		return "permission_denied"
	case 4:
		return "provider_unavailable"
	case 2:
		return "invalid_request"
	default:
		return "execution_failed"
	}
}

func closeApplication(configured *application.Application) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = configured.Close(ctx)
}
