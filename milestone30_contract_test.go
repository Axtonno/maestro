package maestro_test

import (
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone30AbstentionRecoveryContractIsClosed(t *testing.T) {
	encoded, err := os.ReadFile("docs/milestone-30-structured-mutation-abstention-recovery-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var matrix struct {
		Version  int    `yaml:"version"`
		Status   string `yaml:"status"`
		Baseline struct {
			Engine     string `yaml:"mutation_engine_state"`
			Transport  string `yaml:"transport_candidate"`
			Abstention string `yaml:"abstention_state"`
			Authorized bool   `yaml:"v0.5.0_authorized"`
		} `yaml:"baseline"`
		Contract struct {
			Envelope, Proposal                string
			Decisions                         []string `yaml:"decisions"`
			Repair, Fallback, Retry, Invented bool
		} `yaml:"contract"`
		Development []struct{ ID, Class string } `yaml:"development_matrix"`
		Holdout     struct {
			Status  string                       `yaml:"status"`
			Reused  int                          `yaml:"reused_from_development"`
			Visible bool                         `yaml:"visible_during_prompt_design"`
			Cases   []struct{ ID, Class string } `yaml:"cases"`
		} `yaml:"holdout"`
		Gates struct {
			Positive, Abstention, Valid                        float64
			Invalid, Invented, WithoutApproval, FailureEffects int
		} `yaml:"gates"`
		Decision struct {
			Verdict    string `yaml:"verdict"`
			Authorized bool   `yaml:"v0.5.0_candidate_authorized"`
		} `yaml:"decision"`
	}
	if err := yaml.Unmarshal(encoded, &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Version != 1 || matrix.Status != "completed_rejected" || matrix.Baseline.Engine != "controlled_mutation_engine_ready" || matrix.Baseline.Transport != "constrained_structured_output" || matrix.Baseline.Abstention != "structured_mutation_abstention_rejected" || matrix.Baseline.Authorized {
		t.Fatalf("unexpected identity: %#v", matrix.Baseline)
	}
	want := []string{"propose", "abstain_missing_information", "abstain_target_not_found", "abstain_target_ambiguous"}
	if !slices.Equal(matrix.Contract.Decisions, want) || len(matrix.Development) != 9 {
		t.Fatalf("unexpected contract or matrix: %#v", matrix.Contract)
	}
	if matrix.Holdout.Status != "frozen_independent" || len(matrix.Holdout.Cases) != 5 || matrix.Holdout.Reused != 0 || matrix.Holdout.Visible {
		t.Fatalf("holdout is not independent: %#v", matrix.Holdout)
	}
	if matrix.Decision.Verdict != "structured_mutation_abstention_rejected" || matrix.Decision.Authorized {
		t.Fatalf("unexpected decision: %#v", matrix.Decision)
	}
}
