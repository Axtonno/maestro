package runtime

type Event interface {
	Name() string

	Payload() any
}