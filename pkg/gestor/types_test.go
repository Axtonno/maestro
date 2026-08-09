package gestor_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/antonio-cafeo/maestro/pkg/gestor"
)

func componentTarget(id string) gestor.Target {
	return gestor.Target{
		Kind:  gestor.TargetKindComponent,
		ID:    id,
		Scope: gestor.ScopeComponent,
	}
}

func providerTarget(id string, scope gestor.Scope, model string) gestor.Target {
	return gestor.Target{
		Kind:  gestor.TargetKindProvider,
		ID:    id,
		Scope: scope,
		Model: model,
	}
}

func descriptor(capability gestor.CapabilityID, target gestor.Target, source gestor.SourceID) gestor.Descriptor {
	return gestor.Descriptor{
		Capability:   capability,
		Target:       target,
		Availability: gestor.AvailabilityUnknown,
		Source:       source,
	}
}

func TestIdentifiersValidateAndCompare(t *testing.T) {
	validCapabilities := []gestor.CapabilityID{
		"runtime.start",
		"provider.structured_output",
		"plugin.workspace-detection",
		"vendor2.capability3",
	}
	for _, id := range validCapabilities {
		if err := id.Validate(); err != nil {
			t.Errorf("validate capability ID %q: %v", id, err)
		}
	}

	invalidCapabilities := []gestor.CapabilityID{
		"",
		"start",
		" runtime.start",
		"runtime.start ",
		"runtime..start",
		"Runtime.start",
		"runtime._start",
		"runtime.start now",
	}
	for _, id := range invalidCapabilities {
		if err := id.Validate(); !errors.Is(err, gestor.ErrInvalidCapabilityID) {
			t.Errorf("capability ID %q: expected ErrInvalidCapabilityID, got %v", id, err)
		}
	}

	if gestor.CapabilityID("provider.completion").Compare("runtime.start") >= 0 {
		t.Fatal("capability comparison is not lexical")
	}

	validSources := []gestor.SourceID{"runtime.components", "provider.capabilities"}
	for _, id := range validSources {
		if err := id.Validate(); err != nil {
			t.Errorf("validate source ID %q: %v", id, err)
		}
	}
	for _, id := range []gestor.SourceID{"", "runtime", "runtime. components"} {
		if err := id.Validate(); !errors.Is(err, gestor.ErrInvalidSourceID) {
			t.Errorf("source ID %q: expected ErrInvalidSourceID, got %v", id, err)
		}
	}
}

func TestKnownCapabilitiesAreValidOrderedAndDefensive(t *testing.T) {
	capabilities := gestor.KnownCapabilities()
	if len(capabilities) != 17 {
		t.Fatalf("expected 17 known capabilities, got %d", len(capabilities))
	}
	for index, capability := range capabilities {
		if err := capability.Validate(); err != nil {
			t.Fatalf("capability %d: %v", index, err)
		}
		if index > 0 && capabilities[index-1].Compare(capability) >= 0 {
			t.Fatalf("capabilities are not ordered at %d", index)
		}
	}

	capabilities[0] = "changed.value"
	if gestor.KnownCapabilities()[0] == "changed.value" {
		t.Fatal("KnownCapabilities exposed its backing array")
	}
}

func TestTargetValidation(t *testing.T) {
	valid := []gestor.Target{
		componentTarget("workspace"),
		providerTarget("ollama", gestor.ScopeAdapter, ""),
		providerTarget("ollama", gestor.ScopeInstance, ""),
		providerTarget("ollama", gestor.ScopeModel, "llama3.1:8b"),
	}
	for _, target := range valid {
		if err := target.Validate(); err != nil {
			t.Errorf("validate target %#v: %v", target, err)
		}
	}

	invalid := []gestor.Target{
		{},
		{Kind: gestor.TargetKindComponent, ID: "", Scope: gestor.ScopeComponent},
		{Kind: gestor.TargetKindComponent, ID: " workspace", Scope: gestor.ScopeComponent},
		{Kind: gestor.TargetKindComponent, ID: "workspace", Scope: gestor.ScopeModel, Model: "model"},
		{Kind: gestor.TargetKindProvider, ID: "ollama", Scope: gestor.ScopeComponent},
		{Kind: gestor.TargetKindProvider, ID: "ollama", Scope: gestor.ScopeModel},
		{Kind: gestor.TargetKindProvider, ID: "ollama", Scope: gestor.ScopeModel, Model: " model"},
		{Kind: gestor.TargetKindProvider, ID: "ollama", Scope: gestor.ScopeInstance, Model: "model"},
	}
	for _, target := range invalid {
		if err := target.Validate(); !errors.Is(err, gestor.ErrInvalidTarget) {
			t.Errorf("target %#v: expected ErrInvalidTarget, got %v", target, err)
		}
	}

	if componentTarget("a").Compare(componentTarget("b")) >= 0 {
		t.Fatal("target comparison is not lexical")
	}
}

