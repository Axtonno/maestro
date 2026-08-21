package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/buildinfo"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestBenchValidateLoadsVersionedManifest(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"bench", "validate", "--manifest",
			"../../docs/provider-smoke-benchmark-manifest.yaml",
		},
		&stdout,
		&stderr,
	)

	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "providers=2 scenarios=14") {
		t.Fatalf("unexpected validation output: %q", got)
	}
}

func TestBenchSmokeWithoutProviderConfigurationProducesSkippedReport(t *testing.T) {
	t.Setenv("MAESTRO_OLLAMA_BASE_URL", "")
	t.Setenv("MAESTRO_ALLOW_CATALOG_MUTATION", "false")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"bench", "smoke", "--manifest",
			"../../docs/provider-smoke-benchmark-manifest.yaml",
			"--fail-on-failure",
		},
		&stdout,
		&stderr,
	)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	var report pkgBenchmark.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode smoke report: %v\n%s", err, stdout.String())
	}
	if report.SchemaVersion != pkgBenchmark.ReportSchemaVersion ||
		len(report.Scenarios) != 14 {
		t.Fatalf("unexpected report: %#v", report)
	}
	for _, scenario := range report.Scenarios {
		if scenario.State != pkgBenchmark.ResultSkipped ||
			scenario.Samples[0].ReasonCode != "provider_not_configured" {
			t.Fatalf("unexpected scenario: %#v", scenario)
		}
	}
}

func TestBenchRuntimeCommandsWithoutProviderProduceSeparatedSkippedReports(t *testing.T) {
	t.Setenv("MAESTRO_OLLAMA_BASE_URL", "")
	for _, testCase := range []struct {
		command string
		count   int
	}{
		{command: "provider", count: 4},
		{command: "model", count: 7},
	} {
		t.Run(testCase.command, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := run(
				[]string{
					"bench", testCase.command, "--manifest",
					"../../docs/runtime-benchmark-manifest.yaml",
					"--warmup", "0", "--runs", "1", "--fail-on-failure",
				},
				&stdout,
				&stderr,
			)
			if exitCode != 0 || stderr.Len() != 0 {
				t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
			}
			var report pkgBenchmark.Report
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatalf("decode report: %v\n%s", err, stdout.String())
			}
			if len(report.Scenarios) != testCase.count ||
				report.Metadata.Command != "bench "+testCase.command {
				t.Fatalf("unexpected report: %#v", report)
			}
			for _, scenario := range report.Scenarios {
				if scenario.State != pkgBenchmark.ResultSkipped ||
					scenario.Samples[0].ReasonCode != "provider_not_configured" {
					t.Fatalf("unexpected scenario: %#v", scenario)
				}
			}
		})
	}
}

func TestBenchLaravelWithoutProviderLoadsDatasetAndProducesSkippedReport(t *testing.T) {
	t.Setenv("MAESTRO_OLLAMA_BASE_URL", "")
	t.Setenv("MAESTRO_BENCHMARK_GPU", "Fixture GPU")
	t.Setenv("MAESTRO_BENCHMARK_BACKEND", "fixture-backend")
	t.Setenv("MAESTRO_BENCHMARK_VRAM_MB", "4096")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"bench", "laravel", "--manifest",
			"../../docs/developer-benchmark-manifest.yaml",
			"--warmup", "0", "--runs", "1", "--fail-on-failure",
		},
		&stdout,
		&stderr,
	)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	var report pkgBenchmark.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, stdout.String())
	}
	if report.Metadata.Command != "bench laravel" || len(report.Scenarios) != 6 ||
		report.Configuration.Dataset.ID != "maestro-laravel-mini" ||
		report.Configuration.Dataset.Version != "1.0.0" ||
		len(report.Configuration.Plugins) != 1 ||
		report.Configuration.Plugins[0].ID != "laravel" ||
		report.Configuration.Hardware.GPU != "Fixture GPU" ||
		report.Configuration.Hardware.Backend != "fixture-backend" ||
		report.Configuration.Hardware.VRAMMB != 4096 {
		t.Fatalf("unexpected Laravel report: %#v", report)
	}
	for _, scenario := range report.Scenarios {
		if scenario.State != pkgBenchmark.ResultSkipped ||
			scenario.Samples[0].ReasonCode != "provider_not_configured" ||
			scenario.Samples[0].Evaluation != nil {
			t.Fatalf("unexpected scenario: %#v", scenario)
		}
	}
}

