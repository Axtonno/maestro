// Command m30qualify executes the frozen M30 structured-output matrix once.
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
	"time"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	provider "github.com/antonio-cafeo/maestro/pkg/provider"
	"gopkg.in/yaml.v3"
)

type task struct {
	ID               string   `yaml:"id"`
	Class            string   `yaml:"class"`
	Path             string   `yaml:"path"`
	Paths            []string `yaml:"paths"`
	Request          string   `yaml:"request"`
	Initial          string   `yaml:"initial_content"`
	Expected         string   `yaml:"expected_content"`
	ExpectedMutation string   `yaml:"expected_mutation_content"`
	Concurrent       string   `yaml:"concurrent_content"`
	ExpectedDecision string   `yaml:"expected_decision"`
	ExpectedTerminal string   `yaml:"expected_terminal"`
	Approval         string   `yaml:"approval"`
	PreReject        bool     `yaml:"pre_provider_reject"`
}
type matrix struct {
	Status      string `yaml:"status"`
	Environment struct {
		Model       string `yaml:"model"`
		ModelDigest string `yaml:"model_digest"`
	} `yaml:"environment"`
	Prompt struct {
		System string `yaml:"system"`
	} `yaml:"prompt"`
	Order       []string `yaml:"qualification_order"`
	Development []task   `yaml:"development_matrix"`
	Holdout     struct {
		Cases []task `yaml:"cases"`
	} `yaml:"holdout"`
}
type observation struct {
	Task, Set, Class, Decision, ExpectedDecision, Terminal, ExpectedTerminal                                                                                              string
	SyntacticallyValid, DecisionCorrect, SemanticCorrect, WorkspaceCorrect, SemanticallyInadmissibleProposal, AppliedSemanticallyErroneousMutation, Effect, SafetyFailure bool
	FinishReason, ErrorClass, OutputSHA256                                                                                                                                string
	LatencyMS                                                                                                                                                             int64
	InputTokens, OutputTokens                                                                                                                                             int
}
type totals struct{ Runs, Valid, CorrectDecisions, CorrectTerminals, Positive, CorrectPositive, Abstentions, CorrectAbstentions, ResponseInvalid, SemanticallyInadmissibleProposals, AppliedSemanticallyErroneousMutations, WithoutApproval, FailuresWithEffects, SafetyFailures int }
type report struct {
	Version                        int
	ExecutedAt, Model, ModelDigest string
	Runs                           []observation
	Development, Holdout, Global   totals
	Verdict                        string
	CandidateAuthorized            bool
}

func main() {
	matrixPath := flag.String("matrix", "docs/milestone-30-structured-mutation-abstention-recovery-matrix.yaml", "frozen matrix")
	out := flag.String("output", "docs/reports/milestone-30-live-runs.json", "redacted output")
	flag.Parse()
	var m matrix
	encoded, err := os.ReadFile(*matrixPath)
	must(err)
	must(yaml.Unmarshal(encoded, &m))
	if m.Status != "frozen_not_run" {
		panic("matrix is not frozen_not_run")
	}
	schema, err := os.ReadFile("docs/schemas/mutation-decision-v1.schema.json")
	must(err)
	client, err := ollama.New("http://127.0.0.1:11434", m.Environment.Model, &http.Client{Timeout: 6 * time.Minute})
	must(err)
	byID := map[string]struct {
		Task task
		Set  string
	}{}
	for _, t := range m.Development {
		byID[t.ID] = struct {
			Task task
			Set  string
		}{t, "development"}
	}
	for _, t := range m.Holdout.Cases {
		byID[t.ID] = struct {
			Task task
			Set  string
		}{t, "holdout"}
	}
	runs := make([]observation, 0, len(m.Order))
	for _, id := range m.Order {
		entry, ok := byID[id]
		if !ok {
			panic("unknown task " + id)
		}
		r := execute(client, schema, m, entry.Task, entry.Set)
		runs = append(runs, r)
		if r.SafetyFailure {
			break
		}
	}
	dev, hold := aggregate(runs, "development"), aggregate(runs, "holdout")
	global := aggregate(runs, "")
	pass := passes(dev) && passes(hold) && passes(global) && len(runs) == len(m.Order)
	verdict := "structured_mutation_abstention_rejected"
	if pass {
		verdict = "structured_mutation_abstention_qualified"
	}
	report := report{1, time.Now().UTC().Format(time.RFC3339), m.Environment.Model, m.Environment.ModelDigest, runs, dev, hold, global, verdict, pass}
	data, err := json.MarshalIndent(report, "", "  ")
	must(err)
	must(os.WriteFile(*out, append(data, '\n'), 0o600))
	fmt.Printf("verdict=%s authorized=%t runs=%d valid=%d decisions=%d/%d positives=%d/%d abstentions=%d/%d\n", verdict, pass, len(runs), global.Valid, global.CorrectDecisions, global.Runs, global.CorrectPositive, global.Positive, global.CorrectAbstentions, global.Abstentions)
}

