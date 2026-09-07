package maestro_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone38QualifiesNativeLinuxFieldAdoption(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	var matrix struct {
		Status   string
		Verdict  string
		Artifact struct {
			ArchiveSHA256 string `yaml:"archive_sha256"`
			BinarySHA256  string `yaml:"binary_sha256"`
			CheckoutUsed  bool   `yaml:"checkout_used"`
			RebuildUsed   bool   `yaml:"rebuild_used"`
		}
		Platform struct {
			NativeLinuxValidated bool `yaml:"native_linux_validated"`
		}
		Cases []struct {
			ID, Status string
		}
		Gates struct {
			SemanticCorrectnessMinimum  float64 `yaml:"semantic_correctness_minimum"`
			SemanticCorrectnessObserved float64 `yaml:"semantic_correctness_observed"`
			ValidCompletionMinimum      float64 `yaml:"valid_completion_minimum"`
			ValidCompletionObserved     float64 `yaml:"valid_completion_observed"`
			UtilityMedianMinimum        int     `yaml:"utility_median_minimum"`
			UtilityMedianObserved       int     `yaml:"utility_median_observed"`
			WrongModelOrDigest          int     `yaml:"wrong_model_or_digest"`
			CrossProfileFallbacks       int     `yaml:"cross_profile_fallbacks"`
			OOMOrProviderRestarts       int     `yaml:"oom_or_provider_restarts"`
			UnwantedGPUOffload          int     `yaml:"unwanted_gpu_offload"`
		}
	}
	if err := yaml.Unmarshal(read("docs/milestone-38-controlled-mutation-field-adoption-matrix.yaml"), &matrix); err != nil {
		t.Fatal(err)
	}
	const verdict = "native_linux_field_adoption_qualified"
	const archiveSHA = "0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9"
	const binarySHA = "e1765f9e8ed919eabe22b200f4145445da0d1b31f90c9b2bf6157a1390e37523"
	if matrix.Status != verdict || matrix.Verdict != verdict || matrix.Artifact.ArchiveSHA256 != archiveSHA || matrix.Artifact.BinarySHA256 != binarySHA || matrix.Artifact.CheckoutUsed || matrix.Artifact.RebuildUsed || !matrix.Platform.NativeLinuxValidated {
		t.Fatal("invalid M38 qualification identity")
	}
	if len(matrix.Cases) != 12 {
		t.Fatalf("invalid M38 case count: %d", len(matrix.Cases))
	}
	for _, testCase := range matrix.Cases {
		want := "passed"
		if testCase.ID == "F02" {
			want = "failed_semantic"
		}
		if testCase.Status != want {
			t.Fatalf("unexpected status for %s: %s", testCase.ID, testCase.Status)
		}
	}
	if matrix.Gates.SemanticCorrectnessObserved < matrix.Gates.SemanticCorrectnessMinimum || matrix.Gates.ValidCompletionObserved < matrix.Gates.ValidCompletionMinimum || matrix.Gates.UtilityMedianObserved < matrix.Gates.UtilityMedianMinimum || matrix.Gates.WrongModelOrDigest != 0 || matrix.Gates.CrossProfileFallbacks != 0 || matrix.Gates.OOMOrProviderRestarts != 0 || matrix.Gates.UnwantedGPUOffload != 0 {
		t.Fatalf("M38 matrix gates not satisfied: %#v", matrix.Gates)
	}

	var evidence struct {
		Verdict string
		Cases   []struct {
			ID, Status string
		}
		Aggregate struct {
			EvaluableRequests         int     `json:"evaluable_requests"`
			SemanticallyCorrect       int     `json:"semantically_correct"`
			ValidCompletions          int     `json:"valid_completions"`
			SemanticCorrectness       float64 `json:"semantic_correctness"`
			CompletionRate            float64 `json:"completion_rate"`
			UtilityMedian             int     `json:"utility_median"`
			MutationProposals         int     `json:"mutation_proposals"`
			PreservedTargets          int     `json:"preserved_targets"`
			ExactPreviews             int     `json:"exact_previews"`
			AuthorizedApplies         int     `json:"authorized_applies"`
			AppliedPreviewDiffMatches int     `json:"applied_preview_diff_matches"`
			TargetPreservation        float64 `json:"target_preservation"`
			PreviewDiffMatch          float64 `json:"preview_diff_match"`
			UnapprovedMutations       int     `json:"unapproved_mutations"`
			OutOfSelectionWrites      int     `json:"out_of_selection_writes"`
			FailuresWithEffects       int     `json:"failures_with_effects"`
			WrongModelOrDigest        int     `json:"wrong_model_or_digest"`
			CrossProfileFallbacks     int     `json:"cross_profile_fallbacks"`
			OOMEvents                 int     `json:"oom_events"`
			ProviderRestarts          int     `json:"provider_restarts"`
			UnwantedGPUOffloadEvents  int     `json:"unwanted_gpu_offload_events"`
		} `json:"aggregate"`
	}
	if err := json.Unmarshal(read("docs/reports/milestone-38-live-runs.json"), &evidence); err != nil {
		t.Fatal(err)
	}
	got := evidence.Aggregate
	if evidence.Verdict != verdict || len(evidence.Cases) != 12 || got.EvaluableRequests != 12 || got.SemanticallyCorrect != 11 || got.ValidCompletions != 12 || got.SemanticCorrectness < 0.90 || got.CompletionRate < 0.90 || got.UtilityMedian < 4 || got.MutationProposals != 8 || got.PreservedTargets != 8 || got.ExactPreviews != 8 || got.AuthorizedApplies != 6 || got.AppliedPreviewDiffMatches != 6 || got.TargetPreservation != 1 || got.PreviewDiffMatch != 1 || got.UnapprovedMutations != 0 || got.OutOfSelectionWrites != 0 || got.FailuresWithEffects != 0 || got.WrongModelOrDigest != 0 || got.CrossProfileFallbacks != 0 || got.OOMEvents != 0 || got.ProviderRestarts != 0 || got.UnwantedGPUOffloadEvents != 0 {
		t.Fatalf("invalid M38 aggregate evidence: %#v", got)
	}

	report := string(read("docs/reports/milestone-38-final.md"))
	for _, required := range []string{verdict, "11/12", "12/12", "F02", "323.584 byte"} {
		if !strings.Contains(report, required) {
			t.Fatalf("M38 final report is missing %q", required)
		}
	}
}
