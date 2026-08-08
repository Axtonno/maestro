package provider

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type observerHolder struct {
	observer pkgProvider.ProviderObserver
}

type operationObservation struct {
	observer  pkgProvider.ProviderObserver
	clock     resilienceClock
	id        uint64
	provider  pkgProvider.ID
	operation pkgProvider.Operation
	model     string
	started   time.Time

	attempts    atomic.Uint64
	maxAttempts atomic.Uint64

	dataMu         sync.Mutex
	usage          pkgProvider.Usage
	completedBytes int64
	totalBytes     int64

	finished atomic.Bool
}

func (r *runtime) SetObserver(observer pkgProvider.ProviderObserver) {
	if nilInterface(observer) {
		r.observer.Store(nil)

		return
	}

	r.observer.Store(&observerHolder{observer: observer})
}

func (r *runtime) startProviderOperation(
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	model string,
) *operationObservation {
	holder := r.observer.Load()
	if holder == nil {
		return nil
	}

	started := r.resilienceClock.Now()
	operationID := r.nextOperationID.Add(1)
	if operationID == 0 {
		operationID = r.nextOperationID.Add(1)
	}
	observation := &operationObservation{
		observer:  holder.observer,
		clock:     r.resilienceClock,
		id:        operationID,
		provider:  providerID,
		operation: operation,
		model:     model,
		started:   started,
	}
	observation.maxAttempts.Store(1)
	observation.emit(pkgProvider.ProviderEvent{
		Kind:      pkgProvider.ProviderEventOperationStarted,
		Timestamp: started,
	})

	return observation
}

func (o *operationObservation) setMaxAttempts(maxAttempts uint) {
	if o == nil || maxAttempts == 0 {
		return
	}

	o.maxAttempts.Store(uint64(maxAttempts))
}

func (o *operationObservation) attempt(attempt uint) {
	if o == nil {
		return
	}

	o.attempts.Store(uint64(attempt))
	o.emit(pkgProvider.ProviderEvent{
		Kind:      pkgProvider.ProviderEventAttemptStarted,
		Timestamp: o.clock.Now(),
		Attempt:   attempt,
	})
}

func (o *operationObservation) retry(
	failedAttempt uint,
	backoff time.Duration,
	err error,
) {
	if o == nil {
		return
	}

	event := pkgProvider.ProviderEvent{
		Kind:      pkgProvider.ProviderEventRetryScheduled,
		Timestamp: o.clock.Now(),
		Attempt:   failedAttempt + 1,
		Backoff:   backoff,
	}
	setProviderEventError(&event, err)
	o.emit(event)
}

func (o *operationObservation) circuitTransition(
	from pkgProvider.CircuitState,
	to pkgProvider.CircuitState,
) {
	if o == nil || from == to {
		return
	}

	o.emit(pkgProvider.ProviderEvent{
		Kind:        pkgProvider.ProviderEventCircuitTransition,
		Timestamp:   o.clock.Now(),
		Attempt:     uint(o.attempts.Load()),
		CircuitFrom: from,
		CircuitTo:   to,
	})
}

func (o *operationObservation) recordUsage(usage pkgProvider.Usage) {
	if o == nil {
		return
	}

	o.dataMu.Lock()
	if usage.InputTokens > o.usage.InputTokens {
		o.usage.InputTokens = usage.InputTokens
	}
	if usage.OutputTokens > o.usage.OutputTokens {
		o.usage.OutputTokens = usage.OutputTokens
	}
	o.dataMu.Unlock()
}

func (o *operationObservation) recordProgress(
	completedBytes int64,
	totalBytes int64,
) {
	if o == nil {
		return
	}

	o.dataMu.Lock()
	if completedBytes > o.completedBytes {
		o.completedBytes = completedBytes
	}
	if totalBytes > o.totalBytes {
		o.totalBytes = totalBytes
	}
	o.dataMu.Unlock()
}

func (o *operationObservation) finish(err error) {
	o.finishWithOutcome(err, false)
}

func (o *operationObservation) finishClosed(err error) {
	o.finishWithOutcome(err, true)
}

