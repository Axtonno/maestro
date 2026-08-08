package provider

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type fakeResilienceClock struct {
	mu    sync.Mutex
	now   time.Time
	waits []time.Duration
}

func (c *fakeResilienceClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *fakeResilienceClock) Wait(
	ctx context.Context,
	duration time.Duration,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	c.waits = append(c.waits, duration)
	c.now = c.now.Add(duration)
	c.mu.Unlock()

	return nil
}

func (c *fakeResilienceClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	c.mu.Unlock()
}

func (c *fakeResilienceClock) Waits() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]time.Duration(nil), c.waits...)
}

type fixedResilienceJitter float64

func (j fixedResilienceJitter) Float64() float64 { return float64(j) }

type resilienceProvider struct {
	identityProvider

	mu sync.Mutex

	completionErrors []error
	completionCalls  int
	completionStart  chan struct{}
	completionBlock  <-chan struct{}
	cancel           context.CancelFunc

	streams      []pkgProvider.Stream
	streamErrors []error
	streamCalls  int
}

func (p *resilienceProvider) Complete(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	p.mu.Lock()
	index := p.completionCalls
	p.completionCalls++
	var err error
	if index < len(p.completionErrors) {
		err = p.completionErrors[index]
	}
	started := p.completionStart
	block := p.completionBlock
	cancel := p.cancel
	p.mu.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if cancel != nil {
		cancel()
	}
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return pkgProvider.CompletionResponse{}, ctx.Err()
		}
	}

	return pkgProvider.CompletionResponse{Model: request.Model}, err
}

func (p *resilienceProvider) Stream(
	context.Context,
	pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	index := p.streamCalls
	p.streamCalls++
	var stream pkgProvider.Stream
	if index < len(p.streams) {
		stream = p.streams[index]
	}
	var err error
	if index < len(p.streamErrors) {
		err = p.streamErrors[index]
	}

	return stream, err
}

func (p *resilienceProvider) CompletionCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.completionCalls
}

func (p *resilienceProvider) StreamCalls() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.streamCalls
}

type scriptedStream struct {
	mu         sync.Mutex
	results    []streamResult
	index      int
	closeCount int
}

type streamResult struct {
	chunk pkgProvider.StreamChunk
	err   error
}

func (s *scriptedStream) Recv() (pkgProvider.StreamChunk, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.index >= len(s.results) {
		return pkgProvider.StreamChunk{}, io.EOF
	}

	result := s.results[s.index]
	s.index++

	return result.chunk, result.err
}

func (s *scriptedStream) Close() error {
	s.mu.Lock()
	s.closeCount++
	s.mu.Unlock()

	return nil
}

func (s *scriptedStream) Closes() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closeCount
}

func TestResiliencePolicyIsOptInAndValidated(t *testing.T) {
	providerRuntime, _, registered := newResilienceTestRuntime(t)

	_, err := providerRuntime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil || registered.CompletionCalls() != 1 {
		t.Fatalf("default path changed: calls=%d error=%v", registered.CompletionCalls(), err)
	}
	if _, found, err := providerRuntime.ResiliencePolicy(
		"test", pkgProvider.OperationCompletion, "qwen",
	); err != nil || found {
		t.Fatalf("expected absent policy, found=%v error=%v", found, err)
	}

	err = providerRuntime.SetResiliencePolicy(
		context.Background(),
		"test",
		pkgProvider.ResiliencePolicy{
			Operation:         pkgProvider.OperationModelPull,
			MaxAttempts:       2,
			InitialBackoff:    time.Millisecond,
			MaxBackoff:        time.Millisecond,
			BackoffMultiplier: 1,
		},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidResiliencePolicy) {
		t.Fatalf("expected non-idempotent retry rejection, got %v", err)
	}

	policy := pkgProvider.ResiliencePolicy{
		Operation:   pkgProvider.OperationCompletion,
		MaxAttempts: 1,
		CircuitBreaker: pkgProvider.CircuitBreakerPolicy{
			FailureThreshold: 2, OpenDuration: time.Minute,
			HalfOpenMaxAttempts: 1,
		},
	}
	if err := providerRuntime.SetResiliencePolicy(
		context.Background(), "test", policy,
	); err != nil {
		t.Fatalf("set circuit-only policy: %v", err)
	}
	stored, found, err := providerRuntime.ResiliencePolicy(
		"test", pkgProvider.OperationCompletion, "qwen",
	)
	if err != nil || !found || stored.BackoffMultiplier != 1 {
		t.Fatalf("unexpected normalized policy %#v, found=%v error=%v", stored, found, err)
	}
	exact := policy
	exact.Model = "qwen"
	exact.CircuitBreaker.FailureThreshold = 3
	if err := providerRuntime.SetResiliencePolicy(
		context.Background(), "test", exact,
	); err != nil {
		t.Fatalf("set model-specific policy: %v", err)
	}
	stored, found, err = providerRuntime.ResiliencePolicy(
		"test", pkgProvider.OperationCompletion, "qwen",
	)
	if err != nil || !found || stored.Model != "qwen" ||
		stored.CircuitBreaker.FailureThreshold != 3 {
		t.Fatalf("model policy did not take precedence: %#v, found=%v error=%v",
			stored, found, err)
	}
}

