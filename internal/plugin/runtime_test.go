package plugin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type testPlugin struct {
	metadata pkgRuntime.Metadata
	manifest pkgPlugin.Manifest
}

func newTestPlugin(id pkgPlugin.ID) *testPlugin {
	return &testPlugin{
		metadata: pkgRuntime.Metadata{
			ID:      id,
			Name:    string(id),
			Version: "1.0.0",
		},
		manifest: pkgPlugin.Manifest{
			RuntimeAPIVersion: pkgPlugin.RuntimeAPIVersion,
		},
	}
}

func (p *testPlugin) Metadata() pkgRuntime.Metadata {
	return p.metadata
}

func (p *testPlugin) Manifest() pkgPlugin.Manifest {
	return p.manifest
}

type recordingRegistrar struct {
	mu         sync.Mutex
	components map[pkgRuntime.ComponentID]pkgRuntime.Component
	err        error
}

type testLoader struct {
	plugin pkgPlugin.Plugin
	err    error
}

func (l *testLoader) Load(context.Context) (pkgPlugin.Plugin, error) {
	return l.plugin, l.err
}

func newRecordingRegistrar() *recordingRegistrar {
	return &recordingRegistrar{
		components: make(
			map[pkgRuntime.ComponentID]pkgRuntime.Component,
		),
	}
}

func (r *recordingRegistrar) Register(
	component pkgRuntime.Component,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return r.err
	}

	id := component.Metadata().ID
	if _, exists := r.components[id]; exists {
		return pkgRuntime.ErrAlreadyRegistered
	}

	r.components[id] = component

	return nil
}

func (r *recordingRegistrar) resolve(
	id pkgRuntime.ComponentID,
) (pkgRuntime.Component, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	component, exists := r.components[id]

	return component, exists
}

func TestRuntimeRegistersAndResolvesPlugin(t *testing.T) {
	registrar := newRecordingRegistrar()
	pluginRuntime := NewRuntime(registrar)
	registered := newTestPlugin("laravel")

	if err := pluginRuntime.Register(registered); err != nil {
		t.Fatalf("register plugin: %v", err)
	}

	resolved, err := pluginRuntime.Resolve("laravel")
	if err != nil {
		t.Fatalf("resolve plugin: %v", err)
	}

	if resolved != registered {
		t.Fatal("resolved an unexpected plugin")
	}

	component, exists := registrar.resolve("laravel")
	if !exists || component != registered {
		t.Fatal("plugin was not registered in the Runtime Core")
	}

	if !pluginRuntime.Has("laravel") {
		t.Fatal("plugin runtime does not contain registered plugin")
	}
}

func TestRuntimeRejectsInvalidPlugin(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	var typedNil *testPlugin

	tests := []struct {
		name   string
		plugin pkgPlugin.Plugin
	}{
		{name: "nil"},
		{name: "typed nil", plugin: typedNil},
		{name: "empty ID", plugin: newTestPlugin("")},
		{name: "blank ID", plugin: newTestPlugin("   ")},
		{name: "leading whitespace", plugin: newTestPlugin(" laravel")},
		{name: "trailing whitespace", plugin: newTestPlugin("laravel ")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := pluginRuntime.Register(test.plugin)
			if !errors.Is(err, pkgPlugin.ErrInvalidPlugin) {
				t.Fatalf("expected ErrInvalidPlugin, got %v", err)
			}
		})
	}
}

func TestRuntimeValidatesPluginManifest(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())

	missingVersion := newTestPlugin("missing-version")
	missingVersion.manifest.RuntimeAPIVersion = ""
	if err := pluginRuntime.Register(missingVersion); !errors.Is(
		err,
		pkgPlugin.ErrInvalidManifest,
	) {
		t.Fatalf("expected ErrInvalidManifest, got %v", err)
	}

	incompatible := newTestPlugin("incompatible")
	incompatible.manifest.RuntimeAPIVersion = "999"
	if err := pluginRuntime.Register(incompatible); !errors.Is(
		err,
		pkgPlugin.ErrIncompatible,
	) {
		t.Fatalf("expected ErrIncompatible, got %v", err)
	}

	whitespace := newTestPlugin("whitespace-version")
	whitespace.manifest.RuntimeAPIVersion = " 1"
	if err := pluginRuntime.Register(whitespace); !errors.Is(
		err,
		pkgPlugin.ErrIncompatible,
	) {
		t.Fatalf("expected exact version mismatch, got %v", err)
	}

	if pluginRuntime.Has("missing-version") ||
		pluginRuntime.Has("incompatible") ||
		pluginRuntime.Has("whitespace-version") {
		t.Fatal("plugin with invalid manifest was indexed")
	}
}

