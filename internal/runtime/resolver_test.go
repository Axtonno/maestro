package runtime

import (
	"errors"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestResolverBuildsDependencyGraph(t *testing.T) {
	componentRegistry := newRegistry()

	config := newTestComponent("config")

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)

	if err := componentRegistry.Register(config); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := componentRegistry.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	dependencyResolver := newResolver(componentRegistry)

	dependencyGraph, err := dependencyResolver.Resolve()
	if err != nil {
		t.Fatalf("resolve dependencies: %v", err)
	}

	if dependencyGraph.Len() != 2 {
		t.Fatalf(
			"expected 2 graph nodes, got %d",
			dependencyGraph.Len(),
		)
	}

	providerNode, exists := dependencyGraph.Node("provider")
	if !exists {
		t.Fatal("provider node not found")
	}

	if len(providerNode.Dependencies()) != 1 {
		t.Fatalf(
			"expected 1 provider dependency, got %d",
			len(providerNode.Dependencies()),
		)
	}

	dependencyID := providerNode.
		Dependencies()[0].
		Component().
		Metadata().
		ID

	if dependencyID != "config" {
		t.Fatalf(
			"expected config dependency, got %q",
			dependencyID,
		)
	}
}

func TestResolverRejectsMissingRequiredDependency(t *testing.T) {
	componentRegistry := newRegistry()

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)

	if err := componentRegistry.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	dependencyResolver := newResolver(componentRegistry)

	_, err := dependencyResolver.Resolve()
	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestResolverIgnoresMissingOptionalDependency(t *testing.T) {
	componentRegistry := newRegistry()

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "metrics",
			Required: false,
		},
	)

	if err := componentRegistry.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	dependencyResolver := newResolver(componentRegistry)

	dependencyGraph, err := dependencyResolver.Resolve()
	if err != nil {
		t.Fatalf("resolve optional dependency: %v", err)
	}

	providerNode, exists := dependencyGraph.Node("provider")
	if !exists {
		t.Fatal("provider node not found")
	}

	if len(providerNode.Dependencies()) != 0 {
		t.Fatal("missing optional dependency was added to graph")
	}
}

func TestResolverAddsPresentOptionalDependency(t *testing.T) {
	componentRegistry := newRegistry()

	metrics := newTestComponent("metrics")

	provider := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "metrics",
			Required: false,
		},
	)

	if err := componentRegistry.Register(metrics); err != nil {
		t.Fatalf("register metrics: %v", err)
	}

	if err := componentRegistry.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	dependencyResolver := newResolver(componentRegistry)

	dependencyGraph, err := dependencyResolver.Resolve()
	if err != nil {
		t.Fatalf("resolve optional dependency: %v", err)
	}

	providerNode, _ := dependencyGraph.Node("provider")

	if len(providerNode.Dependencies()) != 1 {
		t.Fatalf(
			"expected 1 dependency, got %d",
			len(providerNode.Dependencies()),
		)
	}
}
