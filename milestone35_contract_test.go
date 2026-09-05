package maestro_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type milestone35Totals struct {
	Runs, ProviderRuns, Conforming, Positive, CorrectPositive             int
	NecessaryAbstentions, CorrectAbstentions, Proposals, PreservedTargets int
	Previews, ExactPreviews, ApprovalCases, ReachedApprovals              int
	ExpectedApplies, CompletedApplies, CorrectTerminals                   int
	StaleWrites, OutOfSelectionWrites, IncorrectAppliedMutations          int
	UnapprovedMutations, FailuresWithEffects                              int
	Passed                                                                bool
}

func TestMilestone35SelectsAndQualifiesDedicatedMutationModel(t *testing.T) {
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

	var selection struct {
		Verdict, SelectedModel, SelectedDigest string
	}
	if err := json.Unmarshal(read("docs/reports/milestone-35-selection-runs.json"), &selection); err != nil {
		t.Fatal(err)
	}
	const model = "qwen2.5-coder:14b"
	const digest = "9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849"
	if selection.Verdict != "mutation_specific_model_selected" || selection.SelectedModel != model || selection.SelectedDigest != digest {
		t.Fatal("invalid M35 selection")
	}

	var report struct {
		Verdict, Model, ModelDigest, MatrixSHA256, SchemaSHA256, PromptSHA256 string
		QualificationPassed                                                   bool
		Development, Holdout, Global                                          milestone35Totals
		Runs                                                                  []struct {
			ID, Set, Terminal, Expected                                       string
			ProviderCalled, Conforming, Positive, PositiveCorrect             bool
			NecessaryAbstention, AbstentionCorrect, Proposal, TargetPreserved bool
			Preview, PreviewExact, ApprovalExpected, ApprovalReached          bool
			ExpectedApply, Applied, WorkspaceCorrect, StaleWrite              bool
			OutOfSelectionWrite, IncorrectAppliedMutation, UnapprovedMutation bool
			FailureWithEffect                                                 bool
		}
	}
	if err := json.Unmarshal(read("docs/reports/milestone-35-qualification-runs.json"), &report); err != nil {
		t.Fatal(err)
	}
	if report.Verdict != "mutation_specific_model_qualified" || !report.QualificationPassed || report.Model != model || report.ModelDigest != digest || len(report.Runs) != 30 {
		t.Fatal("invalid M35 qualification identity")
	}
	if report.MatrixSHA256 != hash(read("docs/milestone-35-qualification-matrix.yaml")) || report.SchemaSHA256 != hash(read("docs/schemas/host-bound-mutation-decision-v1.schema.json")) || report.PromptSHA256 != hash(read("docs/prompts/mutation-host-bound-model-selection-v1.txt")) {
		t.Fatal("qualification freeze identity mismatch")
	}
	assertSet := func(name string, got milestone35Totals, runs, provider, positive, abstain, proposals, approvals, applies int) {
		t.Helper()
		if !got.Passed || got.Runs != runs || got.ProviderRuns != provider || got.Conforming != provider || got.Positive != positive || got.CorrectPositive != positive || got.NecessaryAbstentions != abstain || got.CorrectAbstentions != abstain || got.Proposals != proposals || got.PreservedTargets != proposals || got.Previews != proposals || got.ExactPreviews != proposals || got.ApprovalCases != approvals || got.ReachedApprovals != approvals || got.ExpectedApplies != applies || got.CompletedApplies != applies || got.CorrectTerminals != runs || got.StaleWrites != 0 || got.OutOfSelectionWrites != 0 || got.IncorrectAppliedMutations != 0 || got.UnapprovedMutations != 0 || got.FailuresWithEffects != 0 {
			t.Fatalf("invalid %s totals: %#v", name, got)
		}
	}
	assertSet("development", report.Development, 15, 12, 10, 2, 10, 10, 7)
	assertSet("holdout", report.Holdout, 15, 12, 10, 2, 10, 10, 7)
	assertSet("global", report.Global, 30, 24, 20, 4, 20, 20, 14)

	seen := map[string]bool{}
	for _, run := range report.Runs {
		if seen[run.ID] || run.Terminal != run.Expected || !run.WorkspaceCorrect || run.StaleWrite || run.OutOfSelectionWrite || run.IncorrectAppliedMutation || run.UnapprovedMutation || run.FailureWithEffect {
			t.Fatalf("unsafe or inconsistent run %s", run.ID)
		}
		seen[run.ID] = true
		if run.ProviderCalled && !run.Conforming || run.Positive && !run.PositiveCorrect || run.NecessaryAbstention && !run.AbstentionCorrect || run.Proposal && (!run.TargetPreserved || !run.Preview || !run.PreviewExact) || run.ApprovalExpected && !run.ApprovalReached || run.ExpectedApply && !run.Applied {
			t.Fatalf("failed qualification gate in %s", run.ID)
		}
	}

	var matrix struct {
		Status string
		Model  struct{ Name, Digest string }
		Cases  []struct {
			ID, Set, Class, Expected, Approval string
			Provider                           bool
		}
	}
	if err := yaml.Unmarshal(read("docs/milestone-35-qualification-matrix.yaml"), &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Status != "qualification_frozen_not_run" || matrix.Model.Name != model || matrix.Model.Digest != digest || len(matrix.Cases) != 30 {
		t.Fatal("qualification matrix changed")
	}
	counts := map[string]map[string]int{}
	for _, c := range matrix.Cases {
		if counts[c.Set] == nil {
			counts[c.Set] = map[string]int{}
		}
		counts[c.Set]["runs"]++
		if c.Provider {
			counts[c.Set]["provider"]++
		}
		if strings.HasPrefix(c.Class, "insufficient") {
			counts[c.Set]["abstain"]++
		}
		if c.Approval != "" {
			counts[c.Set]["approval"]++
		}
		if c.Expected == "applied" {
			counts[c.Set]["apply"]++
		}
	}
	for _, set := range []string{"development", "holdout"} {
		if counts[set]["runs"] != 15 || counts[set]["provider"] != 12 || counts[set]["abstain"] != 2 || counts[set]["approval"] != 10 || counts[set]["apply"] != 7 {
			t.Fatalf("invalid frozen denominators for %s", set)
		}
	}
}
