package maestro

import (
	"context"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type testProvider struct {
	id pkgProvider.ID
}

type contextComponent struct {
	context pkgRuntime.Context
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

type typedNilConfig struct{}

func (c *typedNilConfig) Get(string) any {
	return nil
}

type typedNilLogger struct{}

func (l *typedNilLogger) Debug(string) {}
func (l *typedNilLogger) Info(string)  {}
func (l *typedNilLogger) Warn(string)  {}
func (l *typedNilLogger) Error(string) {}
