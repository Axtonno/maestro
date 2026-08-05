package plugin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgPlugin.Runtime = (*runtime)(nil)

type componentRegistrar interface {
	Register(pkgRuntime.Component) error
}

type runtime struct {
	mu sync.RWMutex

	registrar componentRegistrar
	eventBus  pkgRuntime.EventBus
	plugins   map[pkgPlugin.ID]pkgPlugin.Plugin
	pluginIDs []pkgPlugin.ID
	loaders   map[pkgPlugin.ID]pkgPlugin.Loader
	loaderIDs []pkgPlugin.ID
}

// NewRuntime constructs a Plugin Runtime backed by the Runtime Core component
// registrar. The registrar owns component registration, state and lifecycle.
func NewRuntime(registrar componentRegistrar) pkgPlugin.Runtime {
	return NewRuntimeWithEventBus(registrar, nil)
}

// NewRuntimeWithEventBus constructs a Plugin Runtime that publishes successful
// catalog, registration and load operations on the shared Runtime Event Bus.
func NewRuntimeWithEventBus(
	registrar componentRegistrar,
	eventBus pkgRuntime.EventBus,
) pkgPlugin.Runtime {
	return &runtime{
		registrar: registrar,
		eventBus:  eventBus,
		plugins:   make(map[pkgPlugin.ID]pkgPlugin.Plugin),
		loaders:   make(map[pkgPlugin.ID]pkgPlugin.Loader),
	}
}

func (r *runtime) Register(plugin pkgPlugin.Plugin) error {
	if nilPlugin(plugin) {
		return fmt.Errorf(
			"register plugin: plugin is nil: %w",
			pkgPlugin.ErrInvalidPlugin,
		)
	}

	pluginID := plugin.Metadata().ID
	if !validPluginID(pluginID) {
		return fmt.Errorf(
			"register plugin: invalid ID %q: %w",
			pluginID,
			pkgPlugin.ErrInvalidPlugin,
		)
	}

	manifest := plugin.Manifest()
	if manifest.RuntimeAPIVersion == "" {
		return fmt.Errorf(
			"register plugin %q: runtime API version is empty: %w",
			pluginID,
			pkgPlugin.ErrInvalidManifest,
		)
	}
	if manifest.RuntimeAPIVersion != pkgPlugin.RuntimeAPIVersion {
		return fmt.Errorf(
			"register plugin %q: requires runtime API %q, current API is %q: %w",
			pluginID,
			manifest.RuntimeAPIVersion,
			pkgPlugin.RuntimeAPIVersion,
			pkgPlugin.ErrIncompatible,
		)
	}

	if err := r.registrar.Register(plugin); err != nil {
		if errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
			return fmt.Errorf(
				"register plugin %q: %w",
				pluginID,
				errors.Join(pkgPlugin.ErrAlreadyRegistered, err),
			)
		}

		return fmt.Errorf(
			"register plugin %q: %w",
			pluginID,
			err,
		)
	}

	r.mu.Lock()
	r.plugins[pluginID] = plugin
	r.pluginIDs = append(r.pluginIDs, pluginID)
	r.mu.Unlock()
	r.publish(pkgPlugin.EventRegistered, pluginID, plugin)

	return nil
}

func (r *runtime) Registered() []pkgPlugin.ID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]pkgPlugin.ID(nil), r.pluginIDs...)
}

func (r *runtime) RegisterLoader(
	pluginID pkgPlugin.ID,
	loader pkgPlugin.Loader,
) error {
	if !validPluginID(pluginID) {
		return fmt.Errorf(
			"register plugin loader: invalid ID %q: %w",
			pluginID,
			pkgPlugin.ErrInvalidLoader,
		)
	}
	if nilLoader(loader) {
		return fmt.Errorf(
			"register plugin loader %q: loader is nil: %w",
			pluginID,
			pkgPlugin.ErrInvalidLoader,
		)
	}

	r.mu.Lock()

	if _, exists := r.loaders[pluginID]; exists {
		r.mu.Unlock()

		return fmt.Errorf(
			"register plugin loader %q: %w",
			pluginID,
			pkgPlugin.ErrLoaderAlreadyRegistered,
		)
	}
	r.loaders[pluginID] = loader
	r.loaderIDs = append(r.loaderIDs, pluginID)
	r.mu.Unlock()
	r.publish(pkgPlugin.EventLoaderRegistered, pluginID, nil)

	return nil
}

