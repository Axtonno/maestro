package gestor

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type graphPlanFixture struct {
	generation   uint64
	dependencies []pkgRuntime.ComponentID
	err          error
}

type dependencyGraphFixture struct {
	mu sync.Mutex

	generation uint64
	current    bool
	plans      map[pkgRuntime.ComponentID]graphPlanFixture
	stateCalls int
	stateFunc  func(int) (uint64, bool)
	planFunc   func(pkgRuntime.ComponentID) (uint64, []pkgRuntime.ComponentID, error)
}

func newDependencyGraphFixture() *dependencyGraphFixture {
	return &dependencyGraphFixture{
		generation: 1,
		current:    true,
		plans:      make(map[pkgRuntime.ComponentID]graphPlanFixture),
	}
}

func (graph *dependencyGraphFixture) State() (uint64, bool) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	graph.stateCalls++
	if graph.stateFunc != nil {
		return graph.stateFunc(graph.stateCalls)
	}

	return graph.generation, graph.current
}

func (graph *dependencyGraphFixture) DependencyPlan(
	componentID pkgRuntime.ComponentID,
) (uint64, []pkgRuntime.ComponentID, error) {
	graph.mu.Lock()
	defer graph.mu.Unlock()

	if graph.planFunc != nil {
		return graph.planFunc(componentID)
	}
	plan, exists := graph.plans[componentID]
	if !exists {
		return graph.generation, nil, pkgRuntime.ErrNotFound
	}

	return plan.generation, slices.Clone(plan.dependencies), plan.err
}

func resolverDescriptor(
	capability pkgGestor.CapabilityID,
	target pkgGestor.Target,
	availability pkgGestor.Availability,
) pkgGestor.Descriptor {
	return pkgGestor.Descriptor{
		Capability:   capability,
		Target:       target,
		Availability: availability,
		Source:       "test.source",
	}
}

func resolverProviderTarget(id string, scope pkgGestor.Scope, model string) pkgGestor.Target {
	return pkgGestor.Target{
		Kind:  pkgGestor.TargetKindProvider,
		ID:    id,
		Scope: scope,
		Model: model,
	}
}

func resolverComponentTarget(id string) pkgGestor.Target {
	return pkgGestor.Target{
		Kind:  pkgGestor.TargetKindComponent,
		ID:    id,
		Scope: pkgGestor.ScopeComponent,
	}
}

func registryForResolver(
	t *testing.T,
	descriptors []pkgGestor.Descriptor,
) *Registry {
	t.Helper()
	registry := NewRegistry()
	source := &memorySource{id: "test.source", descriptors: descriptors}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := registry.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh registry: %v", err)
	}

	return registry
}

func newTestResolver(
	t *testing.T,
	registry *Registry,
	graph dependencyGraphView,
) *Resolver {
	t.Helper()
	resolver, err := NewResolver(registry, graph)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	return resolver
}

func newQuery(
	t *testing.T,
	capability pkgGestor.CapabilityID,
	options pkgGestor.QueryOptions,
) pkgGestor.Query {
	t.Helper()
	query, err := pkgGestor.NewQuery(capability, options)
	if err != nil {
		t.Fatalf("new query: %v", err)
	}

	return query
}

func TestResolverConstructionAndQueryValidation(t *testing.T) {
	graph := newDependencyGraphFixture()
	if _, err := NewResolver(nil, graph); !errors.Is(err, pkgGestor.ErrInvalidResolution) {
		t.Fatalf("nil registry: expected ErrInvalidResolution, got %v", err)
	}
	registry := NewRegistry()
	if _, err := NewResolver(registry, nil); !errors.Is(err, pkgGestor.ErrInvalidResolution) {
		t.Fatalf("nil graph: expected ErrInvalidResolution, got %v", err)
	}
	resolver := newTestResolver(t, registry, graph)
	if _, err := resolver.Candidates(pkgGestor.Query{}); !errors.Is(err, pkgGestor.ErrInvalidQuery) {
		t.Fatalf("zero query: expected ErrInvalidQuery, got %v", err)
	}
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})
	if _, err := resolver.Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("initial snapshot: expected ErrStaleSnapshot, got %v", err)
	}
}

