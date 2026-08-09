package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	internalGestor "github.com/antonio-cafeo/maestro/internal/gestor"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type runtimeGestorPlugin struct {
	metadata pkgRuntime.Metadata
}

func (plugin *runtimeGestorPlugin) Metadata() pkgRuntime.Metadata {
	return plugin.metadata
}

func (plugin *runtimeGestorPlugin) Manifest() pkgPlugin.Manifest {
	return pkgPlugin.Manifest{RuntimeAPIVersion: pkgPlugin.RuntimeAPIVersion}
}

type runtimeResidencyProvider struct {
	unloads int
}

func (*runtimeResidencyProvider) ID() pkgProvider.ID {
	return "residency"
}

func (*runtimeResidencyProvider) DiscoverModels(
	context.Context,
) ([]pkgProvider.ModelInfo, error) {
	return []pkgProvider.ModelInfo{{
		Model: pkgProvider.Model{ID: "qwen"},
		State: pkgProvider.ModelStateAvailable,
	}}, nil
}

func (*runtimeResidencyProvider) LoadModel(
	context.Context,
	pkgProvider.ModelLoadRequest,
) error {
	return nil
}

func (p *runtimeResidencyProvider) UnloadModel(
	context.Context,
	pkgProvider.ModelUnloadRequest,
) error {
	p.unloads++

	return nil
}

func (*runtimeResidencyProvider) Complete(
	context.Context,
	pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	return pkgProvider.CompletionResponse{}, nil
}

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

func TestRuntimeAndPluginRegistrationInvalidateGestorAndShareComponentSource(t *testing.T) {
	rt := newRuntime()
	gestorRegistry := internalGestor.NewRegistry()
	if err := gestorRegistry.Refresh(context.Background()); err != nil {
		t.Fatalf("initial Gestor refresh: %v", err)
	}
	rt.setGestorInvalidator(func() {
		// Re-entering the Runtime proves invalidation happens after its lock is released.
		_ = rt.ready()
		gestorRegistry.Invalidate()
	})

	registered := make(chan error, 1)
	go func() {
		registered <- rt.Register(newTestComponent("component"))
	}()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatalf("register component: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("component invalidation callback appears to run under Runtime lock")
	}
	if gestorRegistry.Snapshot().Metadata().Current {
		t.Fatal("component registration did not invalidate Gestor")
	}

	if err := gestorRegistry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh Gestor before plugin: %v", err)
	}
	plugin := &runtimeGestorPlugin{metadata: pkgRuntime.Metadata{
		ID:           "laravel",
		Name:         "Laravel",
		Version:      "1.0.0",
		Capabilities: []pkgRuntime.Capability{"plugin.workspace_detection"},
	}}
	if err := rt.Plugins().Register(plugin); err != nil {
		t.Fatalf("register plugin: %v", err)
	}
	if gestorRegistry.Snapshot().Metadata().Current {
		t.Fatal("plugin registration did not invalidate Gestor through the component registry")
	}

	componentSource, err := internalGestor.NewRuntimeComponentSource(rt.registry)
	if err != nil {
		t.Fatalf("new component source: %v", err)
	}
	descriptors, err := componentSource.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover Runtime components: %v", err)
	}
	pluginCount := 0
	for _, descriptor := range descriptors {
		if descriptor.Capability == pkgGestor.CapabilityID("plugin.workspace_detection") &&
			descriptor.Target.ID == "laravel" {
			pluginCount++
		}
	}
	if pluginCount != 1 {
		t.Fatalf("expected plugin once through global component source, got %d", pluginCount)
	}

	if err := gestorRegistry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh Gestor before provider: %v", err)
	}
	if err := rt.Providers().Register(&runtimeResidencyProvider{}); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	if gestorRegistry.Snapshot().Metadata().Current {
		t.Fatal("provider registration did not invalidate Gestor")
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

func TestRuntimeStartsComponentsInTopologicalOrder(t *testing.T) {
	rt := newRuntime()
	calls := make([]string, 0)

	config := newLifecycleTestComponent("config", &calls)
	provider := newLifecycleTestComponent(
		"provider",
		&calls,
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)
	agent := newLifecycleTestComponent(
		"agent",
		&calls,
		pkgRuntime.Dependency{
			ID:       "provider",
			Required: true,
		},
	)

	for _, component := range []pkgRuntime.Component{
		agent,
		provider,
		config,
	} {
		if err := rt.Register(component); err != nil {
			t.Fatalf("register component: %v", err)
		}
	}

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	if got, want := calls, []string{
		"config:configure",
		"config:initialize",
		"config:start",
		"provider:configure",
		"provider:initialize",
		"provider:start",
		"agent:configure",
		"agent:initialize",
		"agent:start",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}

	if componentState := rt.StateManager().Get(agent); componentState.State != pkgRuntime.StateRunning {
		t.Fatalf(
			"expected agent running, got %d",
			componentState.State,
		)
	}
}

func TestRuntimeStopsComponentsInReverseTopologicalOrder(t *testing.T) {
	rt := newRuntime()
	calls := make([]string, 0)

	config := newLifecycleTestComponent("config", &calls)
	provider := newLifecycleTestComponent(
		"provider",
		&calls,
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)
	agent := newLifecycleTestComponent(
		"agent",
		&calls,
		pkgRuntime.Dependency{
			ID:       "provider",
			Required: true,
		},
	)

	for _, component := range []pkgRuntime.Component{
		config,
		provider,
		agent,
	} {
		if err := rt.Register(component); err != nil {
			t.Fatalf("register component: %v", err)
		}
	}

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	calls = calls[:0]

	if err := rt.Stop(context.Background()); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}

	if got, want := calls, []string{
		"agent:stop",
		"provider:stop",
		"config:stop",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}

	if componentState := rt.StateManager().Get(config); componentState.State != pkgRuntime.StateStopped {
		t.Fatalf(
			"expected config stopped, got %d",
			componentState.State,
		)
	}
}

