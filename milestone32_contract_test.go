package maestro_test

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone32BinaryDecisionContractIsOpen(t *testing.T) {
	encoded, err := os.ReadFile("docs/milestone-32-mutation-decision-contract-simplification-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var matrix struct {
		Version  int    `yaml:"version"`
		Status   string `yaml:"status"`
		Baseline struct {
			Authorized bool `yaml:"v0.5.0_authorized"`
		} `yaml:"baseline"`
		Contract struct {
			ID              string   `yaml:"id"`
			AbstainTerminal string   `yaml:"abstain_public_terminal"`
			Decisions       []string `yaml:"decisions"`
			Reason          bool     `yaml:"model_reason_authoritative"`
		} `yaml:"contract"`
		Terminals        map[string]string `yaml:"maestro_terminals"`
		Classes          []string          `yaml:"required_classes"`
		Positive, Denial []string
		Holdout          struct {
			Reused  int  `yaml:"reused_from_m30_or_m31"`
			Visible bool `yaml:"visible_during_prompt_design"`
			Minimum int  `yaml:"minimum_cases"`
		} `yaml:"holdout"`
		Stop struct {
			Verdict    string `yaml:"repeated_binary_contract_false_negatives"`
			Exceptions bool   `yaml:"further_prompt_exceptions_allowed"`
		} `yaml:"stop_rule"`
		Decision struct {
			Verdict    string
			Authorized bool `yaml:"v0.5.0_candidate_authorized"`
		} `yaml:"decision"`
	}
	if err := yaml.Unmarshal(encoded, &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Version != 1 || matrix.Status != "design_open_not_frozen" || matrix.Baseline.Authorized || matrix.Decision.Authorized || matrix.Decision.Verdict != "binary_mutation_decision_not_qualified" {
		t.Fatalf("unexpected identity: %#v", matrix)
	}
	if matrix.Contract.ID != "mutation-binary-decision-v1" || !slices.Equal(matrix.Contract.Decisions, []string{"propose", "abstain"}) || matrix.Contract.AbstainTerminal != "insufficient_information" || matrix.Contract.Reason {
		t.Fatalf("model authority not binary: %#v", matrix.Contract)
	}
	for _, terminal := range []string{"target_not_found", "target_ambiguous", "protected_target", "stale_source", "approval_rejected"} {
		if _, ok := matrix.Terminals[terminal]; !ok {
			t.Fatalf("missing Maestro terminal %s", terminal)
		}
	}
	if len(matrix.Classes) != 11 || matrix.Holdout.Minimum < 7 || matrix.Holdout.Reused != 0 || matrix.Holdout.Visible || matrix.Stop.Verdict != "controlled_mutation_model_profile_rejected" || matrix.Stop.Exceptions {
		t.Fatalf("incomplete qualification contract: %#v", matrix)
	}
}
