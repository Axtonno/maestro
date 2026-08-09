package smoke

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	maestro "github.com/antonio-cafeo/maestro"
	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type fixtureProvider struct {
	mu sync.Mutex

	id                 pkgProvider.ID
	unavailable        map[pkgProvider.Capability]bool
	unsupported        map[pkgProvider.Capability]bool
	acquisitionPresent bool
	loadCount          int
	unloadCount        int
	removeCount        int
	streamCloseCount   int
}

func (p *fixtureProvider) ID() pkgProvider.ID { return p.id }

func (p *fixtureProvider) InspectCapabilities(
	_ context.Context,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		support := pkgProvider.CapabilitySupported
		availability := pkgProvider.CapabilityAvailabilityAvailable
		if p.unsupported[capability] {
			support = pkgProvider.CapabilityUnsupported
			availability = pkgProvider.CapabilityAvailabilityUnavailable
		} else if p.unavailable[capability] {
			availability = pkgProvider.CapabilityAvailabilityUnavailable
		}
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{
			Capability: capability, Support: support, Availability: availability,
		})
	}

	return pkgProvider.CapabilityReport{
		Provider: p.id, Target: request.Target, Model: request.Model,
		Capabilities: descriptors,
	}, nil
}

func (p *fixtureProvider) Models(context.Context) ([]pkgProvider.Model, error) {
	return []pkgProvider.Model{{ID: "chat"}, {ID: "embed"}, {ID: "lifecycle"}}, nil
}

func (p *fixtureProvider) DiscoverModels(context.Context) ([]pkgProvider.ModelInfo, error) {
	models := []pkgProvider.ModelInfo{
		{Model: pkgProvider.Model{ID: "chat"}, State: pkgProvider.ModelStateLoaded},
		{Model: pkgProvider.Model{ID: "embed"}, State: pkgProvider.ModelStateLoaded},
		{Model: pkgProvider.Model{ID: "lifecycle"}, State: pkgProvider.ModelStateAvailable},
	}
	if p.acquisitionPresent {
		models = append(models, pkgProvider.ModelInfo{
			Model: pkgProvider.Model{ID: "temporary"},
			State: pkgProvider.ModelStateAvailable,
		})
	}

	return models, nil
}

func (p *fixtureProvider) Complete(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	if err := ctx.Err(); err != nil {
		return pkgProvider.CompletionResponse{}, err
	}
	response := pkgProvider.CompletionResponse{
		Model: request.Model,
		Message: pkgProvider.Message{
			Role: pkgProvider.RoleAssistant, Content: "Maestro smoke complete.",
		},
		FinishReason: pkgProvider.FinishReasonStop,
		Usage:        pkgProvider.Usage{InputTokens: 4, OutputTokens: 3},
	}
	if request.Output != nil {
		response.Message.Content = `{"status":"ok"}`
	}
	if len(request.Tools) > 0 && request.ToolChoice.Mode != pkgProvider.ToolChoiceNone {
		response.Message.Content = ""
		response.Message.ToolCalls = []pkgProvider.ToolCall{{
			ID: "call-1", Name: "echo_message",
			Arguments: []byte(`{"message":"Maestro smoke"}`),
		}}
		response.FinishReason = pkgProvider.FinishReasonToolCalls
	}

	return response, nil
}

func (p *fixtureProvider) Stream(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	if len(request.Tools) > 0 {
		return &fixtureStream{
			provider: p,
			context:  ctx,
			chunks: []pkgProvider.StreamChunk{
				{ToolCalls: []pkgProvider.ToolCallDelta{{
					Index: 0, ID: "call-1", Name: "echo_message",
					Arguments: `{"message":"`,
				}}},
				{ToolCalls: []pkgProvider.ToolCallDelta{{
					Index: 0, Arguments: `Maestro smoke"}`,
				}}, FinishReason: pkgProvider.FinishReasonToolCalls},
			},
		}, nil
	}

	return &fixtureStream{
		provider: p,
		context:  ctx,
		chunks: []pkgProvider.StreamChunk{
			{Content: "Maestro"},
			{Content: " complete", FinishReason: pkgProvider.FinishReasonStop},
		},
	}, nil
}

func (p *fixtureProvider) Embed(
	context.Context,
	pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	return pkgProvider.EmbeddingResponse{
		Model: "embed", Embeddings: [][]float32{{0.1, 0.2, 0.3}},
		Usage: pkgProvider.Usage{InputTokens: 2},
	}, nil
}

func (p *fixtureProvider) LoadModel(
	context.Context,
	pkgProvider.ModelLoadRequest,
) error {
	p.mu.Lock()
	p.loadCount++
	p.mu.Unlock()

	return nil
}

func (p *fixtureProvider) UnloadModel(
	context.Context,
	pkgProvider.ModelUnloadRequest,
) error {
	p.mu.Lock()
	p.unloadCount++
	p.mu.Unlock()

	return nil
}

func (p *fixtureProvider) PullModel(
	context.Context,
	pkgProvider.ModelPullRequest,
) (pkgProvider.ModelPullStream, error) {
	return &fixturePullStream{}, nil
}

