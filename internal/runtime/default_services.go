package runtime

import (
	"fmt"
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type emptyConfig struct{}

func newEmptyConfig() *emptyConfig {
	return &emptyConfig{}
}

func (c *emptyConfig) Get(_ string) any {
	return nil
}

type noopLogger struct{}

func newNoopLogger() *noopLogger {
	return &noopLogger{}
}

func (l *noopLogger) Debug(_ string) {}

func (l *noopLogger) Info(_ string) {}

func (l *noopLogger) Warn(_ string) {}

func (l *noopLogger) Error(_ string) {}

type eventBus struct {
	mu sync.RWMutex

	handlers map[string]pkgRuntime.Handler
}

func newEventBus() *eventBus {
	return &eventBus{
		handlers: make(map[string]pkgRuntime.Handler),
	}
}

func (b *eventBus) Publish(event pkgRuntime.Event) error {
	if event == nil {
		return fmt.Errorf(
			"publish event: event is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	b.mu.RLock()
	handler, exists := b.handlers[event.Name()]
	b.mu.RUnlock()

	if exists {
		handler(event)
	}

	return nil
}

func (b *eventBus) Subscribe(
	topic string,
	handler pkgRuntime.Handler,
) error {
	if topic == "" || handler == nil {
		return fmt.Errorf(
			"subscribe event handler: invalid subscription: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[topic] = handler

	return nil
}

func (b *eventBus) Unsubscribe(topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.handlers, topic)

	return nil
}
