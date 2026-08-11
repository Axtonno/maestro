package laravel

import (
	internalLaravel "github.com/antonio-cafeo/maestro/internal/plugin/laravel"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
)

const (
	ID      = internalLaravel.ID
	Version = internalLaravel.Version
)

var (
	ErrInvalidConfig           = internalLaravel.ErrInvalidConfig
	ErrNotDetected             = internalLaravel.ErrNotDetected
	ErrInvalidComposerManifest = internalLaravel.ErrInvalidComposerManifest
)

// Plugin exposes the Laravel workspace information discovered during
// initialization.
type Plugin interface {
	pkgPlugin.Plugin

	Root() string
	FrameworkVersion() string
	pkgContext.WorkspaceProvider
}

// New constructs a Laravel plugin for a workspace. Detection is performed by
// the Runtime Core during plugin initialization.
func New(config Config) (Plugin, error) {
	return internalLaravel.New(config.Root)
}

// NewLoader constructs a loader suitable for plugin.Runtime.RegisterLoader.
func NewLoader(config Config) pkgPlugin.Loader {
	return internalLaravel.NewLoader(config.Root)
}
