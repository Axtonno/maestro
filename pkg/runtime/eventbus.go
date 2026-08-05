package runtime

type Handler func(Event)

type EventBus interface {
	Publish(Event) error

	Subscribe(topic string, handler Handler) error

	Unsubscribe(topic string) error
}
