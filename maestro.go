package maestro

import (
	internalRuntime "github.com/antonio-cafeo/maestro/internal/runtime"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

// Runtime is Maestro's composition root. It extends the Runtime Core with the
// dedicated Plugin Runtime.
type Runtime interface {
	pkgRuntime.Runtime
	ContextEngine() pkgContext.Engine
	Gestor() pkgGestor.Service
	Plugins() pkgPlugin.Runtime
	Tools() pkgTool.Runtime
	Agents() pkgAgent.Runtime
}

type options struct {
	config pkgRuntime.Config
	logger pkgRuntime.Logger
}

// Option configures the Maestro composition root.
type Option func(*options)

func WithConfig(config pkgRuntime.Config) Option {
	return func(options *options) {
		options.config = config
	}
}

func WithLogger(logger pkgRuntime.Logger) Option {
	return func(options *options) {
		options.logger = logger
	}
}

// New constructs a Runtime with isolated component, event, state, provider and
// plugin services. Nil configurable services are replaced by safe defaults.
func New(runtimeOptions ...Option) Runtime {
	configured := &options{}

	for _, option := range runtimeOptions {
		if option != nil {
			option(configured)
		}
	}

	return internalRuntime.New(configured.config, configured.logger)
}
