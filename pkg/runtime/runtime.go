package runtime

import "context"

type Runtime interface {

    Register(Component) error

    Start(Component) error

    Stop(Component) error

    Restart(Component) error

    Reload(Component) error

    Health(Component) error

    State(Component) State
}