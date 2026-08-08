package provider

import "time"

type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// CircuitBreakerPolicy configures failure counting and bounded half-open
// probes. A zero FailureThreshold disables the circuit breaker.
type CircuitBreakerPolicy struct {
	FailureThreshold    uint
	OpenDuration        time.Duration
	HalfOpenMaxAttempts uint
}

// ResiliencePolicy applies to one provider operation. An empty Model applies
// to every model for that operation; an exact model policy takes precedence.
// MaxAttempts includes the initial attempt.
type ResiliencePolicy struct {
	Operation         Operation
	Model             string
	MaxAttempts       uint
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
	Jitter            float64
	MaxElapsedTime    time.Duration
	CircuitBreaker    CircuitBreakerPolicy
}

// CircuitSnapshot is an instantaneous view of one configured circuit.
type CircuitSnapshot struct {
	Provider            ID
	Operation           Operation
	Model               string
	State               CircuitState
	ConsecutiveFailures uint
	HalfOpenInFlight    uint
	OpenedAt            time.Time
	NextProbeAt         time.Time
}
