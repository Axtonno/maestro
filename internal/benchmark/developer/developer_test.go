package developer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	internalProvider "github.com/antonio-cafeo/maestro/internal/provider"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestEmbeddedDatasetIsVersionedSafeAndMaterializable(t *testing.T) {
	dataset, err := LoadDataset()
	if err != nil {
		t.Fatal(err)
	}
	if dataset.ID != DatasetID || dataset.Version != DatasetVersion || len(dataset.Tasks) != 6 {
		t.Fatalf("unexpected dataset: %#v", dataset)
	}
	for _, task := range dataset.Tasks {
		if _, exists := dataset.Task(task.ID); !exists {
			t.Fatalf("task %q is not indexed", task.ID)
		}
	}
	workspace, cleanup, err := dataset.Materialize()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "artisan")); err != nil {
		t.Fatalf("materialized artisan: %v", err)
	}
	composer, err := os.ReadFile(filepath.Join(workspace, "composer.json"))
	if err != nil || !bytes.Contains(composer, []byte("laravel/framework")) {
		t.Fatalf("materialized composer manifest: %v %s", err, composer)
	}
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("workspace still exists after cleanup: %v", err)
	}
}

func TestDeveloperScenariosSeparateTechnicalSuccessAndQuality(t *testing.T) {
	dataset, err := LoadDataset()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := internalBenchmark.LoadManifest("../../../docs/developer-benchmark-manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	runtime := internalProvider.NewRuntime("")
	provider := &developerProvider{id: "fixture"}
	if err := runtime.Register(provider); err != nil {
		t.Fatal(err)
	}
	config := smoke.Config{
		ProviderID: "fixture", Enabled: true,
		Models: map[string]string{"chat": "chat", "embedding": "embedding"},
	}
	scenarios, err := NewScenarios(manifest, dataset, config, runtime)
	if err != nil {
		t.Fatal(err)
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Runs: 1, Timeout: time.Second, CleanupTimeout: time.Second, Command: "bench laravel",
	})
	if err != nil {
		t.Fatal(err)
	}
	report, err := runner.Run(context.Background(), manifest, pkgBenchmark.ConfigurationProfile{
		Dataset: pkgBenchmark.DatasetProfile{ID: dataset.ID, Version: dataset.Version},
	}, scenarios...)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scenarios) != 6 {
		t.Fatalf("scenario count=%d", len(report.Scenarios))
	}
	for _, scenario := range report.Scenarios {
		sample := scenario.Samples[0]
		if scenario.State != pkgBenchmark.ResultPassed || sample.Evaluation == nil ||
			sample.Evaluation.Score != 3 || sample.Evaluation.MaxScore != 3 {
			t.Fatalf("unexpected scenario result: %#v", scenario)
		}
	}
	var output bytes.Buffer
	if err := internalBenchmark.EncodeReportJSON(&output, report); err != nil {
		t.Fatal(err)
	}
	for _, privateContent := range []string{"customer_id", "Use only the following", "Reply with"} {
		if strings.Contains(output.String(), privateContent) {
			t.Fatalf("report leaked benchmark content %q: %s", privateContent, output.String())
		}
	}
}

func TestGenerationRubricCanScoreZeroWithoutTechnicalFailure(t *testing.T) {
	dataset, err := LoadDataset()
	if err != nil {
		t.Fatal(err)
	}
	task, _ := dataset.Task("developer-explain-controller")
	evaluation := evaluateGeneration(dataset, task, "A concise but unrelated answer.")
	if evaluation.Score != 0 || evaluation.RationaleCode != "criteria_matched_0_of_3" {
		t.Fatalf("unexpected evaluation: %#v", evaluation)
	}
}

func TestCosineRankingRejectsInvalidVectorsAndRanksRelevantFixture(t *testing.T) {
	ranking, err := rankDocuments(
		[]float32{1, 0},
		[][]float32{{0, 1}, {1, 0}, {0.5, 0.5}},
		[]string{"irrelevant", "relevant", "other"},
	)
	if err != nil || bestRelevantRank(ranking, []string{"relevant"}) != 1 {
		t.Fatalf("ranking=%#v err=%v", ranking, err)
	}
	if _, err := cosineSimilarity([]float32{0, 0}, []float32{1, 0}); err == nil {
		t.Fatal("expected zero-norm vector rejection")
	}
}

type developerProvider struct{ id pkgProvider.ID }

func (p *developerProvider) ID() pkgProvider.ID { return p.id }

func (p *developerProvider) Complete(context.Context, pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	return pkgProvider.CompletionResponse{
		Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: strings.Join([]string{
			"validation", "OrderService", "201", "OrderRepository", "PaymentGateway",
			"event dispatcher", "PHPUnit TestCase", "postJson", "mock", "preferred",
			"fallback", "null", "route", "controller", "service", "repository", "payment",
		}, " ")},
		FinishReason: "stop", Usage: pkgProvider.Usage{InputTokens: 100, OutputTokens: 50},
	}, nil
}

func (p *developerProvider) Embed(_ context.Context, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
	embeddings := make([][]float32, len(request.Inputs))
	for index, input := range request.Inputs {
		switch {
		case index == 0:
			embeddings[index] = []float32{1, 0}
		case strings.Contains(input, "routes/api.php"):
			embeddings[index] = []float32{1, 0}
		case strings.Contains(input, "OrderController.php"):
			embeddings[index] = []float32{0.9, 0.1}
		case strings.Contains(input, "OrderService.php"):
			embeddings[index] = []float32{0.8, 0.2}
		default:
			embeddings[index] = []float32{0.1, 0.9}
		}
	}
	return pkgProvider.EmbeddingResponse{Model: request.Model, Embeddings: embeddings}, nil
}

func (p *developerProvider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{
			Capability: capability, Support: pkgProvider.CapabilitySupported,
			Availability: pkgProvider.CapabilityAvailabilityAvailable,
		})
	}
	return pkgProvider.CapabilityReport{Provider: p.id, Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

var (
	_ pkgProvider.Completer           = (*developerProvider)(nil)
	_ pkgProvider.Embedder            = (*developerProvider)(nil)
	_ pkgProvider.CapabilityInspector = (*developerProvider)(nil)
)
