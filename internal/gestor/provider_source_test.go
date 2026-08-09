package gestor

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	internalProvider "github.com/antonio-cafeo/maestro/internal/provider"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type providerReportKey struct {
	provider pkgProvider.ID
	target   pkgProvider.CapabilityTarget
	model    string
}

type capabilityRuntimeFixture struct {
	mu sync.RWMutex

	providers []pkgProvider.ID
	reports   map[providerReportKey]pkgProvider.CapabilityReport
	errors    map[providerReportKey]error
	calls     []providerReportKey
	afterCall func(providerReportKey)
}

func (runtime *capabilityRuntimeFixture) Registered() []pkgProvider.ID {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()

	return slices.Clone(runtime.providers)
}

func (runtime *capabilityRuntimeFixture) Capabilities(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	if err := ctx.Err(); err != nil {
		return pkgProvider.CapabilityReport{}, err
	}
	key := providerReportKey{provider: providerID, target: request.Target, model: request.Model}
	runtime.mu.Lock()
	runtime.calls = append(runtime.calls, key)
	report := runtime.reports[key]
	err := runtime.errors[key]
	afterCall := runtime.afterCall
	runtime.mu.Unlock()
	if afterCall != nil {
		afterCall(key)
	}

	return report, err
}

func (runtime *capabilityRuntimeFixture) setError(key providerReportKey, err error) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	runtime.errors[key] = err
}

func (runtime *capabilityRuntimeFixture) recordedCalls() []providerReportKey {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()

	return slices.Clone(runtime.calls)
}

func capabilityReport(
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
	change func(*pkgProvider.CapabilityDescriptor),
) pkgProvider.CapabilityReport {
	descriptors := make([]pkgProvider.CapabilityDescriptor, 0, len(pkgProvider.KnownCapabilities()))
	for _, capability := range pkgProvider.KnownCapabilities() {
		descriptor := pkgProvider.CapabilityDescriptor{
			Capability:   capability,
			Support:      pkgProvider.CapabilitySupported,
			Availability: pkgProvider.CapabilityAvailabilityUnknown,
		}
		if change != nil {
			change(&descriptor)
		}
		descriptors = append(descriptors, descriptor)
	}

	return pkgProvider.CapabilityReport{
		Provider:     providerID,
		Target:       request.Target,
		Model:        request.Model,
		Capabilities: descriptors,
	}
}

func newCapabilityRuntimeFixture(
	providerID pkgProvider.ID,
	models []string,
) *capabilityRuntimeFixture {
	runtime := &capabilityRuntimeFixture{
		providers: []pkgProvider.ID{providerID},
		reports:   make(map[providerReportKey]pkgProvider.CapabilityReport),
		errors:    make(map[providerReportKey]error),
	}
	requests := []pkgProvider.CapabilityRequest{
		{Target: pkgProvider.CapabilityTargetAdapter},
		{Target: pkgProvider.CapabilityTargetInstance},
	}
	for _, model := range models {
		requests = append(requests, pkgProvider.CapabilityRequest{
			Target: pkgProvider.CapabilityTargetModel,
			Model:  model,
		})
	}
	for _, request := range requests {
		key := providerReportKey{provider: providerID, target: request.Target, model: request.Model}
		runtime.reports[key] = capabilityReport(providerID, request, nil)
	}

	return runtime
}

