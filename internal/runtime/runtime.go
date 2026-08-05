package runtime

import (
	"context"
	"fmt"
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgRuntime.Runtime = (*runtime)(nil)

type runtime struct {
	mu sync.RWMutex

	registry *registry
	builder  *builder

	registryView     pkgRuntime.Registry
	eventBus         pkgRuntime.EventBus
	stateManager     *stateManager
	lifecycleManager *lifecycleManager

	dependencyGraph *graph

	started  bool
	starting bool
	stopping bool
}

func newRuntime() *runtime {
	componentRegistry := newRegistry()
	dependencyResolver := newResolver(componentRegistry)
	graphValidator := newValidator()
	graphBuilder := newBuilder(
		dependencyResolver,
		graphValidator,
	)
	componentStates := newStateManager()
	componentEventBus := newEventBus()

	rt := &runtime{
		registry:     componentRegistry,
		builder:      graphBuilder,
		eventBus:     componentEventBus,
		stateManager: componentStates,
	}

	rt.registryView = newRuntimeRegistry(rt)
	runtimeContext := newRuntimeContext(
		newEmptyConfig(),
		newNoopLogger(),
		componentEventBus,
		rt.registryView,
	)
	rt.lifecycleManager = newLifecycleManager(
		componentStates,
		runtimeContext,
	)

	return rt
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

	if r.started || r.starting {
		return fmt.Errorf(
			"register component: %w",
			pkgRuntime.ErrAlreadyStarted,
		)
	}

	if r.stopping {
		return fmt.Errorf(
			"register component while runtime is stopping: %w",
			pkgRuntime.ErrInvalidState,
		)
	}

	if err := r.registry.Register(component); err != nil {
		return fmt.Errorf(
			"register component: %w",
			err,
		)
	}

	if err := r.stateManager.create(component); err != nil {
		return fmt.Errorf(
			"create component state: %w",
			err,
		)
	}

	// Qualsiasi nuova registrazione rende obsoleto
	// l'eventuale grafo costruito in precedenza.
	r.dependencyGraph = nil

	return nil
}

func (r *runtime) Start(ctx context.Context) error {
	r.mu.Lock()

	if r.started || r.starting {
		r.mu.Unlock()

		return fmt.Errorf(
			"start runtime: %w",
			pkgRuntime.ErrAlreadyStarted,
		)
	}

	if r.stopping {
		r.mu.Unlock()

		return fmt.Errorf(
			"start runtime while stopping: %w",
			pkgRuntime.ErrInvalidState,
		)
	}

	r.starting = true

	if r.dependencyGraph == nil {
		if err := r.buildDependencyGraphLocked(); err != nil {
			r.starting = false
			r.mu.Unlock()

			return err
		}
	}

	orderedNodes, err := r.dependencyGraph.TopologicalOrder()
	if err != nil {
		r.starting = false
		r.mu.Unlock()

		return fmt.Errorf(
			"start runtime: order dependencies: %w",
			err,
		)
	}

	r.mu.Unlock()

	for _, currentNode := range orderedNodes {
		if err := r.lifecycleManager.Start(
			ctx,
			currentNode.Component(),
		); err != nil {
			r.mu.Lock()
			r.starting = false
			r.mu.Unlock()

			return fmt.Errorf(
				"start runtime: %w",
				err,
			)
		}
	}

	r.mu.Lock()
	r.starting = false
	r.started = true
	r.mu.Unlock()

	return nil
}

func (r *runtime) Stop(ctx context.Context) error {
	r.mu.Lock()

	if !r.started {
		r.mu.Unlock()

		return fmt.Errorf(
			"stop runtime: %w",
			pkgRuntime.ErrAlreadyStopped,
		)
	}

	if r.starting || r.stopping {
		r.mu.Unlock()

		return fmt.Errorf(
			"stop runtime from current state: %w",
			pkgRuntime.ErrInvalidState,
		)
	}

	r.stopping = true

	orderedNodes, err := r.dependencyGraph.ReverseTopologicalOrder()
	if err != nil {
		r.stopping = false
		r.mu.Unlock()

		return fmt.Errorf(
			"stop runtime: order dependencies: %w",
			err,
		)
	}

	r.mu.Unlock()

	for _, currentNode := range orderedNodes {
		if err := r.lifecycleManager.Stop(
			ctx,
			currentNode.Component(),
		); err != nil {
			r.mu.Lock()
			r.stopping = false
			r.mu.Unlock()

			return fmt.Errorf(
				"stop runtime: %w",
				err,
			)
		}
	}

	r.mu.Lock()
	r.stopping = false
	r.started = false
	r.mu.Unlock()

	return nil
}

func (r *runtime) Registry() pkgRuntime.Registry {
	return r.registryView
}

func (r *runtime) EventBus() pkgRuntime.EventBus {
	return r.eventBus
}

func (r *runtime) StateManager() pkgRuntime.StateManager {
	return r.stateManager
}

func (r *runtime) buildDependencyGraph() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.buildDependencyGraphLocked()
}

func (r *runtime) buildDependencyGraphLocked() error {
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