func TestDescriptorValidationAndAvailability(t *testing.T) {
	valid := descriptor(gestor.CapabilityRuntimeStart, componentTarget("workspace"), "runtime.components")
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate descriptor: %v", err)
	}

	for _, availability := range []gestor.Availability{
		gestor.AvailabilityUnknown,
		gestor.AvailabilityAvailable,
		gestor.AvailabilityUnavailable,
	} {
		candidate := valid
		candidate.Availability = availability
		if err := candidate.Validate(); err != nil {
			t.Errorf("validate availability %q: %v", availability, err)
		}
	}

	invalid := valid
	invalid.Availability = "maybe"
	err := invalid.Validate()
	if !errors.Is(err, gestor.ErrInvalidAvailability) || !errors.Is(err, gestor.ErrInvalidDescriptor) {
		t.Fatalf("expected availability and descriptor sentinels, got %v", err)
	}
}

func TestQueryValidationAndDefensivePreferences(t *testing.T) {
	preferred := []gestor.Target{
		providerTarget("ollama", gestor.ScopeModel, "llama3.1:8b"),
		providerTarget("llamacpp", gestor.ScopeModel, "llama3.1:8b"),
	}
	query, err := gestor.NewQuery(
		gestor.CapabilityProviderCompletion,
		gestor.QueryOptions{
			TargetKind:       gestor.TargetKindProvider,
			Scope:            gestor.ScopeModel,
			Model:            "llama3.1:8b",
			RequireAvailable: true,
			PreferredTargets: preferred,
		},
	)
	if err != nil {
		t.Fatalf("new query: %v", err)
	}
	if query.Capability() != gestor.CapabilityProviderCompletion ||
		query.TargetKind() != gestor.TargetKindProvider ||
		query.Scope() != gestor.ScopeModel ||
		query.Model() != "llama3.1:8b" ||
		!query.RequireAvailable() {
		t.Fatalf("query getters returned unexpected values")
	}

	preferred[0].ID = "changed-input"
	got := query.PreferredTargets()
	if got[0].ID != "ollama" {
		t.Fatal("query retained the caller's preference slice")
	}
	got[0].ID = "changed-output"
	if query.PreferredTargets()[0].ID != "ollama" {
		t.Fatal("query exposed its preference slice")
	}

	invalidQueries := []gestor.QueryOptions{
		{Scope: gestor.ScopeModel},
		{Model: "model"},
		{TargetKind: gestor.TargetKindComponent, Scope: gestor.ScopeAdapter},
		{PreferredTargets: []gestor.Target{componentTarget("same"), componentTarget("same")}},
	}
	for _, options := range invalidQueries {
		_, err := gestor.NewQuery(gestor.CapabilityRuntimeStart, options)
		if !errors.Is(err, gestor.ErrInvalidQuery) {
			t.Errorf("options %#v: expected ErrInvalidQuery, got %v", options, err)
		}
	}

	if err := (gestor.Query{}).Validate(); !errors.Is(err, gestor.ErrInvalidQuery) {
		t.Fatalf("zero query: expected ErrInvalidQuery, got %v", err)
	}
}