func TestOperationRetryabilityMatrix(t *testing.T) {
	retryable := []pkgProvider.Operation{
		pkgProvider.OperationCompletion,
		pkgProvider.OperationStreaming,
		pkgProvider.OperationEmbedding,
		pkgProvider.OperationModelListing,
		pkgProvider.OperationModelDiscovery,
		pkgProvider.OperationModelLoad,
		pkgProvider.OperationModelUnload,
		pkgProvider.OperationCapabilityIntrospection,
	}
	for _, operation := range retryable {
		if !operationIsRetryable(operation) {
			t.Fatalf("expected %q to be retryable", operation)
		}
	}
	for _, operation := range []pkgProvider.Operation{
		pkgProvider.OperationModelPull,
		pkgProvider.OperationModelRemove,
		pkgProvider.Operation("unknown"),
	} {
		if operationIsRetryable(operation) {
			t.Fatalf("expected %q not to be retryable", operation)
		}
	}
}

func TestResilienceRetriesWithDeterministicBackoff(t *testing.T) {
	providerRuntime, clock, registered := newResilienceTestRuntime(t)
	registered.completionErrors = []error{
		testRetryableProviderError(),
		testRetryableProviderError(),
		nil,
	}
	setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
		Operation:         pkgProvider.OperationCompletion,
		Model:             "qwen",
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        50 * time.Millisecond,
		BackoffMultiplier: 2,
	})

	response, err := providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil || response.Model != "qwen" {
		t.Fatalf("retry completion: response=%#v error=%v", response, err)
	}
	if registered.CompletionCalls() != 3 {
		t.Fatalf("expected three attempts, got %d", registered.CompletionCalls())
	}
	waits := clock.Waits()
	if len(waits) != 2 || waits[0] != 10*time.Millisecond ||
		waits[1] != 20*time.Millisecond {
		t.Fatalf("unexpected backoff sequence: %v", waits)
	}
}

func TestResilienceDoesNotRetryPermanentOrCanceledFailures(t *testing.T) {
	providerRuntime, _, registered := newResilienceTestRuntime(t)
	registered.completionErrors = []error{pkgProvider.NewProviderError(
		pkgProvider.ProviderErrorDetails{
			Kind: pkgProvider.ErrorKindInvalidRequest,
		},
		pkgProvider.ErrInvalidRequest,
	)}
	setCompletionResiliencePolicy(t, providerRuntime, testRetryPolicy())

	_, err := providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidRequest) ||
		registered.CompletionCalls() != 1 {
		t.Fatalf("permanent error was retried: calls=%d error=%v", registered.CompletionCalls(), err)
	}

	providerRuntime, _, registered = newResilienceTestRuntime(t)
	ctx, cancel := context.WithCancel(context.Background())
	registered.cancel = cancel
	registered.completionErrors = []error{testRetryableProviderError()}
	setCompletionResiliencePolicy(t, providerRuntime, testRetryPolicy())
	_, err = providerRuntime.Complete(
		ctx, "test", pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if !errors.Is(err, context.Canceled) || registered.CompletionCalls() != 1 {
		t.Fatalf("context did not stop retry: calls=%d error=%v", registered.CompletionCalls(), err)
	}
}

