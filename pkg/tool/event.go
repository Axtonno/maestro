package tool

const (
	EventInvocationPrepared   = "tool.invocation.prepared"
	EventInvocationAuthorized = "tool.invocation.authorized"
	EventInvocationCompleted  = "tool.invocation.completed"
	EventInvocationFailed     = "tool.invocation.failed"
)

type EventFailure string

const (
	EventFailureNone     EventFailure = ""
	EventFailureInvalid  EventFailure = "invalid"
	EventFailureDenied   EventFailure = "denied"
	EventFailureCanceled EventFailure = "canceled"
	EventFailureDeadline EventFailure = "deadline_exceeded"
	EventFailureLimit    EventFailure = "limit_exceeded"
	EventFailureTool     EventFailure = "tool_failure"
	EventFailureInternal EventFailure = "internal"
)

// EventPayload is an exact telemetry allowlist. It intentionally contains no
// policy ID, workspace, resource, arguments, output, error text, or permit.
type EventPayload struct {
	Run            RunID
	Tool           ID
	Call           CallID
	ActionCount    int
	Decision       DecisionKind
	Disposition    DenyDisposition
	Outcome        ResultOutcome
	Truncated      bool
	DurationMillis int64
	Failure        EventFailure
}

type Event struct {
	Topic string
	Data  EventPayload
}

func (event Event) Name() string { return event.Topic }
func (event Event) Payload() any { return event.Data }