func TestProviderCapabilitySourceMapsAllTargetsAndAvailability(t *testing.T) {
	models := []string{"qwen2.5-coder:7b", "llama3.1:8b"}
	runtime := newCapabilityRuntimeFixture("ollama", models)

	instanceRequest := pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance}
	instanceKey := providerReportKey{provider: "ollama", target: instanceRequest.Target}
	runtime.reports[instanceKey] = capabilityReport("ollama", instanceRequest, func(descriptor *pkgProvider.CapabilityDescriptor) {
		switch descriptor.Capability {
		case pkgProvider.CapabilityCompletion:
			descriptor.Availability = pkgProvider.CapabilityAvailabilityAvailable
		case pkgProvider.CapabilityEmbedding:
			descriptor.Availability = pkgProvider.CapabilityAvailabilityUnavailable
		}
	})
	qwenRequest := pkgProvider.CapabilityRequest{
		Target: pkgProvider.CapabilityTargetModel,
		Model:  "qwen2.5-coder:7b",
	}
	runtime.reports[providerReportKey{provider: "ollama", target: qwenRequest.Target, model: qwenRequest.Model}] = capabilityReport("ollama", qwenRequest, func(descriptor *pkgProvider.CapabilityDescriptor) {
		if descriptor.Capability == pkgProvider.CapabilityToolCalling {
			descriptor.Availability = pkgProvider.CapabilityAvailabilityUnknown
		}
		if descriptor.Capability == pkgProvider.CapabilityEmbedding {
			descriptor.Availability = pkgProvider.CapabilityAvailabilityUnavailable
		}
	})
	llamaRequest := pkgProvider.CapabilityRequest{
		Target: pkgProvider.CapabilityTargetModel,
		Model:  "llama3.1:8b",
	}
	runtime.reports[providerReportKey{provider: "ollama", target: llamaRequest.Target, model: llamaRequest.Model}] = capabilityReport("ollama", llamaRequest, func(descriptor *pkgProvider.CapabilityDescriptor) {
		if descriptor.Capability == pkgProvider.CapabilityToolCalling {
			descriptor.Availability = pkgProvider.CapabilityAvailabilityAvailable
		}
	})

	configuredModels := map[pkgProvider.ID][]string{
		"ollama": {"qwen2.5-coder:7b", "llama3.1:8b"},
	}
	source, err := NewProviderCapabilitySource(runtime, configuredModels)
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	configuredModels["ollama"][0] = "changed"

	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover provider capabilities: %v", err)
	}
	if len(descriptors) != len(pkgProvider.KnownCapabilities())*4 {
		t.Fatalf("expected %d descriptors, got %d", len(pkgProvider.KnownCapabilities())*4, len(descriptors))
	}
	if !slices.IsSortedFunc(descriptors, func(left, right pkgGestor.Descriptor) int {
		return left.Compare(right)
	}) {
		t.Fatal("provider descriptors are not ordered")
	}

	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderCompletion, pkgGestor.ScopeAdapter, "", pkgGestor.AvailabilityUnknown)
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderCompletion, pkgGestor.ScopeInstance, "", pkgGestor.AvailabilityAvailable)
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderEmbedding, pkgGestor.ScopeInstance, "", pkgGestor.AvailabilityUnavailable)
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderToolCalling, pkgGestor.ScopeModel, "qwen2.5-coder:7b", pkgGestor.AvailabilityUnknown)
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderEmbedding, pkgGestor.ScopeModel, "qwen2.5-coder:7b", pkgGestor.AvailabilityUnavailable)
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderToolCalling, pkgGestor.ScopeModel, "llama3.1:8b", pkgGestor.AvailabilityAvailable)

	wantCalls := []providerReportKey{
		{provider: "ollama", target: pkgProvider.CapabilityTargetAdapter},
		{provider: "ollama", target: pkgProvider.CapabilityTargetInstance},
		{provider: "ollama", target: pkgProvider.CapabilityTargetModel, model: "llama3.1:8b"},
		{provider: "ollama", target: pkgProvider.CapabilityTargetModel, model: "qwen2.5-coder:7b"},
	}
	if !slices.Equal(runtime.recordedCalls(), wantCalls) {
		t.Fatalf("expected calls %#v, got %#v", wantCalls, runtime.recordedCalls())
	}
}

func TestProviderCapabilitySourceMapsEveryKnownCapability(t *testing.T) {
	runtime := newCapabilityRuntimeFixture("provider", nil)
	source, err := NewProviderCapabilitySource(runtime, nil)
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover provider capabilities: %v", err)
	}

	found := make(map[pkgGestor.CapabilityID]bool, len(pkgProvider.KnownCapabilities()))
	for _, descriptor := range descriptors {
		if descriptor.Target.Scope == pkgGestor.ScopeAdapter {
			found[descriptor.Capability] = true
		}
	}
	for _, capability := range pkgProvider.KnownCapabilities() {
		mapped, err := providerCapabilityID(capability)
		if err != nil {
			t.Fatalf("map capability %q: %v", capability, err)
		}
		if !found[mapped] {
			t.Errorf("missing adapter descriptor for %q", mapped)
		}
	}
}

