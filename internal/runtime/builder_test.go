package runtime

import (
	"errors"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestBuilderBuildsValidGraph(t *testing.T) {
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
	graphValidator := newValidator()

	graphBuilder := newBuilder(
		dependencyResolver,
		graphValidator,
	)

	dependencyGraph, err := graphBuilder.Build()
	if err != nil {
		t.Fatalf("build graph: %v", err)
	}

	if dependencyGraph == nil {
		t.Fatal("builder returned nil graph")
	}

	if dependencyGraph.Len() != 2 {
		t.Fatalf(
			"expected 2 nodes, got %d",
			dependencyGraph.Len(),
		)
	}
}

func TestBuilderPropagatesResolverError(t *testing.T) {
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

	graphBuilder := newBuilder(
		newResolver(componentRegistry),
		newValidator(),
	)

	_, err := graphBuilder.Build()
	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestBuilderPropagatesValidatorError(t *testing.T) {
	componentRegistry := newRegistry()

	first := newTestComponent(
		"first",
		pkgRuntime.Dependency{
			ID:       "second",
			Required: true,
		},
	)

	second := newTestComponent(
		"second",
		pkgRuntime.Dependency{
			ID:       "first",
			Required: true,
		},
	)

	if err := componentRegistry.Register(first); err != nil {
		t.Fatalf("register first: %v", err)
	}

	if err := componentRegistry.Register(second); err != nil {
		t.Fatalf("register second: %v", err)
	}

	graphBuilder := newBuilder(
		newResolver(componentRegistry),
		newValidator(),
	)

	_, err := graphBuilder.Build()
	if !errors.Is(err, pkgRuntime.ErrCyclicDependency) {
		t.Fatalf(
			"expected ErrCyclicDependency, got %v",
			err,
		)
	}
}