func TestRuntimeRejectsDuplicatePlugin(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	registered := newTestPlugin("laravel")

	if err := pluginRuntime.Register(registered); err != nil {
		t.Fatalf("register plugin: %v", err)
	}

	err := pluginRuntime.Register(registered)
	if !errors.Is(err, pkgPlugin.ErrAlreadyRegistered) {
		t.Fatalf("expected plugin ErrAlreadyRegistered, got %v", err)
	}
	if !errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
		t.Fatalf("expected runtime ErrAlreadyRegistered, got %v", err)
	}
}

func TestRuntimeDoesNotIndexRejectedPlugin(t *testing.T) {
	registrar := newRecordingRegistrar()
	registrar.err = pkgRuntime.ErrAlreadyStarted
	pluginRuntime := NewRuntime(registrar)

	err := pluginRuntime.Register(newTestPlugin("laravel"))
	if !errors.Is(err, pkgRuntime.ErrAlreadyStarted) {
		t.Fatalf("expected ErrAlreadyStarted, got %v", err)
	}

	if pluginRuntime.Has("laravel") {
		t.Fatal("plugin rejected by Runtime Core was indexed")
	}
}

func TestRuntimeResolveErrors(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())

	if _, err := pluginRuntime.Resolve(""); !errors.Is(
		err,
		pkgPlugin.ErrInvalidPlugin,
	) {
		t.Fatalf("expected ErrInvalidPlugin, got %v", err)
	}

	if _, err := pluginRuntime.Resolve("missing"); !errors.Is(
		err,
		pkgPlugin.ErrNotFound,
	) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if pluginRuntime.Has("") {
		t.Fatal("invalid plugin ID must not be present")
	}
}

func TestRuntimeSupportsConcurrentRegistrationAndResolution(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	const pluginCount = 100

	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, pluginCount)

	for index := 0; index < pluginCount; index++ {
		waitGroup.Add(1)

		go func(index int) {
			defer waitGroup.Done()

			id := pkgPlugin.ID(fmt.Sprintf("plugin-%d", index))
			if err := pluginRuntime.Register(newTestPlugin(id)); err != nil {
				errorsChannel <- err
				return
			}

			if _, err := pluginRuntime.Resolve(id); err != nil {
				errorsChannel <- err
			}
		}(index)
	}

	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		t.Errorf("concurrent plugin operation: %v", err)
	}
}

func TestRuntimeReportsRegisteredPluginsInOrder(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())

	for _, id := range []pkgPlugin.ID{"laravel", "symfony", "django"} {
		if err := pluginRuntime.Register(newTestPlugin(id)); err != nil {
			t.Fatalf("register plugin %q: %v", id, err)
		}
	}

	registered := pluginRuntime.Registered()
	want := []pkgPlugin.ID{"laravel", "symfony", "django"}
	if !reflect.DeepEqual(registered, want) {
		t.Fatalf("expected registered plugins %v, got %v", want, registered)
	}

	registered[0] = "changed"
	if got := pluginRuntime.Registered()[0]; got != "laravel" {
		t.Fatalf("registered plugin snapshot changed internal order: %q", got)
	}
}

func TestRuntimeRegistersAndDiscoversLoaders(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	laravelLoader := &testLoader{plugin: newTestPlugin("laravel")}

	if err := pluginRuntime.RegisterLoader("laravel", laravelLoader); err != nil {
		t.Fatalf("register loader: %v", err)
	}
	if err := pluginRuntime.RegisterLoader(
		"symfony",
		pkgPlugin.LoaderFunc(func(context.Context) (pkgPlugin.Plugin, error) {
			return newTestPlugin("symfony"), nil
		}),
	); err != nil {
		t.Fatalf("register function loader: %v", err)
	}

	available := pluginRuntime.Available()
	want := []pkgPlugin.ID{"laravel", "symfony"}
	if !reflect.DeepEqual(available, want) {
		t.Fatalf("expected available plugins %v, got %v", want, available)
	}

	available[0] = "changed"
	if got := pluginRuntime.Available()[0]; got != "laravel" {
		t.Fatalf("available plugin snapshot changed internal order: %q", got)
	}

	if err := pluginRuntime.RegisterLoader("laravel", laravelLoader); !errors.Is(
		err,
		pkgPlugin.ErrLoaderAlreadyRegistered,
	) {
		t.Fatalf("expected ErrLoaderAlreadyRegistered, got %v", err)
	}
}