func (p *fixtureProvider) RemoveModel(
	context.Context,
	pkgProvider.ModelRemoveRequest,
) error {
	p.mu.Lock()
	p.removeCount++
	p.mu.Unlock()

	return nil
}

type fixtureStream struct {
	provider *fixtureProvider
	context  context.Context
	chunks   []pkgProvider.StreamChunk
	index    int
	closed   bool
}

func (s *fixtureStream) Recv() (pkgProvider.StreamChunk, error) {
	if err := s.context.Err(); err != nil {
		return pkgProvider.StreamChunk{}, err
	}
	if s.index >= len(s.chunks) {
		return pkgProvider.StreamChunk{}, io.EOF
	}
	chunk := s.chunks[s.index]
	s.index++

	return chunk, nil
}

func (s *fixtureStream) Close() error {
	if !s.closed {
		s.closed = true
		s.provider.mu.Lock()
		s.provider.streamCloseCount++
		s.provider.mu.Unlock()
	}

	return nil
}

type fixturePullStream struct {
	completed bool
}

func (s *fixturePullStream) Recv() (pkgProvider.ModelPullProgress, error) {
	if s.completed {
		return pkgProvider.ModelPullProgress{}, io.EOF
	}
	s.completed = true

	return pkgProvider.ModelPullProgress{
		Model: "temporary", Stage: pkgProvider.ModelPullStageCompleted,
	}, nil
}

func (*fixturePullStream) Close() error { return nil }

func TestSmokeSuitePassesCompleteManifestThroughProviderRuntime(t *testing.T) {
	manifest := loadTestManifest(t)
	config := enabledTestConfig()
	runtime, provider := newFixtureRuntime(t)
	scenarios, err := NewScenarios(manifest, config, runtime.Providers())
	if err != nil {
		t.Fatalf("construct scenarios: %v", err)
	}
	runner, err := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{
		Runs: 1, Timeout: time.Second, CleanupTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("construct runner: %v", err)
	}

	report, err := runner.Run(
		context.Background(), manifest, config.ConfigurationProfile(), scenarios...,
	)
	if err != nil {
		t.Fatalf("run smoke matrix: %v", err)
	}
	if len(report.Scenarios) != 14 {
		t.Fatalf("expected 14 scenarios, got %d", len(report.Scenarios))
	}
	for _, scenario := range report.Scenarios {
		if scenario.State != pkgBenchmark.ResultPassed {
			t.Fatalf("scenario %q state=%s: %#v", scenario.Scenario.ID, scenario.State, scenario)
		}
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if provider.loadCount != 1 || provider.unloadCount != 1 ||
		provider.removeCount != 1 || provider.streamCloseCount != 3 {
		t.Fatalf(
			"load=%d unload=%d remove=%d stream close=%d",
			provider.loadCount,
			provider.unloadCount,
			provider.removeCount,
			provider.streamCloseCount,
		)
	}
}

func TestSmokeSuiteSkipsEverythingWithoutProviderConfiguration(t *testing.T) {
	manifest := loadTestManifest(t)
	runtime := maestro.New()
	config := Config{ProviderID: ProviderOllama, Models: map[string]string{}}
	scenarios, err := NewScenarios(manifest, config, runtime.Providers())
	if err != nil {
		t.Fatalf("construct scenarios: %v", err)
	}
	runner, _ := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), manifest, config.ConfigurationProfile(), scenarios...,
	)
	if err != nil {
		t.Fatalf("run smoke matrix: %v", err)
	}
	for _, scenario := range report.Scenarios {
		if scenario.State != pkgBenchmark.ResultSkipped ||
			scenario.Samples[0].ReasonCode != "provider_not_configured" {
			t.Fatalf("unexpected disabled scenario: %#v", scenario)
		}
	}
}

func TestAcquisitionFixtureAlreadyPresentIsNeverRemoved(t *testing.T) {
	manifest := loadTestManifest(t)
	config := enabledTestConfig()
	runtime, provider := newFixtureRuntime(t)
	provider.acquisitionPresent = true
	scenarios, err := NewScenarios(manifest, config, runtime.Providers())
	if err != nil {
		t.Fatalf("construct scenarios: %v", err)
	}
	runner, _ := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), manifest, config.ConfigurationProfile(), scenarios...,
	)
	if err != nil {
		t.Fatalf("run smoke matrix: %v", err)
	}
	acquisition := scenarioByID(t, report, "acquisition-pull-remove")
	if acquisition.State != pkgBenchmark.ResultSkipped ||
		acquisition.Samples[0].ReasonCode != "acquisition_fixture_already_present" {
		t.Fatalf("unexpected acquisition result: %#v", acquisition)
	}
	if provider.removeCount != 0 {
		t.Fatal("pre-existing acquisition fixture was removed")
	}
}

