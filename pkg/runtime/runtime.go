package runtime

import "context"

type Runtime interface {
	Register(Component) error

	Start(context.Context) error
	Stop(context.Context) error

	Registry() Registry
	EventBus() EventBus
	StateManager() StateManager
}