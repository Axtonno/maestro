package plugin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

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

type blockingRegistrar struct {
	started chan struct{}
	release chan struct{}
}

func (r *blockingRegistrar) Register(pkgRuntime.Component) error {
	close(r.started)
	<-r.release

	return nil
}

type callbackEventBus struct {
	publish func(pkgRuntime.Event)
}

func (b *callbackEventBus) Publish(event pkgRuntime.Event) error {
	if b.publish != nil {
		b.publish(event)
	}

	return nil
}

func (b *callbackEventBus) Subscribe(string, pkgRuntime.Handler) error {
	return nil
}

func (b *callbackEventBus) Unsubscribe(string) error {
	return nil
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

func timeoutContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	return ctx
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
			_ = pluginRuntime.Registered()
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

func TestRuntimeSupportsConcurrentLoaderRegistrationAndSnapshots(t *testing.T) {
	const loaderCount = 100

	pluginRuntime := NewRuntime(newRecordingRegistrar())
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, loaderCount)

	for index := range loaderCount {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			id := pkgPlugin.ID(fmt.Sprintf("loader-%03d", index))
			if err := pluginRuntime.RegisterLoader(
				id,
				&testLoader{plugin: newTestPlugin(id)},
			); err != nil {
				errorsChannel <- err
				return
			}
			_ = pluginRuntime.Available()
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent loader operation: %v", err)
	}

	available := pluginRuntime.Available()
	if len(available) != loaderCount {
		t.Fatalf("expected %d available loaders, got %d", loaderCount, len(available))
	}
	seen := make(map[pkgPlugin.ID]struct{}, loaderCount)
	for _, id := range available {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate loader ID %q", id)
		}
		seen[id] = struct{}{}
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
		name         string
		id           pkgPlugin.ID
		loader       pkgPlugin.Loader
		ctx          context.Context
		registrarErr error
		want         []error
	}{
		{
			name: "missing loader",
			id:   "missing",
			ctx:  context.Background(),
			want: []error{pkgPlugin.ErrLoaderNotFound},
		},
		{
			name:   "loader error",
			id:     "failure",
			loader: &testLoader{err: loadFailure},
			ctx:    context.Background(),
			want:   []error{pkgPlugin.ErrLoadFailed, loadFailure},
		},
		{
			name:   "nil plugin",
			id:     "nil-plugin",
			loader: &testLoader{},
			ctx:    context.Background(),
			want: []error{
				pkgPlugin.ErrLoadFailed,
				pkgPlugin.ErrInvalidPlugin,
			},
		},
		{
			name:   "mismatched ID",
			id:     "expected",
			loader: &testLoader{plugin: newTestPlugin("other")},
			ctx:    context.Background(),
			want: []error{
				pkgPlugin.ErrLoadFailed,
				pkgPlugin.ErrInvalidPlugin,
			},
		},
		{
			name: "incompatible manifest",
			id:   "incompatible",
			loader: &testLoader{plugin: func() pkgPlugin.Plugin {
				plugin := newTestPlugin("incompatible")
				plugin.manifest.RuntimeAPIVersion = "999"

				return plugin
			}()},
			ctx:  context.Background(),
			want: []error{pkgPlugin.ErrIncompatible},
		},
		{
			name:         "registrar rejection",
			id:           "rejected",
			loader:       &testLoader{plugin: newTestPlugin("rejected")},
			ctx:          context.Background(),
			registrarErr: pkgRuntime.ErrAlreadyStarted,
			want:         []error{pkgRuntime.ErrAlreadyStarted},
		},
		{
			name: "nil context",
			id:   "nil-context",
			ctx:  nil,
			want: []error{pkgPlugin.ErrInvalidLoader},
		},
		{
			name: "invalid ID",
			id:   " invalid",
			ctx:  context.Background(),
			want: []error{pkgPlugin.ErrInvalidLoader},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registrar := newRecordingRegistrar()
			registrar.err = test.registrarErr
			pluginRuntime := NewRuntime(registrar)
			if test.loader != nil {
				if err := pluginRuntime.RegisterLoader(test.id, test.loader); err != nil {
					t.Fatalf("register loader: %v", err)
				}
			}

			_, err := pluginRuntime.Load(test.ctx, test.id)
			for _, want := range test.want {
				if !errors.Is(err, want) {
					t.Fatalf("expected %v, got %v", want, err)
				}
			}
			if pluginRuntime.Has(test.id) {
				t.Fatal("failed load was indexed")
			}
		})
	}
}

