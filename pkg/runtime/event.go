package runtime

// Event is a message published through the runtime EventBus.
// Name identifies the topic used for delivery.
type Event interface {
	Name() string
	Payload() any
}
