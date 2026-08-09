package gestor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
)

type memorySource struct {
	id pkgGestor.SourceID

	mu          sync.RWMutex
	descriptors []pkgGestor.Descriptor
	err         error
	discover    func(context.Context) ([]pkgGestor.Descriptor, error)
}

func (source *memorySource) ID() pkgGestor.SourceID { return source.id }

func (source *memorySource) Discover(ctx context.Context) ([]pkgGestor.Descriptor, error) {
	source.mu.RLock()
	discover := source.discover
	descriptors := slices.Clone(source.descriptors)
	err := source.err
	source.mu.RUnlock()

	if discover != nil {
		return discover(ctx)
	}

	return descriptors, err
}

func (source *memorySource) set(
	descriptors []pkgGestor.Descriptor,
	err error,
) {
	source.mu.Lock()
	defer source.mu.Unlock()

	source.descriptors = slices.Clone(descriptors)
	source.err = err
	source.discover = nil
}

func testComponentTarget(id string) pkgGestor.Target {
	return pkgGestor.Target{
		Kind:  pkgGestor.TargetKindComponent,
		ID:    id,
		Scope: pkgGestor.ScopeComponent,
	}
}

func testDescriptor(
	capability pkgGestor.CapabilityID,
	targetID string,
	source pkgGestor.SourceID,
) pkgGestor.Descriptor {
	return pkgGestor.Descriptor{
		Capability:   capability,
		Target:       testComponentTarget(targetID),
		Availability: pkgGestor.AvailabilityUnknown,
		Source:       source,
	}
}

func TestRegistryRefreshWithZeroSources(t *testing.T) {
	registry := NewRegistry()
	initial := registry.Snapshot().Metadata()
	if initial.Generation != 0 || initial.Current || initial.DescriptorCount != 0 {
		t.Fatalf("unexpected initial metadata: %#v", initial)
	}

	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh empty registry: %v", err)
	}
	metadata := registry.Snapshot().Metadata()
	if metadata.Generation != 1 || !metadata.Current || metadata.DescriptorCount != 0 {
		t.Fatalf("unexpected refreshed metadata: %#v", metadata)
	}
}

func TestRegistryExecutesSourcesInDeterministicOrder(t *testing.T) {
	registry := NewRegistry()
	var orderMu sync.Mutex
	order := make([]pkgGestor.SourceID, 0, 3)

	for _, id := range []pkgGestor.SourceID{"z.source", "a.source", "m.source"} {
		id := id
		source := &memorySource{id: id}
		source.discover = func(context.Context) ([]pkgGestor.Descriptor, error) {
			orderMu.Lock()
			order = append(order, id)
			orderMu.Unlock()
			if id == "m.source" {
				return nil, nil
			}

			return []pkgGestor.Descriptor{
				testDescriptor(pkgGestor.CapabilityRuntimeStart, string(id), id),
			}, nil
		}
		if err := registry.RegisterSource(source); err != nil {
			t.Fatalf("register source %q: %v", id, err)
		}
	}

	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh registry: %v", err)
	}
	wantOrder := []pkgGestor.SourceID{"a.source", "m.source", "z.source"}
	if !slices.Equal(order, wantOrder) {
		t.Fatalf("expected discovery order %v, got %v", wantOrder, order)
	}

	snapshot := registry.Snapshot()
	if !slices.Equal(snapshot.Metadata().Sources(), wantOrder) {
		t.Fatalf("expected snapshot sources %v, got %v", wantOrder, snapshot.Metadata().Sources())
	}
	if snapshot.Metadata().DescriptorCount != 2 {
		t.Fatalf("expected 2 descriptors, got %d", snapshot.Metadata().DescriptorCount)
	}
}