func TestRuntimeLoaderRunsWithoutCatalogLock(t *testing.T) {
	ctx := timeoutContext(t)
	started := make(chan struct{})
	release := make(chan struct{})
	pluginRuntime := NewRuntime(newRecordingRegistrar())

	if err := pluginRuntime.RegisterLoader(
		"laravel",
		pkgPlugin.LoaderFunc(func(context.Context) (pkgPlugin.Plugin, error) {
			close(started)
			<-release

			return newTestPlugin("laravel"), nil
		}),
	); err != nil {
		t.Fatalf("register blocking loader: %v", err)
	}

	loadDone := make(chan error, 1)
	go func() {
		_, err := pluginRuntime.Load(ctx, "laravel")
		loadDone <- err
	}()

	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("loader did not start")
	}

	registerDone := make(chan error, 1)
	go func() {
		registerDone <- pluginRuntime.RegisterLoader(
			"symfony",
			&testLoader{plugin: newTestPlugin("symfony")},
		)
	}()

	select {
	case err := <-registerDone:
		if err != nil {
			t.Fatalf("register loader while factory is blocked: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("catalog remained locked while loader was running")
	}

	if got := pluginRuntime.Available(); !reflect.DeepEqual(
		got,
		[]pkgPlugin.ID{"laravel", "symfony"},
	) {
		t.Fatalf("unexpected available snapshot: %v", got)
	}

	close(release)
	if err := <-loadDone; err != nil {
		t.Fatalf("complete blocked load: %v", err)
	}
}

func TestRuntimeRegistrarRunsWithoutPluginLock(t *testing.T) {
	ctx := timeoutContext(t)
	registrar := &blockingRegistrar{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	pluginRuntime := NewRuntime(registrar)
	registerDone := make(chan error, 1)

	go func() {
		registerDone <- pluginRuntime.Register(newTestPlugin("laravel"))
	}()

	select {
	case <-registrar.started:
	case <-ctx.Done():
		t.Fatal("registrar did not start")
	}

	readDone := make(chan struct{})
	go func() {
		_ = pluginRuntime.Registered()
		_ = pluginRuntime.Has("laravel")
		close(readDone)
	}()

	select {
	case <-readDone:
	case <-ctx.Done():
		t.Fatal("plugin registry remained locked while registrar was running")
	}
	if pluginRuntime.Has("laravel") {
		t.Fatal("plugin became visible before registrar completed")
	}

	close(registrar.release)
	if err := <-registerDone; err != nil {
		t.Fatalf("register plugin: %v", err)
	}
	if !pluginRuntime.Has("laravel") {
		t.Fatal("plugin was not indexed after registrar completed")
	}
}

func TestRuntimeEventsCanReenterPluginRuntime(t *testing.T) {
	ctx := timeoutContext(t)
	eventBus := &callbackEventBus{}
	pluginRuntime := NewRuntimeWithEventBus(newRecordingRegistrar(), eventBus)
	eventBus.publish = func(event pkgRuntime.Event) {
		payload := event.Payload().(pkgPlugin.EventPayload)
		_ = pluginRuntime.Available()
		_ = pluginRuntime.Registered()
		_ = pluginRuntime.Has(payload.ID)
		_, _ = pluginRuntime.Resolve(payload.ID)
	}

	done := make(chan error, 1)
	go func() {
		if err := pluginRuntime.RegisterLoader(
			"laravel",
			&testLoader{plugin: newTestPlugin("laravel")},
		); err != nil {
			done <- err
			return
		}
		_, err := pluginRuntime.Load(ctx, "laravel")
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("complete re-entrant operations: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("event callback deadlocked on Plugin Runtime lock")
	}
}

func TestRuntimeConcurrentLoadsOfSameIDHaveOneWinner(t *testing.T) {
	ctx := timeoutContext(t)
	const attemptCount = 16

	started := make(chan struct{}, attemptCount)
	release := make(chan struct{})
	pluginRuntime := NewRuntime(newRecordingRegistrar())
	if err := pluginRuntime.RegisterLoader(
		"laravel",
		pkgPlugin.LoaderFunc(func(context.Context) (pkgPlugin.Plugin, error) {
			started <- struct{}{}
			<-release

			return newTestPlugin("laravel"), nil
		}),
	); err != nil {
		t.Fatalf("register loader: %v", err)
	}

	results := make(chan error, attemptCount)
	for range attemptCount {
		go func() {
			_, err := pluginRuntime.Load(ctx, "laravel")
			results <- err
		}()
	}
	for range attemptCount {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("not all concurrent loader attempts started")
		}
	}
	close(release)

	successes := 0
	duplicates := 0
	for range attemptCount {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, pkgPlugin.ErrAlreadyRegistered) &&
			errors.Is(err, pkgRuntime.ErrAlreadyRegistered):
			duplicates++
		default:
			t.Fatalf("unexpected concurrent load error: %v", err)
		}
	}

	if successes != 1 || duplicates != attemptCount-1 {
		t.Fatalf(
			"expected one success and %d duplicates, got %d and %d",
			attemptCount-1,
			successes,
			duplicates,
		)
	}
	if got := pluginRuntime.Registered(); !reflect.DeepEqual(
		got,
		[]pkgPlugin.ID{"laravel"},
	) {
		t.Fatalf("unexpected registered snapshot: %v", got)
	}
}

func TestRuntimeSupportsConcurrentLoadsOfDifferentIDs(t *testing.T) {
	ctx := timeoutContext(t)
	const pluginCount = 64

	pluginRuntime := NewRuntime(newRecordingRegistrar())
	for index := range pluginCount {
		id := pkgPlugin.ID(fmt.Sprintf("plugin-%03d", index))
		if err := pluginRuntime.RegisterLoader(
			id,
			&testLoader{plugin: newTestPlugin(id)},
		); err != nil {
			t.Fatalf("register loader %q: %v", id, err)
		}
	}

	results := make(chan error, pluginCount)
	for index := range pluginCount {
		id := pkgPlugin.ID(fmt.Sprintf("plugin-%03d", index))
		go func() {
			_, err := pluginRuntime.Load(ctx, id)
			results <- err
		}()
	}
	for range pluginCount {
		if err := <-results; err != nil {
			t.Fatalf("concurrent load: %v", err)
		}
	}

	registered := pluginRuntime.Registered()
	if len(registered) != pluginCount {
		t.Fatalf("expected %d registered plugins, got %d", pluginCount, len(registered))
	}
	seen := make(map[pkgPlugin.ID]struct{}, pluginCount)
	for _, id := range registered {
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate registered ID %q", id)
		}
		seen[id] = struct{}{}
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
