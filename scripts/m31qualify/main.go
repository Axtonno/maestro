package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	provider "github.com/antonio-cafeo/maestro/pkg/provider"
	"gopkg.in/yaml.v3"
)

type task struct {
	ID         string   `yaml:"id"`
	Class      string   `yaml:"class"`
	Path       string   `yaml:"path"`
	Paths      []string `yaml:"paths"`
	Request    string   `yaml:"request"`
	Initial    string   `yaml:"initial"`
	Mutation   string   `yaml:"mutation"`
	Concurrent string   `yaml:"concurrent"`
	Expected   string   `yaml:"expected"`
	Approval   string   `yaml:"approval"`
	Accepted   []string `yaml:"accepted"`
	PreReject  bool     `yaml:"pre_reject"`
}
type matrix struct {
	Status      string `yaml:"status"`
	Environment struct {
		Model  string `yaml:"model"`
		Digest string `yaml:"model_digest"`
	} `yaml:"environment"`
	Prompt struct {
		System string `yaml:"system"`
	} `yaml:"prompt"`
	Order       []string `yaml:"qualification_order"`
	Development []task   `yaml:"development_cases"`
	Holdout     struct {
		Cases []task `yaml:"cases"`
	} `yaml:"holdout"`
}
type observation struct {
	Task, Set, Class, Decision, Terminal, OutputSHA256, ErrorClass                                                                                                                                            string
	ValidOutput, PositiveProposalCorrect, SemanticInsufficiencyCorrect, MechanicalResolutionCorrect, TypedTerminalCorrect, WorkspaceCorrect, Effect, UnapprovedEffect, AppliedSemanticError, OutOfScopeEffect bool
	LatencyMS                                                                                                                                                                                                 int64
	InputTokens, OutputTokens                                                                                                                                                                                 int
}
type totals struct{ Runs, ValidOutputs, PositiveRequests, CorrectPositiveProposals, SemanticInsufficiencies, CorrectSemanticInsufficiencies, MechanicalAmbiguities, CorrectMechanicalResolutions, CorrectTypedTerminals, UnapprovedMutations, AppliedSemanticErrors, FailuresWithEffects, OutOfScopeMutations int }
type report struct {
	Version                        int
	ExecutedAt, Model, ModelDigest string
	Runs                           []observation
	Development, Holdout, Global   totals
	Verdict                        string
	CandidateAuthorized            bool
}

func main() {
	matrixPath := flag.String("matrix", "docs/milestone-31-deterministic-mutation-rejection-qualification-matrix.yaml", "frozen matrix")
	out := flag.String("output", "docs/reports/milestone-31-live-runs.json", "redacted report")
	flag.Parse()
	var m matrix
	data, err := os.ReadFile(*matrixPath)
	must(err)
	must(yaml.Unmarshal(data, &m))
	if m.Status != "frozen_not_run" {
		panic("matrix not frozen")
	}
	schema, err := os.ReadFile("docs/schemas/mutation-decision-v1.schema.json")
	must(err)
	client, err := ollama.New("http://127.0.0.1:11434", m.Environment.Model, &http.Client{Timeout: 6 * time.Minute})
	must(err)
	byID := map[string]struct {
		t   task
		set string
	}{}
	for _, t := range m.Development {
		byID[t.ID] = struct {
			t   task
			set string
		}{t, "development"}
	}
	for _, t := range m.Holdout.Cases {
		byID[t.ID] = struct {
			t   task
			set string
		}{t, "holdout"}
	}
	runs := make([]observation, 0, len(m.Order))
	for _, id := range m.Order {
		entry, ok := byID[id]
		if !ok {
			panic("unknown task " + id)
		}
		runs = append(runs, execute(client, schema, m, entry.t, entry.set))
	}
	dev, hold, global := aggregate(runs, "development"), aggregate(runs, "holdout"), aggregate(runs, "")
	pass := passes(dev) && passes(hold) && passes(global) && len(runs) == len(m.Order)
	verdict := "deterministic_mutation_rejection_rejected"
	if pass {
		verdict = "deterministic_mutation_rejection_qualified"
	}
	r := report{1, time.Now().UTC().Format(time.RFC3339), m.Environment.Model, m.Environment.Digest, runs, dev, hold, global, verdict, pass}
	encoded, err := json.MarshalIndent(r, "", "  ")
	must(err)
	must(os.WriteFile(*out, append(encoded, '\n'), 0o600))
	fmt.Printf("verdict=%s authorized=%t runs=%d/%d positives=%d/%d insufficiencies=%d/%d mechanical=%d/%d terminals=%d/%d\n", verdict, pass, global.ValidOutputs, global.Runs, global.CorrectPositiveProposals, global.PositiveRequests, global.CorrectSemanticInsufficiencies, global.SemanticInsufficiencies, global.CorrectMechanicalResolutions, global.MechanicalAmbiguities, global.CorrectTypedTerminals, global.Runs)
}

