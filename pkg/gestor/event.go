package gestor

// Stable Event Bus topics emitted by the composed Gestor service.
const (
	EventRefreshStarted      = "gestor.refresh.started"
	EventRefreshCompleted    = "gestor.refresh.completed"
	EventRefreshFailed       = "gestor.refresh.failed"
	EventResolutionCompleted = "gestor.resolution.completed"
	EventResolutionFailed    = "gestor.resolution.failed"
)

type EventFailure string

const (
	EventFailureNone        EventFailure = ""
	EventFailureCanceled    EventFailure = "canceled"
	EventFailureDeadline    EventFailure = "deadline_exceeded"
	EventFailureInvalid     EventFailure = "invalid"
	EventFailureSource      EventFailure = "source_failure"
	EventFailureNotFound    EventFailure = "not_found"
	EventFailureUnavailable EventFailure = "unavailable"
	EventFailureAmbiguous   EventFailure = "ambiguous"
	EventFailureStale       EventFailure = "stale"
	EventFailureInternal    EventFailure = "internal"
)

// EventPayload is redacted by construction. It deliberately has no field for
// error messages, source details, target IDs, model IDs or operational data.
// Fields not relevant to a topic retain their zero value.
type EventPayload struct {
	Capability      CapabilityID
	Generation      uint64
	DescriptorCount int
	SourceCount     int
	TargetKind      TargetKind
	Scope           Scope
	Reason          ResolutionReason
	DependencyCount int
	Failure         EventFailure
}

// Event implements runtime.Event structurally without coupling pkg/gestor to
// the Runtime Core package.
type Event struct {
	Topic string
	Data  EventPayload
}

func (event Event) Name() string {
	return event.Topic
}

func (event Event) Payload() any {
	return event.Data
}
