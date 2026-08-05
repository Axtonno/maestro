package plugin

const (
	EventLoaderRegistered = "plugin.loader.registered"
	EventRegistered       = "plugin.registered"
	EventLoaded           = "plugin.loaded"
)

// EventPayload describes a Plugin Runtime state change.
type EventPayload struct {
	ID     ID
	Plugin Plugin
}

// Event is published on Maestro's shared Event Bus after a Plugin Runtime
// operation has completed successfully.
type Event struct {
	Topic string
	Data  EventPayload
}

func (e Event) Name() string {
	return e.Topic
}

func (e Event) Payload() any {
	return e.Data
}