func TestRegistryRejectsInvalidAndDuplicateSources(t *testing.T) {
	registry := NewRegistry()
	if err := registry.RegisterSource(nil); !errors.Is(err, pkgGestor.ErrInvalidSource) {
		t.Fatalf("nil source: expected ErrInvalidSource, got %v", err)
	}
	var typedNil *memorySource
	if err := registry.RegisterSource(typedNil); !errors.Is(err, pkgGestor.ErrInvalidSource) {
		t.Fatalf("typed nil source: expected ErrInvalidSource, got %v", err)
	}
	invalid := &memorySource{id: "invalid"}
	err := registry.RegisterSource(invalid)
	if !errors.Is(err, pkgGestor.ErrInvalidSource) || !errors.Is(err, pkgGestor.ErrInvalidSourceID) {
		t.Fatalf("invalid source ID: expected source sentinels, got %v", err)
	}

	source := &memorySource{id: "test.source"}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh registry: %v", err)
	}
	before := registry.Snapshot().Metadata()
	err = registry.RegisterSource(&memorySource{id: "test.source"})
	if !errors.Is(err, pkgGestor.ErrSourceAlreadyRegistered) {
		t.Fatalf("duplicate source: expected ErrSourceAlreadyRegistered, got %v", err)
	}
	after := registry.Snapshot().Metadata()
	if after.Generation != before.Generation ||
		after.Current != before.Current ||
		after.DescriptorCount != before.DescriptorCount ||
		!slices.Equal(after.Sources(), before.Sources()) {
		t.Fatalf("duplicate registration changed metadata: before %#v after %#v", before, after)
	}
}

func TestRegistryRejectsInvalidDuplicateAndForeignDescriptors(t *testing.T) {
	tests := []struct {
		name        string
		sources     []*memorySource
		want        error
		wantInvalid error
	}{
		{
			name: "invalid descriptor",
			sources: []*memorySource{{
				id: "a.source",
				descriptors: []pkgGestor.Descriptor{{
					Target:       testComponentTarget("component"),
					Availability: pkgGestor.AvailabilityUnknown,
					Source:       "a.source",
				}},
			}},
			want:        pkgGestor.ErrSourceFailure,
			wantInvalid: pkgGestor.ErrInvalidDescriptor,
		},
		{
			name: "foreign source",
			sources: []*memorySource{{
				id: "a.source",
				descriptors: []pkgGestor.Descriptor{
					testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "b.source"),
				},
			}},
			want:        pkgGestor.ErrSourceFailure,
			wantInvalid: pkgGestor.ErrInvalidDescriptor,
		},
		{
			name: "duplicate within source",
			sources: []*memorySource{{
				id: "a.source",
				descriptors: []pkgGestor.Descriptor{
					testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "a.source"),
					testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "a.source"),
				},
			}},
			want:        pkgGestor.ErrSourceFailure,
			wantInvalid: pkgGestor.ErrInvalidSnapshot,
		},
		{
			name: "collision across sources",
			sources: []*memorySource{
				{
					id: "a.source",
					descriptors: []pkgGestor.Descriptor{
						testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "a.source"),
					},
				},
				{
					id: "b.source",
					descriptors: []pkgGestor.Descriptor{
						testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "b.source"),
					},
				},
			},
			want:        pkgGestor.ErrSourceFailure,
			wantInvalid: pkgGestor.ErrInvalidSnapshot,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := NewRegistry()
			for _, source := range test.sources {
				if err := registry.RegisterSource(source); err != nil {
					t.Fatalf("register source: %v", err)
				}
			}
			err := registry.Refresh(context.Background())
			if !errors.Is(err, test.want) || !errors.Is(err, test.wantInvalid) {
				t.Fatalf("expected %v and %v, got %v", test.want, test.wantInvalid, err)
			}
			metadata := registry.Snapshot().Metadata()
			if metadata.Generation != 0 || metadata.Current {
				t.Fatalf("failed refresh published metadata %#v", metadata)
			}
		})
	}
}

