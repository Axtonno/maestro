package runtime

// Handler receives an event published on its subscribed topic.
type Handler func(Event)

// EventBus delivers events synchronously to the handlers subscribed to their
// topic. Handlers are invoked in subscription order.
type EventBus interface {
	Publish(Event) error

	// Subscribe adds a handler to a topic. A topic may have multiple handlers.
	Subscribe(topic string, handler Handler) error

	// Unsubscribe removes all handlers currently subscribed to a topic.
	Unsubscribe(topic string) error
}
