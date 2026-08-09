package gestor

import (
	"fmt"
	"slices"
)

type SnapshotMetadata struct {
	Generation      uint64
	Current         bool
	DescriptorCount int
	sources         []SourceID
}

func (metadata SnapshotMetadata) Sources() []SourceID {
	return slices.Clone(metadata.sources)
}

func (metadata SnapshotMetadata) Validate() error {
	if metadata.DescriptorCount < 0 {
		return fmt.Errorf("negative descriptor count: %w", ErrInvalidSnapshot)
	}
	for index, source := range metadata.sources {
		if err := source.Validate(); err != nil {
			return fmt.Errorf("snapshot source %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
		if index > 0 && metadata.sources[index-1].Compare(source) >= 0 {
			return fmt.Errorf("snapshot sources are not unique and ordered: %w", ErrInvalidSnapshot)
		}
	}

	return nil
}

// Snapshot is an immutable, deterministically ordered set of descriptors.
type Snapshot struct {
	metadata    SnapshotMetadata
	descriptors []Descriptor
}

func NewSnapshot(generation uint64, current bool, descriptors []Descriptor) (Snapshot, error) {
	sources := make([]SourceID, 0, len(descriptors))
	seenSources := make(map[SourceID]struct{}, len(descriptors))
	for index, descriptor := range descriptors {
		if err := descriptor.Validate(); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot descriptor %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
		if _, exists := seenSources[descriptor.Source]; exists {
			continue
		}
		seenSources[descriptor.Source] = struct{}{}
		sources = append(sources, descriptor.Source)
	}

	return NewSnapshotWithSources(generation, current, sources, descriptors)
}

// NewSnapshotWithSources constructs a snapshot that also records consulted
// sources which produced no descriptors.
func NewSnapshotWithSources(
	generation uint64,
	current bool,
	sources []SourceID,
	descriptors []Descriptor,
) (Snapshot, error) {
	orderedSources := slices.Clone(sources)
	for index, source := range orderedSources {
		if err := source.Validate(); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot source %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
	}
	slices.SortFunc(orderedSources, func(left, right SourceID) int {
		return left.Compare(right)
	})
	for index := 1; index < len(orderedSources); index++ {
		if orderedSources[index-1] == orderedSources[index] {
			return Snapshot{}, fmt.Errorf("duplicate snapshot source %q: %w", orderedSources[index], ErrInvalidSnapshot)
		}
	}
	sourceSet := make(map[SourceID]struct{}, len(orderedSources))
	for _, source := range orderedSources {
		sourceSet[source] = struct{}{}
	}

	ordered := slices.Clone(descriptors)
	for index, descriptor := range ordered {
		if err := descriptor.Validate(); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot descriptor %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
		if _, exists := sourceSet[descriptor.Source]; !exists {
			return Snapshot{}, fmt.Errorf("snapshot descriptor %d source %q was not consulted: %w", index, descriptor.Source, ErrInvalidSnapshot)
		}
	}
	slices.SortFunc(ordered, func(left, right Descriptor) int {
		return left.Compare(right)
	})
	for index := 1; index < len(ordered); index++ {
		if ordered[index-1].Capability == ordered[index].Capability &&
			ordered[index-1].Target == ordered[index].Target {
			return Snapshot{}, fmt.Errorf(
				"duplicate capability %q target %s/%q scope %q model %q from sources %q and %q: %w",
				ordered[index].Capability,
				ordered[index].Target.Kind,
				ordered[index].Target.ID,
				ordered[index].Target.Scope,
				ordered[index].Target.Model,
				ordered[index-1].Source,
				ordered[index].Source,
				ErrInvalidSnapshot,
			)
		}
	}

	metadata := SnapshotMetadata{
		Generation:      generation,
		Current:         current,
		DescriptorCount: len(ordered),
		sources:         orderedSources,
	}

	return Snapshot{metadata: metadata, descriptors: ordered}, nil
}

func (snapshot Snapshot) Metadata() SnapshotMetadata {
	metadata := snapshot.metadata
	metadata.sources = slices.Clone(snapshot.metadata.sources)

	return metadata
}

func (snapshot Snapshot) Descriptors() []Descriptor {
	return slices.Clone(snapshot.descriptors)
}
