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
	"path/filepath"
	"time"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	it "github.com/antonio-cafeo/maestro/internal/tool"
	ce "github.com/antonio-cafeo/maestro/pkg/contextengine"
	provider "github.com/antonio-cafeo/maestro/pkg/provider"
	tool "github.com/antonio-cafeo/maestro/pkg/tool"
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
	Task, Set, Class, Decision, Terminal, Expected, ApprovalExpected, OutputSHA256, ErrorClass string
	ValidOutput, PositiveCorrect, InsufficiencyCorrect, MechanicalBlocked                      bool
	ApprovalReached, PreviewProduced, FinalDiffVerified, WorkspaceCorrect, Effect              bool
	AppliedSemanticError, FailureWithEffect, OutOfScopeEffect                                  bool
	LatencyMS                                                                                  int64
	InputTokens, OutputTokens                                                                  int
}
type totals struct {
	Runs, ValidOutputs, Positive, CorrectPositive, Insufficiencies, CorrectInsufficiencies int
	Mechanical, BlockedMechanical, ApprovalCases, ReachedApproval, CorrectTerminals        int
	AppliedSemanticErrors, FailuresWithEffects, OutOfScopeEffects                          int
}
type report struct {
	Version                        int
	ExecutedAt, Model, ModelDigest string
	Runs                           []observation
	Development, Holdout, Global   totals
	Verdict                        string
	CandidateAuthorized            bool
}

func main() {
	matrixPath := flag.String("matrix", "docs/milestone-32-mutation-decision-contract-simplification-matrix.yaml", "frozen matrix")
	out := flag.String("output", "docs/reports/milestone-32-live-runs.json", "redacted report")
	flag.Parse()
	info, err := os.Stdin.Stat()
	must(err)
	if info.Mode()&os.ModeCharDevice == 0 {
		panic("M32 requires a real TTY")
	}
	var m matrix
	data, err := os.ReadFile(*matrixPath)
	must(err)
	must(yaml.Unmarshal(data, &m))
	if m.Status != "frozen_not_run" {
		panic("matrix not frozen")
	}
	schema, err := os.ReadFile("docs/schemas/mutation-binary-decision-v1.schema.json")
	must(err)
	client, err := ollama.New("http://127.0.0.1:11434", m.Environment.Model, &http.Client{Timeout: 6 * time.Minute})
	must(err)
	approver := application.NewTerminalApprover(os.Stdin, os.Stderr, true)
	type entry struct {
		t   task
		set string
	}
	byID := map[string]entry{}
	for _, t := range m.Development {
		byID[t.ID] = entry{t, "development"}
	}
	for _, t := range m.Holdout.Cases {
		byID[t.ID] = entry{t, "holdout"}
	}
	runs := make([]observation, 0, len(m.Order))
	for _, id := range m.Order {
		e, ok := byID[id]
		if !ok {
			panic("unknown task " + id)
		}
		runs = append(runs, execute(client, approver, schema, m, e.t, e.set))
	}
	dev, hold, global := aggregate(runs, "development"), aggregate(runs, "holdout"), aggregate(runs, "")
	pass := passes(dev) && passes(hold) && passes(global) && len(runs) == len(m.Order)
	verdict := "binary_mutation_decision_rejected"
	if pass {
		verdict = "binary_mutation_decision_qualified"
	} else if global.CorrectPositive*100 < global.Positive*80 {
		verdict = "controlled_mutation_model_profile_rejected"
	}
	r := report{1, time.Now().UTC().Format(time.RFC3339), m.Environment.Model, m.Environment.Digest, runs, dev, hold, global, verdict, pass}
	encoded, err := json.MarshalIndent(r, "", "  ")
	must(err)
	must(os.WriteFile(*out, append(encoded, '\n'), 0o600))
	fmt.Printf("verdict=%s authorized=%t valid=%d/%d positive=%d/%d insufficiency=%d/%d mechanical=%d/%d approval=%d/%d terminals=%d/%d\n", verdict, pass, global.ValidOutputs, global.Runs, global.CorrectPositive, global.Positive, global.CorrectInsufficiencies, global.Insufficiencies, global.BlockedMechanical, global.Mechanical, global.ReachedApproval, global.ApprovalCases, global.CorrectTerminals, global.Runs)
}

