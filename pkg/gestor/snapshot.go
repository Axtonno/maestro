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
	if metadata.DescriptorCount == 0 && len(metadata.sources) != 0 {
		return fmt.Errorf("empty snapshot cannot list sources: %w", ErrInvalidSnapshot)
	}
	if metadata.DescriptorCount > 0 && len(metadata.sources) == 0 {
		return fmt.Errorf("non-empty snapshot must list its sources: %w", ErrInvalidSnapshot)
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
	ordered := slices.Clone(descriptors)
	for index, descriptor := range ordered {
		if err := descriptor.Validate(); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot descriptor %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
	}
	slices.SortFunc(ordered, func(left, right Descriptor) int {
		return left.Compare(right)
	})
	for index := 1; index < len(ordered); index++ {
		if ordered[index-1].Capability == ordered[index].Capability &&
			ordered[index-1].Target == ordered[index].Target {
			return Snapshot{}, fmt.Errorf("duplicate capability %q target %q: %w", ordered[index].Capability, ordered[index].Target.ID, ErrInvalidSnapshot)
		}
	}

	sources := make([]SourceID, 0, len(ordered))
	seenSources := make(map[SourceID]struct{}, len(ordered))
	for _, descriptor := range ordered {
		if _, exists := seenSources[descriptor.Source]; exists {
			continue
		}
		seenSources[descriptor.Source] = struct{}{}
		sources = append(sources, descriptor.Source)
	}
	slices.SortFunc(sources, func(left, right SourceID) int {
		return left.Compare(right)
	})

	metadata := SnapshotMetadata{
		Generation:      generation,
		Current:         current,
		DescriptorCount: len(ordered),
		sources:         sources,
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
