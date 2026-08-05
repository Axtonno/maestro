package runtime

import (
	"errors"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestRuntimeRegisterComponent(t *testing.T) {
	rt := newRuntime()
	component := newTestComponent("config")

	if err := rt.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	if !rt.registry.Has("config") {
		t.Fatal("runtime registry does not contain component")
	}
}

func TestRuntimeRejectsNilComponent(t *testing.T) {
	rt := newRuntime()

	err := rt.Register(nil)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestRuntimeRejectsDuplicateComponent(t *testing.T) {
	rt := newRuntime()
	component := newTestComponent("config")

	if err := rt.Register(component); err != nil {
		t.Fatalf("register first component: %v", err)
	}

	err := rt.Register(component)
	if !errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
		t.Fatalf(
			"expected ErrAlreadyRegistered, got %v",
			err,
		)
	}
}

func TestRuntimeBuildDependencyGraph(t *testing.T) {
	rt := newRuntime()

	config := newTestComponent("config")

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)

	if err := rt.Register(config); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := rt.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}

	if !rt.ready() {
		t.Fatal("runtime is not ready after graph build")
	}

	dependencyGraph, exists := rt.graph()
	if !exists {
		t.Fatal("runtime graph is not available")
	}

	if dependencyGraph.Len() != 2 {
		t.Fatalf(
			"expected 2 graph nodes, got %d",
			dependencyGraph.Len(),
		)
	}
}

func TestRuntimeInvalidatesGraphAfterRegistration(t *testing.T) {
	rt := newRuntime()

	if err := rt.Register(
		newTestComponent("config"),
	); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build graph: %v", err)
	}

	if !rt.ready() {
		t.Fatal("runtime should be ready")
	}

	if err := rt.Register(
		newTestComponent("provider"),
	); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	if rt.ready() {
		t.Fatal("runtime graph was not invalidated")
	}

	if dependencyGraph, exists := rt.graph(); exists ||
		dependencyGraph != nil {
		t.Fatal("runtime returned an invalidated graph")
	}
}

func TestRuntimeDoesNotStoreInvalidGraph(t *testing.T) {
	rt := newRuntime()

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)

	if err := rt.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	err := rt.buildDependencyGraph()
	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}

	if rt.ready() {
		t.Fatal("runtime is ready after failed graph build")
	}

	if dependencyGraph, exists := rt.graph(); exists ||
		dependencyGraph != nil {
		t.Fatal("runtime stored invalid dependency graph")
	}
}

func TestRuntimeCanRebuildInvalidatedGraph(t *testing.T) {
	rt := newRuntime()

	config := newTestComponent("config")

	if err := rt.Register(config); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("build initial graph: %v", err)
	}

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)

	if err := rt.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	if err := rt.buildDependencyGraph(); err != nil {
		t.Fatalf("rebuild dependency graph: %v", err)
	}

	dependencyGraph, exists := rt.graph()
	if !exists {
		t.Fatal("rebuilt graph is not available")
	}

	if dependencyGraph.Len() != 2 {
		t.Fatalf(
			"expected rebuilt graph length 2, got %d",
			dependencyGraph.Len(),
		)
	}
}