func TestRegistryFailureAndCancellationPreserveLastSnapshot(t *testing.T) {
	registry := NewRegistry()
	firstDescriptor := testDescriptor(
		pkgGestor.CapabilityRuntimeStart,
		"component",
		"a.source",
	)
	first := &memorySource{id: "a.source", descriptors: []pkgGestor.Descriptor{firstDescriptor}}
	second := &memorySource{id: "b.source"}
	if err := registry.RegisterSource(first); err != nil {
		t.Fatalf("register first source: %v", err)
	}
	if err := registry.RegisterSource(second); err != nil {
		t.Fatalf("register second source: %v", err)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	want := registry.Snapshot()

	failure := errors.New("discovery failed")
	second.set(nil, failure)
	err := registry.Refresh(context.Background())
	if !errors.Is(err, pkgGestor.ErrSourceFailure) || !errors.Is(err, failure) {
		t.Fatalf("source failure: expected sentinel and cause, got %v", err)
	}
	assertSameSnapshot(t, registry.Snapshot(), want)

	cancelled, cancel := context.WithCancel(context.Background())
	first.mu.Lock()
	first.discover = func(context.Context) ([]pkgGestor.Descriptor, error) {
		cancel()
		return []pkgGestor.Descriptor{firstDescriptor}, nil
	}
	first.mu.Unlock()
	secondCalled := false
	second.mu.Lock()
	second.err = nil
	second.discover = func(context.Context) ([]pkgGestor.Descriptor, error) {
		secondCalled = true
		return nil, nil
	}
	second.mu.Unlock()
	err = registry.Refresh(cancelled)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("canceled refresh: expected cancellation and source failure, got %v", err)
	}
	if secondCalled {
		t.Fatal("refresh executed a later source after cancellation")
	}
	assertSameSnapshot(t, registry.Snapshot(), want)
}

func TestRegistryGenerationInvalidationAndRecovery(t *testing.T) {
	registry := NewRegistry()
	source := &memorySource{id: "test.source"}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}

	for wantGeneration := uint64(1); wantGeneration <= 2; wantGeneration++ {
		if err := registry.Refresh(context.Background()); err != nil {
			t.Fatalf("refresh generation %d: %v", wantGeneration, err)
		}
		metadata := registry.Snapshot().Metadata()
		if metadata.Generation != wantGeneration || !metadata.Current {
			t.Fatalf("unexpected metadata at generation %d: %#v", wantGeneration, metadata)
		}
	}

	registry.Invalidate()
	stale := registry.Snapshot().Metadata()
	if stale.Generation != 2 || stale.Current {
		t.Fatalf("invalidate changed generation or remained current: %#v", stale)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh after invalidate: %v", err)
	}
	recovered := registry.Snapshot().Metadata()
	if recovered.Generation != 3 || !recovered.Current {
		t.Fatalf("unexpected recovered metadata: %#v", recovered)
	}

	newSource := &memorySource{id: "new.source"}
	if err := registry.RegisterSource(newSource); err != nil {
		t.Fatalf("register new source: %v", err)
	}
	invalidated := registry.Snapshot().Metadata()
	if invalidated.Generation != 3 || invalidated.Current {
		t.Fatalf("source registration did not invalidate snapshot: %#v", invalidated)
	}
}

func TestRegistryIndexesAreOrderedAndDefensive(t *testing.T) {
	registry := NewRegistry()
	target := testComponentTarget("component")
	descriptors := []pkgGestor.Descriptor{
		testDescriptor(pkgGestor.CapabilityRuntimeStop, "component", "test.source"),
		testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "test.source"),
		testDescriptor(pkgGestor.CapabilityRuntimeStart, "other", "test.source"),
	}
	source := &memorySource{id: "test.source", descriptors: descriptors}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh registry: %v", err)
	}

	byCapability := registry.descriptorsByCapability(pkgGestor.CapabilityRuntimeStart)
	if len(byCapability) != 2 || byCapability[0].Target.ID != "component" || byCapability[1].Target.ID != "other" {
		t.Fatalf("unexpected capability index: %#v", byCapability)
	}
	byCapability[0].Target.ID = "changed"
	if registry.descriptorsByCapability(pkgGestor.CapabilityRuntimeStart)[0].Target.ID != "component" {
		t.Fatal("capability index exposed its backing slice")
	}

	byTarget := registry.descriptorsByTarget(target)
	if len(byTarget) != 2 || byTarget[0].Capability != pkgGestor.CapabilityRuntimeStart || byTarget[1].Capability != pkgGestor.CapabilityRuntimeStop {
		t.Fatalf("unexpected target index: %#v", byTarget)
	}
	byTarget[0].Availability = pkgGestor.AvailabilityAvailable
	if registry.descriptorsByTarget(target)[0].Availability != pkgGestor.AvailabilityUnknown {
		t.Fatal("target index exposed its backing slice")
	}
}

