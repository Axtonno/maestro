package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestMutationPreflightNormalizesCPUTrademarkMarkers(t *testing.T) {
	if !cpuMeetsLowerBound(
		"Intel(R) Core(TM) i5-8365U CPU @ 1.60GHz",
		"Intel Core i5-8365U",
	) {
		t.Fatal("equivalent CPU identity did not meet the lower bound")
	}
	if cpuMeetsLowerBound("Intel Core i3-1000", "Intel Core i5-8365U") {
		t.Fatal("different CPU identity met the lower bound")
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
		{name: "agent argument", args: []string{"agent", "--config", configPath, "Inspect", "Order"}, contains: []string{"run\trun-cli", "terminal\tcompleted", "model_turns\t1", "result\nCLI completed."}, progress: true},
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

func TestChatCommandUsesV2ProfileAndStableRedactedEnvelope(t *testing.T) {
	configPath, root := newCLIInteractionConfig(t, false)
	provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
		Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "GET /orders."},
		FinishReason: pkgProvider.FinishReasonStop, Usage: pkgProvider.Usage{InputTokens: 10, OutputTokens: 3},
	}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "--file", "routes/api.php", "Which routes exist?"},
		strings.NewReader(""), &stdout, &stderr, cliTestDependencies(provider),
	)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for _, expected := range []string{
		"mode\tchat", "terminal\tcompleted", "model\tchat-model",
		"input_tokens\t10", "output_tokens\t3", "num_ctx_requested\t4096",
		"num_ctx_effective\tunknown", "thinking_requested\tfalse",
		"thinking_effective\tunknown", "result\nGET /orders.",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("output %q does not contain %q", stdout.String(), expected)
		}
	}
	if strings.Contains(stdout.String(), root) || strings.Contains(stderr.String(), root) {
		t.Fatalf("physical root leaked: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if len(provider.requests) != 1 || len(provider.requests[0].Tools) != 0 ||
		provider.requests[0].ToolChoice.Mode != pkgProvider.ToolChoiceNone ||
		provider.requests[0].Options.MaxTokens != 0 || provider.requests[0].KeepAlive != 0 {
		t.Fatalf("chat request was not tool-free: %#v", provider.requests)
	}
}

func TestChatCommandUsesQualificationBudgetAndResidency(t *testing.T) {
	configPath := newCLIQualificationConfig(t, false)
	provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
		Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Qualified response."},
		FinishReason: pkgProvider.FinishReasonStop,
	}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "Question"}, strings.NewReader(""),
		&stdout, &stderr, cliTestDependencies(provider),
	)
	if code != 0 || stderr.Len() != 0 ||
		!strings.Contains(stdout.String(), "num_predict_requested\t512") ||
		!strings.Contains(stdout.String(), "residency_requested\t5m0s") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if len(provider.requests) != 1 || provider.requests[0].Options.MaxTokens != 512 ||
		provider.requests[0].KeepAlive != 5*time.Minute {
		t.Fatalf("qualification request drifted: %#v", provider.requests)
	}

	stdout.Reset()
	if code := runWithIO(
		[]string{"doctor", "--mode", "chat", "--config", configPath}, strings.NewReader(""),
		&stdout, &stderr, cliTestDependencies(provider),
	); code != 0 || !strings.Contains(stdout.String(), "pass\tconfig\tschema_v3_chat_valid") {
		t.Fatalf("doctor code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestChatCommandAndDoctorAcceptChatOnlyProfile(t *testing.T) {
	configPath, _ := newCLIChatOnlyConfig(t, false)
	provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
		Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "No project context supplied."},
		FinishReason: pkgProvider.FinishReasonStop,
	}}}
	dependencies := cliTestDependencies(provider)
	for _, testCase := range []struct {
		name string
		args []string
	}{
		{name: "chat", args: []string{"chat", "--config", configPath, "What is known?"}},
		{name: "doctor", args: []string{"doctor", "--mode", "chat", "--config", configPath}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := runWithIO(testCase.args, strings.NewReader(""), &stdout, &stderr, dependencies); code != 0 || stderr.Len() != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
	if provider.inspectCalls != 2 || len(provider.requests) != 1 || len(provider.streamRequests) != 0 {
		t.Fatalf("unexpected direct path: inspect=%d complete=%d stream=%d", provider.inspectCalls, len(provider.requests), len(provider.streamRequests))
	}
	if _, err := productconfig.Load(configPath); !errors.Is(err, productconfig.ErrInvalid) {
		t.Fatalf("agent loader accepted chat-only profile: %v", err)
	}
}

func TestChatCommandFailsClosedWithSyntheticReasons(t *testing.T) {
	v1 := newCLIConfig(t, "allow")
	v2, _ := newCLIInteractionConfig(t, false)
	for _, testCase := range []struct {
		name     string
		args     []string
		wantCode int
		want     string
	}{
		{name: "v1", args: []string{"chat", "--config", v1, "Question"}, wantCode: 2, want: "chat_profile_required"},
		{name: "absolute", args: []string{"chat", "--config", v2, "--file", "/secret", "Question"}, wantCode: 2, want: "file_not_allowed"},
		{name: "duplicate", args: []string{"chat", "--config", v2, "--file", "a", "--file", "b", "Question"}, wantCode: 2, want: "invalid_request"},
		{name: "empty file", args: []string{"chat", "--config", v2, "--file=", "Question"}, wantCode: 2, want: "invalid_request"},
		{name: "stream disabled", args: []string{"chat", "--config", v2, "--stream", "Question"}, wantCode: 4, want: "capability_unsupported"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(testCase.args, strings.NewReader(""), &stdout, &stderr, cliTestDependencies(&cliProvider{id: "ollama"}))
			if code != testCase.wantCode || stdout.Len() != 0 || !strings.Contains(stderr.String(), "chat failed: "+testCase.want) {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if strings.Contains(stderr.String(), "/secret") {
				t.Fatalf("rejected path leaked: %q", stderr.String())
			}
		})
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := runWithIO([]string{"chat", "--config", v2, "Question"}, strings.NewReader("Concurrent stdin"), &stdout, &stderr, cliTestDependencies(&cliProvider{id: "ollama"})); code != 2 || stdout.Len() != 0 || stderr.String() != "chat failed: invalid_request\n" {
		t.Fatalf("positional/stdin conflict: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestChatCommandUsageErrorsAreCanonicalBeforeComposition(t *testing.T) {
	factoryCalls := 0
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.application.ProviderFactory = func(productconfig.Config, string) (pkgProvider.Provider, error) {
		factoryCalls++
		return &cliProvider{id: "ollama"}, nil
	}
	for _, testCase := range []struct {
		name  string
		args  []string
		stdin string
	}{
		{name: "unknown flag", args: []string{"chat", "--unknown"}},
		{name: "invalid boolean", args: []string{"chat", "--stream=maybe"}},
		{name: "duplicate", args: []string{"chat", "--stream", "--stream", "Question"}},
		{name: "empty file", args: []string{"chat", "--file=", "Question"}},
		{name: "blank", args: []string{"chat"}},
		{name: "oversized", args: []string{"chat", strings.Repeat("q", maxInstructionBytes+1)}},
		{name: "concurrent stdin", args: []string{"chat", "Question"}, stdin: "also stdin"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(testCase.args, strings.NewReader(testCase.stdin), &stdout, &stderr, dependencies)
			if code != 2 || stdout.Len() != 0 || stderr.String() != "chat failed: invalid_request\n" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
	if factoryCalls != 0 {
		t.Fatalf("usage errors composed provider %d times", factoryCalls)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := runWithIO([]string{"chat", "--help"}, strings.NewReader(""), &stdout, &stderr, dependencies); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "usage: maestro chat") {
		t.Fatalf("help code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestChatConfigurationDiagnosticsExposeOnlyKindAndLogicalField(t *testing.T) {
	configPath, _ := newCLIChatOnlyConfig(t, false)
	encoded, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	encoded = bytes.Replace(encoded, []byte("    timeout: 1m"), []byte("    timeout: 1m\n    secret_option: VALUE-SENTINEL"), 1)
	if err := os.WriteFile(configPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	wantDetail := "configuration\tkind=unknown_field field=interaction.chat.secret_option\n"
	for _, testCase := range []struct {
		name string
		args []string
		want string
	}{
		{name: "chat", args: []string{"chat", "--config", configPath, "Question"}, want: "chat failed: invalid_request\n" + wantDetail},
		{name: "doctor", args: []string{"doctor", "--mode", "chat", "--config", configPath}, want: "doctor failed: invalid_request\n" + wantDetail},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(testCase.args, strings.NewReader(""), &stdout, &stderr, cliTestDependencies(&cliProvider{id: "ollama"}))
			if code != 2 || stdout.Len() != 0 || stderr.String() != testCase.want {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			for _, forbidden := range []string{"VALUE-SENTINEL", configPath, "secret-value"} {
				if strings.Contains(stderr.String(), forbidden) {
					t.Fatalf("configuration diagnostic leaked %q: %q", forbidden, stderr.String())
				}
			}
		})
	}
}

func TestChatCommandResponseFailuresUseStableTerminals(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		response pkgProvider.CompletionResponse
		limit    bool
		code     int
		reason   string
	}{
		{name: "empty", response: pkgProvider.CompletionResponse{Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant}, FinishReason: pkgProvider.FinishReasonStop}, code: 1, reason: "response_invalid"},
		{name: "model mismatch", response: pkgProvider.CompletionResponse{Model: "other-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "RESPONSE-SENTINEL"}, FinishReason: pkgProvider.FinishReasonStop}, code: 1, reason: "response_invalid"},
		{name: "output limit", response: pkgProvider.CompletionResponse{Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "RESPONSE-SENTINEL"}, FinishReason: pkgProvider.FinishReasonStop}, limit: true, code: 1, reason: "limit_exceeded"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			configPath, _ := newCLIChatOnlyConfig(t, false)
			if testCase.limit {
				encoded, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatal(err)
				}
				encoded = bytes.Replace(encoded, []byte("max_output_bytes: 1048576"), []byte("max_output_bytes: 4"), 1)
				if err := os.WriteFile(configPath, encoded, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO(
				[]string{"chat", "--config", configPath, "QUESTION-SENTINEL"},
				strings.NewReader(""), &stdout, &stderr,
				cliTestDependencies(&cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{testCase.response}}),
			)
			if code != testCase.code || stdout.Len() != 0 || stderr.String() != "chat failed: "+testCase.reason+"\n" {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if strings.Contains(stderr.String(), "QUESTION-SENTINEL") || strings.Contains(stderr.String(), "RESPONSE-SENTINEL") {
				t.Fatalf("failure leaked payload: %q", stderr.String())
			}
		})
	}
}

func TestChatCommandStreamingUsesAtomicEquivalentEnvelope(t *testing.T) {
	configPath, _ := newCLIInteractionConfig(t, true)
	provider := &cliProvider{id: "ollama", stream: &cliStream{results: []cliStreamResult{
		{chunk: pkgProvider.StreamChunk{Content: "GET /"}},
		{chunk: pkgProvider.StreamChunk{Content: "orders.", FinishReason: pkgProvider.FinishReasonStop, Usage: pkgProvider.Usage{InputTokens: 10, OutputTokens: 3}}},
		{err: io.EOF},
	}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "--stream", "Which routes exist?"},
		strings.NewReader(""), &stdout, &stderr, cliTestDependencies(provider),
	)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "result\nGET /orders.") ||
		len(provider.requests) != 0 || len(provider.streamRequests) != 1 {
		t.Fatalf("code=%d stdout=%q stderr=%q complete=%d stream=%d", code, stdout.String(), stderr.String(), len(provider.requests), len(provider.streamRequests))
	}
}

func TestChatCommandRendersBoundedRedactedHeartbeatAndStopsIt(t *testing.T) {
	configPath, _ := newCLIChatOnlyConfig(t, false)
	provider := &cliProvider{
		id:            "ollama",
		completeDelay: 25 * time.Millisecond,
		responses: []pkgProvider.CompletionResponse{{
			Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "RESPONSE-SENTINEL"},
			FinishReason: pkgProvider.FinishReasonStop,
		}},
	}
	dependencies := cliTestDependencies(provider)
	dependencies.chatHeartbeat = time.Millisecond
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "QUESTION-SENTINEL"},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 0 || !strings.Contains(stdout.String(), "result\nRESPONSE-SENTINEL") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
	if len(lines) == 0 || len(lines) > maxChatHeartbeats {
		t.Fatalf("unexpected heartbeat count %d: %q", len(lines), stderr.String())
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, "progress\tstate=generating elapsed_ms=") {
			t.Fatalf("unexpected heartbeat: %q", line)
		}
	}
	for _, forbidden := range []string{"QUESTION-SENTINEL", "RESPONSE-SENTINEL", configPath, "chat-model"} {
		if strings.Contains(stderr.String(), forbidden) {
			t.Fatalf("heartbeat leaked %q: %q", forbidden, stderr.String())
		}
	}
	stopped := stderr.String()
	time.Sleep(5 * time.Millisecond)
	if stderr.String() != stopped {
		t.Fatalf("heartbeat continued after terminal: before=%q after=%q", stopped, stderr.String())
	}
}

func TestChatCommandFailureDoesNotLeakInputsOrPartialResponse(t *testing.T) {
	configPath, root := newCLIInteractionConfig(t, true)
	provider := &cliProvider{id: "ollama", stream: &cliStream{results: []cliStreamResult{
		{chunk: pkgProvider.StreamChunk{Content: "RESPONSE-SENTINEL"}},
		{err: errors.New("provider broke")},
	}}}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "--file", "routes/api.php", "--stream", "QUESTION-SENTINEL"},
		strings.NewReader(""), &stdout, &stderr, cliTestDependencies(provider),
	)
	if code != 4 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "chat failed: provider_unavailable") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for _, forbidden := range []string{"QUESTION-SENTINEL", "RESPONSE-SENTINEL", "Route::get", "routes/api.php", root} {
		if strings.Contains(stdout.String(), forbidden) || strings.Contains(stderr.String(), forbidden) {
			t.Fatalf("operational output leaked %q: stdout=%q stderr=%q", forbidden, stdout.String(), stderr.String())
		}
	}
}

