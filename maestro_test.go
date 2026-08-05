package maestro

import (
	"context"
	"errors"
	"reflect"
	"testing"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type testProvider struct {
	id pkgProvider.ID
}

type contextComponent struct {
	context pkgRuntime.Context
}

type lifecyclePlugin struct {
	metadata pkgRuntime.Metadata
	calls    *[]string
	context  pkgRuntime.Context
}

func newLifecyclePlugin(
	id pkgPlugin.ID,
	calls *[]string,
	dependencies ...pkgRuntime.Dependency,
) *lifecyclePlugin {
	return &lifecyclePlugin{
		metadata: pkgRuntime.Metadata{
			ID:           id,
			Name:         string(id),
			Version:      "1.0.0",
			Dependencies: dependencies,
			Capabilities: []pkgRuntime.Capability{
				pkgRuntime.CapabilityConfigure,
				pkgRuntime.CapabilityStart,
				pkgRuntime.CapabilityStop,
			},
		},
		calls: calls,
	}
}

func (p *lifecyclePlugin) Metadata() pkgRuntime.Metadata {
	return p.metadata
}

func (p *lifecyclePlugin) Manifest() pkgPlugin.Manifest {
	return pkgPlugin.Manifest{
		RuntimeAPIVersion: pkgPlugin.RuntimeAPIVersion,
	}
}

func (p *lifecyclePlugin) Configure(context pkgRuntime.Context) error {
	p.context = context
	*p.calls = append(*p.calls, string(p.metadata.ID)+":configure")

	return nil
}

func (p *lifecyclePlugin) Start(pkgRuntime.Context) error {
	*p.calls = append(*p.calls, string(p.metadata.ID)+":start")

	return nil
}

func (p *lifecyclePlugin) Stop(pkgRuntime.Context) error {
	*p.calls = append(*p.calls, string(p.metadata.ID)+":stop")

	return nil
}

func (c *contextComponent) Metadata() pkgRuntime.Metadata {
	return pkgRuntime.Metadata{
		ID:      "context-capture",
		Name:    "Context capture",
		Version: "1.0.0",
		Capabilities: []pkgRuntime.Capability{
			pkgRuntime.CapabilityConfigure,
		},
	}
}

func (c *contextComponent) Configure(runtimeContext pkgRuntime.Context) error {
	c.context = runtimeContext

	return nil
}

func (p *testProvider) ID() pkgProvider.ID {
	return p.id
}

func TestNewConfiguresDefaultProvider(t *testing.T) {
	config := pkgRuntime.NewConfig(map[string]any{
		pkgProvider.ConfigDefaultProvider: "ollama",
	})
	runtime := New(WithConfig(config))
	registered := &testProvider{id: "ollama"}

	if err := runtime.Providers().Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	resolved, err := runtime.Providers().Default()
	if err != nil {
		t.Fatalf("resolve configured default: %v", err)
	}

	if resolved != registered {
		t.Fatal("resolved an unexpected default provider")
	}
}

func TestNewSharesConfigurationAndProvidersWithComponents(t *testing.T) {
	config := pkgRuntime.NewConfig(map[string]any{
		"component.setting": "configured",
	})
	runtime := New(WithConfig(config))
	component := &contextComponent{}

	if err := runtime.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	if component.context == nil {
		t.Fatal("component did not receive a runtime context")
	}

	if got := component.context.Config().Get("component.setting"); got != "configured" {
		t.Fatalf("expected injected configuration, got %#v", got)
	}

	if component.context.Providers() != runtime.Providers() {
		t.Fatal("component and runtime received different provider runtimes")
	}
}

func TestNewReplacesTypedNilServicesWithDefaults(t *testing.T) {
	var config *typedNilConfig
	var logger *typedNilLogger

	runtime := New(
		WithConfig(config),
		WithLogger(logger),
	)

	if runtime == nil {
		t.Fatal("expected a runtime")
	}

	if _, err := runtime.Providers().Default(); err == nil {
		t.Fatal("expected an unconfigured default provider")
	}
}

func TestNewRegistersPluginAsRuntimeComponent(t *testing.T) {
	runtime := New()
	calls := make([]string, 0)
	registered := newLifecyclePlugin("laravel", &calls)

	if err := runtime.Plugins().Register(registered); err != nil {
		t.Fatalf("register plugin: %v", err)
	}

	resolvedPlugin, err := runtime.Plugins().Resolve("laravel")
	if err != nil {
		t.Fatalf("resolve plugin: %v", err)
	}
	if resolvedPlugin != registered {
		t.Fatal("resolved an unexpected plugin")
	}

	resolvedComponent, err := runtime.Registry().Resolve("laravel")
	if err != nil {
		t.Fatalf("resolve plugin as component: %v", err)
	}
	if resolvedComponent != registered {
		t.Fatal("plugin and component registries resolved different values")
	}

	if state := runtime.StateManager().Get(registered).State; state != pkgRuntime.StateCreated {
		t.Fatalf("expected plugin state Created, got %d", state)
	}
}

func TestPluginUsesCoreDependencyGraphAndLifecycle(t *testing.T) {
	runtime := New()
	calls := make([]string, 0)
	dependency := newLifecyclePlugin("framework", &calls)
	dependent := newLifecyclePlugin(
		"framework-tools",
		&calls,
		pkgRuntime.Dependency{ID: "framework", Required: true},
	)

	if err := runtime.Plugins().Register(dependent); err != nil {
		t.Fatalf("register dependent plugin: %v", err)
	}
	if err := runtime.Plugins().Register(dependency); err != nil {
		t.Fatalf("register dependency plugin: %v", err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	if dependent.context == nil || dependency.context == nil {
		t.Fatal("plugins did not receive the Runtime Context")
	}

	if got, want := calls, []string{
		"framework:configure",
		"framework:start",
		"framework-tools:configure",
		"framework-tools:start",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}

	calls = calls[:0]
	if err := runtime.Stop(context.Background()); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}

	if got, want := calls, []string{
		"framework-tools:stop",
		"framework:stop",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}
}

func TestDirectComponentRegistrationDoesNotClassifyPlugin(t *testing.T) {
	runtime := New()
	component := &contextComponent{}

	if err := runtime.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	if runtime.Plugins().Has(component.Metadata().ID) {
		t.Fatal("directly registered component was classified as a plugin")
	}

	if _, err := runtime.Plugins().Resolve(component.Metadata().ID); !errors.Is(
		err,
		pkgPlugin.ErrNotFound,
	) {
		t.Fatalf("expected plugin ErrNotFound, got %v", err)
	}
}

func TestPluginRegistrationRejectsComponentIDCollision(t *testing.T) {
	runtime := New()
	component := &contextComponent{}
	calls := make([]string, 0)

	if err := runtime.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	err := runtime.Plugins().Register(
		newLifecyclePlugin(component.Metadata().ID, &calls),
	)
	if !errors.Is(err, pkgPlugin.ErrAlreadyRegistered) {
		t.Fatalf("expected plugin ErrAlreadyRegistered, got %v", err)
	}
	if !errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
		t.Fatalf("expected runtime ErrAlreadyRegistered, got %v", err)
	}

	if runtime.Plugins().Has(component.Metadata().ID) {
		t.Fatal("plugin rejected for component ID collision was indexed")
	}
}

func TestPluginRegistrationRespectsRuntimeState(t *testing.T) {
	runtime := New()
	calls := make([]string, 0)

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	err := runtime.Plugins().Register(
		newLifecyclePlugin("late-plugin", &calls),
	)
	if !errors.Is(err, pkgRuntime.ErrAlreadyStarted) {
		t.Fatalf("expected ErrAlreadyStarted, got %v", err)
	}

	if runtime.Plugins().Has("late-plugin") {
		t.Fatal("plugin rejected by Runtime Core was indexed")
	}
}

func TestPluginRuntimePublishesSuccessfulOperations(t *testing.T) {
	runtime := New()
	calls := make([]string, 0)
	events := make([]pkgPlugin.EventPayload, 0)
	pluginCalls := make([]string, 0)

	for _, topic := range []string{
		pkgPlugin.EventLoaderRegistered,
		pkgPlugin.EventRegistered,
		pkgPlugin.EventLoaded,
	} {
		if err := runtime.EventBus().Subscribe(
			topic,
			func(event pkgRuntime.Event) {
				calls = append(calls, event.Name())
				events = append(events, event.Payload().(pkgPlugin.EventPayload))
			},
		); err != nil {
			t.Fatalf("subscribe to %q: %v", topic, err)
		}
	}

	if err := runtime.Plugins().RegisterLoader(
		"laravel",
		pkgPlugin.LoaderFunc(func(context.Context) (pkgPlugin.Plugin, error) {
			return newLifecyclePlugin("laravel", &pluginCalls), nil
		}),
	); err != nil {
		t.Fatalf("register plugin loader: %v", err)
	}
	loaded, err := runtime.Plugins().Load(context.Background(), "laravel")
	if err != nil {
		t.Fatalf("load plugin: %v", err)
	}

	if want := []string{
		pkgPlugin.EventLoaderRegistered,
		pkgPlugin.EventRegistered,
		pkgPlugin.EventLoaded,
	}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("expected event calls %v, got %v", want, calls)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 plugin events, got %d", len(events))
	}
	for index, event := range events {
		if event.ID != "laravel" {
			t.Fatalf("event %d has unexpected plugin ID %q", index, event.ID)
		}
	}
	if events[0].Plugin != nil {
		t.Fatal("loader registration event unexpectedly contains a plugin")
	}
	if events[1].Plugin != loaded || events[2].Plugin != loaded {
		t.Fatal("registration and load events do not contain loaded plugin")
	}
}

type typedNilConfig struct{}

func (c *typedNilConfig) Get(string) any {
	return nil
}

type typedNilLogger struct{}

func (l *typedNilLogger) Debug(string) {}
func (l *typedNilLogger) Info(string)  {}
func (l *typedNilLogger) Warn(string)  {}
func (l *typedNilLogger) Error(string) {}
