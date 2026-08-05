package runtime

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

type runtimeRegistry struct {
	runtime *runtime
}

func newRuntimeRegistry(rt *runtime) *runtimeRegistry {
	return &runtimeRegistry{
		runtime: rt,
	}
}

func (r *runtimeRegistry) Register(
	component pkgRuntime.Component,
) error {
	return r.runtime.Register(component)
}

func (r *runtimeRegistry) Resolve(
	id pkgRuntime.ComponentID,
) (pkgRuntime.Component, error) {
	return r.runtime.registry.Resolve(id)
}

func (r *runtimeRegistry) Has(
	id pkgRuntime.ComponentID,
) bool {
	return r.runtime.registry.Has(id)
}
