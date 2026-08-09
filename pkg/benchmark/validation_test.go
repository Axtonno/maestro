package benchmark

import (
	"errors"
	"math"
	"testing"
)

func TestManifestValidationAcceptsVersionedContract(t *testing.T) {
	manifest := validTestManifest()

	if err := manifest.Validate(); err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
}

func TestManifestValidationRejectsDuplicateScenariosAndMissingStates(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{
			name: "duplicate scenario",
			mutate: func(manifest *Manifest) {
				manifest.Scenarios = append(manifest.Scenarios, manifest.Scenarios[0])
			},
		},
		{
			name: "missing result state",
			mutate: func(manifest *Manifest) {
				manifest.ResultStates = manifest.ResultStates[:3]
			},
		},
		{
			name: "unsupported version",
			mutate: func(manifest *Manifest) {
				manifest.Version++
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := validTestManifest()
			test.mutate(&manifest)

			if err := manifest.Validate(); !errors.Is(err, ErrInvalidManifest) {
				t.Fatalf("expected invalid manifest, got %v", err)
			}
		})
	}
}

func TestIterationResultValidationProtectsReportSemantics(t *testing.T) {
	tests := []IterationResult{
		{State: ResultFailed},
		{State: ResultSkipped},
		{
			State: ResultPassed,
			Measurements: []Measurement{{
				Name: "latency", Unit: "ms", Value: math.NaN(),
			}},
		},
		{
			State:        ResultUnsupported,
			ReasonCode:   "capability_not_supported",
			Measurements: []Measurement{{Name: "latency", Unit: "ms", Value: 1}},
		},
	}

	for index, result := range tests {
		if err := result.Validate(); err == nil {
			t.Fatalf("case %d unexpectedly passed validation", index)
		}
	}

	valid := IterationResult{
		State:        ResultPassed,
		Measurements: []Measurement{{Name: "latency", Unit: "ms", Value: 1}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate iteration result: %v", err)
	}
}

func validTestManifest() Manifest {
	return Manifest{
		Version: ManifestSchemaVersion,
		Owner:   "benchmark-tests",
		ReportRedaction: RedactionPolicy{Exclude: append(
			[]string(nil), requiredRedactionFields...,
		)},
		Providers: map[string]ProviderManifest{
			"test": {},
		},
		Scenarios: []ScenarioDefinition{{
			ID: "completion", Capability: "completion",
			ModelRole: "chat", Cleanup: "none",
		}},
		ResultStates: []ResultState{
			ResultPassed, ResultFailed, ResultSkipped, ResultUnsupported,
		},
	}
}