func TestBenchMutationValidatesFrozenProfileAndFixture(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"bench", "mutation", "--profile", "../../docs/mutation-qualification-profile.yaml"},
		&stdout,
		&stderr,
	)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	if output := stdout.String(); !strings.Contains(output, "version=1 gates=3 scenarios=15") ||
		!strings.Contains(output, "linux_amd64/ollama\tibm/granite4.1:8b") {
		t.Fatalf("unexpected mutation validation output: %q", output)
	}
}

func TestBenchMutationGateCRejectsNonInteractiveInputBeforeProviderIO(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dependencies := defaultCommandDependencies()
	dependencies.isTerminal = func(io.Reader) bool { return false }
	exitCode := runWithIO(
		[]string{
			"bench", "mutation", "--profile", "../../docs/mutation-qualification-profile.yaml",
			"--mode", "gate-c",
		},
		strings.NewReader("o\n"), &stdout, &stderr, dependencies,
	)
	if exitCode != 3 || !strings.Contains(stderr.String(), "requires an interactive TTY") || stdout.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
}

func TestBenchLaravelWritesMarkdownAndRenderReproducesItFromJSON(t *testing.T) {
	t.Setenv("MAESTRO_OLLAMA_BASE_URL", "")
	directory := t.TempDir()
	jsonPath := filepath.Join(directory, "report.json")
	markdownPath := filepath.Join(directory, "report.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"bench", "laravel", "--manifest", "../../docs/developer-benchmark-manifest.yaml",
			"--warmup", "0", "--runs", "1", "--output", jsonPath,
			"--markdown", markdownPath,
		},
		&stdout,
		&stderr,
	)
	if exitCode != 0 || stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	writtenMarkdown, err := os.ReadFile(markdownPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(writtenMarkdown, []byte("# Maestro Benchmark Report")) ||
		!bytes.Contains(writtenMarkdown, []byte("maestro-laravel-mini@1.0.0")) {
		t.Fatalf("unexpected Markdown report:\n%s", writtenMarkdown)
	}
	for _, name := range []string{jsonPath, markdownPath} {
		info, err := os.Stat(name)
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatalf("report %q mode/error: %v %v", name, info, err)
		}
	}
	stdout.Reset()
	stderr.Reset()
	exitCode = run(
		[]string{"bench", "render", "--input", jsonPath},
		&stdout,
		&stderr,
	)
	if exitCode != 0 || stderr.Len() != 0 || !bytes.Equal(stdout.Bytes(), writtenMarkdown) {
		t.Fatalf("render exit=%d stderr=%q\nwant:\n%s\ngot:\n%s", exitCode, stderr.String(), writtenMarkdown, stdout.String())
	}
}

func TestBenchRenderRejectsInputOutputCollision(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{"bench", "render", "--input", "report.json", "--output", "./report.json"},
		&stdout,
		&stderr,
	)
	if exitCode != 2 || !strings.Contains(stderr.String(), "must be different") {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
}

func TestDeveloperMinimumScoreGateIgnoresWarmupsAndRequiresEvaluations(t *testing.T) {
	report := pkgBenchmark.Report{Scenarios: []pkgBenchmark.ScenarioReport{{
		Samples: []pkgBenchmark.Sample{
			{Iteration: pkgBenchmark.Iteration{Warmup: true}, Evaluation: &pkgBenchmark.QualityEvaluation{Score: 0}},
			{Iteration: pkgBenchmark.Iteration{}, Evaluation: nil},
			{Iteration: pkgBenchmark.Iteration{}, Evaluation: &pkgBenchmark.QualityEvaluation{Score: 2}},
		},
	}}}
	if !reportBelowMinimumScore(report, 2) || !reportBelowMinimumScore(report, 3) {
		t.Fatal("unexpected minimum score gate result")
	}
}

