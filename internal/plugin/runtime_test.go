package plugin

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type testPlugin struct {
	metadata pkgRuntime.Metadata
}

func newTestPlugin(id pkgPlugin.ID) *testPlugin {
	return &testPlugin{
		metadata: pkgRuntime.Metadata{
			ID:      id,
			Name:    string(id),
			Version: "1.0.0",
		},
	}
}

func (p *testPlugin) Metadata() pkgRuntime.Metadata {
	return p.metadata
}

type recordingRegistrar struct {
	mu         sync.Mutex
	components map[pkgRuntime.ComponentID]pkgRuntime.Component
	err        error
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
