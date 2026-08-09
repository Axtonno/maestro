package runtimebench_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	"github.com/antonio-cafeo/maestro/internal/benchmark/runtimebench"
	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	internalProvider "github.com/antonio-cafeo/maestro/internal/provider"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestRuntimeBenchmarkAllScenariosPassWithDeterministicProvider(t *testing.T) {
	manifest, err := internalBenchmark.LoadManifest("../../../docs/runtime-benchmark-manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	runtime := internalProvider.NewRuntime("")
	provider := &benchmarkProvider{id: "fixture"}
	if err := runtime.Register(provider); err != nil {
		t.Fatal(err)
	}
	config := smoke.Config{
		ProviderID: "fixture", Enabled: true, AllowCatalogMutation: true,
		Models: map[string]string{
			"chat": "chat", "embedding": "embedding", "lifecycle": "lifecycle",
			"acquisition_fixture": "acquisition",
		},
	}
	scenarios, err := runtimebench.NewScenarios(manifest, config, runtime, fixtureSampler{})
	if err != nil {
		t.Fatal(err)
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Runs: 1, Timeout: time.Second, CleanupTimeout: time.Second, Command: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	report, err := runner.Run(context.Background(), manifest, pkgBenchmark.ConfigurationProfile{}, scenarios...)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Scenarios) != 11 {
		t.Fatalf("scenario count=%d", len(report.Scenarios))
	}
	for _, scenario := range report.Scenarios {
		if scenario.State != pkgBenchmark.ResultPassed {
			t.Fatalf("scenario %s: %#v", scenario.Scenario.ID, scenario)
		}
	}
}

func TestSelectManifestSeparatesCommands(t *testing.T) {
	manifest, err := internalBenchmark.LoadManifest("../../../docs/runtime-benchmark-manifest.yaml")
	if err != nil {
		t.Fatal(err)
	}
	providerManifest, err := runtimebench.SelectManifest(manifest, runtimebench.KindProvider)
	if err != nil || len(providerManifest.Scenarios) != 4 {
		t.Fatalf("provider scenarios=%d err=%v", len(providerManifest.Scenarios), err)
	}
	modelManifest, err := runtimebench.SelectManifest(manifest, runtimebench.KindModel)
	if err != nil || len(modelManifest.Scenarios) != 7 {
		t.Fatalf("model scenarios=%d err=%v", len(modelManifest.Scenarios), err)
	}
}

type fixtureSampler struct{}
type fixtureSession struct{}

func (fixtureSampler) Start() runtimebench.SampleSession { return fixtureSession{} }
func (fixtureSession) Stop() []pkgBenchmark.Measurement {
	return []pkgBenchmark.Measurement{{Name: "peak_memory_mb", Value: 1, Unit: "MiB", Scope: "fixture_process", Method: "fixture"}}
}

type benchmarkProvider struct{ id pkgProvider.ID }

func (p *benchmarkProvider) ID() pkgProvider.ID { return p.id }

func (p *benchmarkProvider) Complete(context.Context, pkgProvider.CompletionRequest) (pkgProvider.CompletionResponse, error) {
	return pkgProvider.CompletionResponse{
		Message:      pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Maestro fixture."},
		FinishReason: "stop", Usage: pkgProvider.Usage{InputTokens: 4, OutputTokens: 3},
	}, nil
}

func (p *benchmarkProvider) Stream(ctx context.Context, _ pkgProvider.CompletionRequest) (pkgProvider.Stream, error) {
	return &benchmarkStream{ctx: ctx}, nil
}

func (p *benchmarkProvider) Embed(_ context.Context, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
	embeddings := make([][]float32, len(request.Inputs))
	for index := range embeddings {
		embeddings[index] = []float32{0.1, 0.2, 0.3}
	}
	return pkgProvider.EmbeddingResponse{Model: request.Model, Embeddings: embeddings}, nil
}

func (p *benchmarkProvider) Models(context.Context) ([]pkgProvider.Model, error) {
	return []pkgProvider.Model{{ID: "chat"}, {ID: "embedding"}, {ID: "lifecycle"}}, nil
}

func (p *benchmarkProvider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	return []pkgProvider.ModelInfo{
		{Model: pkgProvider.Model{ID: "chat"}, State: pkgProvider.ModelStateLoaded},
		{Model: pkgProvider.Model{ID: "embedding"}, State: pkgProvider.ModelStateLoaded},
		{Model: pkgProvider.Model{ID: "lifecycle"}, State: pkgProvider.ModelStateLoaded},
	}, nil
}

func (p *benchmarkProvider) LoadModel(context.Context, pkgProvider.ModelLoadRequest) error {
	return nil
}
func (p *benchmarkProvider) UnloadModel(context.Context, pkgProvider.ModelUnloadRequest) error {
	return nil
}
func (p *benchmarkProvider) RemoveModel(context.Context, pkgProvider.ModelRemoveRequest) error {
	return nil
}

func (p *benchmarkProvider) PullModel(ctx context.Context, _ pkgProvider.ModelPullRequest) (pkgProvider.ModelPullStream, error) {
	return &benchmarkPullStream{ctx: ctx}, nil
}

func (p *benchmarkProvider) InspectCapabilities(_ context.Context, request pkgProvider.CapabilityRequest) (pkgProvider.CapabilityReport, error) {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{
			Capability: capability, Support: pkgProvider.CapabilitySupported,
			Availability: pkgProvider.CapabilityAvailabilityAvailable,
		})
	}
	return pkgProvider.CapabilityReport{Provider: p.id, Target: request.Target, Model: request.Model, Capabilities: descriptors}, nil
}

type benchmarkStream struct {
	ctx   context.Context
	index int
}

func (s *benchmarkStream) Recv() (pkgProvider.StreamChunk, error) {
	if err := s.ctx.Err(); err != nil {
		return pkgProvider.StreamChunk{}, err
	}
	s.index++
	switch s.index {
	case 1:
		return pkgProvider.StreamChunk{Content: "Maestro "}, nil
	case 2:
		return pkgProvider.StreamChunk{Content: "fixture.", FinishReason: "stop", Usage: pkgProvider.Usage{OutputTokens: 3}}, nil
	default:
		return pkgProvider.StreamChunk{}, io.EOF
	}
}

func (*benchmarkStream) Close() error { return nil }

type benchmarkPullStream struct {
	ctx   context.Context
	first bool
}

func (s *benchmarkPullStream) Recv() (pkgProvider.ModelPullProgress, error) {
	if !s.first {
		s.first = true
		return pkgProvider.ModelPullProgress{Stage: pkgProvider.ModelPullStageDownloading}, nil
	}
	if err := s.ctx.Err(); err != nil {
		return pkgProvider.ModelPullProgress{}, err
	}
	return pkgProvider.ModelPullProgress{}, errors.New("expected cancellation")
}

func (*benchmarkPullStream) Close() error { return nil }
