package runtime

type Configurer interface {
	Configure(Context) error
}

type Initializer interface {
	Initialize(Context) error
}

type Starter interface {
	Start(Context) error
}

type Stopper interface {
	Stop(Context) error
}

type Reloader interface {
	Reload(Context) error
}

type HealthChecker interface {
	Health(Context) error
}
