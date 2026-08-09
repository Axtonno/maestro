package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func TestEncodeReportJSONRedactsCredentialsURLsAndUserPaths(t *testing.T) {
	definition := testScenarioDefinition()
	runner := deterministicRunner(t, RunnerOptions{Runs: 1, Command: "bench smoke"})
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(_ context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{
				State: pkgBenchmark.ResultFailed,
				Error: &pkgBenchmark.ErrorRecord{
					Kind: "provider", Code: "unavailable",
					Model: "/home/antonio-cafeo/private/model.gguf",
				},
			}, nil
		},
	}
	configuration := pkgBenchmark.ConfigurationProfile{
		Provider: pkgBenchmark.ProviderProfile{
			ID: "ollama", Endpoint: "https://user:secret@example.test/private?token=secret",
		},
		Model: pkgBenchmark.ModelProfile{ID: "/home/antonio-cafeo/models/private.gguf"},
		Models: map[string]pkgBenchmark.ModelProfile{
			"lifecycle": {ID: "/home/antonio-cafeo/models/lifecycle.gguf"},
		},
	}
	report, err := runner.Run(
		context.Background(), testManifest(definition), configuration, scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}

	var output bytes.Buffer
	if err := EncodeReportJSON(&output, report); err != nil {
		t.Fatalf("encode report: %v", err)
	}
	encoded := output.String()
	for _, sensitive := range []string{
		"user:secret", "token=secret", "/private", "/home/antonio-cafeo",
	} {
		if strings.Contains(encoded, sensitive) {
			t.Fatalf("report leaked %q: %s", sensitive, encoded)
		}
	}
	if !strings.Contains(encoded, `"endpoint": "https://example.test"`) ||
		!strings.Contains(encoded, `"id": "[redacted-path]"`) {
		t.Fatalf("expected redacted report, got %s", encoded)
	}
	if report.Configuration.Provider.Endpoint == "https://example.test" ||
		report.Configuration.Model.ID == "[redacted-path]" ||
		report.Configuration.Models["lifecycle"].ID == "[redacted-path]" {
		t.Fatal("encoding mutated the source report")
	}
}

func TestWriteReportJSONPublishesProtectedAtomicFile(t *testing.T) {
	definition := testScenarioDefinition()
	runner := deterministicRunner(t, RunnerOptions{Runs: 1})
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultPassed}, nil
		},
	}
	report, err := runner.Run(
		context.Background(), testManifest(definition),
		pkgBenchmark.ConfigurationProfile{}, scenario,
	)
	if err != nil {
		t.Fatalf("run benchmark: %v", err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "report.json")

	if err := WriteReportJSON(path, report); err != nil {
		t.Fatalf("write report: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat report: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("report mode=%o", info.Mode().Perm())
	}
	encoded, err := os.ReadFile(path)
	if err != nil || !bytes.Contains(encoded, []byte(pkgBenchmark.ReportSchemaVersion)) {
		t.Fatalf("read report: %v, %s", err, encoded)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary report was not cleaned up: %v, %#v", err, entries)
	}
}

func TestPublishedReportSchemaMatchesCurrentVersion(t *testing.T) {
	encoded, err := os.ReadFile("../../docs/schemas/benchmark-report-v1.schema.json")
	if err != nil {
		t.Fatalf("read report schema: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(encoded, &schema); err != nil {
		t.Fatalf("decode report schema: %v", err)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties: %#v", schema)
	}
	version, ok := properties["schema_version"].(map[string]any)
	if !ok || version["const"] != pkgBenchmark.ReportSchemaVersion {
		t.Fatalf("schema version does not match runtime: %#v", version)
	}
}

func TestProviderHandoffEndToEndProducesVersionedRedactedReport(t *testing.T) {
	manifest, err := LoadManifest("../../docs/provider-smoke-benchmark-manifest.yaml")
	if err != nil {
		t.Fatalf("load provider handoff: %v", err)
	}
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: manifest.Scenarios[0],
		RunFunc: func(_ context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{
				State: pkgBenchmark.ResultPassed,
				Measurements: []pkgBenchmark.Measurement{{
					Name: "total_latency_ms", Unit: "ms", Value: 5,
				}},
			}, nil
		},
	}
	runner := deterministicRunner(t, RunnerOptions{Runs: 1, Command: "bench smoke"})
	report, err := runner.Run(
		context.Background(),
		manifest,
		pkgBenchmark.ConfigurationProfile{
			Provider: pkgBenchmark.ProviderProfile{
				ID: "ollama", Endpoint: "http://user:secret@localhost:11434/private",
			},
		},
		scenario,
	)
	if err != nil {
		t.Fatalf("run provider handoff: %v", err)
	}
	if len(report.Scenarios) != 14 ||
		report.Scenarios[0].State != pkgBenchmark.ResultPassed ||
		report.Scenarios[1].State != pkgBenchmark.ResultSkipped {
		t.Fatalf("unexpected handoff report: %#v", report.Scenarios)
	}

	var output bytes.Buffer
	if err := EncodeReportJSON(&output, report); err != nil {
		t.Fatalf("encode provider handoff: %v", err)
	}
	if !strings.Contains(
		output.String(),
		`"schema_version": "`+pkgBenchmark.ReportSchemaVersion+`"`,
	) ||
		strings.Contains(output.String(), "secret") ||
		strings.Contains(output.String(), "/private") {
		t.Fatalf("unexpected serialized handoff: %s", output.String())
	}
}