func TestSnapshotIsOrderedValidatedAndDefensive(t *testing.T) {
	descriptors := []gestor.Descriptor{
		descriptor(gestor.CapabilityRuntimeStop, componentTarget("workspace"), "runtime.components"),
		descriptor(gestor.CapabilityProviderCompletion, providerTarget("ollama", gestor.ScopeAdapter, ""), "provider.capabilities"),
		descriptor(gestor.CapabilityRuntimeStart, componentTarget("workspace"), "runtime.components"),
	}
	snapshot, err := gestor.NewSnapshot(4, true, descriptors)
	if err != nil {
		t.Fatalf("new snapshot: %v", err)
	}

	descriptors[0].Capability = "changed.value"
	got := snapshot.Descriptors()
	if !slices.IsSortedFunc(got, func(left, right gestor.Descriptor) int { return left.Compare(right) }) {
		t.Fatal("snapshot descriptors are not deterministic")
	}
	if got[0].Capability == "changed.value" {
		t.Fatal("snapshot retained the caller's descriptor slice")
	}
	got[0].Capability = "changed.output"
	if snapshot.Descriptors()[0].Capability == "changed.output" {
		t.Fatal("snapshot exposed its descriptor slice")
	}

	metadata := snapshot.Metadata()
	if metadata.Generation != 4 || !metadata.Current || metadata.DescriptorCount != 3 {
		t.Fatalf("unexpected snapshot metadata: %#v", metadata)
	}
	sources := metadata.Sources()
	wantSources := []gestor.SourceID{"provider.capabilities", "runtime.components"}
	if !slices.Equal(sources, wantSources) {
		t.Fatalf("expected sources %v, got %v", wantSources, sources)
	}
	sources[0] = "changed.source"
	if snapshot.Metadata().Sources()[0] != "provider.capabilities" {
		t.Fatal("snapshot metadata exposed its sources")
	}

	duplicate := []gestor.Descriptor{descriptors[1], descriptors[1]}
	if _, err := gestor.NewSnapshot(5, true, duplicate); !errors.Is(err, gestor.ErrInvalidSnapshot) {
		t.Fatalf("duplicate snapshot: expected ErrInvalidSnapshot, got %v", err)
	}
}

func TestResolutionIsValidatedAndDefensive(t *testing.T) {
	selected := descriptor(gestor.CapabilityRuntimeStart, componentTarget("app"), "runtime.components")
	snapshot, err := gestor.NewSnapshot(7, true, []gestor.Descriptor{selected})
	if err != nil {
		t.Fatalf("new snapshot: %v", err)
	}
	dependencies := []gestor.Target{componentTarget("config"), componentTarget("logger")}
	resolution, err := gestor.NewResolution(
		selected,
		snapshot.Metadata(),
		gestor.ResolutionSingleCandidate,
		dependencies,
	)
	if err != nil {
		t.Fatalf("new resolution: %v", err)
	}
	if resolution.Descriptor() != selected || resolution.Reason() != gestor.ResolutionSingleCandidate {
		t.Fatal("resolution getters returned unexpected values")
	}

	dependencies[0].ID = "changed-input"
	got := resolution.Dependencies()
	if got[0].ID != "config" {
		t.Fatal("resolution retained the caller's dependency slice")
	}
	got[0].ID = "changed-output"
	if resolution.Dependencies()[0].ID != "config" {
		t.Fatal("resolution exposed its dependency slice")
	}

	stale, err := gestor.NewSnapshot(7, false, []gestor.Descriptor{selected})
	if err != nil {
		t.Fatalf("new stale snapshot: %v", err)
	}
	_, err = gestor.NewResolution(selected, stale.Metadata(), gestor.ResolutionSingleCandidate, nil)
	if !errors.Is(err, gestor.ErrStaleSnapshot) || !errors.Is(err, gestor.ErrInvalidResolution) {
		t.Fatalf("expected stale and invalid resolution sentinels, got %v", err)
	}
}

func TestSentinelsRemainCompatibleThroughWrapping(t *testing.T) {
	for _, sentinel := range []error{
		gestor.ErrInvalidQuery,
		gestor.ErrSourceFailure,
		gestor.ErrNotFound,
		gestor.ErrUnavailable,
		gestor.ErrAmbiguous,
		gestor.ErrStaleSnapshot,
		gestor.ErrInvalidDescriptor,
	} {
		wrapped := fmt.Errorf("operation failed: %w", sentinel)
		if !errors.Is(wrapped, sentinel) {
			t.Errorf("errors.Is did not preserve %v", sentinel)
		}
	}
}

type contractSource struct{}

func (contractSource) ID() gestor.SourceID { return "test.source" }
func (contractSource) Discover(context.Context) ([]gestor.Descriptor, error) {
	return nil, nil
}

type contractRegistry struct{}

func (contractRegistry) RegisterSource(gestor.Source) error { return nil }
func (contractRegistry) Refresh(context.Context) error      { return nil }
func (contractRegistry) Invalidate()                        {}
func (contractRegistry) Snapshot() gestor.Snapshot          { return gestor.Snapshot{} }

type contractResolver struct{}

func (contractResolver) Candidates(gestor.Query) ([]gestor.Descriptor, error) {
	return nil, nil
}
func (contractResolver) Resolve(gestor.Query) (gestor.Resolution, error) {
	return gestor.Resolution{}, nil
}

var (
	_ gestor.Source   = contractSource{}
	_ gestor.Registry = contractRegistry{}
	_ gestor.Resolver = contractResolver{}
)