func (r *runtime) Load(
	ctx context.Context,
	pluginID pkgPlugin.ID,
) (pkgPlugin.Plugin, error) {
	if !validPluginID(pluginID) {
		return nil, fmt.Errorf(
			"load plugin: invalid ID %q: %w",
			pluginID,
			pkgPlugin.ErrInvalidLoader,
		)
	}
	if ctx == nil {
		return nil, fmt.Errorf(
			"load plugin %q: context is nil: %w",
			pluginID,
			pkgPlugin.ErrInvalidLoader,
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"load plugin %q: %w",
			pluginID,
			err,
		)
	}

	r.mu.RLock()
	loader, exists := r.loaders[pluginID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf(
			"load plugin %q: %w",
			pluginID,
			pkgPlugin.ErrLoaderNotFound,
		)
	}

	loaded, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"load plugin %q: %w",
			pluginID,
			errors.Join(pkgPlugin.ErrLoadFailed, err),
		)
	}
	if nilPlugin(loaded) {
		return nil, fmt.Errorf(
			"load plugin %q: loader returned a nil plugin: %w",
			pluginID,
			errors.Join(
				pkgPlugin.ErrLoadFailed,
				pkgPlugin.ErrInvalidPlugin,
			),
		)
	}
	if loaded.Metadata().ID != pluginID {
		return nil, fmt.Errorf(
			"load plugin %q: loader returned plugin %q: %w",
			pluginID,
			loaded.Metadata().ID,
			errors.Join(
				pkgPlugin.ErrLoadFailed,
				pkgPlugin.ErrInvalidPlugin,
			),
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"load plugin %q: %w",
			pluginID,
			err,
		)
	}
	if err := r.Register(loaded); err != nil {
		return nil, fmt.Errorf(
			"load plugin %q: %w",
			pluginID,
			err,
		)
	}
	r.publish(pkgPlugin.EventLoaded, pluginID, loaded)

	return loaded, nil
}

func (r *runtime) Available() []pkgPlugin.ID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]pkgPlugin.ID(nil), r.loaderIDs...)
}

func (r *runtime) Resolve(
	pluginID pkgPlugin.ID,
) (pkgPlugin.Plugin, error) {
	if !validPluginID(pluginID) {
		return nil, fmt.Errorf(
			"resolve plugin: invalid ID %q: %w",
			pluginID,
			pkgPlugin.ErrInvalidPlugin,
		)
	}

	r.mu.RLock()
	plugin, exists := r.plugins[pluginID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf(
			"resolve plugin %q: %w",
			pluginID,
			pkgPlugin.ErrNotFound,
		)
	}

	return plugin, nil
}

func (r *runtime) Has(pluginID pkgPlugin.ID) bool {
	if !validPluginID(pluginID) {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[pluginID]

	return exists
}

func validPluginID(pluginID pkgPlugin.ID) bool {
	value := string(pluginID)

	return value != "" && strings.TrimSpace(value) == value
}

func nilPlugin(plugin pkgPlugin.Plugin) bool {
	if plugin == nil {
		return true
	}

	value := reflect.ValueOf(plugin)

	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func nilLoader(loader pkgPlugin.Loader) bool {
	if loader == nil {
		return true
	}

	value := reflect.ValueOf(loader)

	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (r *runtime) publish(
	topic string,
	pluginID pkgPlugin.ID,
	registered pkgPlugin.Plugin,
) {
	if r.eventBus == nil {
		return
	}

	// Topics are constants and Event always implements runtime.Event, so the
	// in-process Event Bus cannot reject these notifications.
	_ = r.eventBus.Publish(pkgPlugin.Event{
		Topic: topic,
		Data: pkgPlugin.EventPayload{
			ID:     pluginID,
			Plugin: registered,
		},
	})
}
