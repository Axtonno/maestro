package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antonio-cafeo/maestro/internal/buildinfo"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	"gopkg.in/yaml.v3"
)

type setupProvider interface {
	pkgProvider.Provider
	pkgProvider.ModelDiscoverer
	pkgProvider.ModelPuller
}

var errSetupWorkspaceMismatch = errors.New("existing configuration targets a different workspace")

type setupDocument struct {
	Version   int                    `yaml:"version"`
	Provider  setupProviderDocument  `yaml:"provider"`
	Workspace setupWorkspaceDocument `yaml:"workspace"`
	Chat      setupChatDocument      `yaml:"direct_chat"`
	Mutation  setupMutationDocument  `yaml:"controlled_mutation"`
}

type setupProviderDocument struct {
	ID        string `yaml:"id"`
	BaseURL   string `yaml:"base_url"`
	Timeout   string `yaml:"timeout"`
	APIKeyEnv string `yaml:"api_key_env"`
}

type setupWorkspaceDocument struct {
	ID        string `yaml:"id"`
	Root      string `yaml:"root"`
	Framework string `yaml:"framework"`
}

type setupChatDocument struct {
	Model          string `yaml:"model"`
	Digest         string `yaml:"digest"`
	Timeout        string `yaml:"timeout"`
	Streaming      bool   `yaml:"streaming"`
	NumCtx         int    `yaml:"num_ctx"`
	NumPredict     int    `yaml:"num_predict"`
	Thinking       string `yaml:"thinking"`
	Residency      string `yaml:"residency"`
	MaxFileBytes   int    `yaml:"max_file_bytes"`
	MaxOutputBytes int    `yaml:"max_output_bytes"`
}

type setupMutationDocument struct {
	Enabled        bool   `yaml:"enabled"`
	Model          string `yaml:"model"`
	Digest         string `yaml:"digest"`
	Timeout        string `yaml:"timeout"`
	NumCtx         int    `yaml:"num_ctx"`
	NumPredict     int    `yaml:"num_predict"`
	Thinking       string `yaml:"thinking"`
	Residency      string `yaml:"residency"`
	Prompt         string `yaml:"prompt"`
	PromptSHA256   string `yaml:"prompt_sha256"`
	Schema         string `yaml:"schema"`
	SchemaSHA256   string `yaml:"schema_sha256"`
	MaxOutputBytes int    `yaml:"max_output_bytes"`
}

func runSetup(arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, dependencies commandDependencies) int {
	flags := flag.NewFlagSet("maestro setup", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	usage := func() { fmt.Fprintln(stdout, "usage: maestro setup [--config path] [--workspace path] [--pull]") }
	configFlag := flags.String("config", "", "configuration path")
	workspaceFlag := flags.String("workspace", ".", "workspace root")
	pull := flags.Bool("pull", false, "download missing recommended models without prompting")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			usage()
			return 0
		}
		fmt.Fprintln(stderr, "setup failed: invalid_request")
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "setup failed: invalid_request")
		return 2
	}

	getenv := dependencies.application.Getenv
	configPath, err := productconfig.ResolvePath(*configFlag, getenv)
	if err != nil {
		fmt.Fprintln(stderr, "setup failed: configuration_path_unavailable")
		return 2
	}
	workspace, err := filepath.Abs(*workspaceFlag)
	if err != nil {
		fmt.Fprintln(stderr, "setup failed: workspace_unavailable")
		return 2
	}
	workspace = filepath.Clean(workspace)
	info, err := os.Lstat(workspace)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		fmt.Fprintln(stderr, "setup failed: workspace_unavailable")
		return 2
	}

	fmt.Fprintln(stdout, "Maestro setup")
	fmt.Fprintln(stdout)
	current := buildinfo.Current()
	if dependencies.buildInfo != nil {
		current = dependencies.buildInfo()
	}
	fmt.Fprintf(stdout, "✓ Maestro %s\n", current.Version)

	config, created, err := ensureSetupConfig(configPath, workspace)
	if err != nil {
		if errors.Is(err, errSetupWorkspaceMismatch) {
			fmt.Fprintln(stderr, "setup failed: workspace_mismatch")
			fmt.Fprintln(stderr, "use --config to create a separate project configuration")
			return 2
		}
		fmt.Fprintln(stderr, "setup failed: configuration_invalid")
		return 2
	}
	if created {
		fmt.Fprintf(stdout, "✓ Configuration created: %s\n", configPath)
	} else {
		fmt.Fprintf(stdout, "✓ Configuration already valid: %s\n", configPath)
	}

	factory := dependencies.application.ProviderFactory
	if factory == nil {
		fmt.Fprintln(stderr, "setup failed: provider_unavailable")
		return 4
	}
	secret, err := config.Secret(getenv)
	if err != nil {
		fmt.Fprintln(stderr, "setup failed: secret_environment_missing")
		return 2
	}
	candidate, err := factory(config, secret)
	provider, ok := candidate.(setupProvider)
	if err != nil || !ok || provider == nil || provider.ID() != "ollama" {
		fmt.Fprintln(stderr, "setup failed: provider_unavailable")
		return 4
	}
	ctx, cancel := commandContext(dependencies)
	defer cancel()
	models, err := provider.DiscoverModels(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return 130
		}
		fmt.Fprintln(stderr, "setup failed: ollama_unreachable")
		return 4
	}
	fmt.Fprintln(stdout, "✓ Ollama detected and responding")

	required := []struct{ model, digest, use string }{
		{productconfig.QualifiedDirectChatModel, productconfig.QualifiedDirectChatDigest, "chat"},
		{productconfig.QualifiedMutationModel, productconfig.QualifiedMutationDigest, "controlled mutation"},
	}
	missing := make([]string, 0, len(required))
	for _, want := range required {
		state := setupModelState(models, want.model, want.digest)
		switch state {
		case "available":
			fmt.Fprintf(stdout, "✓ %s available for %s\n", want.model, want.use)
		case "digest_mismatch":
			fmt.Fprintf(stderr, "setup failed: model_digest_mismatch model=%s\n", want.model)
			return 4
		default:
			missing = append(missing, want.model)
		}
	}

	shouldPull := *pull
	if len(missing) > 0 && !shouldPull && dependencies.isTerminal != nil && dependencies.isTerminal(stdin) {
		fmt.Fprintf(stderr, "Download the missing recommended models (%s)? [Y/n]: ", strings.Join(missing, ", "))
		line, readErr := bufio.NewReaderSize(stdin, 64).ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		shouldPull = readErr == nil && (answer == "" || answer == "y" || answer == "yes")
	}
	if len(missing) > 0 && !shouldPull {
		for _, model := range missing {
			fmt.Fprintf(stdout, "! %s is missing\n", model)
		}
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Run:")
		fmt.Fprintln(stdout, "  maestro setup --pull")
		return 4
	}
	for _, model := range missing {
		fmt.Fprintf(stdout, "→ Downloading %s\n", model)
		if err := pullSetupModel(ctx, provider, model, stderr); err != nil {
			if ctx.Err() != nil {
				return 130
			}
			fmt.Fprintf(stderr, "setup failed: model_pull_failed model=%s\n", model)
			return 4
		}
	}
	if len(missing) > 0 {
		models, err = provider.DiscoverModels(ctx)
		if err != nil {
			fmt.Fprintln(stderr, "setup failed: model_verification_failed")
			return 4
		}
		for _, want := range required {
			if setupModelState(models, want.model, want.digest) != "available" {
				fmt.Fprintf(stderr, "setup failed: model_verification_failed model=%s\n", want.model)
				return 4
			}
			fmt.Fprintf(stdout, "✓ %s available for %s\n", want.model, want.use)
		}
	}
	fmt.Fprintln(stdout, "✓ Workspace configured")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Try:")
	fmt.Fprintln(stdout, "  maestro chat \"Come puoi aiutarmi?\"")
	return 0
}

