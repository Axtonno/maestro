package provider

import (
	"fmt"
	"strings"
)

// Capability identifies one provider-neutral operation. KnownCapabilities
// returns the canonical order used in capability reports.
type Capability string

const (
	CapabilityCompletion       Capability = "completion"
	CapabilityStreaming        Capability = "streaming"
	CapabilityEmbedding        Capability = "embedding"
	CapabilityModelListing     Capability = "model_listing"
	CapabilityModelDiscovery   Capability = "model_discovery"
	CapabilityModelLoad        Capability = "model_load"
	CapabilityModelUnload      Capability = "model_unload"
	CapabilityModelPull        Capability = "model_pull"
	CapabilityModelRemove      Capability = "model_remove"
	CapabilityStructuredOutput Capability = "structured_output"
	CapabilityToolCalling      Capability = "tool_calling"
)

var knownCapabilities = [...]Capability{
	CapabilityCompletion,
	CapabilityStreaming,
	CapabilityEmbedding,
	CapabilityModelListing,
	CapabilityModelDiscovery,
	CapabilityModelLoad,
	CapabilityModelUnload,
	CapabilityModelPull,
	CapabilityModelRemove,
	CapabilityStructuredOutput,
	CapabilityToolCalling,
}

func KnownCapabilities() []Capability {
	capabilities := make([]Capability, len(knownCapabilities))
	copy(capabilities, knownCapabilities[:])

	return capabilities
}

type CapabilityTarget string

const (
	// CapabilityTargetAdapter reports structural support without remote I/O.
	CapabilityTargetAdapter CapabilityTarget = "adapter"
	// CapabilityTargetInstance probes the configured provider instance.
	CapabilityTargetInstance CapabilityTarget = "instance"
	// CapabilityTargetModel probes one exact model on the configured instance.
	CapabilityTargetModel CapabilityTarget = "model"
)

type CapabilitySupport string

const (
	CapabilityUnsupported CapabilitySupport = "unsupported"
	CapabilitySupported   CapabilitySupport = "supported"
)

type CapabilityAvailability string

const (
	CapabilityAvailabilityUnknown     CapabilityAvailability = "unknown"
	CapabilityAvailabilityAvailable   CapabilityAvailability = "available"
	CapabilityAvailabilityUnavailable CapabilityAvailability = "unavailable"
)

type CapabilityRequest struct {
	Target CapabilityTarget
	Model  string
}

func (r CapabilityRequest) Validate() error {
	switch r.Target {
	case CapabilityTargetAdapter, CapabilityTargetInstance:
		if r.Model != "" {
			return fmt.Errorf(
				"capability target %q does not accept a model: %w",
				r.Target,
				ErrInvalidRequest,
			)
		}
	case CapabilityTargetModel:
		if strings.TrimSpace(r.Model) == "" ||
			strings.TrimSpace(r.Model) != r.Model {
			return fmt.Errorf(
				"capability model target requires an exact model ID: %w",
				ErrInvalidRequest,
			)
		}
	default:
		return fmt.Errorf(
			"unknown capability target %q: %w",
			r.Target,
			ErrInvalidRequest,
		)
	}

	return nil
}

type CapabilityDescriptor struct {
	Capability   Capability
	Support      CapabilitySupport
	Availability CapabilityAvailability
}

type CapabilityReport struct {
	Provider     ID
	Target       CapabilityTarget
	Model        string
	Capabilities []CapabilityDescriptor
}