func TestRuntimeRejectsDuplicateStart(t *testing.T) {
	rt := newRuntime()

	if err := rt.Register(
		newTestComponent("config"),
	); err != nil {
		t.Fatalf("register component: %v", err)
	}

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	err := rt.Start(context.Background())
	if !errors.Is(err, pkgRuntime.ErrAlreadyStarted) {
		t.Fatalf(
			"expected ErrAlreadyStarted, got %v",
			err,
		)
	}
}

func TestRuntimeRejectsStopBeforeStart(t *testing.T) {
	rt := newRuntime()

	err := rt.Stop(context.Background())
	if !errors.Is(err, pkgRuntime.ErrAlreadyStopped) {
		t.Fatalf(
			"expected ErrAlreadyStopped, got %v",
			err,
		)
	}
}

func TestRuntimeStartFailureMarksComponentFailed(t *testing.T) {
	rt := newRuntime()
	calls := make([]string, 0)

	config := newLifecycleTestComponent("config", &calls)
	provider := newLifecycleTestComponent(
		"provider",
		&calls,
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
	)
	provider.failOn = "start"

	if err := rt.Register(config); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := rt.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	err := rt.Start(context.Background())
	if err == nil {
		t.Fatal("expected runtime start error")
	}

	componentState := rt.StateManager().Get(provider)
	if componentState.State != pkgRuntime.StateFailed {
		t.Fatalf(
			"expected provider failed, got %d",
			componentState.State,
		)
	}
}

func TestRuntimeRejectsRegistrationAfterStart(t *testing.T) {
	rt := newRuntime()

	if err := rt.Register(
		newTestComponent("config"),
	); err != nil {
		t.Fatalf("register config: %v", err)
	}

	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	err := rt.Register(newTestComponent("provider"))
	if !errors.Is(err, pkgRuntime.ErrAlreadyStarted) {
		t.Fatalf(
			"expected ErrAlreadyStarted, got %v",
			err,
		)
	}
}

func TestRuntimeStopShutsDownProviderResidencies(t *testing.T) {
	rt := newRuntime()
	provider := &runtimeResidencyProvider{}
	if err := rt.Providers().Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	if err := rt.Providers().SetModelResidencyPolicy(
		context.Background(),
		"residency",
		pkgProvider.ModelResidencyPolicy{
			Model: "qwen", Autoload: true, Persistent: true,
		},
	); err != nil {
		t.Fatalf("set residency policy: %v", err)
	}
	if _, err := rt.Providers().Complete(
		context.Background(),
		"residency",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	if err := rt.Stop(context.Background()); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}
	if provider.unloads != 1 {
		t.Fatalf("expected one provider unload at shutdown, got %d", provider.unloads)
	}
}
