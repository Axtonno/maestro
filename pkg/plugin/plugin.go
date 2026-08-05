package plugin

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

// RuntimeAPIVersion identifies the Plugin Runtime contract implemented by this
// release of Maestro.
const RuntimeAPIVersion = "1"

// ID is the component identifier used to register and resolve a plugin.
type ID = pkgRuntime.ComponentID

// Plugin is a Runtime component classified as a plugin by registration through
// plugin.Runtime. Its metadata, dependencies and lifecycle capabilities are
// interpreted by the Runtime Core.
type Plugin interface {
	pkgRuntime.Component

	Manifest() Manifest
}

// Manifest declares the compatibility requirements of a plugin. Component
// identity and lifecycle declarations remain in runtime.Metadata.
type Manifest struct {
	RuntimeAPIVersion string
}
