package runtime

import "context"

type LifecycleManager interface {
	Start(context.Context, Component) error

	Stop(context.Context, Component) error

	Restart(context.Context, Component) error

	Reload(context.Context, Component) error

	Health(context.Context, Component) error
}