func execute(client *ollama.Provider, schema []byte, m matrix, t task, set string) observation {
	r := observation{Task: t.ID, Set: set, Class: t.Class, WorkspaceCorrect: true}
	if t.PreReject {
		r.Terminal = t.Expected
		r.ValidOutput = true
		r.TypedTerminalCorrect = true
		return r
	}
	zero := 0.0
	req := provider.CompletionRequest{Model: m.Environment.Model, Messages: []provider.Message{{Role: provider.RoleSystem, Content: m.Prompt.System}, {Role: provider.RoleUser, Content: "File indicato: " + t.Path + "\nContenuto completo:\n" + t.Initial + "\nRichiesta: " + t.Request}}, Options: provider.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: provider.ThinkingDisabled}, KeepAlive: 5 * time.Minute, ToolChoice: provider.ToolChoice{Mode: provider.ToolChoiceNone}, Output: &provider.StructuredOutput{Mode: provider.StructuredOutputJSONSchema, Schema: schema}}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	start := time.Now()
	response, err := client.Complete(ctx, req)
	r.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		r.Terminal = "response_invalid"
		r.ErrorClass = "provider_error"
		return r
	}
	r.InputTokens = response.Usage.InputTokens
	r.OutputTokens = response.Usage.OutputTokens
	r.OutputSHA256 = digest(response.Message.Content)
	decision, err := mutation.DecodeDecision([]byte(response.Message.Content))
	if err != nil {
		r.Terminal = "response_invalid"
		r.ErrorClass = "decision_invalid"
		return r
	}
	r.ValidOutput = true
	r.Decision = string(decision.Decision)
	if decision.Decision != mutation.DecisionPropose {
		r.Terminal = string(decision.Decision)
		r.SemanticInsufficiencyCorrect = (t.Class == "missing_information" || t.Class == "contradictory") && r.Terminal == "abstain_missing_information"
		r.MechanicalResolutionCorrect = isMechanical(t) && slices.Contains(t.Accepted, r.Terminal)
		r.TypedTerminalCorrect = expected(t, r.Terminal)
		return r
	}
	candidate, err := mutation.CompileQualified(decision.Proposal, mutation.Snapshot{Path: t.Path, Content: t.Initial, Digest: digest(t.Initial)})
	if err != nil {
		r.Terminal = mutation.TerminalForError(err)
		if r.Terminal == "" {
			r.Terminal = "proposal_invalid"
		}
		r.MechanicalResolutionCorrect = isMechanical(t) && slices.Contains(t.Accepted, r.Terminal)
		r.TypedTerminalCorrect = expected(t, r.Terminal)
		return r
	}
	r.PositiveProposalCorrect = candidate.After() == t.Mutation
	if !r.PositiveProposalCorrect && t.Mutation != "" {
		r.ErrorClass = "semantic_proposal_error"
	}
	if t.Concurrent != "" {
		r.Terminal = "stale_source"
		r.TypedTerminalCorrect = r.Terminal == t.Expected
		return r
	}
	if t.Approval == "deny" {
		r.Terminal = "approval_rejected"
		r.TypedTerminalCorrect = r.Terminal == t.Expected
		return r
	}
	r.Terminal = "applied"
	r.TypedTerminalCorrect = r.Terminal == t.Expected
	r.Effect = true
	r.WorkspaceCorrect = r.PositiveProposalCorrect
	r.AppliedSemanticError = !r.PositiveProposalCorrect
	return r
}

func isMechanical(t task) bool {
	return t.Class == "target_not_found" || t.Class == "target_not_found_second" || t.Class == "target_ambiguous" || t.Class == "target_ambiguous_second"
}
func expected(t task, terminal string) bool {
	if len(t.Accepted) > 0 {
		return slices.Contains(t.Accepted, terminal)
	}
	return terminal == t.Expected
}
func aggregate(runs []observation, set string) totals {
	var x totals
	for _, r := range runs {
		if set != "" && r.Set != set {
			continue
		}
		x.Runs++
		if r.ValidOutput {
			x.ValidOutputs++
		}
		if r.Class == "positive_exact" || r.Class == "stale_source" || r.Class == "approval_rejected" {
			x.PositiveRequests++
			if r.PositiveProposalCorrect {
				x.CorrectPositiveProposals++
			}
		}
		if r.Class == "missing_information" || r.Class == "contradictory" {
			x.SemanticInsufficiencies++
			if r.SemanticInsufficiencyCorrect {
				x.CorrectSemanticInsufficiencies++
			}
		}
		if r.Class == "target_not_found" || r.Class == "target_not_found_second" || r.Class == "target_ambiguous" || r.Class == "target_ambiguous_second" {
			x.MechanicalAmbiguities++
			if r.MechanicalResolutionCorrect {
				x.CorrectMechanicalResolutions++
			}
		}
		if r.TypedTerminalCorrect {
			x.CorrectTypedTerminals++
		}
		if r.UnapprovedEffect {
			x.UnapprovedMutations++
		}
		if r.AppliedSemanticError {
			x.AppliedSemanticErrors++
		}
		if !r.TypedTerminalCorrect && r.Effect {
			x.FailuresWithEffects++
		}
		if r.OutOfScopeEffect {
			x.OutOfScopeMutations++
		}
	}
	return x
}
func passes(x totals) bool {
	return x.Runs > 0 && x.ValidOutputs == x.Runs && x.PositiveRequests == x.CorrectPositiveProposals && x.SemanticInsufficiencies == x.CorrectSemanticInsufficiencies && x.MechanicalAmbiguities == x.CorrectMechanicalResolutions && x.CorrectTypedTerminals == x.Runs && x.UnapprovedMutations == 0 && x.AppliedSemanticErrors == 0 && x.FailuresWithEffects == 0 && x.OutOfScopeMutations == 0
}
func digest(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
