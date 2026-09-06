package maestro_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone36ResidencyQualifiedWithSwapDisabledReference(t *testing.T) {
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
	var decision struct {
		Status               string
		AcceptedReport       string `yaml:"accepted_report"`
		AcceptedReportSHA256 string `yaml:"accepted_report_sha256"`
		Strategy             struct {
			Simultaneous bool   `yaml:"simultaneous_residency"`
			Handoff      string `yaml:"explicit_handoff"`
		}
		Blocking struct {
			Swap bool `yaml:"wsl_swap_growth"`
			Peak int  `yaml:"peak_observed_kib"`
		} `yaml:"blocking_observation"`
		Decision struct {
			Completed  bool `yaml:"milestone36_completed"`
			Authorized bool `yaml:"v0.5.0_authorized"`
			Tuning     bool `yaml:"model_or_protocol_tuning_allowed"`
		} `yaml:"decision"`
	}
	if err := yaml.Unmarshal(read("docs/milestone-36-residency-decision.yaml"), &decision); err != nil {
		t.Fatal(err)
	}
	if decision.Status != "residency_gate_qualified_swap_disabled_reference" || decision.AcceptedReport != "docs/reports/milestone-36-residency-runs-v4.json" || decision.AcceptedReportSHA256 != hash(read(decision.AcceptedReport)) || decision.Strategy.Simultaneous || decision.Strategy.Handoff != "unload_previous_wait_empty_then_generate" || decision.Blocking.Swap || decision.Blocking.Peak != 0 || decision.Decision.Completed || decision.Decision.Authorized || decision.Decision.Tuning {
		t.Fatalf("invalid residency decision: %#v", decision)
	}
	var report struct {
		Verdict, ProfileSHA256, PromptSHA256, SchemaSHA256 string
		Cycles                                             int
		Requests                                           []struct {
			Cycle, Step                          int
			Kind, Model, ObservedModel, Terminal string
			Correct, WithinLatencyLimit          bool
			ExplicitUnload                       bool
			UnloadedModel                        string
		}
		Gates struct {
			PatternsIdentical, ExpectedModelsLoaded, LatenciesWithinLimits bool
			ProviderProcessObserved                                        bool
			CrossProfileFallbacks, UnexpectedCPUOffloads, SwapGrowthEvents int
			OOMEvents, ProviderRestarts, FailedTransitions                 int
			Passed                                                         bool
		}
	}
	if err := json.Unmarshal(read(decision.AcceptedReport), &report); err != nil {
		t.Fatal(err)
	}
	if report.Verdict != "dual_model_residency_qualified" || report.Cycles != 3 || len(report.Requests) != 18 || report.ProfileSHA256 != hash(read("docs/milestone-36-residency-profile-v4.yaml")) || !report.Gates.PatternsIdentical || !report.Gates.ExpectedModelsLoaded || !report.Gates.LatenciesWithinLimits || !report.Gates.ProviderProcessObserved || report.Gates.CrossProfileFallbacks != 0 || report.Gates.UnexpectedCPUOffloads != 0 || report.Gates.SwapGrowthEvents != 0 || report.Gates.OOMEvents != 0 || report.Gates.ProviderRestarts != 0 || report.Gates.FailedTransitions != 0 || !report.Gates.Passed {
		t.Fatalf("invalid residency evidence: %#v", report.Gates)
	}
	for _, request := range report.Requests {
		if request.Terminal != "completed" || !request.Correct || !request.WithinLatencyLimit || request.Model != request.ObservedModel {
			t.Fatalf("failed transition: %#v", request)
		}
		wantUnload := request.Kind == "chat_to_mutation" || request.Kind == "mutation_to_chat"
		if request.ExplicitUnload != wantUnload || wantUnload && request.UnloadedModel == "" {
			t.Fatalf("invalid handoff evidence: %#v", request)
		}
	}
}
