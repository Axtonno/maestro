package gestor

import (
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestGestorEventImplementsRuntimeEventWithValuePayload(t *testing.T) {
	payload := EventPayload{
		Capability: CapabilityProviderCompletion,
		Generation: 3,
		Failure:    EventFailureNone,
	}
	event := Event{Topic: EventResolutionCompleted, Data: payload}
	var runtimeEvent pkgRuntime.Event = event

	if runtimeEvent.Name() != EventResolutionCompleted {
		t.Fatalf("unexpected event name %q", runtimeEvent.Name())
	}
	if got, ok := runtimeEvent.Payload().(EventPayload); !ok || got != payload {
		t.Fatalf("unexpected event payload: %#v", runtimeEvent.Payload())
	}
}

func TestGestorEventTopicsAndFailuresAreStable(t *testing.T) {
	topics := []string{
		EventRefreshStarted,
		EventRefreshCompleted,
		EventRefreshFailed,
		EventResolutionCompleted,
		EventResolutionFailed,
	}
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		if topic == "" {
			t.Fatal("Gestor event topic is empty")
		}
		if _, exists := seen[topic]; exists {
			t.Fatalf("duplicate Gestor event topic %q", topic)
		}
		seen[topic] = struct{}{}
	}

	failures := []EventFailure{
		EventFailureCanceled,
		EventFailureDeadline,
		EventFailureInvalid,
		EventFailureSource,
		EventFailureNotFound,
		EventFailureUnavailable,
		EventFailureAmbiguous,
		EventFailureStale,
		EventFailureInternal,
	}
	seenFailures := make(map[EventFailure]struct{}, len(failures))
	for _, failure := range failures {
		if failure == EventFailureNone {
			t.Fatal("failure category is empty")
		}
		if _, exists := seenFailures[failure]; exists {
			t.Fatalf("duplicate failure category %q", failure)
		}
		seenFailures[failure] = struct{}{}
	}
}
