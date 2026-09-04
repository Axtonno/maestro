package maestro_test

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone31DeterministicRejectionContractIsClosed(t *testing.T) {
	encoded, err := os.ReadFile("docs/milestone-31-deterministic-mutation-rejection-qualification-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var matrix struct {
		Version  int    `yaml:"version"`
		Status   string `yaml:"status"`
		Baseline struct {
			Transport, Positive, Abstention, Safety string
			Authorized                              bool `yaml:"v0.5.0_authorized"`
		} `yaml:"baseline"`
		Required    map[string]string            `yaml:"required_terminals"`
		Accepted    map[string][]string          `yaml:"accepted_mechanical_paths"`
		Development []struct{ ID, Class string } `yaml:"development_cases"`
		Holdout     struct {
			Status  string
			Reused  int                          `yaml:"reused_from_m30"`
			Visible bool                         `yaml:"visible_during_implementation"`
			Cases   []struct{ ID, Class string } `yaml:"cases"`
		} `yaml:"holdout"`
		Decision struct {
			Verdict    string
			Authorized bool `yaml:"v0.5.0_candidate_authorized"`
		} `yaml:"decision"`
	}
	if err := yaml.Unmarshal(encoded, &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Version != 1 || matrix.Status != "completed_rejected" || matrix.Baseline.Authorized || matrix.Decision.Authorized || matrix.Decision.Verdict != "deterministic_mutation_rejection_rejected" {
		t.Fatalf("unexpected M31 identity: %#v", matrix)
	}
	want := []string{"target_not_found", "target_ambiguous", "stale_source", "protected_target", "approval_rejected"}
	for _, terminal := range want {
		if _, ok := matrix.Required[terminal]; !ok {
			t.Fatalf("missing terminal %s", terminal)
		}
	}
	if !slices.Equal(matrix.Accepted["target_ambiguous"], []string{"abstain_target_ambiguous", "target_ambiguous"}) {
		t.Fatalf("unexpected ambiguity paths: %v", matrix.Accepted)
	}
	if len(matrix.Development) != 11 || len(matrix.Holdout.Cases) != 6 || matrix.Holdout.Reused != 0 || matrix.Holdout.Visible {
		t.Fatalf("matrix or holdout incomplete: %#v", matrix.Holdout)
	}
}
