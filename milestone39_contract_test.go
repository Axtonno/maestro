package maestro_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone39FreezesCrediblePublicBaseline(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	hash := func(data []byte) string {
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}

	const verdict = "documentation_onboarding_public_trial_ready"
	var matrix struct {
		Status          string
		Verdict         string
		Cases           map[string]string
		ExecutionPolicy struct {
			PublicAssetOnly       bool `yaml:"public_asset_only"`
			CheckoutAllowed       bool `yaml:"checkout_allowed"`
			RebuildAllowed        bool `yaml:"rebuild_allowed"`
			AttemptsPerCompletion int  `yaml:"attempts_per_completion"`
		} `yaml:"execution_policy"`
	}
	if err := yaml.Unmarshal(read("docs/milestone-39-public-trial-matrix.yaml"), &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Status != "completed" || matrix.Verdict != verdict || len(matrix.Cases) != 8 || !matrix.ExecutionPolicy.PublicAssetOnly || matrix.ExecutionPolicy.CheckoutAllowed || matrix.ExecutionPolicy.RebuildAllowed || matrix.ExecutionPolicy.AttemptsPerCompletion != 1 {
		t.Fatalf("invalid M39 matrix: %#v", matrix)
	}
	for name, status := range matrix.Cases {
		if status != "passed" {
			t.Fatalf("M39 case %s is %s", name, status)
		}
	}

	var evidence struct {
		Verdict string
		Doctor  struct {
			Passed int
			Total  int
			Failed int
		}
		Chat struct {
			Terminal        string
			Model           string
			SemanticCorrect bool `json:"semantic_correct"`
		}
		ControlledMutation struct {
			Terminal     string
			Model        string
			Approval     string
			PreviewExact bool `json:"preview_exact"`
		} `json:"controlled_mutation"`
		GitVerification struct {
			ChangedFiles         []string `json:"changed_files"`
			OutOfSelectionWrites int      `json:"out_of_selection_writes"`
		} `json:"git_verification"`
		Cases      map[string]string
		Violations map[string]int
	}
	if err := json.Unmarshal(read("docs/reports/milestone-39-public-trial.json"), &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.Verdict != verdict || evidence.Doctor.Passed != 14 || evidence.Doctor.Total != 14 || evidence.Doctor.Failed != 0 || evidence.Chat.Terminal != "completed" || evidence.Chat.Model != "qwen3.5:9b" || !evidence.Chat.SemanticCorrect || evidence.ControlledMutation.Terminal != "applied" || evidence.ControlledMutation.Model != "qwen2.5-coder:14b" || evidence.ControlledMutation.Approval != "allow_once" || !evidence.ControlledMutation.PreviewExact || len(evidence.GitVerification.ChangedFiles) != 1 || len(evidence.Cases) != 8 {
		t.Fatalf("invalid M39 trial evidence: %#v", evidence)
	}
	for name, count := range evidence.Violations {
		if count != 0 {
			t.Fatalf("M39 violation %s=%d", name, count)
		}
	}

	var publicFreeze struct {
		Baseline string
		Files    map[string]string
		Policy   struct {
			AssetImmutable      bool `yaml:"public_v0.5.0_asset_is_immutable"`
			EvidenceImmutable   bool `yaml:"evidence_m38_is_immutable"`
			NewFreezeRequired   bool `yaml:"baseline_changes_require_new_freeze"`
			QualificationNeeded bool `yaml:"future_capabilities_require_new_qualification"`
		}
	}
	if err := yaml.Unmarshal(read("docs/v0.5.0-public-baseline-freeze.yaml"), &publicFreeze); err != nil {
		t.Fatal(err)
	}
	if publicFreeze.Baseline != "v0.5.0_public_baseline_credible" || len(publicFreeze.Files) != 11 || !publicFreeze.Policy.AssetImmutable || !publicFreeze.Policy.EvidenceImmutable || !publicFreeze.Policy.NewFreezeRequired || !publicFreeze.Policy.QualificationNeeded {
		t.Fatal("invalid v0.5.0 public baseline freeze")
	}
	for path, expected := range publicFreeze.Files {
		if expected != hash(read(path)) {
			t.Fatalf("public baseline hash mismatch for %s", path)
		}
	}

	var freeze struct {
		Verdict       string
		AcceptedFiles map[string]string `yaml:"accepted_files"`
		Policy        struct {
			EvidenceImmutable               bool `yaml:"evidence_is_immutable"`
			ReinterpretationAllowed         bool `yaml:"reinterpretation_allowed"`
			CorrectionsRequireNewRecord     bool `yaml:"corrections_require_new_record"`
			FutureCapabilitiesNeedMilestone bool `yaml:"future_capabilities_require_new_milestone"`
		}
	}
	if err := yaml.Unmarshal(read("docs/milestone-39-public-trial-freeze.yaml"), &freeze); err != nil {
		t.Fatal(err)
	}
	if freeze.Verdict != verdict || len(freeze.AcceptedFiles) != 5 || !freeze.Policy.EvidenceImmutable || freeze.Policy.ReinterpretationAllowed || !freeze.Policy.CorrectionsRequireNewRecord || !freeze.Policy.FutureCapabilitiesNeedMilestone {
		t.Fatal("invalid M39 evidence freeze")
	}
	for path, expected := range freeze.AcceptedFiles {
		if expected != hash(read(path)) {
			t.Fatalf("M39 frozen evidence mismatch for %s", path)
		}
	}
}