func TestAcquisitionRequiresExplicitMutationPermission(t *testing.T) {
	manifest := loadTestManifest(t)
	config := enabledTestConfig()
	config.AllowCatalogMutation = false
	runtime, provider := newFixtureRuntime(t)
	scenarios, err := NewScenarios(manifest, config, runtime.Providers())
	if err != nil {
		t.Fatalf("construct scenarios: %v", err)
	}
	runner, _ := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), manifest, config.ConfigurationProfile(), scenarios...,
	)
	if err != nil {
		t.Fatalf("run smoke matrix: %v", err)
	}
	acquisition := scenarioByID(t, report, "acquisition-pull-remove")
	if acquisition.State != pkgBenchmark.ResultSkipped ||
		acquisition.Samples[0].ReasonCode != "catalog_mutation_not_allowed" {
		t.Fatalf("unexpected acquisition result: %#v", acquisition)
	}
	if provider.removeCount != 0 {
		t.Fatal("acquisition cleanup mutated the catalog without permission")
	}
}

func TestCapabilitySupportAndAvailabilityProduceDistinctStates(t *testing.T) {
	manifest := loadTestManifest(t)
	config := enabledTestConfig()
	runtime, provider := newFixtureRuntime(t)
	provider.unsupported[pkgProvider.CapabilityEmbedding] = true
	provider.unavailable[pkgProvider.CapabilityStructuredOutput] = true
	scenarios, err := NewScenarios(manifest, config, runtime.Providers())
	if err != nil {
		t.Fatalf("construct scenarios: %v", err)
	}
	runner, _ := internalBenchmark.NewRunner(internalBenchmark.RunnerOptions{Runs: 1})

	report, err := runner.Run(
		context.Background(), manifest, config.ConfigurationProfile(), scenarios...,
	)
	if err != nil {
		t.Fatalf("run smoke matrix: %v", err)
	}
	if got := scenarioByID(t, report, "embedding"); got.State != pkgBenchmark.ResultUnsupported {
		t.Fatalf("embedding state=%s", got.State)
	}
	if got := scenarioByID(t, report, "structured-json"); got.State != pkgBenchmark.ResultSkipped {
		t.Fatalf("structured output state=%s", got.State)
	}
}

func TestSmokeSuiteRejectsChangedMutationGuard(t *testing.T) {
	manifest := loadTestManifest(t)
	for index := range manifest.Scenarios {
		if manifest.Scenarios[index].ID == "acquisition-pull-remove" {
			manifest.Scenarios[index].MutationGuard = "UNSAFE=true"
		}
	}
	runtime, _ := newFixtureRuntime(t)

	if _, err := NewScenarios(
		manifest,
		enabledTestConfig(),
		runtime.Providers(),
	); err == nil {
		t.Fatal("expected unsupported mutation guard rejection")
	}
}

func newFixtureRuntime(t *testing.T) (maestro.Runtime, *fixtureProvider) {
	t.Helper()
	runtime := maestro.New()
	provider := &fixtureProvider{
		id:          ProviderOllama,
		unavailable: make(map[pkgProvider.Capability]bool),
		unsupported: make(map[pkgProvider.Capability]bool),
	}
	if err := runtime.Providers().Register(provider); err != nil {
		t.Fatalf("register fixture provider: %v", err)
	}
	if err := runtime.Providers().SetDefault(provider.id); err != nil {
		t.Fatalf("set fixture default: %v", err)
	}

	return runtime, provider
}

func enabledTestConfig() Config {
	return Config{
		ProviderID: ProviderOllama, BaseURL: "http://localhost:11434",
		Enabled: true, AllowCatalogMutation: true, AdapterTimeout: time.Second,
		Models: map[string]string{
			"chat": "chat", "embedding": "embed",
			"lifecycle": "lifecycle", "acquisition_fixture": "temporary",
		},
	}
}

func loadTestManifest(t *testing.T) pkgBenchmark.Manifest {
	t.Helper()
	manifest, err := internalBenchmark.LoadManifest("../../../docs/provider-smoke-benchmark-manifest.yaml")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	return manifest
}

func scenarioByID(
	t *testing.T,
	report pkgBenchmark.Report,
	id string,
) pkgBenchmark.ScenarioReport {
	t.Helper()
	for _, scenario := range report.Scenarios {
		if scenario.Scenario.ID == id {
			return scenario
		}
	}
	t.Fatalf("scenario %q not found", id)

	return pkgBenchmark.ScenarioReport{}
}

var (
	_ pkgProvider.Completer           = (*fixtureProvider)(nil)
	_ pkgProvider.Streamer            = (*fixtureProvider)(nil)
	_ pkgProvider.Embedder            = (*fixtureProvider)(nil)
	_ pkgProvider.ModelLister         = (*fixtureProvider)(nil)
	_ pkgProvider.ModelDiscoverer     = (*fixtureProvider)(nil)
	_ pkgProvider.ModelLoader         = (*fixtureProvider)(nil)
	_ pkgProvider.ModelUnloader       = (*fixtureProvider)(nil)
	_ pkgProvider.ModelPuller         = (*fixtureProvider)(nil)
	_ pkgProvider.ModelRemover        = (*fixtureProvider)(nil)
	_ pkgProvider.CapabilityInspector = (*fixtureProvider)(nil)
	_ pkgProvider.Stream              = (*fixtureStream)(nil)
	_ pkgProvider.ModelPullStream     = (*fixturePullStream)(nil)
)
