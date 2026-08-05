package plugin

import (
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
	plugins   map[pkgPlugin.ID]pkgPlugin.Plugin
}

// NewRuntime constructs a Plugin Runtime backed by the Runtime Core component
// registrar. The registrar owns component registration, state and lifecycle.
func NewRuntime(registrar componentRegistrar) pkgPlugin.Runtime {
	return &runtime{
		registrar: registrar,
		plugins:   make(map[pkgPlugin.ID]pkgPlugin.Plugin),
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
	r.mu.Unlock()

	return nil
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
