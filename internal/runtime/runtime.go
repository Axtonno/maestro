package runtime

import (
	"fmt"
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type runtime struct {
	mu sync.RWMutex

	registry *registry
	builder  *builder

	dependencyGraph *graph
}

func newRuntime() *runtime {
	componentRegistry := newRegistry()
	dependencyResolver := newResolver(componentRegistry)
	graphValidator := newValidator()
	graphBuilder := newBuilder(
		dependencyResolver,
		graphValidator,
	)

	return &runtime{
		registry: componentRegistry,
		builder:  graphBuilder,
	}
}

func (r *runtime) Register(
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"register component: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.registry.Register(component); err != nil {
		return fmt.Errorf(
			"register component: %w",
			err,
		)
	}

	// Qualsiasi nuova registrazione rende obsoleto
	// l'eventuale grafo costruito in precedenza.
	r.dependencyGraph = nil

	return nil
}

func (r *runtime) buildDependencyGraph() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dependencyGraph, err := r.builder.Build()
	if err != nil {
		return fmt.Errorf(
			"prepare runtime dependencies: %w",
			err,
		)
	}

	r.dependencyGraph = dependencyGraph

	return nil
}

func (r *runtime) graph() (*graph, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.dependencyGraph == nil {
		return nil, false
	}

	return r.dependencyGraph, true
}

func (r *runtime) ready() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.dependencyGraph != nil
}