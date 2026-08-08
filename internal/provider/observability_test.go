package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type recordingProviderObserver struct {
	mu     sync.Mutex
	events []pkgProvider.ProviderEvent
}

func TestProviderObservationDisabledFastPathDoesNotAllocate(t *testing.T) {
	providerRuntime := NewRuntime("").(*runtime)
	allocations := testing.AllocsPerRun(1000, func() {
		if observation := providerRuntime.startProviderOperation(
			"test", pkgProvider.OperationCompletion, "qwen",
		); observation != nil {
			t.Fatal("unexpected observation without observer")
		}
	})
	if allocations != 0 {
		t.Fatalf("disabled observation allocated %.2f objects", allocations)
	}
	if providerRuntime.nextOperationID.Load() != 0 {
		t.Fatal("disabled observation consumed an operation ID")
	}
}

func (o *recordingProviderObserver) ObserveProviderEvent(
	event pkgProvider.ProviderEvent,
) error {
	o.mu.Lock()
	o.events = append(o.events, event)
	o.mu.Unlock()

	return nil
}

func (o *recordingProviderObserver) Events() []pkgProvider.ProviderEvent {
	o.mu.Lock()
	defer o.mu.Unlock()

	return append([]pkgProvider.ProviderEvent(nil), o.events...)
}

func TestProviderObservationIsOrderedCorrelatedAndRedacted(t *testing.T) {
	providerRuntime, _, registered := newResilienceTestRuntime(t)
	registered.completionResponse = pkgProvider.CompletionResponse{
		Model: "qwen",
		Message: pkgProvider.Message{
			Role: pkgProvider.RoleAssistant, Content: "secret response",
		},
		Usage: pkgProvider.Usage{InputTokens: 7, OutputTokens: 11},
	}
	observer := &recordingProviderObserver{}
	providerRuntime.SetObserver(observer)

	response, err := providerRuntime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{
			Model: "qwen",
			Messages: []pkgProvider.Message{{
				Role: pkgProvider.RoleUser, Content: "secret prompt",
			}},
		},
	)
	if err != nil || response.Message.Content != "secret response" {
		t.Fatalf("complete: response=%#v error=%v", response, err)
	}

	events := observer.Events()
	assertEventKinds(t, events,
		pkgProvider.ProviderEventOperationStarted,
		pkgProvider.ProviderEventAttemptStarted,
		pkgProvider.ProviderEventOperationCompleted,
	)
	operationID := events[0].OperationID
	for _, event := range events {
		if event.OperationID != operationID || operationID == 0 ||
			event.Provider != "test" ||
			event.Operation != pkgProvider.OperationCompletion ||
			event.Model != "qwen" {
			t.Fatalf("uncorrelated event: %#v", event)
		}
	}
	terminal := events[len(events)-1]
	if terminal.Outcome != pkgProvider.ProviderOperationOutcomeSuccess ||
		terminal.Attempt != 1 || terminal.MaxAttempts != 1 ||
		terminal.Usage.InputTokens != 7 || terminal.Usage.OutputTokens != 11 ||
		terminal.Duration < 0 {
		t.Fatalf("unexpected terminal event: %#v", terminal)
	}
	encoded := fmt.Sprintf("%#v", events)
	for _, secret := range []string{"secret prompt", "secret response"} {
		if strings.Contains(encoded, secret) {
			t.Fatalf("observation leaked %q: %s", secret, encoded)
		}
	}

	providerRuntime.SetObserver(nil)
	_, err = providerRuntime.Complete(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil || len(observer.Events()) != len(events) {
		t.Fatalf("cleared observer still received events: %v", err)
	}
}

