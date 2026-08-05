package runtime

import (
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type runtimeContext struct {
	config    pkgRuntime.Config
	logger    pkgRuntime.Logger
	eventBus  pkgRuntime.EventBus
	registry  pkgRuntime.Registry
	providers pkgProvider.Runtime
}

func newRuntimeContext(
	config pkgRuntime.Config,
	logger pkgRuntime.Logger,
	eventBus pkgRuntime.EventBus,
	registry pkgRuntime.Registry,
	providers pkgProvider.Runtime,
) *runtimeContext {
	return &runtimeContext{
		config:    config,
		logger:    logger,
		eventBus:  eventBus,
		registry:  registry,
		providers: providers,
	}
}

func (c *runtimeContext) Config() pkgRuntime.Config {
	return c.config
}

func (c *runtimeContext) Logger() pkgRuntime.Logger {
	return c.logger
}

func (c *runtimeContext) EventBus() pkgRuntime.EventBus {
	return c.eventBus
}

func (c *runtimeContext) Registry() pkgRuntime.Registry {
	return c.registry
}

func (c *runtimeContext) Providers() pkgProvider.Runtime {
	return c.providers
}
