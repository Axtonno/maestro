package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type resilienceKey struct {
	providerID pkgProvider.ID
	operation  pkgProvider.Operation
	model      string
}

type resilienceClock interface {
	Now() time.Time
	Wait(context.Context, time.Duration) error
}

type realResilienceClock struct{}

func (realResilienceClock) Now() time.Time { return time.Now() }

func (realResilienceClock) Wait(
	ctx context.Context,
	duration time.Duration,
) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type resilienceJitter interface {
	Float64() float64
}

type lockedResilienceJitter struct {
	mu     sync.Mutex
	random *rand.Rand
}

func newLockedResilienceJitter() *lockedResilienceJitter {
	return &lockedResilienceJitter{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (j *lockedResilienceJitter) Float64() float64 {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.random.Float64()
}

type circuitBreaker struct {
	mu sync.Mutex

	state               pkgProvider.CircuitState
	consecutiveFailures uint
	halfOpenInFlight    uint
	openedAt            time.Time
	generation          uint64
}

type circuitPermit struct {
	breaker    *circuitBreaker
	generation uint64
	halfOpen   bool
}

type resilienceExecution struct {
	runtime *runtime
	policy  pkgProvider.ResiliencePolicy
	permit  circuitPermit
	started time.Time

	finishOnce sync.Once
}

func (r *runtime) SetResiliencePolicy(
	ctx context.Context,
	providerID pkgProvider.ID,
	policy pkgProvider.ResiliencePolicy,
) error {
	if ctx == nil {
		return fmt.Errorf(
			"set provider resilience policy: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	policy = normalizeResiliencePolicy(policy)
	if err := validateResiliencePolicy(policy); err != nil {
		return err
	}

	selected, err := r.Resolve(providerID)
	if err != nil {
		return fmt.Errorf(
			"set resilience policy with provider %q: %w",
			providerID,
			err,
		)
	}

	key := resilienceKey{
		providerID: selected.ID(),
		operation:  policy.Operation,
		model:      policy.Model,
	}
	r.resilienceMu.Lock()
	r.resiliencePolicies[key] = policy
	r.circuits[key] = &circuitBreaker{state: pkgProvider.CircuitStateClosed}
	r.resilienceMu.Unlock()

	return nil
}

func (r *runtime) ResiliencePolicy(
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	model string,
) (pkgProvider.ResiliencePolicy, bool, error) {
	selected, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.ResiliencePolicy{}, false, fmt.Errorf(
			"get resilience policy with provider %q: %w",
			providerID,
			err,
		)
	}
	if !validResilienceTarget(operation, model) {
		return pkgProvider.ResiliencePolicy{}, false, fmt.Errorf(
			"get resilience policy for operation %q model %q: %w",
			operation,
			model,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}

	_, policy, _, ok := r.lookupResilience(selected.ID(), operation, model)

	return policy, ok, nil
}

func (r *runtime) CircuitState(
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	model string,
) (pkgProvider.CircuitSnapshot, bool, error) {
	selected, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.CircuitSnapshot{}, false, fmt.Errorf(
			"get circuit state with provider %q: %w",
			providerID,
			err,
		)
	}
	if !validResilienceTarget(operation, model) {
		return pkgProvider.CircuitSnapshot{}, false, fmt.Errorf(
			"get circuit state for operation %q model %q: %w",
			operation,
			model,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}

	key, policy, breaker, ok := r.lookupResilience(
		selected.ID(),
		operation,
		model,
	)
	if !ok || policy.CircuitBreaker.FailureThreshold == 0 {
		return pkgProvider.CircuitSnapshot{}, false, nil
	}

	return breaker.snapshot(key, policy.CircuitBreaker), true, nil
}

func (r *runtime) beginResilience(
	ctx context.Context,
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	model string,
) (*resilienceExecution, error) {
	_, policy, breaker, ok := r.lookupResilience(
		providerID,
		operation,
		model,
	)
	if !ok {
		return nil, nil
	}
	if ctx == nil {
		return nil, fmt.Errorf(
			"execute resilient provider operation: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	permit, allowed := breaker.allow(r.resilienceClock.Now(), policy.CircuitBreaker)
	if !allowed {
		return nil, pkgProvider.NewProviderError(
			pkgProvider.ProviderErrorDetails{
				Kind:      pkgProvider.ErrorKindUnavailable,
				Operation: operation,
				Provider:  providerID,
				Model:     model,
				Retryable: true,
				Message:   "circuit breaker is open",
			},
			pkgProvider.ErrCircuitOpen,
		)
	}

	return &resilienceExecution{
		runtime: r,
		policy:  policy,
		permit:  permit,
		started: r.resilienceClock.Now(),
	}, nil
}

func (r *runtime) lookupResilience(
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	model string,
) (
	resilienceKey,
	pkgProvider.ResiliencePolicy,
	*circuitBreaker,
	bool,
) {
	r.resilienceMu.RLock()
	defer r.resilienceMu.RUnlock()

	key := resilienceKey{providerID: providerID, operation: operation, model: model}
	policy, ok := r.resiliencePolicies[key]
	if !ok && model != "" {
		key.model = ""
		policy, ok = r.resiliencePolicies[key]
	}
	if !ok {
		return resilienceKey{}, pkgProvider.ResiliencePolicy{}, nil, false
	}

	return key, policy, r.circuits[key], true
}

func executeWithResilience[T any](
	ctx context.Context,
	execution *resilienceExecution,
	call func() (T, error),
) (T, error) {
	if execution == nil {
		return call()
	}

	var value T
	var err error
	for attempt := uint(1); attempt <= execution.policy.MaxAttempts; attempt++ {
		value, err = call()
		if err == nil || !execution.canRetry(attempt, err) {
			break
		}
		if waitError := execution.waitBeforeRetry(ctx, attempt); waitError != nil {
			if !errors.Is(waitError, errRetryBudgetExhausted) {
				err = waitError
			}
			break
		}
	}
	execution.finish(err)

	return value, err
}

func executeErrorWithResilience(
	ctx context.Context,
	execution *resilienceExecution,
	call func() error,
) error {
	_, err := executeWithResilience(ctx, execution, func() (struct{}, error) {
		return struct{}{}, call()
	})

	return err
}

func (e *resilienceExecution) canRetry(attempt uint, err error) bool {
	return attempt < e.policy.MaxAttempts &&
		operationIsRetryable(e.policy.Operation) &&
		errorIsRetryable(err)
}

func (e *resilienceExecution) waitBeforeRetry(
	ctx context.Context,
	attempt uint,
) error {
	delay := retryBackoff(e.policy, attempt, e.runtime.resilienceJitter)
	if e.policy.MaxElapsedTime > 0 &&
		e.runtime.resilienceClock.Now().Sub(e.started)+delay > e.policy.MaxElapsedTime {
		return errRetryBudgetExhausted
	}

	return e.runtime.resilienceClock.Wait(ctx, delay)
}

func (e *resilienceExecution) finish(err error) {
	if e == nil {
		return
	}

	e.finishOnce.Do(func() {
		e.permit.breaker.finish(
			e.permit,
			circuitOutcomeForError(err),
			e.runtime.resilienceClock.Now(),
			e.policy.CircuitBreaker,
		)
	})
}

var errRetryBudgetExhausted = errors.New("provider retry time budget exhausted")

type circuitOutcome uint8

const (
	circuitOutcomeSuccess circuitOutcome = iota
	circuitOutcomeFailure
	circuitOutcomeNeutral
)

func circuitOutcomeForError(err error) circuitOutcome {
	if err == nil || errors.Is(err, io.EOF) {
		return circuitOutcomeSuccess
	}
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, errRetryBudgetExhausted) {
		return circuitOutcomeNeutral
	}
	if errorIsRetryable(err) {
		return circuitOutcomeFailure
	}

	return circuitOutcomeSuccess
}

func errorIsRetryable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var providerError *pkgProvider.ProviderError

	return errors.As(err, &providerError) && providerError.Retryable
}

func retryBackoff(
	policy pkgProvider.ResiliencePolicy,
	failedAttempt uint,
	jitter resilienceJitter,
) time.Duration {
	factor := math.Pow(policy.BackoffMultiplier, float64(failedAttempt-1))
	scaled := float64(policy.InitialBackoff) * factor
	delay := policy.MaxBackoff
	if !math.IsInf(scaled, 1) && scaled < float64(policy.MaxBackoff) {
		delay = time.Duration(scaled)
	}
	if policy.Jitter == 0 {
		return delay
	}

	random := jitter.Float64()
	multiplier := 1 + ((2*random)-1)*policy.Jitter
	delay = time.Duration(float64(delay) * multiplier)
	if delay < 0 {
		return 0
	}
	if delay > policy.MaxBackoff {
		return policy.MaxBackoff
	}

	return delay
}

func validateResiliencePolicy(policy pkgProvider.ResiliencePolicy) error {
	if !validResilienceTarget(policy.Operation, policy.Model) {
		return fmt.Errorf(
			"set resilience policy for operation %q model %q: invalid target: %w",
			policy.Operation,
			policy.Model,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}
	if policy.MaxAttempts == 0 {
		return fmt.Errorf(
			"set resilience policy for operation %q: max attempts must be positive: %w",
			policy.Operation,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}
	if policy.MaxAttempts > 1 && !operationIsRetryable(policy.Operation) {
		return fmt.Errorf(
			"set resilience policy for operation %q: operation is not automatically retryable: %w",
			policy.Operation,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}
	if policy.InitialBackoff < 0 || policy.MaxBackoff < 0 ||
		policy.MaxElapsedTime < 0 || policy.BackoffMultiplier < 1 ||
		policy.Jitter < 0 || policy.Jitter > 1 {
		return fmt.Errorf(
			"set resilience policy for operation %q: invalid retry bounds: %w",
			policy.Operation,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}
	if policy.MaxAttempts > 1 && (policy.InitialBackoff <= 0 ||
		policy.MaxBackoff < policy.InitialBackoff) {
		return fmt.Errorf(
			"set resilience policy for operation %q: retry backoff is invalid: %w",
			policy.Operation,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}
	breaker := policy.CircuitBreaker
	if breaker.FailureThreshold == 0 {
		if breaker.OpenDuration != 0 || breaker.HalfOpenMaxAttempts != 0 {
			return fmt.Errorf(
				"set resilience policy for operation %q: disabled circuit has settings: %w",
				policy.Operation,
				pkgProvider.ErrInvalidResiliencePolicy,
			)
		}

		return nil
	}
	if breaker.OpenDuration <= 0 || breaker.HalfOpenMaxAttempts == 0 {
		return fmt.Errorf(
			"set resilience policy for operation %q: circuit bounds are invalid: %w",
			policy.Operation,
			pkgProvider.ErrInvalidResiliencePolicy,
		)
	}

	return nil
}

func normalizeResiliencePolicy(
	policy pkgProvider.ResiliencePolicy,
) pkgProvider.ResiliencePolicy {
	if policy.MaxAttempts == 1 && policy.BackoffMultiplier == 0 {
		policy.BackoffMultiplier = 1
	}

	return policy
}

func validResilienceTarget(operation pkgProvider.Operation, model string) bool {
	if model != "" && !validModelID(model) {
		return false
	}
	if model != "" && (operation == pkgProvider.OperationModelListing ||
		operation == pkgProvider.OperationModelDiscovery) {
		return false
	}

	switch operation {
	case pkgProvider.OperationCompletion,
		pkgProvider.OperationStreaming,
		pkgProvider.OperationEmbedding,
		pkgProvider.OperationModelListing,
		pkgProvider.OperationModelDiscovery,
		pkgProvider.OperationModelLoad,
		pkgProvider.OperationModelUnload,
		pkgProvider.OperationModelPull,
		pkgProvider.OperationModelRemove,
		pkgProvider.OperationCapabilityIntrospection:
		return true
	default:
		return false
	}
}

func operationIsRetryable(operation pkgProvider.Operation) bool {
	switch operation {
	case pkgProvider.OperationCompletion,
		pkgProvider.OperationStreaming,
		pkgProvider.OperationEmbedding,
		pkgProvider.OperationModelListing,
		pkgProvider.OperationModelDiscovery,
		pkgProvider.OperationModelLoad,
		pkgProvider.OperationModelUnload,
		pkgProvider.OperationCapabilityIntrospection:
		return true
	default:
		return false
	}
}

func (b *circuitBreaker) allow(
	now time.Time,
	policy pkgProvider.CircuitBreakerPolicy,
) (circuitPermit, bool) {
	if policy.FailureThreshold == 0 {
		return circuitPermit{breaker: b}, true
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == "" {
		b.state = pkgProvider.CircuitStateClosed
	}
	if b.state == pkgProvider.CircuitStateOpen {
		if now.Before(b.openedAt.Add(policy.OpenDuration)) {
			return circuitPermit{}, false
		}
		b.state = pkgProvider.CircuitStateHalfOpen
		b.halfOpenInFlight = 0
		b.generation++
	}
	if b.state == pkgProvider.CircuitStateHalfOpen {
		if b.halfOpenInFlight >= policy.HalfOpenMaxAttempts {
			return circuitPermit{}, false
		}
		b.halfOpenInFlight++

		return circuitPermit{
			breaker: b, generation: b.generation, halfOpen: true,
		}, true
	}

	return circuitPermit{breaker: b, generation: b.generation}, true
}

func (b *circuitBreaker) finish(
	permit circuitPermit,
	outcome circuitOutcome,
	now time.Time,
	policy pkgProvider.CircuitBreakerPolicy,
) {
	if policy.FailureThreshold == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if permit.generation != b.generation {
		return
	}
	if permit.halfOpen && b.halfOpenInFlight > 0 {
		b.halfOpenInFlight--
	}

	switch outcome {
	case circuitOutcomeNeutral:
		return
	case circuitOutcomeSuccess:
		wasHalfOpen := b.state == pkgProvider.CircuitStateHalfOpen
		b.state = pkgProvider.CircuitStateClosed
		b.consecutiveFailures = 0
		b.openedAt = time.Time{}
		b.halfOpenInFlight = 0
		if wasHalfOpen {
			b.generation++
		}
	case circuitOutcomeFailure:
		if b.state == pkgProvider.CircuitStateHalfOpen {
			b.open(now)

			return
		}
		b.consecutiveFailures++
		if b.consecutiveFailures >= policy.FailureThreshold {
			b.open(now)
		}
	}
}

func (b *circuitBreaker) open(now time.Time) {
	b.state = pkgProvider.CircuitStateOpen
	b.openedAt = now
	b.halfOpenInFlight = 0
	b.generation++
}

func (b *circuitBreaker) snapshot(
	key resilienceKey,
	policy pkgProvider.CircuitBreakerPolicy,
) pkgProvider.CircuitSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	snapshot := pkgProvider.CircuitSnapshot{
		Provider:            key.providerID,
		Operation:           key.operation,
		Model:               key.model,
		State:               b.state,
		ConsecutiveFailures: b.consecutiveFailures,
		HalfOpenInFlight:    b.halfOpenInFlight,
		OpenedAt:            b.openedAt,
	}
	if snapshot.State == "" {
		snapshot.State = pkgProvider.CircuitStateClosed
	}
	if snapshot.State == pkgProvider.CircuitStateOpen {
		snapshot.NextProbeAt = b.openedAt.Add(policy.OpenDuration)
	}

	return snapshot
}