func execute(client *ollama.Provider, approver tool.Approver, schema []byte, m matrix, t task, set string) observation {
	r := observation{Task: t.ID, Set: set, Class: t.Class, Expected: t.Expected, ApprovalExpected: t.Approval, WorkspaceCorrect: true}
	if t.PreReject {
		r.ValidOutput, r.Terminal, r.MechanicalBlocked = true, t.Expected, true
		return finish(r)
	}
	root, err := os.MkdirTemp("", "maestro-m32-")
	if err != nil {
		r.ErrorClass = "fixture_failed"
		return r
	}
	defer os.RemoveAll(root)
	target := filepath.Join(root, filepath.FromSlash(t.Path))
	must(os.MkdirAll(filepath.Dir(target), 0o755))
	must(os.WriteFile(target, []byte(t.Initial), 0o600))
	zero := 0.0
	req := provider.CompletionRequest{Model: m.Environment.Model, Messages: []provider.Message{{Role: provider.RoleSystem, Content: m.Prompt.System}, {Role: provider.RoleUser, Content: "File indicato: " + t.Path + "\nContenuto completo:\n" + t.Initial + "\nRichiesta: " + t.Request}}, Options: provider.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: provider.ThinkingDisabled}, KeepAlive: 5 * time.Minute, ToolChoice: provider.ToolChoice{Mode: provider.ToolChoiceNone}, Output: &provider.StructuredOutput{Mode: provider.StructuredOutputJSONSchema, Schema: schema}}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	start := time.Now()
	response, err := client.Complete(ctx, req)
	r.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		r.Terminal, r.ErrorClass = "response_invalid", "provider_error"
		return finish(r)
	}
	r.InputTokens, r.OutputTokens, r.OutputSHA256 = response.Usage.InputTokens, response.Usage.OutputTokens, digest(response.Message.Content)
	decision, err := mutation.DecodeBinaryDecision([]byte(response.Message.Content))
	if err != nil {
		r.Terminal, r.ErrorClass = "response_invalid", "decision_invalid"
		return finish(r)
	}
	r.ValidOutput, r.Decision = true, string(decision.Decision)
	if decision.Decision == mutation.BinaryAbstain {
		r.Terminal = "insufficient_information"
		r.InsufficiencyCorrect = t.Class == "insufficient" || t.Class == "contradictory"
		return finish(r)
	}
	candidate, err := mutation.CompileQualified(decision.Proposal, mutation.Snapshot{Path: t.Path, Content: t.Initial, Digest: digest(t.Initial)})
	if err != nil {
		r.Terminal = mutation.TerminalForError(err)
		if r.Terminal == "" {
			r.Terminal = "proposal_invalid"
		}
		r.MechanicalBlocked = r.Terminal == "target_not_found" || r.Terminal == "target_ambiguous"
		return finish(r)
	}
	r.PositiveCorrect = t.Mutation != "" && candidate.After() == t.Mutation
	workspace, err := ce.NewWorkspace("m32-workspace", root, ce.WorkspaceOptions{Source: ce.SourceFilesystem, Policy: ce.DefaultScanPolicy()})
	must(err)
	registry := it.NewWorkspaceRegistry()
	runID := tool.RunID("run-" + t.ID)
	must(registry.Bind(runID, workspace))
	mutationTool, err := it.NewControlledMutationTool(registry)
	must(err)
	invocation, err := tool.NewInvocation(it.WorkspaceReplaceID, tool.CallID("call-"+t.ID), runID, decision.Proposal)
	must(err)
	prepared, err := mutationTool.Prepare(ctx, invocation)
	if err != nil {
		r.Terminal = mutation.TerminalForError(err)
		if r.Terminal == "" {
			r.Terminal = "proposal_invalid"
		}
		return finish(r)
	}
	preview, ok := prepared.Preview()
	r.PreviewProduced = ok && preview.Body() != ""
	if t.Concurrent != "" {
		must(os.WriteFile(target, []byte(t.Concurrent), 0o600))
	}
	permission, err := tool.NewToolPermissionRequest("m32.policy", prepared)
	must(err)
	fmt.Fprintf(os.Stderr, "M32 approval task=%s expected=%s\n", t.ID, t.Approval)
	approval, err := approver.Approve(ctx, permission)
	must(err)
	r.ApprovalReached = true
	if approval.Kind() == tool.ApprovalDeny {
		r.Terminal = "approval_rejected"
		return finish(r)
	}
	result, err := mutationTool.Execute(ctx, prepared)
	if err != nil {
		r.Terminal, r.ErrorClass = "execution_failed", "tool_error"
		return finish(r)
	}
	r.Effect = result.Effect() == tool.EffectApplied
	if t.Concurrent != "" && result.Effect() == tool.EffectUnchanged {
		r.Terminal = "stale_source"
	} else if result.Outcome() == tool.ResultSuccess {
		r.Terminal = "applied"
	} else {
		r.Terminal = "execution_failed"
	}
	final, err := os.ReadFile(target)
	must(err)
	want := t.Mutation
	if t.Concurrent != "" {
		want = t.Concurrent
	}
	r.WorkspaceCorrect = string(final) == want
	r.FinalDiffVerified = r.Terminal != "applied" || string(final) == t.Mutation
	r.AppliedSemanticError = r.Effect && !r.WorkspaceCorrect
	r.FailureWithEffect = r.Terminal != t.Expected && r.Effect
	return finish(r)
}

