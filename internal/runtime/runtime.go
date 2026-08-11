package runtime

import (
	"context"
	"fmt"
	"sync"

	internalContext "github.com/antonio-cafeo/maestro/internal/contextengine"
	internalGestor "github.com/antonio-cafeo/maestro/internal/gestor"
	internalPlugin "github.com/antonio-cafeo/maestro/internal/plugin"
	internalProvider "github.com/antonio-cafeo/maestro/internal/provider"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgRuntime.Runtime = (*runtime)(nil)

type providerRegistrationInvalidator interface {
	SetRegistrationInvalidator(func())
}

type gestorProviderRuntime interface {
	Registered() []pkgProvider.ID
	Capabilities(
		context.Context,
		pkgProvider.ID,
		pkgProvider.CapabilityRequest,
	) (pkgProvider.CapabilityReport, error)
}

// Runtime extends the Runtime Core contract with the services composed by the
// Maestro entry point.
type Runtime interface {
	pkgRuntime.Runtime
	ContextEngine() pkgContext.Engine
	Gestor() pkgGestor.Service
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
	gestorService     pkgGestor.Service
	contextEngine     pkgContext.Engine
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
	contextEngine, err := internalContext.NewWithOptions(internalContext.Options{
		Embedding: providerRuntime,
		Cache:     pkgContext.DefaultCachePolicy(),
		EventBus:  componentEventBus,
	})
	if err != nil {
		panic(fmt.Sprintf("compose Context Engine: %v", err))
	}
	rt.contextEngine = contextEngine

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
	rt.composeGestor()

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

	if r.stopping {
		r.mu.Unlock()
		return fmt.Errorf(
			"register component while runtime is stopping: %w",
			pkgRuntime.ErrInvalidState,
		)
	}

	if r.started || r.starting {
		r.mu.Unlock()
		return fmt.Errorf(
			"register component: %w",
			pkgRuntime.ErrAlreadyStarted,
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

func (r *runtime) Gestor() pkgGestor.Service {
	return r.gestorService
}

func (r *runtime) ContextEngine() pkgContext.Engine {
	return r.contextEngine
}

func (r *runtime) composeGestor() {
	providerRuntime, ok := r.providerRuntime.(gestorProviderRuntime)
	if !ok {
		panic("compose Gestor: Provider Runtime lacks internal discovery capabilities")
	}

	componentSource, err := internalGestor.NewRuntimeComponentSource(r.registry)
	if err != nil {
		panic(fmt.Sprintf("compose Gestor component source: %v", err))
	}
	// The built-in policy probes adapter and instance targets. Model targets are
	// never inferred; applications can add exact declarations through the
	// public Source extension point.
	providerSource, err := internalGestor.NewProviderCapabilitySource(
		providerRuntime,
		nil,
	)
	if err != nil {
		panic(fmt.Sprintf("compose Gestor provider source: %v", err))
	}

	registry := internalGestor.NewRegistry()
	if err := registry.RegisterSource(componentSource); err != nil {
		panic(fmt.Sprintf("compose Gestor component source registration: %v", err))
	}
	if err := registry.RegisterSource(providerSource); err != nil {
		panic(fmt.Sprintf("compose Gestor provider source registration: %v", err))
	}

	resolver, err := internalGestor.NewResolver(registry, newGestorGraphView(r))
	if err != nil {
		panic(fmt.Sprintf("compose Gestor resolver: %v", err))
	}
	service, err := internalGestor.NewService(registry, resolver, r.eventBus)
	if err != nil {
		panic(fmt.Sprintf("compose Gestor service: %v", err))
	}
	if err := service.Refresh(context.Background()); err != nil {
		panic(fmt.Sprintf("compose Gestor initial refresh: %v", err))
	}

	r.gestorService = service
	r.setGestorInvalidator(registry.Invalidate)
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