func TestResilienceHonorsElapsedBudgetAndJitterBounds(t *testing.T) {
	providerRuntime, clock, registered := newResilienceTestRuntime(t)
	registered.completionErrors = []error{
		testRetryableProviderError(),
		testRetryableProviderError(),
		testRetryableProviderError(),
	}
	setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
		Operation:         pkgProvider.OperationCompletion,
		Model:             "qwen",
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        40 * time.Millisecond,
		BackoffMultiplier: 2,
		MaxElapsedTime:    15 * time.Millisecond,
	})

	_, err := providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if !errorIsRetryable(err) || registered.CompletionCalls() != 2 {
		t.Fatalf("elapsed budget was not enforced: calls=%d error=%v",
			registered.CompletionCalls(), err)
	}
	if waits := clock.Waits(); len(waits) != 1 || waits[0] != 10*time.Millisecond {
		t.Fatalf("unexpected elapsed-budget waits: %v", waits)
	}

	policy := pkgProvider.ResiliencePolicy{
		InitialBackoff:    20 * time.Millisecond,
		MaxBackoff:        30 * time.Millisecond,
		BackoffMultiplier: 2,
		Jitter:            0.5,
	}
	if delay := retryBackoff(policy, 1, fixedResilienceJitter(0)); delay != 10*time.Millisecond {
		t.Fatalf("unexpected lower jitter bound: %v", delay)
	}
	if delay := retryBackoff(policy, 1, fixedResilienceJitter(1)); delay != 30*time.Millisecond {
		t.Fatalf("unexpected capped upper jitter bound: %v", delay)
	}
}

func TestCircuitBreakerOpensAndRecoversThroughHalfOpen(t *testing.T) {
	providerRuntime, clock, registered := newResilienceTestRuntime(t)
	registered.completionErrors = []error{
		testRetryableProviderError(),
		testRetryableProviderError(),
		nil,
	}
	setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
		Operation:   pkgProvider.OperationCompletion,
		Model:       "qwen",
		MaxAttempts: 1,
		CircuitBreaker: pkgProvider.CircuitBreakerPolicy{
			FailureThreshold:    2,
			OpenDuration:        time.Minute,
			HalfOpenMaxAttempts: 1,
		},
	})

	for range 2 {
		_, _ = providerRuntime.Complete(
			context.Background(), "test",
			pkgProvider.CompletionRequest{Model: "qwen"},
		)
	}
	snapshot := requireCircuitSnapshot(t, providerRuntime)
	if snapshot.State != pkgProvider.CircuitStateOpen ||
		snapshot.ConsecutiveFailures != 2 || snapshot.NextProbeAt.IsZero() {
		t.Fatalf("unexpected open circuit: %#v", snapshot)
	}

	_, err := providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if !errors.Is(err, pkgProvider.ErrCircuitOpen) ||
		registered.CompletionCalls() != 2 {
		t.Fatalf("open circuit reached provider: calls=%d error=%v", registered.CompletionCalls(), err)
	}

	clock.Advance(time.Minute)
	_, err = providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil {
		t.Fatalf("half-open probe failed: %v", err)
	}
	snapshot = requireCircuitSnapshot(t, providerRuntime)
	if snapshot.State != pkgProvider.CircuitStateClosed ||
		snapshot.ConsecutiveFailures != 0 {
		t.Fatalf("circuit did not recover: %#v", snapshot)
	}
}

func TestCircuitBreakerBoundsConcurrentHalfOpenProbes(t *testing.T) {
	providerRuntime, clock, registered := newResilienceTestRuntime(t)
	registered.completionErrors = []error{testRetryableProviderError(), nil}
	setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
		Operation:   pkgProvider.OperationCompletion,
		Model:       "qwen",
		MaxAttempts: 1,
		CircuitBreaker: pkgProvider.CircuitBreakerPolicy{
			FailureThreshold:    1,
			OpenDuration:        time.Minute,
			HalfOpenMaxAttempts: 1,
		},
	})
	_, _ = providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	clock.Advance(time.Minute)

	started := make(chan struct{}, 1)
	allow := make(chan struct{})
	registered.mu.Lock()
	registered.completionStart = started
	registered.completionBlock = allow
	registered.mu.Unlock()
	probeResult := make(chan error, 1)
	go func() {
		_, err := providerRuntime.Complete(
			context.Background(), "test",
			pkgProvider.CompletionRequest{Model: "qwen"},
		)
		probeResult <- err
	}()
	<-started
	snapshot := requireCircuitSnapshot(t, providerRuntime)
	if snapshot.State != pkgProvider.CircuitStateHalfOpen ||
		snapshot.HalfOpenInFlight != 1 {
		t.Fatalf("unexpected half-open state: %#v", snapshot)
	}

	_, err := providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if !errors.Is(err, pkgProvider.ErrCircuitOpen) {
		t.Fatalf("expected bounded half-open rejection, got %v", err)
	}
	close(allow)
	if err := <-probeResult; err != nil {
		t.Fatalf("half-open probe: %v", err)
	}
	if registered.CompletionCalls() != 2 {
		t.Fatalf("unexpected provider calls: %d", registered.CompletionCalls())
	}
}