func TestProviderObservationCorrelatesRetryAndCircuitTransitions(t *testing.T) {
	t.Run("retry", func(t *testing.T) {
		providerRuntime, _, registered := newResilienceTestRuntime(t)
		registered.completionErrors = []error{testRetryableProviderError(), nil}
		observer := &recordingProviderObserver{}
		providerRuntime.SetObserver(observer)
		setCompletionResiliencePolicy(t, providerRuntime, pkgProvider.ResiliencePolicy{
			Operation:         pkgProvider.OperationCompletion,
			Model:             "qwen",
			MaxAttempts:       2,
			InitialBackoff:    5 * time.Millisecond,
			MaxBackoff:        5 * time.Millisecond,
			BackoffMultiplier: 1,
		})

		_, err := providerRuntime.Complete(
			context.Background(), "test",
			pkgProvider.CompletionRequest{Model: "qwen"},
		)
		if err != nil {
			t.Fatalf("retry completion: %v", err)
		}
		events := observer.Events()
		assertEventKinds(t, events,
			pkgProvider.ProviderEventOperationStarted,
			pkgProvider.ProviderEventAttemptStarted,
			pkgProvider.ProviderEventRetryScheduled,
			pkgProvider.ProviderEventAttemptStarted,
			pkgProvider.ProviderEventOperationCompleted,
		)
		if events[2].Attempt != 2 || events[2].Backoff != 5*time.Millisecond ||
			events[2].ErrorKind != pkgProvider.ErrorKindUnavailable ||
			!events[2].Retryable || events[4].Attempt != 2 ||
			events[4].MaxAttempts != 2 {
			t.Fatalf("unexpected retry events: %#v", events)
		}
		assertSingleOperationID(t, events)
	})

	t.Run("circuit", func(t *testing.T) {
		providerRuntime, clock, registered := newResilienceTestRuntime(t)
		registered.completionErrors = []error{testRetryableProviderError(), nil}
		observer := &recordingProviderObserver{}
		providerRuntime.SetObserver(observer)
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
		first := observer.Events()
		assertEventKinds(t, first,
			pkgProvider.ProviderEventOperationStarted,
			pkgProvider.ProviderEventAttemptStarted,
			pkgProvider.ProviderEventCircuitTransition,
			pkgProvider.ProviderEventOperationCompleted,
		)
		if first[2].CircuitFrom != pkgProvider.CircuitStateClosed ||
			first[2].CircuitTo != pkgProvider.CircuitStateOpen {
			t.Fatalf("unexpected opening transition: %#v", first[2])
		}
		assertSingleOperationID(t, first)

		clock.Advance(time.Minute)
		_, err := providerRuntime.Complete(
			context.Background(), "test",
			pkgProvider.CompletionRequest{Model: "qwen"},
		)
		if err != nil {
			t.Fatalf("half-open recovery: %v", err)
		}
		all := observer.Events()
		second := all[len(first):]
		assertEventKinds(t, second,
			pkgProvider.ProviderEventOperationStarted,
			pkgProvider.ProviderEventCircuitTransition,
			pkgProvider.ProviderEventAttemptStarted,
			pkgProvider.ProviderEventCircuitTransition,
			pkgProvider.ProviderEventOperationCompleted,
		)
		if second[1].CircuitFrom != pkgProvider.CircuitStateOpen ||
			second[1].CircuitTo != pkgProvider.CircuitStateHalfOpen ||
			second[3].CircuitFrom != pkgProvider.CircuitStateHalfOpen ||
			second[3].CircuitTo != pkgProvider.CircuitStateClosed {
			t.Fatalf("unexpected recovery transitions: %#v", second)
		}
		assertSingleOperationID(t, second)
		if first[0].OperationID == second[0].OperationID {
			t.Fatal("separate operations reused an operation ID")
		}
	})
}

