package provider

import "time"

type ProviderEventKind string

const (
	ProviderEventOperationStarted   ProviderEventKind = "operation_started"
	ProviderEventAttemptStarted     ProviderEventKind = "attempt_started"
	ProviderEventRetryScheduled     ProviderEventKind = "retry_scheduled"
	ProviderEventCircuitTransition  ProviderEventKind = "circuit_transition"
	ProviderEventOperationCompleted ProviderEventKind = "operation_completed"
)

type ProviderOperationOutcome string

const (
	ProviderOperationOutcomeSuccess          ProviderOperationOutcome = "success"
	ProviderOperationOutcomeError            ProviderOperationOutcome = "error"
	ProviderOperationOutcomeCanceled         ProviderOperationOutcome = "canceled"
	ProviderOperationOutcomeDeadlineExceeded ProviderOperationOutcome = "deadline_exceeded"
	ProviderOperationOutcomeClosed           ProviderOperationOutcome = "closed"
)

// ProviderEvent is redacted by construction. It never contains request or
// response content, embeddings, credentials, or remote error messages.
type ProviderEvent struct {
	Kind        ProviderEventKind
	OperationID uint64
	Provider    ID
	Operation   Operation
	Model       string
	Timestamp   time.Time
	Duration    time.Duration

	Attempt     uint
	MaxAttempts uint
	Backoff     time.Duration

	Outcome     ProviderOperationOutcome
	ErrorKind   ErrorKind
	StatusCode  int
	Retryable   bool
	CircuitOpen bool

	CircuitFrom CircuitState
	CircuitTo   CircuitState

	Usage          Usage
	CompletedBytes int64
	TotalBytes     int64
}

// ProviderObserver receives immutable event values. Implementations must be
// safe for concurrent calls. Returned errors do not affect provider results.
type ProviderObserver interface {
	ObserveProviderEvent(ProviderEvent) error
}

type ProviderObserverFunc func(ProviderEvent) error

func (f ProviderObserverFunc) ObserveProviderEvent(event ProviderEvent) error {
	return f(event)
}
