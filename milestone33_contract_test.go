package maestro_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone33HostBoundContractIsOpen(t *testing.T) {
	encoded, err := os.ReadFile("docs/milestone-33-host-bound-target-mutation-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Version  int
		Status   string
		Contract struct {
			ID                string
			Fields, Forbidden []string `yaml:",omitempty"`
			TargetAuthority   string   `yaml:"target_authority"`
		}
		Required []string `yaml:"required_cases"`
		Gates    struct {
			Target   float64 `yaml:"host_bound_target_preserved_rate"`
			Positive float64 `yaml:"correct_positive_rate_minimum"`
			Preview  float64 `yaml:"preview_matches_selected_target_rate"`
		}
		Decision struct {
			Verdict    string
			Authorized bool `yaml:"v0.5.0_candidate_authorized"`
		}
	}
	if err := yaml.Unmarshal(encoded, &m); err != nil {
		t.Fatal(err)
	}
	// Decode the differently named model-field lists explicitly to keep the
	// assertions readable.
	var lists struct {
		Contract struct {
			Fields    []string `yaml:"model_fields"`
			Forbidden []string `yaml:"forbidden_model_fields"`
		}
	}
	if err := yaml.Unmarshal(encoded, &lists); err != nil {
		t.Fatal(err)
	}
	if m.Version != 1 || m.Status != "design_open_not_frozen" || m.Decision.Authorized || m.Decision.Verdict != "host_bound_mutation_not_yet_qualified" {
		t.Fatalf("unexpected identity: %#v", m)
	}
	if m.Contract.ID != "host-bound-mutation-decision-v1" || m.Contract.TargetAuthority != "maestro_and_user" || !slices.Equal(lists.Contract.Fields, []string{"decision", "new_text"}) {
		t.Fatalf("target authority is not host-bound")
	}
	for _, forbidden := range []string{"path", "old_text", "start_line", "end_line", "operation"} {
		if !slices.Contains(lists.Contract.Forbidden, forbidden) {
			t.Fatalf("missing forbidden model field %s", forbidden)
		}
	}
	if len(m.Required) != 10 || m.Gates.Target != 1 || m.Gates.Positive != .8 || m.Gates.Preview != 1 {
		t.Fatalf("incomplete qualification contract: %#v", m)
	}
	schema, err := os.ReadFile("docs/schemas/host-bound-mutation-decision-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(schema, &document); err != nil {
		t.Fatal(err)
	}
}