func TestProviderCapabilitySourceOmitsUnsupportedDeclarations(t *testing.T) {
	runtime := newCapabilityRuntimeFixture("provider", nil)
	request := pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetAdapter}
	key := providerReportKey{provider: "provider", target: request.Target}
	runtime.reports[key] = capabilityReport("provider", request, func(descriptor *pkgProvider.CapabilityDescriptor) {
		if descriptor.Capability == pkgProvider.CapabilityToolCalling {
			descriptor.Support = pkgProvider.CapabilityUnsupported
			descriptor.Availability = pkgProvider.CapabilityAvailabilityUnavailable
		}
	})
	source, err := NewProviderCapabilitySource(runtime, nil)
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover provider capabilities: %v", err)
	}
	for _, descriptor := range descriptors {
		if descriptor.Target.Scope == pkgGestor.ScopeAdapter &&
			descriptor.Capability == pkgGestor.CapabilityProviderToolCalling {
			t.Fatal("unsupported capability was incorrectly declared in Gestor")
		}
	}
}

func TestProviderCapabilitySourceRejectsInvalidInputsAndReports(t *testing.T) {
	if _, err := NewProviderCapabilitySource(nil, nil); !errors.Is(err, pkgGestor.ErrInvalidSource) {
		t.Fatalf("nil runtime: expected ErrInvalidSource, got %v", err)
	}
	runtime := newCapabilityRuntimeFixture("provider", nil)
	for name, models := range map[string]map[pkgProvider.ID][]string{
		"invalid provider": {" bad": {"model"}},
		"invalid model":    {"provider": {" model"}},
		"duplicate model":  {"provider": {"model", "model"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewProviderCapabilitySource(runtime, models); !errors.Is(err, pkgGestor.ErrInvalidSource) {
				t.Fatalf("expected ErrInvalidSource, got %v", err)
			}
		})
	}

	request := pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetAdapter}
	key := providerReportKey{provider: "provider", target: request.Target}
	invalidReports := []pkgProvider.CapabilityReport{
		{Provider: "other", Target: request.Target},
		{Provider: "provider", Target: request.Target},
	}
	wrongOrder := capabilityReport("provider", request, nil)
	wrongOrder.Capabilities[0], wrongOrder.Capabilities[1] = wrongOrder.Capabilities[1], wrongOrder.Capabilities[0]
	invalidReports = append(invalidReports, wrongOrder)
	invalidSupport := capabilityReport("provider", request, nil)
	invalidSupport.Capabilities[0].Support = "maybe"
	invalidReports = append(invalidReports, invalidSupport)
	invalidAvailability := capabilityReport("provider", request, nil)
	invalidAvailability.Capabilities[0].Availability = "maybe"
	invalidReports = append(invalidReports, invalidAvailability)
	unsupportedUnknown := capabilityReport("provider", request, nil)
	unsupportedUnknown.Capabilities[0].Support = pkgProvider.CapabilityUnsupported
	invalidReports = append(invalidReports, unsupportedUnknown)

	for index, report := range invalidReports {
		runtime.reports[key] = report
		source, err := NewProviderCapabilitySource(runtime, nil)
		if err != nil {
			t.Fatalf("new source %d: %v", index, err)
		}
		if _, err := source.Discover(context.Background()); !errors.Is(err, pkgGestor.ErrInvalidDescriptor) {
			t.Errorf("report %d: expected ErrInvalidDescriptor, got %v", index, err)
		}
	}
}

func TestProviderCapabilitySourcePropagatesErrorsCancellationAndPreventsPartialRefresh(t *testing.T) {
	runtime := newCapabilityRuntimeFixture("provider", nil)
	source, err := NewProviderCapabilitySource(runtime, nil)
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	registry := NewRegistry()
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register provider source: %v", err)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	want := registry.Snapshot()

	failure := errors.New("introspection failed")
	instanceKey := providerReportKey{provider: "provider", target: pkgProvider.CapabilityTargetInstance}
	runtime.setError(instanceKey, failure)
	err = registry.Refresh(context.Background())
	if !errors.Is(err, failure) || !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("expected introspection cause and source failure, got %v", err)
	}
	assertSameSnapshot(t, registry.Snapshot(), want)

	runtime.setError(instanceKey, nil)
	ctx, cancel := context.WithCancel(context.Background())
	runtime.mu.Lock()
	runtime.afterCall = func(key providerReportKey) {
		if key.target == pkgProvider.CapabilityTargetAdapter {
			cancel()
		}
	}
	runtime.mu.Unlock()
	err = registry.Refresh(ctx)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("expected cancellation and source failure, got %v", err)
	}
	assertSameSnapshot(t, registry.Snapshot(), want)
}

type inspectorProvider struct {
	id pkgProvider.ID
}

func (provider *inspectorProvider) ID() pkgProvider.ID { return provider.id }