func TestRuntimeRejectsInvalidLoader(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	var typedNil *testLoader

	tests := []struct {
		name   string
		id     pkgPlugin.ID
		loader pkgPlugin.Loader
	}{
		{name: "empty ID", loader: &testLoader{}},
		{name: "blank ID", id: "   ", loader: &testLoader{}},
		{name: "nil", id: "laravel"},
		{name: "typed nil", id: "laravel", loader: typedNil},
		{name: "nil function", id: "laravel", loader: pkgPlugin.LoaderFunc(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := pluginRuntime.RegisterLoader(test.id, test.loader)
			if !errors.Is(err, pkgPlugin.ErrInvalidLoader) {
				t.Fatalf("expected ErrInvalidLoader, got %v", err)
			}
		})
	}
}

func TestRuntimeLoadsAndRegistersPlugin(t *testing.T) {
	registrar := newRecordingRegistrar()
	pluginRuntime := NewRuntime(registrar)
	loaded := newTestPlugin("laravel")

	if err := pluginRuntime.RegisterLoader(
		"laravel",
		&testLoader{plugin: loaded},
	); err != nil {
		t.Fatalf("register loader: %v", err)
	}

	plugin, err := pluginRuntime.Load(context.Background(), "laravel")
	if err != nil {
		t.Fatalf("load plugin: %v", err)
	}
	if plugin != loaded {
		t.Fatal("loaded an unexpected plugin")
	}
	if !pluginRuntime.Has("laravel") {
		t.Fatal("loaded plugin was not registered")
	}
	if component, exists := registrar.resolve("laravel"); !exists || component != loaded {
		t.Fatal("loaded plugin was not registered as a Runtime component")
	}
}

func TestRuntimeReportsLoaderFailures(t *testing.T) {
	loadFailure := errors.New("factory failed")

	tests := []struct {
		name   string
		id     pkgPlugin.ID
		loader pkgPlugin.Loader
		ctx    context.Context
		want   error
	}{
		{
			name: "missing loader",
			id:   "missing",
			ctx:  context.Background(),
			want: pkgPlugin.ErrLoaderNotFound,
		},
		{
			name:   "loader error",
			id:     "failure",
			loader: &testLoader{err: loadFailure},
			ctx:    context.Background(),
			want:   pkgPlugin.ErrLoadFailed,
		},
		{
			name:   "nil plugin",
			id:     "nil-plugin",
			loader: &testLoader{},
			ctx:    context.Background(),
			want:   pkgPlugin.ErrInvalidPlugin,
		},
		{
			name:   "mismatched ID",
			id:     "expected",
			loader: &testLoader{plugin: newTestPlugin("other")},
			ctx:    context.Background(),
			want:   pkgPlugin.ErrInvalidPlugin,
		},
		{
			name: "nil context",
			id:   "nil-context",
			ctx:  nil,
			want: pkgPlugin.ErrInvalidLoader,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pluginRuntime := NewRuntime(newRecordingRegistrar())
			if test.loader != nil {
				if err := pluginRuntime.RegisterLoader(test.id, test.loader); err != nil {
					t.Fatalf("register loader: %v", err)
				}
			}

			_, err := pluginRuntime.Load(test.ctx, test.id)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if test.want == pkgPlugin.ErrLoadFailed && !errors.Is(err, loadFailure) {
				t.Fatalf("expected wrapped loader error, got %v", err)
			}
		})
	}
}

func TestRuntimeLoadHonorsCanceledContext(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	if err := pluginRuntime.RegisterLoader(
		"laravel",
		&testLoader{plugin: newTestPlugin("laravel")},
	); err != nil {
		t.Fatalf("register loader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := pluginRuntime.Load(ctx, "laravel"); !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if pluginRuntime.Has("laravel") {
		t.Fatal("plugin was registered after context cancellation")
	}
}

func TestRuntimeLoadChecksCancellationAfterLoader(t *testing.T) {
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	ctx, cancel := context.WithCancel(context.Background())

	if err := pluginRuntime.RegisterLoader(
		"laravel",
		pkgPlugin.LoaderFunc(func(context.Context) (pkgPlugin.Plugin, error) {
			cancel()

			return newTestPlugin("laravel"), nil
		}),
	); err != nil {
		t.Fatalf("register loader: %v", err)
	}

	if _, err := pluginRuntime.Load(ctx, "laravel"); !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if pluginRuntime.Has("laravel") {
		t.Fatal("plugin was registered after loader canceled context")
	}
}