func TestRegistryDoesNotExecuteSourceUnderLock(t *testing.T) {
	registry := NewRegistry()
	entered := make(chan struct{})
	release := make(chan struct{})
	blocking := &memorySource{id: "a.source"}
	blocking.discover = func(context.Context) ([]pkgGestor.Descriptor, error) {
		close(entered)
		<-release
		return nil, nil
	}
	if err := registry.RegisterSource(blocking); err != nil {
		t.Fatalf("register blocking source: %v", err)
	}

	refreshDone := make(chan error, 1)
	go func() {
		refreshDone <- registry.Refresh(context.Background())
	}()
	<-entered
	during := registry.Snapshot().Metadata()
	if during.Generation != 0 || during.Current || during.DescriptorCount != 0 {
		t.Fatalf("refresh exposed a partial snapshot: %#v", during)
	}

	registerDone := make(chan error, 1)
	go func() {
		registerDone <- registry.RegisterSource(&memorySource{id: "b.source"})
	}()
	select {
	case err := <-registerDone:
		if err != nil {
			t.Fatalf("register while source blocked: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("source Discover appears to hold the registry lock")
	}

	close(release)
	err := <-refreshDone
	if !errors.Is(err, pkgGestor.ErrStaleSnapshot) {
		t.Fatalf("refresh over changed catalog: expected ErrStaleSnapshot, got %v", err)
	}
}

func TestRegistryConcurrentRefreshAndReads(t *testing.T) {
	registry := NewRegistry()
	source := &memorySource{
		id: "test.source",
		descriptors: []pkgGestor.Descriptor{
			testDescriptor(pkgGestor.CapabilityRuntimeStart, "component", "test.source"),
		},
	}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}

	const writers = 8
	const readers = 8
	const iterations = 40
	var wait sync.WaitGroup
	errorsChannel := make(chan error, writers)
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range iterations {
				if err := registry.Refresh(context.Background()); err != nil {
					errorsChannel <- err
					return
				}
			}
		}()
	}
	for range readers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range iterations * 2 {
				snapshot := registry.Snapshot()
				_ = snapshot.Metadata().Sources()
				_ = snapshot.Descriptors()
				_ = registry.descriptorsByCapability(pkgGestor.CapabilityRuntimeStart)
				_ = registry.descriptorsByTarget(testComponentTarget("component"))
			}
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent refresh: %v", err)
	}

	metadata := registry.Snapshot().Metadata()
	wantGeneration := uint64(writers * iterations)
	if metadata.Generation != wantGeneration || !metadata.Current {
		t.Fatalf("expected current generation %d, got %#v", wantGeneration, metadata)
	}
}

func TestRegistryRejectsNilAndCanceledRefreshContexts(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Refresh(nil); !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("nil context: expected ErrSourceFailure, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := registry.Refresh(ctx)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("canceled context: expected cancellation and source failure, got %v", err)
	}
	if registry.Snapshot().Metadata().Generation != 0 {
		t.Fatal("canceled refresh changed the generation")
	}
}

func assertSameSnapshot(t *testing.T, got, want pkgGestor.Snapshot) {
	t.Helper()
	if got.Metadata().Generation != want.Metadata().Generation ||
		got.Metadata().Current != want.Metadata().Current ||
		got.Metadata().DescriptorCount != want.Metadata().DescriptorCount ||
		!slices.Equal(got.Metadata().Sources(), want.Metadata().Sources()) ||
		!slices.Equal(got.Descriptors(), want.Descriptors()) {
		t.Fatalf("snapshot changed\nwant: %#v %#v\ngot:  %#v %#v", want.Metadata(), want.Descriptors(), got.Metadata(), got.Descriptors())
	}
}

func ExampleRegistry() {
	registry := NewRegistry()
	source := &memorySource{id: "example.source"}
	_ = registry.RegisterSource(source)
	_ = registry.Refresh(context.Background())

	fmt.Println(registry.Snapshot().Metadata().Generation)
	// Output: 1
}