func TestProviderObservationStreamsHaveOneTerminalEvent(t *testing.T) {
	tests := []struct {
		name    string
		results []streamResult
		consume func(*testing.T, pkgProvider.Stream)
		outcome pkgProvider.ProviderOperationOutcome
		usage   pkgProvider.Usage
	}{
		{
			name: "completed",
			results: []streamResult{
				{chunk: pkgProvider.StreamChunk{
					Content: "secret chunk",
					Usage:   pkgProvider.Usage{InputTokens: 3, OutputTokens: 5},
				}},
				{err: io.EOF},
			},
			consume: func(t *testing.T, stream pkgProvider.Stream) {
				t.Helper()
				if _, err := stream.Recv(); err != nil {
					t.Fatalf("receive chunk: %v", err)
				}
				if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
					t.Fatalf("receive EOF: %v", err)
				}
				_ = stream.Close()
			},
			outcome: pkgProvider.ProviderOperationOutcomeSuccess,
			usage:   pkgProvider.Usage{InputTokens: 3, OutputTokens: 5},
		},
		{
			name: "closed early",
			results: []streamResult{{chunk: pkgProvider.StreamChunk{
				Content: "unobserved secret",
			}}},
			consume: func(t *testing.T, stream pkgProvider.Stream) {
				t.Helper()
				if err := stream.Close(); err != nil {
					t.Fatalf("close stream: %v", err)
				}
				if err := stream.Close(); err != nil {
					t.Fatalf("close stream twice: %v", err)
				}
			},
			outcome: pkgProvider.ProviderOperationOutcomeClosed,
		},
		{
			name:    "canceled",
			results: []streamResult{{err: context.Canceled}},
			consume: func(t *testing.T, stream pkgProvider.Stream) {
				t.Helper()
				if _, err := stream.Recv(); !errors.Is(err, context.Canceled) {
					t.Fatalf("receive cancellation: %v", err)
				}
				_ = stream.Close()
			},
			outcome: pkgProvider.ProviderOperationOutcomeCanceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			providerRuntime, _, registered := newResilienceTestRuntime(t)
			registered.streams = []pkgProvider.Stream{&scriptedStream{
				results: test.results,
			}}
			observer := &recordingProviderObserver{}
			providerRuntime.SetObserver(observer)

			stream, err := providerRuntime.Stream(
				context.Background(), "test",
				pkgProvider.CompletionRequest{
					Model:    "qwen",
					Messages: []pkgProvider.Message{{Content: "secret prompt"}},
				},
			)
			if err != nil {
				t.Fatalf("open stream: %v", err)
			}
			test.consume(t, stream)

			events := observer.Events()
			assertEventKinds(t, events,
				pkgProvider.ProviderEventOperationStarted,
				pkgProvider.ProviderEventAttemptStarted,
				pkgProvider.ProviderEventOperationCompleted,
			)
			terminal := events[2]
			if terminal.Outcome != test.outcome || terminal.Usage != test.usage {
				t.Fatalf("unexpected stream terminal: %#v", terminal)
			}
			if test.outcome == pkgProvider.ProviderOperationOutcomeCanceled &&
				terminal.ErrorKind != pkgProvider.ErrorKindCanceled {
				t.Fatalf("unexpected cancellation classification: %#v", terminal)
			}
			encoded := fmt.Sprintf("%#v", events)
			if strings.Contains(encoded, "secret") {
				t.Fatalf("stream observation leaked content: %s", encoded)
			}
		})
	}
}

func TestProviderObservationIncludesResidencyReleaseFailure(t *testing.T) {
	providerError := pkgProvider.NewProviderError(
		pkgProvider.ProviderErrorDetails{
			Kind:      pkgProvider.ErrorKindUnavailable,
			Operation: pkgProvider.OperationModelUnload,
			Provider:  "test",
			Model:     "qwen",
			Retryable: true,
		},
		pkgProvider.ErrProviderUnavailable,
	)
	registered := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
		stream:           &scriptedStream{results: []streamResult{{err: io.EOF}}},
		unloadError:      providerError,
	}
	providerRuntime := newResidencyRuntime(
		t,
		&fakeResidencyScheduler{},
		registered,
	)
	setResidencyPolicy(t, providerRuntime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true,
	})
	observer := &recordingProviderObserver{}
	providerRuntime.SetObserver(observer)

	stream, err := providerRuntime.Stream(
		context.Background(), "test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil {
		t.Fatalf("open resident stream: %v", err)
	}
	if _, err := stream.Recv(); !errors.Is(err, pkgProvider.ErrProviderUnavailable) {
		t.Fatalf("expected residency release failure, got %v", err)
	}

	events := observer.Events()
	assertEventKinds(t, events,
		pkgProvider.ProviderEventOperationStarted,
		pkgProvider.ProviderEventAttemptStarted,
		pkgProvider.ProviderEventOperationCompleted,
	)
	terminal := events[2]
	if terminal.Outcome != pkgProvider.ProviderOperationOutcomeError ||
		terminal.ErrorKind != pkgProvider.ErrorKindUnavailable ||
		!terminal.Retryable {
		t.Fatalf("unexpected stream terminal after release failure: %#v", terminal)
	}
}