func TestResolverDistinguishesNotFoundAndUnavailable(t *testing.T) {
	unknown := resolverDescriptor(
		pkgGestor.CapabilityProviderCompletion,
		resolverProviderTarget("unknown", pkgGestor.ScopeAdapter, ""),
		pkgGestor.AvailabilityUnknown,
	)
	unavailable := resolverDescriptor(
		pkgGestor.CapabilityProviderCompletion,
		resolverProviderTarget("down", pkgGestor.ScopeInstance, ""),
		pkgGestor.AvailabilityUnavailable,
	)
	resolver := newTestResolver(t, registryForResolver(t, []pkgGestor.Descriptor{unknown, unavailable}), newDependencyGraphFixture())

	missing := newQuery(t, pkgGestor.CapabilityProviderEmbedding, pkgGestor.QueryOptions{})
	if _, err := resolver.Resolve(missing); !errors.Is(err, pkgGestor.ErrNotFound) {
		t.Fatalf("missing capability: expected ErrNotFound, got %v", err)
	}
	wrongScope := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		TargetKind: pkgGestor.TargetKindProvider,
		Scope:      pkgGestor.ScopeModel,
		Model:      "model",
	})
	if _, err := resolver.Resolve(wrongScope); !errors.Is(err, pkgGestor.ErrNotFound) {
		t.Fatalf("missing exact target: expected ErrNotFound, got %v", err)
	}
	instance := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		TargetKind: pkgGestor.TargetKindProvider,
		Scope:      pkgGestor.ScopeInstance,
	})
	if _, err := resolver.Resolve(instance); !errors.Is(err, pkgGestor.ErrUnavailable) {
		t.Fatalf("unavailable declaration: expected ErrUnavailable, got %v", err)
	}
	requireAvailable := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		TargetKind:       pkgGestor.TargetKindProvider,
		Scope:            pkgGestor.ScopeAdapter,
		RequireAvailable: true,
	})
	if _, err := resolver.Resolve(requireAvailable); !errors.Is(err, pkgGestor.ErrUnavailable) {
		t.Fatalf("unknown operational state: expected ErrUnavailable, got %v", err)
	}

	allowUnknown := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		TargetKind: pkgGestor.TargetKindProvider,
		Scope:      pkgGestor.ScopeAdapter,
	})
	resolution, err := resolver.Resolve(allowUnknown)
	if err != nil {
		t.Fatalf("resolve unknown declaration without operational requirement: %v", err)
	}
	if resolution.Descriptor() != unknown {
		t.Fatalf("expected unknown descriptor, got %#v", resolution.Descriptor())
	}
	if resolution.Reason() != pkgGestor.ResolutionSingleCandidate ||
		resolution.Snapshot().Generation != 1 {
		t.Fatalf("unexpected resolution explanation: reason=%q snapshot=%#v", resolution.Reason(), resolution.Snapshot())
	}
}

func TestResolverAppliesExactKindScopeAndModelFilters(t *testing.T) {
	descriptors := []pkgGestor.Descriptor{
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("ollama", pkgGestor.ScopeAdapter, ""), pkgGestor.AvailabilityUnknown),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("ollama", pkgGestor.ScopeInstance, ""), pkgGestor.AvailabilityAvailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("ollama", pkgGestor.ScopeModel, "llama3.1:8b"), pkgGestor.AvailabilityAvailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("ollama", pkgGestor.ScopeModel, "qwen2.5-coder:7b"), pkgGestor.AvailabilityAvailable),
	}
	resolver := newTestResolver(t, registryForResolver(t, descriptors), newDependencyGraphFixture())
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		TargetKind:       pkgGestor.TargetKindProvider,
		Scope:            pkgGestor.ScopeModel,
		Model:            "llama3.1:8b",
		RequireAvailable: true,
	})
	candidates, err := resolver.Candidates(query)
	if err != nil {
		t.Fatalf("filter candidates: %v", err)
	}
	if len(candidates) != 1 || candidates[0].Target.Model != "llama3.1:8b" {
		t.Fatalf("unexpected exact candidates: %#v", candidates)
	}
	candidates[0].Target.Model = "changed"
	again, err := resolver.Candidates(query)
	if err != nil {
		t.Fatalf("filter candidates again: %v", err)
	}
	if again[0].Target.Model != "llama3.1:8b" {
		t.Fatal("Candidates exposed internal descriptor storage")
	}
}

