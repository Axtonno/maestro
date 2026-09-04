package maestro_test

import (
	"gopkg.in/yaml.v3"
	"os"
	"slices"
	"testing"
)

func TestMilestone32BinaryDecisionContractIsFrozen(t *testing.T) {
	encoded, err := os.ReadFile("docs/milestone-32-mutation-decision-contract-simplification-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Version  int
		Status   string
		Baseline struct {
			Authorized bool `yaml:"v0.5.0_authorized"`
		}
		Contract struct {
			ID              string
			AbstainTerminal string `yaml:"abstain_public_terminal"`
			Decisions       []string
			Reason          bool `yaml:"model_reason_authoritative"`
		}
		Terminals   map[string]string     `yaml:"maestro_terminals"`
		Order       []string              `yaml:"qualification_order"`
		Development []struct{ ID string } `yaml:"development_cases"`
		Holdout     struct {
			Reused  int  `yaml:"reused_from_m30_or_m31"`
			Visible bool `yaml:"visible_during_prompt_design"`
			Cases   []struct{ ID string }
		}
		Stop struct {
			Threshold  float64 `yaml:"frequent_false_negative_threshold"`
			Verdict    string  `yaml:"repeated_binary_contract_false_negatives"`
			Exceptions bool    `yaml:"further_prompt_exceptions_allowed"`
		} `yaml:"stop_rule"`
		Decision struct {
			Verdict    string
			Authorized bool `yaml:"v0.5.0_candidate_authorized"`
		}
	}
	if err := yaml.Unmarshal(encoded, &m); err != nil {
		t.Fatal(err)
	}
	if m.Version != 1 || m.Status != "completed_rejected" || m.Baseline.Authorized || m.Decision.Authorized || m.Decision.Verdict != "binary_mutation_decision_rejected" {
		t.Fatalf("unexpected identity: %#v", m)
	}
	if m.Contract.ID != "mutation-binary-decision-v1" || !slices.Equal(m.Contract.Decisions, []string{"propose", "abstain"}) || m.Contract.AbstainTerminal != "insufficient_information" || m.Contract.Reason {
		t.Fatalf("model authority not binary: %#v", m.Contract)
	}
	for _, terminal := range []string{"target_not_found", "target_ambiguous", "protected_target", "stale_source", "approval_rejected"} {
		if _, ok := m.Terminals[terminal]; !ok {
			t.Fatalf("missing Maestro terminal %s", terminal)
		}
	}
	if len(m.Development) != 11 || len(m.Holdout.Cases) != 7 || len(m.Order) != 18 || m.Holdout.Reused != 0 || m.Holdout.Visible || m.Stop.Threshold != 0.80 || m.Stop.Verdict != "controlled_mutation_model_profile_rejected" || m.Stop.Exceptions {
		t.Fatalf("incomplete frozen qualification contract: %#v", m)
	}
}