func ensureSetupConfig(path, workspace string) (productconfig.Config, bool, error) {
	if _, err := os.Stat(path); err == nil {
		config, loadErr := productconfig.LoadMutation(path)
		if loadErr == nil && config.Workspace.Root != workspace {
			return productconfig.Config{}, false, errSetupWorkspaceMismatch
		}
		return config, false, loadErr
	} else if !os.IsNotExist(err) {
		return productconfig.Config{}, false, err
	}
	document := setupDocument{
		Version:   4,
		Provider:  setupProviderDocument{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: "5m", APIKeyEnv: ""},
		Workspace: setupWorkspaceDocument{ID: "laravel", Root: workspace, Framework: "laravel"},
		Chat:      setupChatDocument{Model: productconfig.QualifiedDirectChatModel, Digest: productconfig.QualifiedDirectChatDigest, Timeout: "5m", Streaming: false, NumCtx: 4096, NumPredict: 1024, Thinking: "false", Residency: "5m", MaxFileBytes: 1 << 20, MaxOutputBytes: 1 << 20},
		Mutation:  setupMutationDocument{Enabled: true, Model: productconfig.QualifiedMutationModel, Digest: productconfig.QualifiedMutationDigest, Timeout: "5m", NumCtx: 4096, NumPredict: 1024, Thinking: "false", Residency: "5m", Prompt: productconfig.MutationPromptID, PromptSHA256: productconfig.MutationPromptSHA256, Schema: productconfig.MutationSchemaID, SchemaSHA256: productconfig.MutationSchemaSHA256, MaxOutputBytes: 1 << 20},
	}
	encoded, err := yaml.Marshal(document)
	if err != nil {
		return productconfig.Config{}, false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return productconfig.Config{}, false, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return productconfig.Config{}, false, err
	}
	written := false
	defer func() {
		_ = file.Close()
		if !written {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(encoded); err != nil {
		return productconfig.Config{}, false, err
	}
	if err := file.Sync(); err != nil {
		return productconfig.Config{}, false, err
	}
	if err := file.Close(); err != nil {
		return productconfig.Config{}, false, err
	}
	written = true
	config, err := productconfig.LoadMutation(path)
	if err != nil {
		written = false
		return productconfig.Config{}, false, err
	}
	return config, true, nil
}

func setupModelState(models []pkgProvider.ModelInfo, model, digest string) string {
	for _, candidate := range models {
		if candidate.Model.ID != model {
			continue
		}
		if candidate.Digest != digest {
			return "digest_mismatch"
		}
		return "available"
	}
	return "missing"
}

func pullSetupModel(ctx context.Context, provider setupProvider, model string, stderr io.Writer) error {
	stream, err := provider.PullModel(ctx, pkgProvider.ModelPullRequest{Model: model})
	if err != nil {
		return err
	}
	defer stream.Close()
	last := pkgProvider.ModelPullStage("")
	for {
		progress, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if progress.Stage != last {
			fmt.Fprintf(stderr, "model\t%s\tstage=%s\n", model, progress.Stage)
			last = progress.Stage
		}
		if progress.Stage == pkgProvider.ModelPullStageCompleted {
			return nil
		}
	}
}