func TestChatCommandMapsCanceledContextWithoutProviderIO(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		ctx      func() context.Context
		code     int
		codeName string
	}{
		{name: "cancel", ctx: func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }, code: 130, codeName: "canceled"},
		{name: "deadline", ctx: func() context.Context {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			defer cancel()
			return ctx
		}, code: 4, codeName: "deadline_exceeded"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			configPath, _ := newCLIInteractionConfig(t, false)
			provider := &cliProvider{id: "ollama"}
			dependencies := cliTestDependencies(provider)
			dependencies.context = func() (context.Context, context.CancelFunc) { return testCase.ctx(), func() {} }
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := runWithIO([]string{"chat", "--config", configPath, "Question"}, strings.NewReader(""), &stdout, &stderr, dependencies)
			if code != testCase.code || stdout.Len() != 0 || stderr.String() != "chat failed: "+testCase.codeName+"\n" ||
				len(provider.requests) != 0 || len(provider.streamRequests) != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q complete=%d stream=%d", code, stdout.String(), stderr.String(), len(provider.requests), len(provider.streamRequests))
			}
		})
	}
}

func TestProductCommandHelpVersionAndErrorExitCodes(t *testing.T) {
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.buildInfo = func() buildinfo.Info { return buildinfo.Info{Version: "v0.1.0", Commit: "abc123", Status: "release"} }
	dependencies.binaryIdentity = func() (binaryIdentity, error) {
		return binaryIdentity{Executable: "/opt/maestro/bin/maestro", SHA256: strings.Repeat("a", 64)}, nil
	}
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
		{name: "version diagnostic", args: []string{"version", "--diagnostic"}, wantCode: 0, output: "mode\tbinary_identity\nversion\tv0.1.0\nstatus\trelease\ncommit\tabc123\ndirty\tfalse\nexecutable\t\"/opt/maestro/bin/maestro\"\nsha256\t" + strings.Repeat("a", 64)},
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

func TestVersionDiagnosticFailsClosedWithoutLeakingIdentityError(t *testing.T) {
	dependencies := cliTestDependencies(&cliProvider{id: "ollama"})
	dependencies.binaryIdentity = func() (binaryIdentity, error) {
		return binaryIdentity{}, errors.New("PATH-SENTINEL identity failure")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := runWithIO([]string{"version", "--diagnostic"}, strings.NewReader(""), &stdout, &stderr, dependencies)
	if code != 1 || stdout.Len() != 0 || stderr.String() != "version failed: identity_unavailable\n" || strings.Contains(stderr.String(), "PATH-SENTINEL") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
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
	id             pkgProvider.ID
	responses      []pkgProvider.CompletionResponse
	requests       []pkgProvider.CompletionRequest
	stream         pkgProvider.Stream
	streamErr      error
	streamRequests []pkgProvider.CompletionRequest
	inspectCalls   int
	completeDelay  time.Duration
}

func (provider *cliProvider) Stream(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.Stream, error) {
	provider.streamRequests = append(provider.streamRequests, request)
	return provider.stream, provider.streamErr
}

func (provider *cliProvider) ID() pkgProvider.ID { return provider.id }

func (provider *cliProvider) Complete(_ context.Context, request pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	provider.requests = append(provider.requests, request)
	if provider.completeDelay > 0 {
		time.Sleep(provider.completeDelay)
	}
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
	provider.inspectCalls++
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

func newCLIInteractionConfig(t *testing.T, streaming bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "routes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "routes", "api.php"), []byte("<?php Route::get('/orders');\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf(`version: 2
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 1m
  api_key_env: ""
models:
  embedding: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
interaction:
  chat:
    model: chat-model
    timeout: 1m
    streaming: %t
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576
  agent:
    model: agent-model
    timeout: 1m
    streaming: false
    num_ctx: 8192
    thinking: default
agent:
  id: agent.reference
  tools:
    - workspace.read
policy:
  id: policy.test
  model: allow
  workspace_inspect: allow
  workspace_mutate: deny
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
`, root, streaming)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, root
}

func newCLIChatOnlyConfig(t *testing.T, streaming bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	content := fmt.Sprintf(`version: 2
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 1m
  api_key_env: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
interaction:
  chat:
    model: chat-model
    timeout: 1m
    streaming: %t
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576
policy:
  workspace_mutate: deny
`, root, streaming)
	path := filepath.Join(t.TempDir(), "chat.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, root
}

func newCLIQualificationConfig(t *testing.T, streaming bool) string {
	t.Helper()
	root := t.TempDir()
	content := fmt.Sprintf(`version: 3
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 5m
  api_key_env: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
interaction:
  chat:
    model: chat-model
    timeout: 5m
    streaming: %t
    num_ctx: 4096
    num_predict: 512
    thinking: "false"
    residency: 5m
    max_file_bytes: 1048576
    max_output_bytes: 1048576
policy:
  workspace_mutate: deny
`, root, streaming)
	path := filepath.Join(t.TempDir(), "qualification.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

type cliStreamResult struct {
	chunk pkgProvider.StreamChunk
	err   error
}

type cliStream struct {
	results []cliStreamResult
}

func (stream *cliStream) Recv() (pkgProvider.StreamChunk, error) {
	if len(stream.results) == 0 {
		return pkgProvider.StreamChunk{}, io.EOF
	}
	result := stream.results[0]
	stream.results = stream.results[1:]
	return result.chunk, result.err
}

func (*cliStream) Close() error { return nil }
