package agent

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

const (
	CapabilityRun            pkgRuntime.Capability = "agent.run"
	CapabilityPlanning       pkgRuntime.Capability = "agent.planning"
	CapabilityWorkspaceAware pkgRuntime.Capability = "agent.workspace-aware"
)

type Descriptor struct {
	id           ID
	version      Version
	description  string
	capabilities []pkgRuntime.Capability
}

func NewDescriptor(id ID, version Version, description string, capabilities []pkgRuntime.Capability) (Descriptor, error) {
	if err := id.Validate(); err != nil {
		return Descriptor{}, fmt.Errorf("agent descriptor identity: %w: %w", err, ErrInvalidDescriptor)
	}
	if err := version.Validate(); err != nil {
		return Descriptor{}, fmt.Errorf("agent descriptor version: %w: %w", err, ErrInvalidDescriptor)
	}
	if strings.TrimSpace(description) == "" || len(description) > 4096 ||
		!utf8.ValidString(description) || strings.ContainsRune(description, 0) {
		return Descriptor{}, fmt.Errorf("agent description is blank or unsafe: %w", ErrInvalidDescriptor)
	}
	cloned := slices.Clone(capabilities)
	if len(cloned) == 0 {
		return Descriptor{}, fmt.Errorf("agent requires at least one capability: %w", ErrInvalidDescriptor)
	}
	seen := make(map[pkgRuntime.Capability]struct{}, len(cloned))
	for _, capability := range cloned {
		if !namespacedID(string(capability)) {
			return Descriptor{}, fmt.Errorf("agent capability %q is not namespaced: %w", capability, ErrInvalidDescriptor)
		}
		if _, exists := seen[capability]; exists {
			return Descriptor{}, fmt.Errorf("agent capability %q is duplicated: %w", capability, ErrInvalidDescriptor)
		}
		seen[capability] = struct{}{}
	}
	slices.Sort(cloned)
	return Descriptor{id: id, version: version, description: description, capabilities: cloned}, nil
}

func (descriptor Descriptor) ID() ID              { return descriptor.id }
func (descriptor Descriptor) Version() Version    { return descriptor.version }
func (descriptor Descriptor) Description() string { return descriptor.description }
func (descriptor Descriptor) Capabilities() []pkgRuntime.Capability {
	return slices.Clone(descriptor.capabilities)
}

func (descriptor Descriptor) Validate() error {
	_, err := NewDescriptor(descriptor.id, descriptor.version, descriptor.description, descriptor.capabilities)
	return err
}
