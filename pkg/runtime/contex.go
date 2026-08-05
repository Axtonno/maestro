package runtime

import "github.com/antonio-cafeo/maestro/pkg/provider"

type Context interface {
	Config() Config
	Logger() Logger
	EventBus() EventBus
	Registry() Registry
	Providers() provider.Runtime
}
