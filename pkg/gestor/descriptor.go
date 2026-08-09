package gestor

import (
	"cmp"
	"fmt"
)

type Availability string

const (
	AvailabilityUnknown     Availability = "unknown"
	AvailabilityAvailable   Availability = "available"
	AvailabilityUnavailable Availability = "unavailable"
)

func (availability Availability) Valid() bool {
	switch availability {
	case AvailabilityUnknown, AvailabilityAvailable, AvailabilityUnavailable:
		return true
	default:
		return false
	}
}

// Descriptor records a declaration and its separately observed availability.
// Snapshot generation belongs to SnapshotMetadata rather than being copied
// into every descriptor.
type Descriptor struct {
	Capability   CapabilityID
	Target       Target
	Availability Availability
	Source       SourceID
}

func (descriptor Descriptor) Validate() error {
	if err := descriptor.Capability.Validate(); err != nil {
		return fmt.Errorf("descriptor capability: %w: %w", err, ErrInvalidDescriptor)
	}
	if err := descriptor.Target.Validate(); err != nil {
		return fmt.Errorf("descriptor target: %w: %w", err, ErrInvalidDescriptor)
	}
	if !descriptor.Availability.Valid() {
		return fmt.Errorf("descriptor availability %q: %w: %w", descriptor.Availability, ErrInvalidAvailability, ErrInvalidDescriptor)
	}
	if err := descriptor.Source.Validate(); err != nil {
		return fmt.Errorf("descriptor source: %w: %w", err, ErrInvalidDescriptor)
	}

	return nil
}

func (descriptor Descriptor) Compare(other Descriptor) int {
	if result := descriptor.Capability.Compare(other.Capability); result != 0 {
		return result
	}
	if result := descriptor.Target.Compare(other.Target); result != 0 {
		return result
	}
	if result := descriptor.Source.Compare(other.Source); result != 0 {
		return result
	}

	return cmp.Compare(descriptor.Availability, other.Availability)
}