func TestStreamRetriesOnlyBeforeFirstDeliveredChunk(t *testing.T) {
	providerRuntime, clock, registered := newResilienceTestRuntime(t)
	first := &scriptedStream{results: []streamResult{{
		err: testRetryableProviderError(),
	}}}
	second := &scriptedStream{results: []streamResult{
		{chunk: pkgProvider.StreamChunk{Content: "ready"}},
		{err: testRetryableProviderError()},
	}}
	registered.streams = []pkgProvider.Stream{first, second}
	setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
		Operation:         pkgProvider.OperationStreaming,
		Model:             "qwen",
		MaxAttempts:       3,
		InitialBackoff:    time.Millisecond,
		MaxBackoff:        time.Millisecond,
		BackoffMultiplier: 1,
		CircuitBreaker: pkgProvider.CircuitBreakerPolicy{
			FailureThreshold:    1,
			OpenDuration:        time.Minute,
			HalfOpenMaxAttempts: 1,
		},
	})

	stream, err := providerRuntime.Stream(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil {
		t.Fatalf("open resilient stream: %v", err)
	}
	chunk, err := stream.Recv()
	if err != nil || chunk.Content != "ready" {
		t.Fatalf("receive retried stream: chunk=%#v error=%v", chunk, err)
	}
	if registered.StreamCalls() != 2 || first.Closes() != 1 ||
		len(clock.Waits()) != 1 {
		t.Fatalf("stream was not reopened once: calls=%d closes=%d waits=%v",
			registered.StreamCalls(), first.Closes(), clock.Waits())
	}
	if _, err := stream.Recv(); !errorIsRetryable(err) {
		t.Fatalf("expected post-delivery error, got %v", err)
	}
	if registered.StreamCalls() != 2 {
		t.Fatalf("stream retried after delivery: %d calls", registered.StreamCalls())
	}
	snapshot, found, stateError := providerRuntime.CircuitState(
		"test", pkgProvider.OperationStreaming, "qwen",
	)
	if stateError != nil || !found || snapshot.State != pkgProvider.CircuitStateClosed {
		t.Fatalf("post-delivery error changed circuit: %#v found=%v error=%v",
			snapshot, found, stateError)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}
}

func newResilienceTestRuntime(
	t *testing.T,
) (*runtime, *fakeResilienceClock, *resilienceProvider) {
	t.Helper()
	clock := &fakeResilienceClock{now: time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)}
	providerRuntime := newRuntimeWithDependencies(
		"",
		&fakeResidencyScheduler{},
		clock,
		fixedResilienceJitter(0.5),
	)
	registered := &resilienceProvider{identityProvider: identityProvider{id: "test"}}
	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	return providerRuntime, clock, registered
}

func setCompletionResiliencePolicy(
	t *testing.T,
	providerRuntime *runtime,
	policy pkgProvider.ResiliencePolicy,
) {
	t.Helper()
	if err := providerRuntime.SetResiliencePolicy(
		context.Background(), "test", policy,
	); err != nil {
		t.Fatalf("set resilience policy: %v", err)
	}
}

func testRetryPolicy() pkgProvider.ResiliencePolicy {
	return pkgProvider.ResiliencePolicy{
		Operation:         pkgProvider.OperationCompletion,
		Model:             "qwen",
		MaxAttempts:       3,
		InitialBackoff:    time.Millisecond,
		MaxBackoff:        time.Millisecond,
		BackoffMultiplier: 1,
	}
}

func testRetryableProviderError() error {
	return pkgProvider.NewProviderError(
		pkgProvider.ProviderErrorDetails{
			Kind:      pkgProvider.ErrorKindUnavailable,
			Retryable: true,
		},
		errors.New("temporary provider failure"),
	)
}

func requireCircuitSnapshot(
	t *testing.T,
	providerRuntime *runtime,
) pkgProvider.CircuitSnapshot {
	t.Helper()
	snapshot, found, err := providerRuntime.CircuitState(
		"test", pkgProvider.OperationCompletion, "qwen",
	)
	if err != nil || !found {
		t.Fatalf("get circuit state: found=%v error=%v", found, err)
	}

	return snapshot
}
