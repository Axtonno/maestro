package provider

import (
	"errors"
	"reflect"
	"testing"
)

func TestKnownCapabilitiesReturnsCanonicalCopy(t *testing.T) {
	first := KnownCapabilities()
	second := KnownCapabilities()
	if len(first) != 11 || !reflect.DeepEqual(first, second) {
		t.Fatalf("unexpected known capabilities %#v", first)
	}

	first[0] = "mutated"
	if second[0] != CapabilityCompletion ||
		KnownCapabilities()[0] != CapabilityCompletion {
		t.Fatal("known capability storage was mutated")
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
