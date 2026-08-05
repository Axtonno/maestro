package runtime

import (
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type registry struct {
	mu sync.RWMutex

	components map[pkgRuntime.ComponentID]pkgRuntime.Component
}

func newRegistry() *registry {
	return &registry{
		components: make(map[pkgRuntime.ComponentID]pkgRuntime.Component),
	}
}

func (r *registry) Register(component pkgRuntime.Component) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := component.Metadata().ID

	if _, exists := r.components[id]; exists {
		return pkgRuntime.ErrAlreadyRegistered
	}

	r.components[id] = component

	return nil
}

func (r *registry) Resolve(
	id pkgRuntime.ComponentID,
) (pkgRuntime.Component, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	component, exists := r.components[id]
	if !exists {
		return nil, pkgRuntime.ErrNotFound
	}

	return component, nil
}

func (r *registry) Has(id pkgRuntime.ComponentID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.components[id]

	return exists
}

func (r *registry) Components() []pkgRuntime.Component {
	r.mu.RLock()
	defer r.mu.RUnlock()

	components := make(
		[]pkgRuntime.Component,
		0,
		len(r.components),
	)

	for _, component := range r.components {
		components = append(components, component)
	}

	return components
}
