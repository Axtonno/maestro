package plugin

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

// ID is the component identifier used to register and resolve a plugin.
type ID = pkgRuntime.ComponentID

// Plugin is a Runtime component classified as a plugin by registration through
// plugin.Runtime. Its metadata, dependencies and lifecycle capabilities are
// interpreted by the Runtime Core.
type Plugin interface {
	pkgRuntime.Component
}
