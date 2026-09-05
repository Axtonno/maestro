package maestro_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone33HostBoundContractIsClosedRejected(t *testing.T) {
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
	if m.Version != 1 || m.Status != "completed_rejected" || m.Decision.Authorized || m.Decision.Verdict != "host_bound_mutation_rejected" {
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

func TestMilestone33EvidenceMatchesFrozenCases(t *testing.T) {
	data, err := os.ReadFile("docs/milestone-33-cases.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var cases struct {
		Cases []struct{ ID, Set, Expected, Approval, Raw string }
	}
	if err := yaml.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("docs/reports/milestone-33-live-runs.json")
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Verdict             string
		CandidateAuthorized bool
		Runs                []struct {
			ID, Set, Expected, Terminal                                                                                                                                                    string
			ProviderCalled, PositiveCorrect, ApprovalReached, Preview, PreviewExact, TargetPreserved, WorkspaceCorrect, Applied, FailureWithEffect, UnapprovedEffect, OutOfSelectionEffect bool
		}
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Verdict != "host_bound_mutation_rejected" || report.CandidateAuthorized || len(report.Runs) != 19 || len(cases.Cases) != 19 {
		t.Fatal("invalid conclusion")
	}
	live, positive, approval, terminals, applied := 0, 0, 0, 0, 0
	for i, r := range report.Runs {
		c := cases.Cases[i]
		if r.ID != c.ID || r.Set != c.Set || r.Expected != c.Expected {
			t.Fatalf("case mismatch %s", r.ID)
		}
		if !r.TargetPreserved || !r.WorkspaceCorrect || r.Preview && !r.PreviewExact || r.FailureWithEffect || r.UnapprovedEffect || r.OutOfSelectionEffect {
			t.Fatalf("safety failure %s", r.ID)
		}
		if r.ProviderCalled {
			live++
		}
		if r.PositiveCorrect {
			positive++
		}
		if r.ApprovalReached {
			approval++
		}
		if r.Terminal == r.Expected {
			terminals++
		}
		if r.Applied {
			applied++
		}
	}
	if live != 12 || positive != 7 || approval != 7 || terminals != 16 || applied != 3 {
		t.Fatalf("evidence changed: %d %d %d %d %d", live, positive, approval, terminals, applied)
	}
}
