package gestor

import (
	"context"
	"errors"
	"fmt"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ pkgGestor.Service = (*Service)(nil)

type eventPublisher interface {
	Publish(pkgRuntime.Event) error
}

// Service is the composed Gestor facade. It adds redacted observability around
// Registry and Resolver without changing their state semantics.
type Service struct {
	registry *Registry
	resolver *Resolver
	events   eventPublisher
}

func NewService(
	registry *Registry,
	resolver *Resolver,
	events eventPublisher,
) (*Service, error) {
	if registry == nil {
		return nil, fmt.Errorf("Gestor service registry is nil: %w", pkgGestor.ErrInvalidSnapshot)
	}
	if resolver == nil {
		return nil, fmt.Errorf("Gestor service resolver is nil: %w", pkgGestor.ErrInvalidResolution)
	}

	return &Service{
		registry: registry,
		resolver: resolver,
		events:   events,
	}, nil
}

func (service *Service) RegisterSource(source pkgGestor.Source) error {
	return service.registry.RegisterSource(source)
}

func (service *Service) Refresh(ctx context.Context) error {
	service.publish(
		pkgGestor.EventRefreshStarted,
		refreshEventPayload(service.registry.Snapshot(), pkgGestor.EventFailureNone),
	)

	err := service.registry.Refresh(ctx)
	snapshot := service.registry.Snapshot()
	if err != nil {
		service.publish(
			pkgGestor.EventRefreshFailed,
			refreshEventPayload(snapshot, classifyEventFailure(err)),
		)

		return err
	}

	service.publish(
		pkgGestor.EventRefreshCompleted,
		refreshEventPayload(snapshot, pkgGestor.EventFailureNone),
	)

	return nil
}

func (service *Service) Invalidate() {
	service.registry.Invalidate()
}

func (service *Service) Snapshot() pkgGestor.Snapshot {
	return service.registry.Snapshot()
}

func (service *Service) Candidates(
	query pkgGestor.Query,
) ([]pkgGestor.Descriptor, error) {
	return service.resolver.Candidates(query)
}

func (service *Service) Resolve(
	query pkgGestor.Query,
) (pkgGestor.Resolution, error) {
	resolution, err := service.resolver.Resolve(query)
	if err != nil {
		metadata := service.registry.Snapshot().Metadata()
		service.publish(pkgGestor.EventResolutionFailed, pkgGestor.EventPayload{
			Capability: query.Capability(),
			Generation: metadata.Generation,
			Failure:    classifyEventFailure(err),
		})

		return pkgGestor.Resolution{}, err
	}

	selected := resolution.Descriptor()
	service.publish(pkgGestor.EventResolutionCompleted, pkgGestor.EventPayload{
		Capability:      selected.Capability,
		Generation:      resolution.Snapshot().Generation,
		TargetKind:      selected.Target.Kind,
		Scope:           selected.Target.Scope,
		Reason:          resolution.Reason(),
		DependencyCount: len(resolution.Dependencies()),
	})

	return resolution, nil
}

func (service *Service) publish(topic string, payload pkgGestor.EventPayload) {
	if nilInterface(service.events) {
		return
	}

	// Event delivery is deliberately best-effort. Observer errors and panics do
	// not alter an already completed refresh or resolution.
	func() {
		defer func() {
			_ = recover()
		}()

		_ = service.events.Publish(pkgGestor.Event{Topic: topic, Data: payload})
	}()
}

func refreshEventPayload(
	snapshot pkgGestor.Snapshot,
	failure pkgGestor.EventFailure,
) pkgGestor.EventPayload {
	metadata := snapshot.Metadata()

	return pkgGestor.EventPayload{
		Generation:      metadata.Generation,
		DescriptorCount: metadata.DescriptorCount,
		SourceCount:     len(metadata.Sources()),
		Failure:         failure,
	}
}

func classifyEventFailure(err error) pkgGestor.EventFailure {
	switch {
	case err == nil:
		return pkgGestor.EventFailureNone
	case errors.Is(err, context.Canceled):
		return pkgGestor.EventFailureCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return pkgGestor.EventFailureDeadline
	case errors.Is(err, pkgGestor.ErrStaleSnapshot):
		return pkgGestor.EventFailureStale
	case errors.Is(err, pkgGestor.ErrSourceFailure):
		return pkgGestor.EventFailureSource
	case errors.Is(err, pkgGestor.ErrNotFound):
		return pkgGestor.EventFailureNotFound
	case errors.Is(err, pkgGestor.ErrUnavailable):
		return pkgGestor.EventFailureUnavailable
	case errors.Is(err, pkgGestor.ErrAmbiguous):
		return pkgGestor.EventFailureAmbiguous
	case errors.Is(err, pkgGestor.ErrInvalidCapabilityID),
		errors.Is(err, pkgGestor.ErrInvalidTarget),
		errors.Is(err, pkgGestor.ErrInvalidQuery),
		errors.Is(err, pkgGestor.ErrInvalidResolution):
		return pkgGestor.EventFailureInvalid
	default:
		return pkgGestor.EventFailureInternal
	}
}
