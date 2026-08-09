package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
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
		report.Configuration.Plugins[0].ID != "laravel" {
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
	if strings.Count(stdout.String(), "Maestro AI Runtime") != 1 || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