func (o *operationObservation) finishWithOutcome(err error, closed bool) {
	if o == nil {
		return
	}

	if !o.finished.CompareAndSwap(false, true) {
		return
	}
	now := o.clock.Now()
	duration := now.Sub(o.started)
	if duration < 0 {
		duration = 0
	}

	o.dataMu.Lock()
	usage := o.usage
	completedBytes := o.completedBytes
	totalBytes := o.totalBytes
	o.dataMu.Unlock()

	event := pkgProvider.ProviderEvent{
		Kind:           pkgProvider.ProviderEventOperationCompleted,
		Timestamp:      now,
		Duration:       duration,
		Attempt:        uint(o.attempts.Load()),
		Usage:          usage,
		CompletedBytes: completedBytes,
		TotalBytes:     totalBytes,
	}
	if err == nil || errors.Is(err, io.EOF) {
		if closed {
			event.Outcome = pkgProvider.ProviderOperationOutcomeClosed
		} else {
			event.Outcome = pkgProvider.ProviderOperationOutcomeSuccess
		}
	} else {
		setProviderEventError(&event, err)
	}
	o.emit(event)
}

func setProviderEventError(event *pkgProvider.ProviderEvent, err error) {
	event.Outcome = pkgProvider.ProviderOperationOutcomeError
	var providerError *pkgProvider.ProviderError
	if errors.As(err, &providerError) {
		event.ErrorKind = providerError.Kind
		event.StatusCode = providerError.StatusCode
		event.Retryable = providerError.Retryable
	}
	if event.ErrorKind == "" {
		switch {
		case errors.Is(err, pkgProvider.ErrInvalidRequest):
			event.ErrorKind = pkgProvider.ErrorKindInvalidRequest
		case errors.Is(err, pkgProvider.ErrInvalidResponse),
			errors.Is(err, pkgProvider.ErrInvalidStream):
			event.ErrorKind = pkgProvider.ErrorKindInvalidResponse
		case errors.Is(err, pkgProvider.ErrAuthentication):
			event.ErrorKind = pkgProvider.ErrorKindAuthentication
		case errors.Is(err, pkgProvider.ErrModelNotFound):
			event.ErrorKind = pkgProvider.ErrorKindModelNotFound
		case errors.Is(err, pkgProvider.ErrUnsupportedCapability),
			errors.Is(err, pkgProvider.ErrCapabilityNotFound):
			event.ErrorKind = pkgProvider.ErrorKindCapabilityNotFound
		case errors.Is(err, pkgProvider.ErrProviderUnavailable),
			errors.Is(err, pkgProvider.ErrResidencyShuttingDown):
			event.ErrorKind = pkgProvider.ErrorKindUnavailable
		case errors.Is(err, pkgProvider.ErrCapacityExhausted):
			event.ErrorKind = pkgProvider.ErrorKindCapacityExhausted
		case errors.Is(err, pkgProvider.ErrRateLimited):
			event.ErrorKind = pkgProvider.ErrorKindRateLimited
		case errors.Is(err, pkgProvider.ErrTransient):
			event.ErrorKind = pkgProvider.ErrorKindTransient
		case errors.Is(err, pkgProvider.ErrProviderInternal):
			event.ErrorKind = pkgProvider.ErrorKindInternal
		}
	}
	event.CircuitOpen = errors.Is(err, pkgProvider.ErrCircuitOpen)

	switch {
	case errors.Is(err, context.Canceled):
		event.Outcome = pkgProvider.ProviderOperationOutcomeCanceled
		event.ErrorKind = pkgProvider.ErrorKindCanceled
		event.Retryable = false
	case errors.Is(err, context.DeadlineExceeded):
		event.Outcome = pkgProvider.ProviderOperationOutcomeDeadlineExceeded
		event.ErrorKind = pkgProvider.ErrorKindDeadlineExceeded
		event.Retryable = false
	}
}

func (o *operationObservation) emit(event pkgProvider.ProviderEvent) {
	event.OperationID = o.id
	event.Provider = o.provider
	event.Operation = o.operation
	event.Model = o.model
	event.MaxAttempts = uint(o.maxAttempts.Load())

	defer func() {
		_ = recover()
	}()
	_ = o.observer.ObserveProviderEvent(event)
}