func (provider *inspectorProvider) InspectCapabilities(
	_ context.Context,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	return capabilityReport(provider.id, request, func(descriptor *pkgProvider.CapabilityDescriptor) {
		if request.Target == pkgProvider.CapabilityTargetModel {
			descriptor.Availability = pkgProvider.CapabilityAvailabilityAvailable
		}
	}), nil
}

func TestProviderCapabilitySourceUsesInternalRuntimeAdditiveInterface(t *testing.T) {
	providerRuntime := internalProvider.NewRuntime("")
	sourceRuntime, ok := providerRuntime.(providerCapabilityRuntime)
	if !ok {
		t.Fatal("internal Provider Runtime does not implement Gestor source interface")
	}
	if err := providerRuntime.Register(&inspectorProvider{id: "fixture"}); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	source, err := NewProviderCapabilitySource(sourceRuntime, map[pkgProvider.ID][]string{
		"fixture": {"llama3.1:8b"},
	})
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover through real Provider Runtime: %v", err)
	}
	assertProviderDescriptor(t, descriptors, pkgGestor.CapabilityProviderToolCalling, pkgGestor.ScopeModel, "llama3.1:8b", pkgGestor.AvailabilityAvailable)
}

func TestProviderCapabilitySourceConcurrentDiscovery(t *testing.T) {
	runtime := newCapabilityRuntimeFixture("provider", []string{"model"})
	source, err := NewProviderCapabilitySource(runtime, map[pkgProvider.ID][]string{"provider": {"model"}})
	if err != nil {
		t.Fatalf("new source: %v", err)
	}

	const workers = 12
	var wait sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			descriptors, err := source.Discover(context.Background())
			if err == nil && len(descriptors) != len(pkgProvider.KnownCapabilities())*3 {
				err = errors.New("unexpected descriptor count")
			}
			results <- err
		}()
	}
	wait.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Errorf("concurrent discovery: %v", err)
		}
	}
}

func TestDiscoverySourcesComposeInSnapshotRegistry(t *testing.T) {
	componentSource, err := NewRuntimeComponentSource(&componentMemoryCatalog{
		components: []pkgRuntime.Component{&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID:           "workspace",
			Capabilities: []pkgRuntime.Capability{pkgRuntime.CapabilityHealth},
		}}},
	})
	if err != nil {
		t.Fatalf("new component source: %v", err)
	}
	providerRuntime := newCapabilityRuntimeFixture("provider", nil)
	providerSource, err := NewProviderCapabilitySource(providerRuntime, nil)
	if err != nil {
		t.Fatalf("new provider source: %v", err)
	}
	registry := NewRegistry()
	if err := registry.RegisterSource(componentSource); err != nil {
		t.Fatalf("register component source: %v", err)
	}
	if err := registry.RegisterSource(providerSource); err != nil {
		t.Fatalf("register provider source: %v", err)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh composed sources: %v", err)
	}

	snapshot := registry.Snapshot()
	wantSources := []pkgGestor.SourceID{providerCapabilitySourceID, runtimeComponentSourceID}
	if !slices.Equal(snapshot.Metadata().Sources(), wantSources) {
		t.Fatalf("expected sources %v, got %v", wantSources, snapshot.Metadata().Sources())
	}
	wantDescriptors := 1 + len(pkgProvider.KnownCapabilities())*2
	if snapshot.Metadata().DescriptorCount != wantDescriptors || !snapshot.Metadata().Current {
		t.Fatalf("unexpected composed snapshot metadata: %#v", snapshot.Metadata())
	}
}

func assertProviderDescriptor(
	t *testing.T,
	descriptors []pkgGestor.Descriptor,
	capability pkgGestor.CapabilityID,
	scope pkgGestor.Scope,
	model string,
	availability pkgGestor.Availability,
) {
	t.Helper()
	for _, descriptor := range descriptors {
		if descriptor.Capability == capability && descriptor.Target.Scope == scope && descriptor.Target.Model == model {
			if descriptor.Target.Kind != pkgGestor.TargetKindProvider ||
				descriptor.Target.ID != "ollama" && descriptor.Target.ID != "fixture" ||
				descriptor.Availability != availability ||
				descriptor.Source != providerCapabilitySourceID {
				t.Fatalf("unexpected provider descriptor: %#v", descriptor)
			}
			return
		}
	}
	t.Fatalf("missing descriptor capability %q scope %q model %q", capability, scope, model)
}
