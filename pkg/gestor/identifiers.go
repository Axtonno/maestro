package gestor

import (
	"cmp"
	"fmt"
	"strings"
)

// CapabilityID is a stable, namespaced capability identifier.
type CapabilityID string

// SourceID is a stable, namespaced discovery-source identifier.
type SourceID string

func (id CapabilityID) Validate() error {
	if !validNamespacedID(string(id)) {
		return fmt.Errorf("capability ID %q must be a normalized namespace and name: %w", id, ErrInvalidCapabilityID)
	}

	return nil
}

func (id CapabilityID) Compare(other CapabilityID) int {
	return cmp.Compare(id, other)
}

func (id SourceID) Validate() error {
	if !validNamespacedID(string(id)) {
		return fmt.Errorf("source ID %q must be a normalized namespace and name: %w", id, ErrInvalidSourceID)
	}

	return nil
}

func (id SourceID) Compare(other SourceID) int {
	return cmp.Compare(id, other)
}

func validNamespacedID(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}

	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for index, character := range part {
			letter := character >= 'a' && character <= 'z'
			digit := character >= '0' && character <= '9'
			separator := index > 0 && (character == '_' || character == '-')
			if !letter && !digit && !separator {
				return false
			}
		}
	}

	return true
}