func TestResolverNeverUsesLexicalOrderAsImplicitRanking(t *testing.T) {
	descriptors := []pkgGestor.Descriptor{
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("zeta", pkgGestor.ScopeAdapter, ""), pkgGestor.AvailabilityAvailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, resolverProviderTarget("alpha", pkgGestor.ScopeAdapter, ""), pkgGestor.AvailabilityAvailable),
	}
	resolver := newTestResolver(t, registryForResolver(t, descriptors), newDependencyGraphFixture())
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})
	firstError := ""
	for range 5 {
		_, err := resolver.Resolve(query)
		if !errors.Is(err, pkgGestor.ErrAmbiguous) {
			t.Fatalf("expected ErrAmbiguous, got %v", err)
		}
		if firstError == "" {
			firstError = err.Error()
		} else if err.Error() != firstError {
			t.Fatalf("ambiguity diagnostics changed: %q != %q", err, firstError)
		}
	}
}

func TestResolverAppliesOrderedExplicitPreferences(t *testing.T) {
	alpha := resolverProviderTarget("alpha", pkgGestor.ScopeAdapter, "")
	beta := resolverProviderTarget("beta", pkgGestor.ScopeAdapter, "")
	gamma := resolverProviderTarget("gamma", pkgGestor.ScopeAdapter, "")
	descriptors := []pkgGestor.Descriptor{
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, alpha, pkgGestor.AvailabilityUnavailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, beta, pkgGestor.AvailabilityAvailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, gamma, pkgGestor.AvailabilityAvailable),
	}
	resolver := newTestResolver(t, registryForResolver(t, descriptors), newDependencyGraphFixture())
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		PreferredTargets: []pkgGestor.Target{alpha, gamma, beta},
	})
	resolution, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("resolve explicit preference: %v", err)
	}
	if resolution.Descriptor().Target != gamma ||
		resolution.Reason() != pkgGestor.ResolutionPreferredTarget {
		t.Fatalf("unexpected preferred resolution: %#v", resolution)
	}

	missingPreference := resolverProviderTarget("missing", pkgGestor.ScopeAdapter, "")
	query = newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
		PreferredTargets: []pkgGestor.Target{missingPreference},
	})
	if _, err := resolver.Resolve(query); !errors.Is(err, pkgGestor.ErrAmbiguous) {
		t.Fatalf("missing preference: expected ErrAmbiguous, got %v", err)
	}

	onlyBeta := registryForResolver(t, []pkgGestor.Descriptor{
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, alpha, pkgGestor.AvailabilityUnavailable),
		resolverDescriptor(pkgGestor.CapabilityProviderCompletion, beta, pkgGestor.AvailabilityAvailable),
	})
	resolution, err = newTestResolver(t, onlyBeta, newDependencyGraphFixture()).Resolve(
		newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{
			PreferredTargets: []pkgGestor.Target{alpha},
		}),
	)
	if err != nil {
		t.Fatalf("resolve sole eligible fallback: %v", err)
	}
	if resolution.Descriptor().Target != beta ||
		resolution.Reason() != pkgGestor.ResolutionSingleCandidate {
		t.Fatalf("unexpected sole eligible resolution: %#v", resolution)
	}
}

func TestResolverUsesDependencyEligibilityAndPlan(t *testing.T) {
	graph := newDependencyGraphFixture()
	graph.plans["agent"] = graphPlanFixture{
		generation:   1,
		dependencies: []pkgRuntime.ComponentID{"config", "provider"},
	}
	graph.plans["missing"] = graphPlanFixture{generation: 1, err: pkgRuntime.ErrNotFound}
	descriptors := []pkgGestor.Descriptor{
		resolverDescriptor(pkgGestor.CapabilityRuntimeStart, resolverComponentTarget("agent"), pkgGestor.AvailabilityUnknown),
		resolverDescriptor(pkgGestor.CapabilityRuntimeStart, resolverComponentTarget("missing"), pkgGestor.AvailabilityUnknown),
	}
	resolver := newTestResolver(t, registryForResolver(t, descriptors), graph)
	query := newQuery(t, pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{
		TargetKind: pkgGestor.TargetKindComponent,
	})
	resolution, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("resolve eligible component: %v", err)
	}
	if resolution.Descriptor().Target.ID != "agent" {
		t.Fatalf("expected agent, got %#v", resolution.Descriptor())
	}
	wantDependencies := []pkgGestor.Target{
		resolverComponentTarget("config"),
		resolverComponentTarget("provider"),
	}
	if !slices.Equal(resolution.Dependencies(), wantDependencies) {
		t.Fatalf("expected dependencies %#v, got %#v", wantDependencies, resolution.Dependencies())
	}

	onlyMissing := newTestResolver(t, registryForResolver(t, descriptors[1:]), graph)
	if _, err := onlyMissing.Resolve(query); !errors.Is(err, pkgGestor.ErrUnavailable) {
		t.Fatalf("component absent from graph: expected ErrUnavailable, got %v", err)
	}
}