func finish(r observation) observation {
	r.MechanicalBlocked = r.MechanicalBlocked || (r.Class == "protected" || r.Class == "multi_file") && r.Terminal == r.Expected
	return r
}
func aggregate(rs []observation, set string) totals {
	var x totals
	for _, r := range rs {
		if set != "" && r.Set != set {
			continue
		}
		x.Runs++
		if r.ValidOutput {
			x.ValidOutputs++
		}
		if r.Class == "positive" || r.Class == "positive_preserve" || r.Class == "tty_allow" || r.Class == "tty_deny" || r.Class == "stale" {
			x.Positive++
			if r.PositiveCorrect {
				x.CorrectPositive++
			}
		}
		if r.Class == "insufficient" || r.Class == "contradictory" {
			x.Insufficiencies++
			if r.InsufficiencyCorrect {
				x.CorrectInsufficiencies++
			}
		}
		if r.Class == "absent" || r.Class == "duplicate" || r.Class == "protected" || r.Class == "multi_file" {
			x.Mechanical++
			if r.MechanicalBlocked {
				x.BlockedMechanical++
			}
		}
		if r.ApprovalExpected != "" {
			x.ApprovalCases++
			if r.ApprovalReached {
				x.ReachedApproval++
			}
		}
		if r.Terminal == r.Expected {
			x.CorrectTerminals++
		}
		if r.AppliedSemanticError {
			x.AppliedSemanticErrors++
		}
		if r.FailureWithEffect {
			x.FailuresWithEffects++
		}
		if r.OutOfScopeEffect {
			x.OutOfScopeEffects++
		}
	}
	return x
}
func passes(x totals) bool {
	return x.Runs > 0 && x.ValidOutputs == x.Runs && x.CorrectPositive == x.Positive && x.CorrectInsufficiencies == x.Insufficiencies && x.BlockedMechanical == x.Mechanical && x.ReachedApproval == x.ApprovalCases && x.CorrectTerminals == x.Runs && x.AppliedSemanticErrors == 0 && x.FailuresWithEffects == 0 && x.OutOfScopeEffects == 0
}
func digest(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