func TestProviderObservationCoversOperationalSurface(t *testing.T) {
	providerRuntime := NewRuntime("").(*runtime)
	registered := &capableProvider{
		identityProvider: identityProvider{id: "all"},
		completionResponse: pkgProvider.CompletionResponse{
			Usage: pkgProvider.Usage{InputTokens: 1, OutputTokens: 2},
		},
		embeddingResponse: pkgProvider.EmbeddingResponse{
			Usage: pkgProvider.Usage{InputTokens: 4},
		},
		stream: &testStream{chunks: []pkgProvider.StreamChunk{{}}},
		pullStream: &testModelPullStream{progress: []pkgProvider.ModelPullProgress{{
			Stage:          pkgProvider.ModelPullStageCompleted,
			CompletedBytes: 8,
			TotalBytes:     10,
		}}},
	}
	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}
	capabilityRequest := pkgProvider.CapabilityRequest{
		Target: pkgProvider.CapabilityTargetAdapter,
	}
	inspector := &capabilityProvider{
		identityProvider: identityProvider{id: "inspector"},
		report: testCapabilityReport(
			"inspector",
			capabilityRequest,
		),
	}
	if err := providerRuntime.Register(inspector); err != nil {
		t.Fatalf("register capability provider: %v", err)
	}
	observer := &recordingProviderObserver{}
	providerRuntime.SetObserver(observer)
	ctx := context.Background()

	_, _ = providerRuntime.Complete(ctx, "all", pkgProvider.CompletionRequest{Model: "m"})
	stream, _ := providerRuntime.Stream(ctx, "all", pkgProvider.CompletionRequest{Model: "m"})
	_, _ = stream.Recv()
	_, _ = stream.Recv()
	_ = stream.Close()
	_, _ = providerRuntime.Embed(ctx, "all", pkgProvider.EmbeddingRequest{Model: "m"})
	_, _ = providerRuntime.Models(ctx, "all")
	_, _ = providerRuntime.DiscoverModels(ctx, "all")
	_ = providerRuntime.LoadModel(ctx, "all", pkgProvider.ModelLoadRequest{Model: "m"})
	_ = providerRuntime.UnloadModel(ctx, "all", pkgProvider.ModelUnloadRequest{Model: "m"})
	pull, _ := providerRuntime.PullModel(ctx, "all", pkgProvider.ModelPullRequest{Model: "m"})
	_, _ = pull.Recv()
	_, _ = pull.Recv()
	_ = pull.Close()
	_ = providerRuntime.RemoveModel(ctx, "all", pkgProvider.ModelRemoveRequest{Model: "m"})
	_, _ = providerRuntime.Capabilities(ctx, "inspector", capabilityRequest)

	terminalByOperation := make(map[pkgProvider.Operation]int)
	terminalEvents := make(map[pkgProvider.Operation]pkgProvider.ProviderEvent)
	for _, event := range observer.Events() {
		if event.Kind == pkgProvider.ProviderEventOperationCompleted {
			terminalByOperation[event.Operation]++
			terminalEvents[event.Operation] = event
		}
	}
	for _, operation := range []pkgProvider.Operation{
		pkgProvider.OperationCompletion,
		pkgProvider.OperationStreaming,
		pkgProvider.OperationEmbedding,
		pkgProvider.OperationModelListing,
		pkgProvider.OperationModelDiscovery,
		pkgProvider.OperationModelLoad,
		pkgProvider.OperationModelUnload,
		pkgProvider.OperationModelPull,
		pkgProvider.OperationModelRemove,
		pkgProvider.OperationCapabilityIntrospection,
	} {
		if terminalByOperation[operation] != 1 {
			t.Fatalf("operation %q has %d terminal events", operation,
				terminalByOperation[operation])
		}
	}
	if terminalEvents[pkgProvider.OperationEmbedding].Usage.InputTokens != 4 {
		t.Fatalf("embedding usage was not observed: %#v",
			terminalEvents[pkgProvider.OperationEmbedding])
	}
	pullTerminal := terminalEvents[pkgProvider.OperationModelPull]
	if pullTerminal.CompletedBytes != 8 || pullTerminal.TotalBytes != 10 {
		t.Fatalf("pull progress was not observed: %#v", pullTerminal)
	}
}

