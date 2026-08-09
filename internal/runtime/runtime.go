package runtime

import (
	"context"
	"fmt"
	"sync"

	internalPlugin "github.com/antonio-cafeo/maestro/internal/plugin"
	internalProvider "github.com/antonio-cafeo/maestro/internal/provider"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgRuntime.Runtime = (*runtime)(nil)

type providerRegistrationInvalidator interface {
	SetRegistrationInvalidator(func())
}

// Runtime extends the Runtime Core contract with the services composed by the
// Maestro entry point.
type Runtime interface {
	pkgRuntime.Runtime
	Plugins() pkgPlugin.Runtime
}

var _ Runtime = (*runtime)(nil)

type runtime struct {
	mu sync.RWMutex

	registry *registry
	builder  *builder

	registryView      pkgRuntime.Registry
	eventBus          pkgRuntime.EventBus
	stateManager      *stateManager
	lifecycleManager  *lifecycleManager
	pluginRuntime     pkgPlugin.Runtime
	providerRuntime   pkgProvider.Runtime
	gestorInvalidator func()

	dependencyGraph          *graph
	componentGeneration      uint64
	graphGeneration          uint64
	graphComponentGeneration uint64

	started  bool
	starting bool
	stopping bool
}

func newRuntime() *runtime {
	return newRuntimeWithServices(nil, nil)
}

func newRuntimeWithServices(
	config pkgRuntime.Config,
	logger pkgRuntime.Logger,
) *runtime {
	if nilService(config) {
		config = newEmptyConfig()
	}

	if nilService(logger) {
		logger = newNoopLogger()
	}

	componentRegistry := newRegistry()
	dependencyResolver := newResolver(componentRegistry)
	graphValidator := newValidator()
	graphBuilder := newBuilder(
		dependencyResolver,
		graphValidator,
	)
	componentStates := newStateManager()
	componentEventBus := newEventBus()
	providerRuntime := internalProvider.NewRuntime(
		defaultProviderID(config),
	)

	rt := &runtime{
		registry:        componentRegistry,
		builder:         graphBuilder,
		eventBus:        componentEventBus,
		stateManager:    componentStates,
		providerRuntime: providerRuntime,
	}

	rt.registryView = newRuntimeRegistry(rt)
	rt.pluginRuntime = internalPlugin.NewRuntimeWithEventBus(
		rt.registryView,
		componentEventBus,
	)
	runtimeContext := newRuntimeContext(
		config,
		logger,
		componentEventBus,
		rt.registryView,
		providerRuntime,
	)
	rt.lifecycleManager = newLifecycleManager(
		componentStates,
		runtimeContext,
	)

	return rt
}

// New is the internal composition entry point used by the public maestro
// package. Concrete services remain hidden behind the public contracts.
func New(
	config pkgRuntime.Config,
	logger pkgRuntime.Logger,
) Runtime {
	return newRuntimeWithServices(config, logger)
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

	if r.started || r.starting {
		r.mu.Unlock()
		return fmt.Errorf(
			"register component: %w",
			pkgRuntime.ErrAlreadyStarted,
		)
	}

	if r.stopping {
		r.mu.Unlock()
		return fmt.Errorf(
			"register component while runtime is stopping: %w",
			pkgRuntime.ErrInvalidState,
		)
	}

	if err := r.registry.Register(component); err != nil {
		r.mu.Unlock()
		return fmt.Errorf(
			"register component: %w",
			err,
		)
	}

	if err := r.stateManager.create(component); err != nil {
		r.mu.Unlock()
		return fmt.Errorf(
			"create component state: %w",
			err,
		)
	}

	// Qualsiasi nuova registrazione rende obsoleto
	// l'eventuale grafo costruito in precedenza.
	r.componentGeneration++
	r.dependencyGraph = nil
	invalidator := r.gestorInvalidator
	r.mu.Unlock()

	if invalidator != nil {
		invalidator()
	}

	return nil
}

func (r *runtime) setGestorInvalidator(invalidator func()) {
	r.mu.Lock()
	r.gestorInvalidator = invalidator
	providerRuntime := r.providerRuntime
	r.mu.Unlock()

	if provider, ok := providerRuntime.(providerRegistrationInvalidator); ok {
		provider.SetRegistrationInvalidator(invalidator)
	}
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

	providerErr := r.providerRuntime.Shutdown(ctx)

	r.mu.Lock()
	r.stopping = false
	r.started = false
	r.mu.Unlock()

	if providerErr != nil {
		return fmt.Errorf(
			"stop runtime: shutdown providers: %w",
			providerErr,
		)
	}

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

func (r *runtime) Providers() pkgProvider.Runtime {
	return r.providerRuntime
}

func (r *runtime) Plugins() pkgPlugin.Runtime {
	return r.pluginRuntime
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
	r.graphGeneration++
	r.graphComponentGeneration = r.componentGeneration

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

func defaultProviderID(config pkgRuntime.Config) pkgProvider.ID {
	value := config.Get(pkgProvider.ConfigDefaultProvider)

	switch configured := value.(type) {
	case pkgProvider.ID:
		return configured
	case string:
		return pkgProvider.ID(configured)
	default:
		return ""
	}
}

func (r *runtime) ready() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.dependencyGraph != nil
}
