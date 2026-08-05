package runtime

import (
	"context"

	"github.com/antonio-cafeo/maestro/pkg/provider"
)

type Runtime interface {
	Register(Component) error

	Start(context.Context) error
	Stop(context.Context) error

	Registry() Registry
	EventBus() EventBus
	StateManager() StateManager
	Providers() provider.Runtime
}