func TestBenchHelpAndUnknownCommandHaveStableExitCodes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := run([]string{"bench", "--help"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("help exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "maestro bench <command>") {
		t.Fatalf("unexpected help: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := run([]string{"bench", "unknown"}, &stdout, &stderr); exitCode != 2 {
		t.Fatalf("unknown command exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "unknown bench command") {
		t.Fatalf("unexpected error: %q", stderr.String())
	}
}

func TestRootOutputIsWrittenOnlyToProvidedWriter(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := run(nil, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("root exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "usage: maestro <command>") ||
		!strings.Contains(stdout.String(), "doctor") || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestProductCommandsUseStrictConfigAndStableOutputs(t *testing.T) {
	configPath := newCLIConfig(t, "allow")
	tests := []struct {
		name     string
		args     []string
		stdin    string
		contains []string
		progress bool
	}{
		{name: "doctor", args: []string{"doctor", "--config", configPath}, contains: []string{"pass\tconfig", "pass\tprovider", "pass\tlaravel"}},
		{name: "models", args: []string{"models", "--config", configPath}, contains: []string{"provider\tollama", "fixture-model"}},
		{name: "agents", args: []string{"agents", "--config", configPath}, contains: []string{"agent.reference", "agent.run"}},
		{name: "run argument", args: []string{"run", "--config", configPath, "Inspect", "Order"}, contains: []string{"run\trun-cli", "terminal\tcompleted", "model_turns\t1", "result\nCLI completed."}, progress: true},
		{name: "run stdin", args: []string{"run", "--config", configPath}, stdin: "Inspect Order\n", contains: []string{"terminal\tcompleted", "result\nCLI completed."}, progress: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "CLI completed."}, FinishReason: pkgProvider.FinishReasonStop}}}
			dependencies := cliTestDependencies(provider)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := runWithIO(testCase.args, strings.NewReader(testCase.stdin), &stdout, &stderr, dependencies)
			if exitCode != 0 {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
			}
			if testCase.progress {
				if !strings.Contains(stderr.String(), "plan\trun=run-cli") || !strings.Contains(stderr.String(), "terminal\trun=run-cli reason=completed") {
					t.Fatalf("missing run progress: %q", stderr.String())
				}
			} else if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
			for _, expected := range testCase.contains {
				if !strings.Contains(stdout.String(), expected) {
					t.Fatalf("output %q does not contain %q", stdout.String(), expected)
				}
			}
		})
	}
}

func TestProductCommandHelpVersionAndErrorExitCodes(t *testing.T) {
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.buildInfo = func() buildinfo.Info { return buildinfo.Info{Version: "v0.1.0", Commit: "abc123"} }
	for _, testCase := range []struct {
		name     string
		args     []string
		wantCode int
		output   string
	}{
		{name: "root help", args: []string{"--help"}, wantCode: 0, output: "maestro <command>"},
		{name: "command help", args: []string{"doctor", "--help"}, wantCode: 0, output: "maestro doctor"},
		{name: "help command", args: []string{"help", "run"}, wantCode: 0, output: "maestro run"},
		{name: "version", args: []string{"version"}, wantCode: 0, output: "maestro v0.1.0\ncommit abc123"},
		{name: "unknown", args: []string{"unknown"}, wantCode: 2, output: "unknown command"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(testCase.args, strings.NewReader(""), &stdout, &stderr, dependencies)
			combined := stdout.String() + stderr.String()
			if code != testCase.wantCode || !strings.Contains(combined, testCase.output) {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}

	invalidPath := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte("version: 1\nunknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := runWithIO([]string{"doctor", "--config", invalidPath}, strings.NewReader(""), &stdout, &stderr, dependencies); code != 2 || !strings.Contains(stderr.String(), "configuration invalid") {
		t.Fatalf("invalid config code=%d stderr=%q", code, stderr.String())
	}
}

func TestRunMapsPermissionAndProviderFailures(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		policyMode string
		provider   *cliProvider
		wantCode   int
	}{
		{name: "prompt without approver", policyMode: "prompt", provider: &cliProvider{id: "ollama"}, wantCode: 3},
		{name: "provider failure", policyMode: "allow", provider: &cliProvider{id: "ollama"}, wantCode: 4},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(
				[]string{"run", "--config", newCLIConfig(t, testCase.policyMode), "Inspect Order"},
				strings.NewReader(""), &stdout, &stderr, cliTestDependencies(testCase.provider),
			)
			if code != testCase.wantCode || !strings.Contains(stderr.String(), "run failed") {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunInteractiveApprovalAndFailClosedInputs(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		input    string
		terminal bool
		wantCode int
		approved bool
	}{
		{name: "one shot", input: "once\n", terminal: true, wantCode: 0, approved: true},
		{name: "deny", input: "deny\n", terminal: true, wantCode: 3},
		{name: "invalid", input: "yes\n", terminal: true, wantCode: 3},
		{name: "eof", terminal: true, wantCode: 3},
		{name: "non tty", input: "once\n", wantCode: 3},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
				Message:      pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Safe result."},
				FinishReason: pkgProvider.FinishReasonStop,
			}}}
			dependencies := cliTestDependencies(provider)
			dependencies.isTerminal = func(io.Reader) bool { return testCase.terminal }
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(
				[]string{"run", "--config", newCLIConfig(t, "prompt"), "TOP-SECRET-INSTRUCTION"},
				strings.NewReader(testCase.input), &stdout, &stderr, dependencies,
			)
			if code != testCase.wantCode {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if testCase.terminal != strings.Contains(stderr.String(), "approval required") {
				t.Fatalf("unexpected approval rendering: %q", stderr.String())
			}
			if strings.Contains(stderr.String(), "TOP-SECRET-INSTRUCTION") {
				t.Fatalf("instruction leaked to stderr: %q", stderr.String())
			}
			if testCase.approved {
				if !strings.Contains(stdout.String(), "terminal\tcompleted") || !strings.Contains(stdout.String(), "result\nSafe result.") {
					t.Fatalf("unexpected success output: %q", stdout.String())
				}
			} else if stdout.Len() != 0 {
				t.Fatalf("failed run wrote stdout: %q", stdout.String())
			}
		})
	}
}