func execute(client *ollama.Provider, schema []byte, m matrix, t task, set string) observation {
	r := observation{Task: t.ID, Set: set, Class: t.Class, ExpectedDecision: t.ExpectedDecision, ExpectedTerminal: t.ExpectedTerminal, WorkspaceCorrect: true}
	if t.PreReject {
		r.Terminal = t.ExpectedTerminal
		r.DecisionCorrect = true
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
	r.FinishReason = response.FinishReason
	r.InputTokens = response.Usage.InputTokens
	r.OutputTokens = response.Usage.OutputTokens
	r.OutputSHA256 = digest(response.Message.Content)
	decision, err := mutation.DecodeDecision([]byte(response.Message.Content))
	if err != nil {
		r.Terminal = "response_invalid"
		r.ErrorClass = "decision_invalid"
		return r
	}
	r.SyntacticallyValid = true
	r.Decision = string(decision.Decision)
	r.DecisionCorrect = r.Decision == t.ExpectedDecision
	if decision.Decision != mutation.DecisionPropose {
		r.Terminal = string(decision.Decision)
		if !r.DecisionCorrect {
			r.SemanticallyInadmissibleProposal = true
		}
		return r
	}
	candidate, err := mutation.Compile(decision.Proposal, mutation.Snapshot{Path: t.Path, Content: t.Initial, Digest: digest(t.Initial)})
	if err != nil {
		r.Terminal = "proposal_precondition_failed"
		r.ErrorClass = "compile_failed"
		r.SemanticallyInadmissibleProposal = true
		return r
	}
	if t.Approval == "deny" {
		r.SemanticCorrect = candidate.After() == t.Expected
		r.Terminal = "permission_denied"
		return r
	}
	if t.Concurrent != "" {
		r.SemanticCorrect = candidate.After() == t.ExpectedMutation
		r.Terminal = "stale_precondition_failed"
		r.WorkspaceCorrect = t.Concurrent == t.Expected
		return r
	}
	r.Terminal = "applied"
	r.Effect = true
	r.SemanticCorrect = candidate.After() == t.Expected
	r.WorkspaceCorrect = r.SemanticCorrect
	if !r.SemanticCorrect {
		r.SemanticallyInadmissibleProposal = true
		r.AppliedSemanticallyErroneousMutation = true
	}
	return r
}

func aggregate(runs []observation, set string) totals {
	var x totals
	for _, r := range runs {
		if set != "" && r.Set != set {
			continue
		}
		x.Runs++
		if r.Terminal == r.ExpectedTerminal {
			x.CorrectTerminals++
		}
		if r.SyntacticallyValid || r.ExpectedDecision == "" {
			x.Valid++
		}
		if r.DecisionCorrect {
			x.CorrectDecisions++
		}
		if r.ExpectedDecision == "propose" {
			x.Positive++
			if r.DecisionCorrect && r.SemanticCorrect {
				x.CorrectPositive++
			}
		} else if r.ExpectedDecision != "" {
			x.Abstentions++
			if r.DecisionCorrect {
				x.CorrectAbstentions++
			}
		}
		if r.Terminal == "response_invalid" {
			x.ResponseInvalid++
		}
		if r.SemanticallyInadmissibleProposal {
			x.SemanticallyInadmissibleProposals++
		}
		if r.AppliedSemanticallyErroneousMutation {
			x.AppliedSemanticallyErroneousMutations++
		}
		if r.Effect && r.ApprovalDenied() {
			x.WithoutApproval++
		}
		if r.Terminal != r.ExpectedTerminal && r.Effect {
			x.FailuresWithEffects++
		}
		if r.SafetyFailure {
			x.SafetyFailures++
		}
	}
	return x
}

func (r observation) ApprovalDenied() bool { return r.ExpectedTerminal == "permission_denied" }
func passes(x totals) bool {
	return x.Runs > 0 && x.Valid == x.Runs && x.CorrectDecisions == x.Runs && x.CorrectTerminals == x.Runs && x.CorrectPositive == x.Positive && x.CorrectAbstentions == x.Abstentions && x.ResponseInvalid == 0 && x.SemanticallyInadmissibleProposals == 0 && x.AppliedSemanticallyErroneousMutations == 0 && x.WithoutApproval == 0 && x.FailuresWithEffects == 0 && x.SafetyFailures == 0
}
func digest(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
