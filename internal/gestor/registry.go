package gestor

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"slices"
	"sync"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
)

var _ pkgGestor.Registry = (*Registry)(nil)

type sourceEntry struct {
	id     pkgGestor.SourceID
	source pkgGestor.Source
}

type indexedSnapshot struct {
	value        pkgGestor.Snapshot
	byCapability map[pkgGestor.CapabilityID][]pkgGestor.Descriptor
	byTarget     map[pkgGestor.Target][]pkgGestor.Descriptor
}

// Registry is the in-process Gestor snapshot registry. Its concrete type stays
// in internal/ while the implemented contract belongs to pkg/gestor.
type Registry struct {
	mu sync.RWMutex

	sources map[pkgGestor.SourceID]pkgGestor.Source
	epoch   uint64
	current *indexedSnapshot
}

func NewRegistry() *Registry {
	empty, err := pkgGestor.NewSnapshot(0, false, nil)
	if err != nil {
		panic(fmt.Sprintf("construct initial Gestor snapshot: %v", err))
	}

	return &Registry{
		sources: make(map[pkgGestor.SourceID]pkgGestor.Source),
		current: indexSnapshot(empty),
	}
}

func (registry *Registry) RegisterSource(source pkgGestor.Source) error {
	if nilSource(source) {
		return fmt.Errorf("register source: source is nil: %w", pkgGestor.ErrInvalidSource)
	}

	sourceID := source.ID()
	if err := sourceID.Validate(); err != nil {
		return fmt.Errorf("register source ID %q: %w: %w", sourceID, err, pkgGestor.ErrInvalidSource)
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.sources[sourceID]; exists {
		return fmt.Errorf("register source %q: %w", sourceID, pkgGestor.ErrSourceAlreadyRegistered)
	}

	registry.sources[sourceID] = source
	registry.epoch++
	registry.invalidateLocked()

	return nil
}

func (registry *Registry) Refresh(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("refresh with nil context: %w", pkgGestor.ErrSourceFailure)
	}
	if err := ctx.Err(); err != nil {
		return refreshContextError(err)
	}

	sources, epoch := registry.sourceSnapshot()
	descriptors := make([]pkgGestor.Descriptor, 0)
	sourceIDs := make([]pkgGestor.SourceID, 0, len(sources))

	for _, entry := range sources {
		if err := ctx.Err(); err != nil {
			return refreshContextError(err)
		}

		discovered, err := entry.source.Discover(ctx)
		if err != nil {
			return fmt.Errorf("discover source %q: %w: %w", entry.id, err, pkgGestor.ErrSourceFailure)
		}
		if err := ctx.Err(); err != nil {
			return refreshContextError(err)
		}

		for index, descriptor := range discovered {
			if err := descriptor.Validate(); err != nil {
				return fmt.Errorf(
					"discover source %q descriptor %d: %w: %w",
					entry.id,
					index,
					err,
					pkgGestor.ErrSourceFailure,
				)
			}
			if descriptor.Source != entry.id {
				return fmt.Errorf(
					"discover source %q descriptor %d declares source %q: %w: %w",
					entry.id,
					index,
					descriptor.Source,
					pkgGestor.ErrInvalidDescriptor,
					pkgGestor.ErrSourceFailure,
				)
			}
		}

		sourceIDs = append(sourceIDs, entry.id)
		descriptors = append(descriptors, discovered...)
	}

	candidate, err := pkgGestor.NewSnapshotWithSources(0, true, sourceIDs, descriptors)
	if err != nil {
		return fmt.Errorf("compose refresh snapshot: %w: %w", err, pkgGestor.ErrSourceFailure)
	}
	if err := ctx.Err(); err != nil {
		return refreshContextError(err)
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return refreshContextError(err)
	}
	if registry.epoch != epoch {
		return fmt.Errorf("source catalog or snapshot changed during refresh: %w", pkgGestor.ErrStaleSnapshot)
	}

	metadata := registry.current.value.Metadata()
	if metadata.Generation == math.MaxUint64 {
		return fmt.Errorf("snapshot generation overflow: %w", pkgGestor.ErrInvalidSnapshot)
	}
	published, err := pkgGestor.NewSnapshotWithSources(
		metadata.Generation+1,
		true,
		candidate.Metadata().Sources(),
		candidate.Descriptors(),
	)
	if err != nil {
		return fmt.Errorf("publish refresh snapshot: %w", err)
	}

	registry.current = indexSnapshot(published)

	return nil
}

func (registry *Registry) Invalidate() {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.epoch++
	registry.invalidateLocked()
}

func (registry *Registry) Snapshot() pkgGestor.Snapshot {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return registry.current.value
}

func (registry *Registry) descriptorsByCapability(
	capability pkgGestor.CapabilityID,
) []pkgGestor.Descriptor {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return slices.Clone(registry.current.byCapability[capability])
}

func (registry *Registry) descriptorsByTarget(
	target pkgGestor.Target,
) []pkgGestor.Descriptor {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	return slices.Clone(registry.current.byTarget[target])
}

func (registry *Registry) sourceSnapshot() ([]sourceEntry, uint64) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	sources := make([]sourceEntry, 0, len(registry.sources))
	for id, source := range registry.sources {
		sources = append(sources, sourceEntry{id: id, source: source})
	}
	slices.SortFunc(sources, func(left, right sourceEntry) int {
		return left.id.Compare(right.id)
	})

	return sources, registry.epoch
}

func (registry *Registry) invalidateLocked() {
	metadata := registry.current.value.Metadata()
	if !metadata.Current {
		return
	}

	stale, err := pkgGestor.NewSnapshotWithSources(
		metadata.Generation,
		false,
		metadata.Sources(),
		registry.current.value.Descriptors(),
	)
	if err != nil {
		panic(fmt.Sprintf("invalidate valid Gestor snapshot: %v", err))
	}
	registry.current = indexSnapshot(stale)
}

func indexSnapshot(snapshot pkgGestor.Snapshot) *indexedSnapshot {
	descriptors := snapshot.Descriptors()
	indexed := &indexedSnapshot{
		value:        snapshot,
		byCapability: make(map[pkgGestor.CapabilityID][]pkgGestor.Descriptor),
		byTarget:     make(map[pkgGestor.Target][]pkgGestor.Descriptor),
	}
	for _, descriptor := range descriptors {
		indexed.byCapability[descriptor.Capability] = append(
			indexed.byCapability[descriptor.Capability],
			descriptor,
		)
		indexed.byTarget[descriptor.Target] = append(
			indexed.byTarget[descriptor.Target],
			descriptor,
		)
	}

	return indexed
}

func refreshContextError(err error) error {
	return fmt.Errorf("refresh canceled: %w: %w", err, pkgGestor.ErrSourceFailure)
}

func nilSource(source pkgGestor.Source) bool {
	if source == nil {
		return true
	}

	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
