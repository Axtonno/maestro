package gestor

import (
	"fmt"
	"slices"
)

type ResolutionReason string

const (
	ResolutionSingleCandidate ResolutionReason = "single_candidate"
	ResolutionPreferredTarget ResolutionReason = "preferred_target"
)

func (reason ResolutionReason) Valid() bool {
	switch reason {
	case ResolutionSingleCandidate, ResolutionPreferredTarget:
		return true
	default:
		return false
	}
}

// Resolution is an immutable explanation of a selection. Dependencies are
// component targets in dependency-first topological order.
type Resolution struct {
	descriptor   Descriptor
	snapshot     SnapshotMetadata
	reason       ResolutionReason
	dependencies []Target
}

func NewResolution(
	descriptor Descriptor,
	snapshot SnapshotMetadata,
	reason ResolutionReason,
	dependencies []Target,
) (Resolution, error) {
	resolution := Resolution{
		descriptor:   descriptor,
		snapshot:     snapshot,
		reason:       reason,
		dependencies: slices.Clone(dependencies),
	}
	resolution.snapshot.sources = slices.Clone(snapshot.sources)
	if err := resolution.Validate(); err != nil {
		return Resolution{}, err
	}

	return resolution, nil
}

func (resolution Resolution) Descriptor() Descriptor   { return resolution.descriptor }
func (resolution Resolution) Reason() ResolutionReason { return resolution.reason }

func (resolution Resolution) Snapshot() SnapshotMetadata {
	metadata := resolution.snapshot
	metadata.sources = slices.Clone(resolution.snapshot.sources)

	return metadata
}

func (resolution Resolution) Dependencies() []Target {
	return slices.Clone(resolution.dependencies)
}

func (resolution Resolution) Validate() error {
	if err := resolution.descriptor.Validate(); err != nil {
		return fmt.Errorf("resolution descriptor: %w: %w", err, ErrInvalidResolution)
	}
	if err := resolution.snapshot.Validate(); err != nil {
		return fmt.Errorf("resolution snapshot: %w: %w", err, ErrInvalidResolution)
	}
	if !resolution.snapshot.Current {
		return fmt.Errorf("resolution snapshot is stale: %w: %w", ErrStaleSnapshot, ErrInvalidResolution)
	}
	if !resolution.reason.Valid() {
		return fmt.Errorf("resolution reason %q is unknown: %w", resolution.reason, ErrInvalidResolution)
	}
	foundSource := false
	for _, source := range resolution.snapshot.sources {
		if source == resolution.descriptor.Source {
			foundSource = true
			break
		}
	}
	if !foundSource {
		return fmt.Errorf("resolution descriptor source is absent from snapshot: %w", ErrInvalidResolution)
	}

	seen := make(map[Target]struct{}, len(resolution.dependencies))
	for index, dependency := range resolution.dependencies {
		if err := dependency.Validate(); err != nil {
			return fmt.Errorf("resolution dependency %d: %w: %w", index, err, ErrInvalidResolution)
		}
		if dependency.Kind != TargetKindComponent {
			return fmt.Errorf("resolution dependency %d is not a component: %w", index, ErrInvalidResolution)
		}
		if _, exists := seen[dependency]; exists {
			return fmt.Errorf("resolution dependency %d is duplicated: %w", index, ErrInvalidResolution)
		}
		seen[dependency] = struct{}{}
	}

	return nil
}
