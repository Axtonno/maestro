package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/antonio-cafeo/maestro/internal/directchat"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func runChat(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	if duplicateChatFlag(arguments) {
		fmt.Fprintln(stderr, "chat failed: invalid_request")
		return 2
	}
	flags := flag.NewFlagSet("maestro chat", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stdout, "usage: maestro chat [--config path] [--file logical-path] [--stream] [question]")
	}
	configPath := flags.String("config", "", "path to Maestro configuration")
	logical := flags.String("file", "", "single logical workspace file")
	stream := flags.Bool("stream", false, "stream the direct response")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if chatFlagPresent(arguments, "--file") && *logical == "" {
		fmt.Fprintln(stderr, "chat failed: invalid_request")
		return 2
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	interactive := dependencies.isTerminal != nil && dependencies.isTerminal(stdin)
	question := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if question != "" && !interactive {
		concurrent, err := readInstruction(ctx, bufio.NewReader(stdin), false, stderr)
		if err != nil || strings.TrimSpace(concurrent) != "" {
			if ctx.Err() != nil {
				fmt.Fprintf(stderr, "chat failed: %s\n", chatFailureCode(ctx, ctx.Err()))
				return exitCodeForChatError(ctx, ctx.Err())
			}
			fmt.Fprintln(stderr, "chat failed: invalid_request")
			return 2
		}
	}
	if question == "" {
		encoded, err := readInstruction(ctx, bufio.NewReader(stdin), interactive, stderr)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Fprintf(stderr, "chat failed: %s\n", chatFailureCode(ctx, ctx.Err()))
				return exitCodeForChatError(ctx, ctx.Err())
			}
			fmt.Fprintln(stderr, "chat failed: invalid_request")
			return 2
		}
		question = strings.TrimSpace(encoded)
	}
	if question == "" {
		fmt.Fprintln(stderr, "chat failed: invalid_request")
		return 2
	}
	config, err := resolveAndLoad(*configPath, dependencies)
	if err != nil {
		fmt.Fprintln(stderr, "chat failed: invalid_request")
		return 2
	}
	service, err := directchat.Build(config, directChatDependencies(dependencies))
	if err != nil {
		fmt.Fprintf(stderr, "chat failed: %s\n", chatFailureCode(ctx, err))
		return exitCodeForChatError(ctx, err)
	}
	result, err := service.Execute(ctx, directchat.Request{Question: question, File: *logical, Stream: *stream})
	if err != nil {
		fmt.Fprintf(stderr, "chat failed: %s\n", chatFailureCode(ctx, err))
		return exitCodeForChatError(ctx, err)
	}
	fmt.Fprintln(stdout, "mode\tchat")
	fmt.Fprintln(stdout, "terminal\tcompleted")
	fmt.Fprintf(stdout, "model\t%s\n", result.Model)
	fmt.Fprintf(stdout, "duration_ms\t%d\n", result.Duration.Milliseconds())
	fmt.Fprintf(stdout, "input_tokens\t%d\n", result.Usage.InputTokens)
	fmt.Fprintf(stdout, "output_tokens\t%d\n", result.Usage.OutputTokens)
	fmt.Fprintf(stdout, "num_ctx_requested\t%d\n", result.RequestedNumCtx)
	fmt.Fprintln(stdout, "num_ctx_effective\tunknown")
	fmt.Fprintf(stdout, "thinking_requested\t%s\n", result.RequestedThinking)
	fmt.Fprintln(stdout, "thinking_effective\tunknown")
	fmt.Fprintln(stdout, "truncated\tunknown")
	fmt.Fprintf(stdout, "finish_reason\t%s\n", result.FinishReason)
	fmt.Fprintln(stdout, "result")
	fmt.Fprintln(stdout, result.Content)
	return 0
}

func directChatDependencies(dependencies commandDependencies) directchat.Dependencies {
	configured := dependencies.application
	result := directchat.Dependencies{Getenv: configured.Getenv}
	if configured.ProviderFactory != nil {
		result.ProviderFactory = func(config productconfig.Config, secret string) (pkgProvider.Provider, error) {
			return configured.ProviderFactory(config, secret)
		}
	}
	return result
}

func duplicateChatFlag(arguments []string) bool {
	seen := make(map[string]struct{}, 3)
	for _, argument := range arguments {
		for _, name := range []string{"--config", "--file", "--stream"} {
			if argument == name || strings.HasPrefix(argument, name+"=") {
				if _, exists := seen[name]; exists {
					return true
				}
				seen[name] = struct{}{}
			}
		}
	}
	return false
}

func chatFlagPresent(arguments []string, name string) bool {
	for _, argument := range arguments {
		if argument == name || strings.HasPrefix(argument, name+"=") {
			return true
		}
	}
	return false
}

func exitCodeForChatError(ctx context.Context, err error) int {
	switch {
	case errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled):
		return 130
	case errors.Is(err, context.DeadlineExceeded), errors.Is(ctx.Err(), context.DeadlineExceeded):
		return 4
	case errors.Is(err, directchat.ErrInvalidRequest),
		errors.Is(err, directchat.ErrProfileRequired),
		errors.Is(err, directchat.ErrFileNotAllowed),
		errors.Is(err, productconfig.ErrInvalid),
		errors.Is(err, productconfig.ErrSecretMissing):
		return 2
	case errors.Is(err, directchat.ErrProviderUnavailable),
		errors.Is(err, directchat.ErrCapabilityUnsupported):
		return 4
	default:
		return 1
	}
}

func chatFailureCode(ctx context.Context, err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(ctx.Err(), context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, context.Canceled), errors.Is(ctx.Err(), context.Canceled):
		return "canceled"
	case errors.Is(err, directchat.ErrProfileRequired):
		return "chat_profile_required"
	case errors.Is(err, directchat.ErrFileNotAllowed):
		return "file_not_allowed"
	case errors.Is(err, directchat.ErrProviderUnavailable):
		return "provider_unavailable"
	case errors.Is(err, directchat.ErrCapabilityUnsupported):
		return "capability_unsupported"
	case errors.Is(err, directchat.ErrResponseInvalid):
		return "response_invalid"
	case errors.Is(err, directchat.ErrLimitExceeded):
		return "limit_exceeded"
	case errors.Is(err, directchat.ErrInvalidRequest), errors.Is(err, productconfig.ErrInvalid),
		errors.Is(err, productconfig.ErrSecretMissing):
		return "invalid_request"
	default:
		return "execution_failed"
	}
}
