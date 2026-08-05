package runtime

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type testEvent struct {
	name    string
	payload any
}

func (e testEvent) Name() string {
	return e.name
}

func (e testEvent) Payload() any {
	return e.payload
}

func TestEventBusPublishesToHandlersInSubscriptionOrder(t *testing.T) {
	bus := newEventBus()
	calls := make([]string, 0, 2)
	event := testEvent{
		name:    "component.started",
		payload: "provider",
	}

	if err := bus.Subscribe(event.Name(), func(received pkgRuntime.Event) {
		calls = append(calls, "first:"+received.Payload().(string))
	}); err != nil {
		t.Fatalf("subscribe first handler: %v", err)
	}

	if err := bus.Subscribe(event.Name(), func(received pkgRuntime.Event) {
		calls = append(calls, "second:"+received.Payload().(string))
	}); err != nil {
		t.Fatalf("subscribe second handler: %v", err)
	}

	if err := bus.Publish(event); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	if got, want := calls, []string{
		"first:provider",
		"second:provider",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}
}

func TestEventBusOnlyPublishesToMatchingTopic(t *testing.T) {
	bus := newEventBus()
	called := false

	if err := bus.Subscribe("component.stopped", func(pkgRuntime.Event) {
		called = true
	}); err != nil {
		t.Fatalf("subscribe handler: %v", err)
	}

	if err := bus.Publish(testEvent{name: "component.started"}); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	if called {
		t.Fatal("handler for another topic was called")
	}
}

func TestEventBusUnsubscribesAllHandlersFromTopic(t *testing.T) {
	bus := newEventBus()
	called := false

	for range 2 {
		if err := bus.Subscribe("component.started", func(pkgRuntime.Event) {
			called = true
		}); err != nil {
			t.Fatalf("subscribe handler: %v", err)
		}
	}

	if err := bus.Unsubscribe("component.started"); err != nil {
		t.Fatalf("unsubscribe handlers: %v", err)
	}

	// Unsubscribing a valid topic is idempotent.
	if err := bus.Unsubscribe("component.started"); err != nil {
		t.Fatalf("unsubscribe missing handlers: %v", err)
	}

	if err := bus.Publish(testEvent{name: "component.started"}); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	if called {
		t.Fatal("unsubscribed handler was called")
	}
}

func TestEventBusRejectsInvalidEvent(t *testing.T) {
	bus := newEventBus()

	tests := []struct {
		name  string
		event pkgRuntime.Event
	}{
		{name: "nil", event: nil},
		{name: "empty topic", event: testEvent{}},
		{name: "blank topic", event: testEvent{name: " \t\n"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := bus.Publish(test.event)
			if !errors.Is(err, pkgRuntime.ErrInvalidEvent) {
				t.Fatalf("expected ErrInvalidEvent, got %v", err)
			}
		})
	}
}

func TestEventBusRejectsInvalidSubscription(t *testing.T) {
	bus := newEventBus()
	handler := func(pkgRuntime.Event) {}

	tests := []struct {
		name    string
		topic   string
		handler pkgRuntime.Handler
	}{
		{name: "empty topic", handler: handler},
		{name: "blank topic", topic: " \t\n", handler: handler},
		{name: "nil handler", topic: "component.started"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := bus.Subscribe(test.topic, test.handler)
			if !errors.Is(err, pkgRuntime.ErrInvalidSubscription) {
				t.Fatalf(
					"expected ErrInvalidSubscription, got %v",
					err,
				)
			}
		})
	}
}

func TestEventBusRejectsInvalidUnsubscription(t *testing.T) {
	bus := newEventBus()

	for _, topic := range []string{"", " \t\n"} {
		err := bus.Unsubscribe(topic)
		if !errors.Is(err, pkgRuntime.ErrInvalidSubscription) {
			t.Fatalf(
				"expected ErrInvalidSubscription, got %v",
				err,
			)
		}
	}
}

func TestEventBusUsesSubscriptionSnapshotDuringPublish(t *testing.T) {
	bus := newEventBus()
	calls := make([]string, 0, 3)
	const topic = "component.started"

	if err := bus.Subscribe(topic, func(pkgRuntime.Event) {
		calls = append(calls, "first")

		if err := bus.Unsubscribe(topic); err != nil {
			t.Errorf("unsubscribe from handler: %v", err)
		}

		if err := bus.Subscribe(topic, func(pkgRuntime.Event) {
			calls = append(calls, "third")
		}); err != nil {
			t.Errorf("subscribe from handler: %v", err)
		}
	}); err != nil {
		t.Fatalf("subscribe first handler: %v", err)
	}

	if err := bus.Subscribe(topic, func(pkgRuntime.Event) {
		calls = append(calls, "second")
	}); err != nil {
		t.Fatalf("subscribe second handler: %v", err)
	}

	if err := bus.Publish(testEvent{name: topic}); err != nil {
		t.Fatalf("publish first event: %v", err)
	}

	if got, want := calls, []string{"first", "second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected first publication calls %v, got %v", want, got)
	}

	calls = calls[:0]

	if err := bus.Publish(testEvent{name: topic}); err != nil {
		t.Fatalf("publish second event: %v", err)
	}

	if got, want := calls, []string{"third"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected second publication calls %v, got %v", want, got)
	}
}

func TestEventBusSupportsConcurrentUse(t *testing.T) {
	bus := newEventBus()
	const (
		topic        = "component.started"
		handlerCount = 32
		publishCount = 32
	)

	var calls atomic.Int64
	var subscribers sync.WaitGroup
	for range handlerCount {
		subscribers.Add(1)

		go func() {
			defer subscribers.Done()

			if err := bus.Subscribe(topic, func(pkgRuntime.Event) {
				calls.Add(1)
			}); err != nil {
				t.Errorf("subscribe handler: %v", err)
			}
		}()
	}
	subscribers.Wait()

	var publishers sync.WaitGroup
	for range publishCount {
		publishers.Add(1)

		go func() {
			defer publishers.Done()

			if err := bus.Publish(testEvent{name: topic}); err != nil {
				t.Errorf("publish event: %v", err)
			}
		}()
	}
	publishers.Wait()

	if got, want := calls.Load(), int64(handlerCount*publishCount); got != want {
		t.Fatalf("expected %d handler calls, got %d", want, got)
	}
}
