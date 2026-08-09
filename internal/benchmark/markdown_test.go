package benchmark

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func TestMarkdownReportIsDerivedFromStrictRedactedJSON(t *testing.T) {
	report := markdownTestReport(t)
	var encodedJSON bytes.Buffer
	if err := EncodeReportJSON(&encodedJSON, report); err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeReportJSON(bytes.NewReader(encodedJSON.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	var markdown bytes.Buffer
	if err := EncodeReportMarkdown(&markdown, decoded); err != nil {
		t.Fatal(err)
	}
	content := markdown.String()
	for _, expected := range []string{
		"# Maestro Benchmark Report", "Maestro Test CPU", "completion",
		"3/3", "criteria_matched_3_of_3", "https://example.test",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("Markdown omitted %q:\n%s", expected, content)
		}
	}
	for _, sensitive := range []string{"user:secret", "token=secret", "/private"} {
		if strings.Contains(content, sensitive) {
			t.Fatalf("Markdown leaked %q:\n%s", sensitive, content)
		}
	}
}

func TestDecodeReportJSONRejectsUnknownFieldsAndMultipleDocuments(t *testing.T) {
	report := markdownTestReport(t)
	var encoded bytes.Buffer
	if err := EncodeReportJSON(&encoded, report); err != nil {
		t.Fatal(err)
	}
	unknown := bytes.Replace(encoded.Bytes(), []byte(`"run_id"`), []byte(`"unknown":true,"run_id"`), 1)
	if _, err := DecodeReportJSON(bytes.NewReader(unknown)); err == nil {
		t.Fatal("expected unknown field rejection")
	}
	multiple := append(append([]byte(nil), encoded.Bytes()...), []byte("{}")...)
	if _, err := DecodeReportJSON(bytes.NewReader(multiple)); err == nil {
		t.Fatal("expected multiple JSON document rejection")
	}
}

func TestWriteReportMarkdownPublishesProtectedAtomicFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.md")
	if err := WriteReportMarkdown(path, markdownTestReport(t)); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("Markdown report mode=%o", info.Mode().Perm())
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary Markdown file leaked: %v %#v", err, entries)
	}
}

func markdownTestReport(t *testing.T) pkgBenchmark.Report {
	t.Helper()
	definition := testScenarioDefinition()
	scenario := pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(context.Context, pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			return pkgBenchmark.IterationResult{
				State:        pkgBenchmark.ResultPassed,
				Measurements: []pkgBenchmark.Measurement{{Name: "total_latency_ms", Value: 12.5, Unit: "ms"}},
				Evaluation: &pkgBenchmark.QualityEvaluation{
					Evaluator: "fixture@1.0:completion", Method: "deterministic_checklist",
					Score: 3, MaxScore: 3, RationaleCode: "criteria_matched_3_of_3",
				},
			}, nil
		},
	}
	runner := deterministicRunner(t, RunnerOptions{Runs: 1, Command: "bench test"})
	report, err := runner.Run(
		context.Background(),
		testManifest(definition),
		pkgBenchmark.ConfigurationProfile{
			Hardware: pkgBenchmark.HardwareProfile{OS: "linux", CPU: "Maestro Test CPU", LogicalCPUs: 8, MemoryMB: 16384},
			Provider: pkgBenchmark.ProviderProfile{ID: "ollama", Endpoint: "https://user:secret@example.test/private?token=secret"},
			Model:    pkgBenchmark.ModelProfile{ID: "fixture"},
		},
		scenario,
	)
	if err != nil {
		t.Fatal(err)
	}
	return report
}
