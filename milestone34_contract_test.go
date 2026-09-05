package maestro_test

import (
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone34ClosesBranchBWithoutAuthorizingMutation(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var matrix struct {
		Status      string
		Attribution struct {
			Branch string `yaml:"selected_branch"`
		}
		Gates    map[string]float64
		Decision struct {
			Verdict     string
			Authorized  bool   `yaml:"v0.5.0_candidate_authorized"`
			Tuning      bool   `yaml:"further_same_profile_tuning_allowed"`
			Generations int    `yaml:"new_provider_generations"`
			Reruns      int    `yaml:"m33_reruns"`
			Next        string `yaml:"next_plan"`
		}
	}
	if err := yaml.Unmarshal(read("docs/milestone-34-host-bound-mutation-failure-attribution-matrix.yaml"), &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Status != "completed_profile_rejected" || matrix.Attribution.Branch != "correct_prompt_and_payload" || matrix.Decision.Verdict != "qwen3.5_9b_host_bound_mutation_profile_rejected" || matrix.Decision.Authorized || matrix.Decision.Tuning || matrix.Decision.Generations != 0 || matrix.Decision.Reruns != 0 {
		t.Fatal("inconsistent M34 decision")
	}
	if matrix.Gates["correct_positive_proposal_rate_minimum"] != .9 || matrix.Gates["holdout_apply_completed_rate"] != 1 {
		t.Fatal("qualification thresholds changed")
	}
	read(matrix.Decision.Next)
	var evidence struct {
		Kind, ModelDigest, PromptSHA256, SchemaSHA256, MatrixSHA256 string
		ProviderGenerations                                         int
		Cases                                                       []struct {
			ID                    string
			ExpectedSpliceCorrect bool
		}
	}
	if err := json.Unmarshal(read("docs/reports/milestone-34-offline-reconstruction.json"), &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Kind != "offline_reconstruction_not_historical_wire_capture" || evidence.ProviderGenerations != 0 || len(evidence.Cases) != 3 {
		t.Fatal("invalid reconstruction provenance")
	}
	for i, id := range []string{"M33-D03", "M33-H01", "M33-H02"} {
		if evidence.Cases[i].ID != id || !evidence.Cases[i].ExpectedSpliceCorrect {
			t.Fatal("incomplete attribution cases")
		}
	}
	var m33 struct{ ModelDigest, PromptSHA256, SchemaSHA256, MatrixSHA256 string }
	if err := json.Unmarshal(read("docs/reports/milestone-33-live-runs.json"), &m33); err != nil {
		t.Fatal(err)
	}
	if evidence.ModelDigest != m33.ModelDigest || evidence.PromptSHA256 != m33.PromptSHA256 || evidence.SchemaSHA256 != m33.SchemaSHA256 || evidence.MatrixSHA256 != m33.MatrixSHA256 {
		t.Fatal("M33 identity mismatch")
	}
	var tags struct {
		Models []struct{ Name, Digest string }
	}
	if err := json.Unmarshal(read("docs/reports/milestone-34-model-tags.json"), &tags); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, model := range tags.Models {
		if model.Name == "qwen3.5:9b" && model.Digest == m33.ModelDigest {
			found = true
		}
	}
	if !found {
		t.Fatal("model identity not substantiated")
	}
}
