package maestro

import (
	internalRuntime "github.com/antonio-cafeo/maestro/internal/runtime"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

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

// New constructs a Runtime with isolated component, event, state and provider
// services. Nil services are replaced by safe defaults.
func New(runtimeOptions ...Option) pkgRuntime.Runtime {
	configured := &options{}

	for _, option := range runtimeOptions {
		if option != nil {
			option(configured)
		}
	}

	return internalRuntime.New(configured.config, configured.logger)
}
