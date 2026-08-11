package plugin

import "errors"

var (
	// ErrInvalidPlugin reports a nil plugin, an invalid plugin ID or an invalid
	// plugin returned by a loader.
	ErrInvalidPlugin = errors.New("invalid plugin")
	// ErrInvalidManifest reports a manifest without a Runtime API version.
	ErrInvalidManifest = errors.New("invalid plugin manifest")
	// ErrIncompatible reports an exact Runtime API version mismatch.
	ErrIncompatible = errors.New("incompatible plugin")
	// ErrAlreadyRegistered reports a plugin or component ID collision in the
	// Runtime Core registry.
	ErrAlreadyRegistered = errors.New("plugin already registered")
	// ErrNotFound reports that an ID is not present in the plugin registry.
	ErrNotFound = errors.New("plugin not found")
	// ErrInvalidLoader reports an invalid loader, loader ID or Load context.
	ErrInvalidLoader = errors.New("invalid plugin loader")
	// ErrLoaderAlreadyRegistered reports an ID collision in the loader catalog.
	ErrLoaderAlreadyRegistered = errors.New("plugin loader already registered")
	// ErrLoaderNotFound reports that no loader is available for an ID.
	ErrLoaderNotFound = errors.New("plugin loader not found")
	// ErrLoadFailed reports a loader failure or an invalid loader result.
	ErrLoadFailed = errors.New("plugin load failed")
)
