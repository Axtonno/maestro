package provider

import (
	"errors"
	"reflect"
	"testing"
)

func TestKnownCapabilitiesReturnsCanonicalCopy(t *testing.T) {
	first := KnownCapabilities()
	second := KnownCapabilities()
	if len(first) != 14 || !reflect.DeepEqual(first, second) {
		t.Fatalf("unexpected known capabilities %#v", first)
	}

	first[0] = "mutated"
	if second[0] != CapabilityCompletion ||
		KnownCapabilities()[0] != CapabilityCompletion {
		t.Fatal("known capability storage was mutated")
	}
}

func TestValidateGenerationCapabilitiesRequiresExplicitControls(t *testing.T) {
	report := CapabilityReport{
		Target: CapabilityTargetModel, Model: "qwen",
		Capabilities: []CapabilityDescriptor{
			{Capability: CapabilityContextWindowControl, Support: CapabilitySupported, Availability: CapabilityAvailabilityAvailable},
			{Capability: CapabilityThinkingControl, Support: CapabilitySupported, Availability: CapabilityAvailabilityAvailable},
			{Capability: CapabilityThinking, Support: CapabilitySupported, Availability: CapabilityAvailabilityUnavailable},
		},
	}
	if err := ValidateGenerationCapabilities(report, GenerationOptions{ContextWindow: 4096, Thinking: ThinkingDisabled}); err != nil {
		t.Fatalf("explicit disabled thinking should be supported: %v", err)
	}
	if err := ValidateGenerationCapabilities(report, GenerationOptions{Thinking: ThinkingEnabled}); !errors.Is(err, ErrUnsupportedCapability) {
		t.Fatalf("enabled thinking should require model support: %v", err)
	}
	if err := ValidateGenerationCapabilities(CapabilityReport{Target: CapabilityTargetAdapter}, GenerationOptions{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("adapter report should be rejected: %v", err)
	}
}

func TestCapabilityRequestValidation(t *testing.T) {
	valid := []CapabilityRequest{
		{Target: CapabilityTargetAdapter},
		{Target: CapabilityTargetInstance},
		{Target: CapabilityTargetModel, Model: "qwen"},
	}
	for _, request := range valid {
		if err := request.Validate(); err != nil {
			t.Fatalf("request %#v: %v", request, err)
		}
	}

	invalid := []CapabilityRequest{
		{},
		{Target: "future"},
		{Target: CapabilityTargetAdapter, Model: "qwen"},
		{Target: CapabilityTargetInstance, Model: "qwen"},
		{Target: CapabilityTargetModel},
		{Target: CapabilityTargetModel, Model: " qwen"},
	}
	for _, request := range invalid {
		if err := request.Validate(); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("request %#v: expected ErrInvalidRequest, got %v", request, err)
		}
	}
}