func TestResolverRejectsStaleOrChangingGraphGenerations(t *testing.T) {
	descriptor := resolverDescriptor(
		pkgGestor.CapabilityRuntimeStart,
		resolverComponentTarget("agent"),
		pkgGestor.AvailabilityUnknown,
	)
	registry := registryForResolver(t, []pkgGestor.Descriptor{descriptor})
	query := newQuery(t, pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{})

	stale := newDependencyGraphFixture()
	stale.current = false
	if _, err := newTestResolver(t, registry, stale).Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("stale graph: expected ErrStaleSnapshot, got %v", err)
	}

	mismatched := newDependencyGraphFixture()
	mismatched.plans["agent"] = graphPlanFixture{generation: 2}
	if _, err := newTestResolver(t, registry, mismatched).Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("mismatched plan generation: expected ErrStaleSnapshot, got %v", err)
	}

	changing := newDependencyGraphFixture()
	changing.plans["agent"] = graphPlanFixture{generation: 1}
	changing.stateFunc = func(call int) (uint64, bool) {
		if call == 1 {
			return 1, true
		}
		return 2, true
	}
	if _, err := newTestResolver(t, registry, changing).Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("changing graph generation: expected ErrStaleSnapshot, got %v", err)
	}
}

func TestResolverRejectsSnapshotChangedDuringResolution(t *testing.T) {
	descriptor := resolverDescriptor(
		pkgGestor.CapabilityRuntimeStart,
		resolverComponentTarget("agent"),
		pkgGestor.AvailabilityUnknown,
	)
	registry := registryForResolver(t, []pkgGestor.Descriptor{descriptor})
	graph := newDependencyGraphFixture()
	graph.planFunc = func(pkgRuntime.ComponentID) (uint64, []pkgRuntime.ComponentID, error) {
		if err := registry.Refresh(t.Context()); err != nil {
			return 1, nil, err
		}
		return 1, nil, nil
	}
	query := newQuery(t, pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{})
	if _, err := newTestResolver(t, registry, graph).Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("changed snapshot: expected ErrStaleSnapshot, got %v", err)
	}
}

func TestResolverProviderCandidatesDoNotRequireCurrentGraph(t *testing.T) {
	descriptor := resolverDescriptor(
		pkgGestor.CapabilityProviderCompletion,
		resolverProviderTarget("provider", pkgGestor.ScopeAdapter, ""),
		pkgGestor.AvailabilityAvailable,
	)
	graph := newDependencyGraphFixture()
	graph.current = false
	resolver := newTestResolver(t, registryForResolver(t, []pkgGestor.Descriptor{descriptor}), graph)
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})
	if _, err := resolver.Resolve(query); err != nil {
		t.Fatalf("provider-only resolution unexpectedly required graph: %v", err)
	}
}

func TestResolverConcurrentResolutions(t *testing.T) {
	descriptor := resolverDescriptor(
		pkgGestor.CapabilityProviderCompletion,
		resolverProviderTarget("provider", pkgGestor.ScopeAdapter, ""),
		pkgGestor.AvailabilityAvailable,
	)
	resolver := newTestResolver(t, registryForResolver(t, []pkgGestor.Descriptor{descriptor}), newDependencyGraphFixture())
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})

	const workers = 24
	const iterations = 40
	var wait sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range iterations {
				resolution, err := resolver.Resolve(query)
				if err != nil {
					results <- err
					return
				}
				if resolution.Descriptor() != descriptor {
					results <- fmt.Errorf("unexpected descriptor %#v", resolution.Descriptor())
					return
				}
			}
			results <- nil
		}()
	}
	wait.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Errorf("concurrent resolution: %v", err)
		}
	}
}
