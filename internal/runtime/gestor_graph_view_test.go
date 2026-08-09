package runtime

import (
	"errors"
	"slices"
	"sync/atomic"
	"testing"

	internalGestor "github.com/antonio-cafeo/maestro/internal/gestor"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type metadataCountingComponent struct {
	metadata pkgRuntime.Metadata
	calls    atomic.Int64
}

func (component *metadataCountingComponent) Metadata() pkgRuntime.Metadata {
	component.calls.Add(1)

	return component.metadata
}

func TestGestorGraphViewUsesRuntimeGraphAndGeneration(t *testing.T) {
	rt := newRuntime()
	config := newTestComponent("config")
	provider := newTestComponent("provider", pkgRuntime.Dependency{
		ID: "config", Required: true,
	})
	agent := newTestComponent(
		"agent",
		pkgRuntime.Dependency{ID: "provider", Required: true},
		pkgRuntime.Dependency{ID: "missing-optional", Required: false},
	)
	for _, component := range []pkgRuntime.Component{agent, provider, config} {
		if err := rt.Register(component); err != nil {
			t.Fatalf("register component %q: %v", component.Metadata().ID, err)
		}
	}
	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}

	view := newGestorGraphView(rt)
	generation, current := view.State()
	if generation != 1 || !current {
		t.Fatalf("expected current graph generation 1, got %d current=%t", generation, current)
	}
	dependencyGraph, exists := rt.graph()
	if !exists {
		t.Fatal("Runtime graph is absent")
	}
	beforeNodes := dependencyGraph.Len()
	planGeneration, plan, err := view.DependencyPlan("agent")
	if err != nil {
		t.Fatalf("dependency plan: %v", err)
	}
	if planGeneration != generation {
		t.Fatalf("expected plan generation %d, got %d", generation, planGeneration)
	}
	want := []pkgRuntime.ComponentID{"config", "provider"}
	if !slices.Equal(plan, want) {
		t.Fatalf("expected dependency plan %v, got %v", want, plan)
	}
	if dependencyGraph.Len() != beforeNodes {
		t.Fatal("Gestor graph view modified the Runtime graph")
	}
	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("rebuild unchanged dependency graph: %v", err)
	}
	generation, current = view.State()
	if generation != 2 || !current {
		t.Fatalf("expected current rebuilt graph generation 2, got %d current=%t", generation, current)
	}

	if err := rt.Register(newTestComponent("later")); err != nil {
		t.Fatalf("register later component: %v", err)
	}
	generation, current = view.State()
	if generation != 2 || current {
		t.Fatalf("expected stale graph generation 2, got %d current=%t", generation, current)
	}
	if _, _, err := view.DependencyPlan("agent"); !errors.Is(err, pkgRuntime.ErrInvalidState) {
		t.Fatalf("stale dependency plan: expected ErrInvalidState, got %v", err)
	}
}

func TestGestorResolverUsesUniqueRuntimeGraphWithOptionalDependencies(t *testing.T) {
	rt := newRuntime()
	config := newTestComponent("config")
	provider := newTestComponent("provider", pkgRuntime.Dependency{
		ID: "config", Required: true,
	})
	agent := newTestComponent(
		"agent",
		pkgRuntime.Dependency{ID: "provider", Required: true},
		pkgRuntime.Dependency{ID: "missing-optional", Required: false},
	)
	agent.metadata.Capabilities = []pkgRuntime.Capability{pkgRuntime.CapabilityStart}
	for _, component := range []pkgRuntime.Component{agent, provider, config} {
		if err := rt.Register(component); err != nil {
			t.Fatalf("register component %q: %v", component.Metadata().ID, err)
		}
	}
	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}

	registry := internalGestor.NewRegistry()
	source, err := internalGestor.NewRuntimeComponentSource(rt.registry)
	if err != nil {
		t.Fatalf("new Runtime component source: %v", err)
	}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register Runtime component source: %v", err)
	}
	if err := registry.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh Gestor: %v", err)
	}
	resolver, err := internalGestor.NewResolver(registry, newGestorGraphView(rt))
	if err != nil {
		t.Fatalf("new Gestor resolver: %v", err)
	}
	query, err := pkgGestor.NewQuery(pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{
		TargetKind: pkgGestor.TargetKindComponent,
	})
	if err != nil {
		t.Fatalf("new query: %v", err)
	}
	resolution, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("resolve Runtime capability: %v", err)
	}
	wantPlan := []pkgGestor.Target{
		{Kind: pkgGestor.TargetKindComponent, ID: "config", Scope: pkgGestor.ScopeComponent},
		{Kind: pkgGestor.TargetKindComponent, ID: "provider", Scope: pkgGestor.ScopeComponent},
	}
	if !slices.Equal(resolution.Dependencies(), wantPlan) {
		t.Fatalf("expected dependency plan %#v, got %#v", wantPlan, resolution.Dependencies())
	}
	if resolution.Descriptor().Target.ID != "agent" {
		t.Fatalf("expected agent resolution, got %#v", resolution.Descriptor())
	}
}

func TestGestorResolverRejectsGraphWithMissingRequiredDependency(t *testing.T) {
	rt := newRuntime()
	agent := newTestComponent("agent", pkgRuntime.Dependency{
		ID: "missing-required", Required: true,
	})
	agent.metadata.Capabilities = []pkgRuntime.Capability{pkgRuntime.CapabilityStart}
	if err := rt.Register(agent); err != nil {
		t.Fatalf("register agent: %v", err)
	}
	if err := rt.buildDependencyGraph(); !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf("build graph: expected ErrNotFound, got %v", err)
	}

	registry := internalGestor.NewRegistry()
	source, err := internalGestor.NewRuntimeComponentSource(rt.registry)
	if err != nil {
		t.Fatalf("new Runtime component source: %v", err)
	}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := registry.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh Gestor: %v", err)
	}
	resolver, err := internalGestor.NewResolver(registry, newGestorGraphView(rt))
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	query, err := pkgGestor.NewQuery(pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{})
	if err != nil {
		t.Fatalf("new query: %v", err)
	}
	if _, err := resolver.Resolve(query); !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("resolve with invalid graph: expected ErrStaleSnapshot, got %v", err)
	}
}

func TestGestorResolverDoesNotExecuteSelectedComponent(t *testing.T) {
	rt := newRuntime()
	component := &metadataCountingComponent{metadata: pkgRuntime.Metadata{
		ID:           "agent",
		Name:         "Agent",
		Version:      "1.0.0",
		Capabilities: []pkgRuntime.Capability{pkgRuntime.CapabilityStart},
	}}
	if err := rt.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}
	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}
	registry := internalGestor.NewRegistry()
	source, err := internalGestor.NewRuntimeComponentSource(rt.registry)
	if err != nil {
		t.Fatalf("new component source: %v", err)
	}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := registry.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh Gestor: %v", err)
	}
	resolver, err := internalGestor.NewResolver(registry, newGestorGraphView(rt))
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	query, err := pkgGestor.NewQuery(pkgGestor.CapabilityRuntimeStart, pkgGestor.QueryOptions{})
	if err != nil {
		t.Fatalf("new query: %v", err)
	}
	before := component.calls.Load()
	if _, err := resolver.Resolve(query); err != nil {
		t.Fatalf("resolve capability: %v", err)
	}
	if after := component.calls.Load(); after != before {
		t.Fatalf("Resolve executed component Metadata: before %d after %d", before, after)
	}
}
