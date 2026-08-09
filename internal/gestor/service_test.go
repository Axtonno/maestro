package gestor

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type gestorEventRecorder struct {
	mu sync.Mutex

	events []pkgGestor.Event
	err    error
	hook   func(pkgRuntime.Event)
}

func (recorder *gestorEventRecorder) Publish(event pkgRuntime.Event) error {
	recorder.mu.Lock()
	recorder.events = append(recorder.events, event.(pkgGestor.Event))
	hook := recorder.hook
	err := recorder.err
	recorder.mu.Unlock()

	if hook != nil {
		hook(event)
	}

	return err
}

func (recorder *gestorEventRecorder) snapshot() []pkgGestor.Event {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return append([]pkgGestor.Event(nil), recorder.events...)
}

func newGestorServiceFixture(
	t *testing.T,
	recorder eventPublisher,
) (*Service, *memorySource) {
	t.Helper()

	registry := NewRegistry()
	source := &memorySource{id: "test.source", descriptors: []pkgGestor.Descriptor{
		resolverDescriptor(
			pkgGestor.CapabilityProviderCompletion,
			resolverProviderTarget("secret-target", pkgGestor.ScopeAdapter, ""),
			pkgGestor.AvailabilityAvailable,
		),
	}}
	if err := registry.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	resolver, err := NewResolver(registry, newDependencyGraphFixture())
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	service, err := NewService(registry, resolver, recorder)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	return service, source
}

func TestGestorServiceRejectsMissingCollaborators(t *testing.T) {
	registry := NewRegistry()
	resolver, err := NewResolver(registry, newDependencyGraphFixture())
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	if _, err := NewService(nil, resolver, nil); !errors.Is(err, pkgGestor.ErrInvalidSnapshot) {
		t.Fatalf("expected invalid snapshot for nil registry, got %v", err)
	}
	if _, err := NewService(registry, nil, nil); !errors.Is(err, pkgGestor.ErrInvalidResolution) {
		t.Fatalf("expected invalid resolution for nil resolver, got %v", err)
	}
}

func TestGestorServicePublishesRedactedRefreshEvents(t *testing.T) {
	recorder := &gestorEventRecorder{}
	service, source := newGestorServiceFixture(t, recorder)

	if err := service.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh service: %v", err)
	}
	events := recorder.snapshot()
	if len(events) != 2 ||
		events[0].Topic != pkgGestor.EventRefreshStarted ||
		events[1].Topic != pkgGestor.EventRefreshCompleted {
		t.Fatalf("unexpected refresh events: %#v", events)
	}
	completed := events[1].Data
	if completed.Generation != 1 || completed.DescriptorCount != 1 ||
		completed.SourceCount != 1 || completed.Failure != pkgGestor.EventFailureNone {
		t.Fatalf("unexpected completed payload: %#v", completed)
	}

	source.set(nil, fmt.Errorf("remote secret-token response"))
	service.Invalidate()
	if err := service.Refresh(t.Context()); !errors.Is(err, pkgGestor.ErrSourceFailure) {
		t.Fatalf("expected source failure, got %v", err)
	}
	events = recorder.snapshot()
	failed := events[len(events)-1]
	if failed.Topic != pkgGestor.EventRefreshFailed ||
		failed.Data.Failure != pkgGestor.EventFailureSource {
		t.Fatalf("unexpected failed event: %#v", failed)
	}
	if strings.Contains(fmt.Sprintf("%#v", failed.Data), "secret-token") {
		t.Fatal("failure event leaked the source error")
	}
	if metadata := service.Snapshot().Metadata(); metadata.Generation != 1 || metadata.Current {
		t.Fatalf("failed refresh corrupted or published a snapshot: %#v", metadata)
	}
}

func TestGestorServicePublishesResolutionOutcomeWithoutTargetIdentity(t *testing.T) {
	recorder := &gestorEventRecorder{}
	service, _ := newGestorServiceFixture(t, recorder)
	if err := service.Refresh(t.Context()); err != nil {
		t.Fatalf("refresh service: %v", err)
	}

	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})
	resolution, err := service.Resolve(query)
	if err != nil {
		t.Fatalf("resolve capability: %v", err)
	}
	if resolution.Descriptor().Target.ID != "secret-target" {
		t.Fatalf("unexpected selected target: %#v", resolution.Descriptor().Target)
	}
	events := recorder.snapshot()
	completed := events[len(events)-1]
	if completed.Topic != pkgGestor.EventResolutionCompleted ||
		completed.Data.Capability != pkgGestor.CapabilityProviderCompletion ||
		completed.Data.TargetKind != pkgGestor.TargetKindProvider ||
		completed.Data.Scope != pkgGestor.ScopeAdapter ||
		completed.Data.Reason != pkgGestor.ResolutionSingleCandidate {
		t.Fatalf("unexpected resolution event: %#v", completed)
	}
	if strings.Contains(fmt.Sprintf("%#v", completed.Data), "secret-target") {
		t.Fatal("resolution event leaked the target ID")
	}

	missing := newQuery(t, pkgGestor.CapabilityProviderEmbedding, pkgGestor.QueryOptions{})
	if _, err := service.Resolve(missing); !errors.Is(err, pkgGestor.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	events = recorder.snapshot()
	failed := events[len(events)-1]
	if failed.Topic != pkgGestor.EventResolutionFailed ||
		failed.Data.Failure != pkgGestor.EventFailureNotFound {
		t.Fatalf("unexpected resolution failure event: %#v", failed)
	}
}

func TestGestorServiceObserverFailuresPanicsAndReentryDoNotChangeResults(t *testing.T) {
	recorder := &gestorEventRecorder{err: errors.New("observer failed")}
	service, _ := newGestorServiceFixture(t, recorder)
	recorder.hook = func(pkgRuntime.Event) {
		// A synchronous observer can re-enter read APIs because publication does
		// not happen while the Registry is locked.
		_ = service.Snapshot().Metadata()
	}
	if err := service.Refresh(t.Context()); err != nil {
		t.Fatalf("observer error changed refresh result: %v", err)
	}

	recorder.hook = func(pkgRuntime.Event) { panic("observer panic") }
	query := newQuery(t, pkgGestor.CapabilityProviderCompletion, pkgGestor.QueryOptions{})
	resolution, err := service.Resolve(query)
	if err != nil {
		t.Fatalf("observer panic changed resolution: %v", err)
	}
	if resolution.Descriptor().Target.ID != "secret-target" {
		t.Fatalf("unexpected resolution after observer panic: %#v", resolution)
	}
}

func TestClassifyGestorEventFailure(t *testing.T) {
	tests := []struct {
		err  error
		want pkgGestor.EventFailure
	}{
		{nil, pkgGestor.EventFailureNone},
		{pkgGestor.ErrInvalidQuery, pkgGestor.EventFailureInvalid},
		{pkgGestor.ErrSourceFailure, pkgGestor.EventFailureSource},
		{pkgGestor.ErrNotFound, pkgGestor.EventFailureNotFound},
		{pkgGestor.ErrUnavailable, pkgGestor.EventFailureUnavailable},
		{pkgGestor.ErrAmbiguous, pkgGestor.EventFailureAmbiguous},
		{pkgGestor.ErrStaleSnapshot, pkgGestor.EventFailureStale},
		{errors.New("unexpected"), pkgGestor.EventFailureInternal},
	}
	for _, test := range tests {
		if got := classifyEventFailure(test.err); got != test.want {
			t.Errorf("classify %v: expected %q, got %q", test.err, test.want, got)
		}
	}
}
