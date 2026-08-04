package runtime

type Context interface {
	Config() Config
	Logger() Logger
	EventBus() EventBus
	Registry() Registry
}