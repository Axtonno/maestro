package runtime

import (
	"fmt"
	"strings"
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgRuntime.EventBus = (*eventBus)(nil)

type eventBus struct {
	mu sync.RWMutex

	handlers map[string][]pkgRuntime.Handler
}

func newEventBus() *eventBus {
	return &eventBus{
		handlers: make(map[string][]pkgRuntime.Handler),
	}
}

func (b *eventBus) Publish(event pkgRuntime.Event) error {
	if event == nil {
		return fmt.Errorf(
			"publish event: event is nil: %w",
			pkgRuntime.ErrInvalidEvent,
		)
	}

	topic := event.Name()
	if !validEventTopic(topic) {
		return fmt.Errorf(
			"publish event: topic is empty: %w",
			pkgRuntime.ErrInvalidEvent,
		)
	}

	// Publish works on a snapshot so handlers can safely subscribe or
	// unsubscribe while an event is being delivered.
	b.mu.RLock()
	handlers := append(
		[]pkgRuntime.Handler(nil),
		b.handlers[topic]...,
	)
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}

	return nil
}

func (b *eventBus) Subscribe(
	topic string,
	handler pkgRuntime.Handler,
) error {
	if !validEventTopic(topic) || handler == nil {
		return fmt.Errorf(
			"subscribe event handler: %w",
			pkgRuntime.ErrInvalidSubscription,
		)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = append(b.handlers[topic], handler)

	return nil
}

func (b *eventBus) Unsubscribe(topic string) error {
	if !validEventTopic(topic) {
		return fmt.Errorf(
			"unsubscribe event handlers: %w",
			pkgRuntime.ErrInvalidSubscription,
		)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.handlers, topic)

	return nil
}

func validEventTopic(topic string) bool {
	return strings.TrimSpace(topic) != ""
}