func TestProviderObserverFailureAndPanicDoNotChangeResults(t *testing.T) {
	for _, test := range []struct {
		name     string
		observer pkgProvider.ProviderObserver
	}{
		{
			name: "error",
			observer: pkgProvider.ProviderObserverFunc(func(pkgProvider.ProviderEvent) error {
				return errors.New("observer failed")
			}),
		},
		{
			name: "panic",
			observer: pkgProvider.ProviderObserverFunc(func(pkgProvider.ProviderEvent) error {
				panic("observer panic")
			}),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			providerRuntime, _, _ := newResilienceTestRuntime(t)
			providerRuntime.SetObserver(test.observer)
			if _, err := providerRuntime.Complete(
				context.Background(), "test",
				pkgProvider.CompletionRequest{Model: "qwen"},
			); err != nil {
				t.Fatalf("observer changed provider result: %v", err)
			}
		})
	}
}

func TestProviderObservationIsConcurrent(t *testing.T) {
	providerRuntime, _, _ := newResilienceTestRuntime(t)
	observer := &recordingProviderObserver{}
	providerRuntime.SetObserver(observer)

	const operations = 64
	var waitGroup sync.WaitGroup
	waitGroup.Add(operations)
	for range operations {
		go func() {
			defer waitGroup.Done()
			_, err := providerRuntime.Complete(
				context.Background(), "test",
				pkgProvider.CompletionRequest{Model: "qwen"},
			)
			if err != nil {
				t.Errorf("complete concurrently: %v", err)
			}
		}()
	}
	waitGroup.Wait()

	counts := make(map[uint64]map[pkgProvider.ProviderEventKind]int)
	for _, event := range observer.Events() {
		if counts[event.OperationID] == nil {
			counts[event.OperationID] = make(map[pkgProvider.ProviderEventKind]int)
		}
		counts[event.OperationID][event.Kind]++
	}
	if len(counts) != operations {
		t.Fatalf("expected %d operation IDs, got %d", operations, len(counts))
	}
	for operationID, kinds := range counts {
		for _, kind := range []pkgProvider.ProviderEventKind{
			pkgProvider.ProviderEventOperationStarted,
			pkgProvider.ProviderEventAttemptStarted,
			pkgProvider.ProviderEventOperationCompleted,
		} {
			if kinds[kind] != 1 {
				t.Fatalf("operation %d has events %v", operationID, kinds)
			}
		}
	}
}

func assertEventKinds(
	t *testing.T,
	events []pkgProvider.ProviderEvent,
	kinds ...pkgProvider.ProviderEventKind,
) {
	t.Helper()
	if len(events) != len(kinds) {
		t.Fatalf("expected %d events, got %d: %#v", len(kinds), len(events), events)
	}
	for index, kind := range kinds {
		if events[index].Kind != kind {
			t.Fatalf("event %d is %q, expected %q: %#v",
				index, events[index].Kind, kind, events)
		}
	}
}

func assertSingleOperationID(
	t *testing.T,
	events []pkgProvider.ProviderEvent,
) {
	t.Helper()
	if len(events) == 0 || events[0].OperationID == 0 {
		t.Fatalf("missing operation ID: %#v", events)
	}
	operationID := events[0].OperationID
	for _, event := range events[1:] {
		if event.OperationID != operationID {
			t.Fatalf("events are not correlated: %#v", events)
		}
	}
}