func TestRunMapsCanceledCommandContextToInterrupt(t *testing.T) {
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.context = func() (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx, func() {}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"run", "--config", newCLIConfig(t, "allow"), "Inspect Order"},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 130 || !strings.Contains(stderr.String(), "run failed") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunCancelsWhileWaitingForInstruction(t *testing.T) {
	for _, interactive := range []bool{true, false} {
		t.Run(fmt.Sprintf("interactive=%t", interactive), func(t *testing.T) {
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
			dependencies.isTerminal = func(io.Reader) bool { return interactive }
			dependencies.context = func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(
				[]string{"run", "--config", newCLIConfig(t, "allow")}, reader,
				&stdout, &stderr, dependencies,
			)
			if code != 130 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "context canceled") {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunBoundsInteractiveInstruction(t *testing.T) {
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.isTerminal = func(io.Reader) bool { return true }
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"run", "--config", newCLIConfig(t, "allow")},
		strings.NewReader(strings.Repeat("x", maxInstructionBytes+1)+"\n"),
		&stdout, &stderr, dependencies,
	)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "instruction exceeds") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

type cliProvider struct {
	id        pkgProvider.ID
	responses []pkgProvider.CompletionResponse
}

func (provider *cliProvider) ID() pkgProvider.ID { return provider.id }

func (provider *cliProvider) Complete(context.Context, pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	if len(provider.responses) == 0 {
		return pkgProvider.CompletionResponse{}, fmt.Errorf("fixture provider failure")
	}
	response := provider.responses[0]
	provider.responses = provider.responses[1:]
	return response, nil
}

func (provider *cliProvider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	return []pkgProvider.ModelInfo{{Model: pkgProvider.Model{ID: "fixture-model"}, State: pkgProvider.ModelStateLoaded}}, nil
}

func (provider *cliProvider) Models(context.Context) ([]pkgProvider.Model, error) {
	return []pkgProvider.Model{{ID: "fixture-model"}}, nil
}

func (provider *cliProvider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{Capability: capability, Support: pkgProvider.CapabilitySupported, Availability: pkgProvider.CapabilityAvailabilityAvailable})
	}
	return pkgProvider.CapabilityReport{Provider: provider.id, Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

func cliTestDependencies(provider pkgProvider.Provider) commandDependencies {
	return commandDependencies{
		application: application.Dependencies{
			Getenv:          func(string) string { return "" },
			ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil },
			RunID:           func() (pkgAgent.RunID, error) { return "run-cli", nil },
		},
		buildInfo: func() buildinfo.Info { return buildinfo.Info{Version: "devel", Commit: "fixture"} },
		context:   func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) },
	}
}

func newCLIConfig(t *testing.T, modelPolicy string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"artisan":       "#!/usr/bin/env php\n",
		"composer.json": `{"require":{"laravel/framework":"^12.0"}}`,
		"app/Order.php": "<?php\nclass Order {}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	content := fmt.Sprintf(`version: 1
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 1m
  api_key_env: ""
models:
  chat: fixture-model
  embedding: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
agent:
  id: agent.reference
  streaming: false
  tools:
    - workspace.patch
    - workspace.read
policy:
  id: policy.test
  model: %s
  workspace_inspect: allow
  workspace_mutate: prompt
limits:
  duration: 1m
  model_turns: 5
  tool_calls: 4
  tool_calls_per_turn: 2
  plan_steps: 3
  plan_revisions: 2
  tool_result_bytes: 65536
  session_bytes: 1048576
  input_tokens: 10000
  output_tokens: 10000
context:
  retrieval: lexical
  top_k: 5
  max_tokens: 1024
  reserved_tokens: 128
  safety_tokens: 64
`, root, modelPolicy)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
