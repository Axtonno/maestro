// m35qualify executes one frozen qualification of the selected model.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	it "github.com/antonio-cafeo/maestro/internal/tool"
	ce "github.com/antonio-cafeo/maestro/pkg/contextengine"
	p "github.com/antonio-cafeo/maestro/pkg/provider"
	tool "github.com/antonio-cafeo/maestro/pkg/tool"
	"gopkg.in/yaml.v3"
)

type model struct{ Name, Digest string }
type task struct {
	ID, Set, Class, Path, Request, Initial, Replacement, After, Concurrent, Expected, Approval, Raw string
	Paths                                                                                           []string
	Start, End                                                                                      int
	Provider                                                                                        bool
}
type matrix struct {
	Version int
	Status  string
	Model   model
	Cases   []task
}
type observation struct {
	ID, Set, Class, Terminal, Expected, Decision, OutputSHA256, FinalSHA256, ErrorClass                                                                                                                                                                                                                                            string
	ProviderCalled, Conforming, Positive, PositiveCorrect, NecessaryAbstention, AbstentionCorrect, Proposal, TargetPreserved, Preview, PreviewExact, ApprovalExpected, ApprovalReached, ExpectedApply, Applied, WorkspaceCorrect, StaleWrite, OutOfSelectionWrite, IncorrectAppliedMutation, UnapprovedMutation, FailureWithEffect bool
	LatencyMS                                                                                                                                                                                                                                                                                                                      int64
	InputTokens, OutputTokens                                                                                                                                                                                                                                                                                                      int
}
type totals struct {
	Runs, ProviderRuns, Conforming, Positive, CorrectPositive, NecessaryAbstentions, CorrectAbstentions, Proposals, PreservedTargets, Previews, ExactPreviews, ApprovalCases, ReachedApprovals, ExpectedApplies, CompletedApplies, CorrectTerminals int
	StaleWrites, OutOfSelectionWrites, IncorrectAppliedMutations, UnapprovedMutations, FailuresWithEffects                                                                                                                                          int
	Passed                                                                                                                                                                                                                                          bool
}
type report struct {
	Version                                                                                            int
	ExecutedAt, ProviderVersion, Model, ModelDigest, MatrixSHA256, SchemaSHA256, PromptSHA256, Verdict string
	QualificationPassed                                                                                bool
	Runs                                                                                               []observation
	Development, Holdout, Global                                                                       totals
}

func main() {
	matrixPath := flag.String("matrix", "docs/milestone-35-qualification-matrix.yaml", "frozen matrix")
	out := flag.String("output", "docs/reports/milestone-35-qualification-runs.json", "exclusive report")
	preflight := flag.Bool("preflight", false, "validate without generations")
	flag.Parse()
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		panic("requires Linux amd64")
	}
	data := read(*matrixPath)
	var m matrix
	must(yaml.Unmarshal(data, &m))
	if m.Status != "qualification_frozen_not_run" || m.Model.Name == "" || len(m.Cases) < 2 {
		panic("qualification not frozen")
	}
	sets := map[string]int{}
	ids := map[string]bool{}
	for _, tc := range m.Cases {
		if ids[tc.ID] || (tc.Set != "development" && tc.Set != "holdout") {
			panic("invalid qualification case")
		}
		ids[tc.ID] = true
		sets[tc.Set]++
	}
	if sets["development"] == 0 || sets["holdout"] == 0 {
		panic("empty set")
	}
	validateMatrix(m)
	prompt := string(read("docs/prompts/mutation-host-bound-model-selection-v1.txt"))
	schema := read("docs/schemas/host-bound-mutation-decision-v1.schema.json")
	var version struct{ Version string }
	get("http://127.0.0.1:11434/api/version", &version)
	var tags struct {
		Models []struct{ Name, Digest string }
	}
	get("http://127.0.0.1:11434/api/tags", &tags)
	found := false
	for _, x := range tags.Models {
		if x.Name == m.Model.Name && x.Digest == m.Model.Digest {
			found = true
		}
	}
	if !found {
		panic("selected model identity mismatch")
	}
	if *preflight {
		fmt.Printf("PASS model=%s cases=%d matrix=%s schema=%s prompt=%s\n", m.Model.Name, len(m.Cases), hash(data), hash(schema), hash([]byte(prompt)))
		return
	}
	info, err := os.Stdin.Stat()
	must(err)
	if info.Mode()&os.ModeCharDevice == 0 {
		panic("requires real TTY")
	}
	f, err := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	must(err)
	must(f.Close())
	r := report{Version: 1, ExecutedAt: time.Now().UTC().Format(time.RFC3339), ProviderVersion: version.Version, Model: m.Model.Name, ModelDigest: m.Model.Digest, MatrixSHA256: hash(data), SchemaSHA256: hash(schema), PromptSHA256: hash([]byte(prompt)), Verdict: "in_progress"}
	write(*out, r)
	client, err := ollama.New("http://127.0.0.1:11434", m.Model.Name, &http.Client{Timeout: 6 * time.Minute})
	must(err)
	approver := application.NewTerminalApprover(os.Stdin, os.Stderr, true)
	for _, tc := range m.Cases {
		r.Runs = append(r.Runs, execute(client, approver, m.Model.Name, prompt, schema, tc))
		write(*out, r)
		fmt.Printf("M35 result %s %s\n", tc.ID, r.Runs[len(r.Runs)-1].Terminal)
	}
	r.Development = aggregate(r.Runs, "development")
	r.Holdout = aggregate(r.Runs, "holdout")
	r.Global = aggregate(r.Runs, "")
	r.Verdict = "mutation_specific_model_qualification_rejected"
	if r.Development.Passed && r.Holdout.Passed && r.Global.Passed {
		r.Verdict = "mutation_specific_model_qualified"
		r.QualificationPassed = true
	}
	write(*out, r)
	fmt.Printf("verdict=%s qualification_passed=%t\n", r.Verdict, r.QualificationPassed)
}

func validateMatrix(m matrix) {
	type counts struct {
		runs, provider, positive, abstain, approvals, applies, host int
	}
	bySet := map[string]*counts{
		"development": {},
		"holdout":     {},
	}
	for _, tc := range m.Cases {
		c := bySet[tc.Set]
		c.runs++
		if !tc.Provider {
			c.host++
			continue
		}
		c.provider++
		if strings.HasPrefix(tc.Class, "insufficient") {
			c.abstain++
		} else {
			c.positive++
		}
		if tc.Approval != "" {
			c.approvals++
		}
		if tc.Expected == "applied" {
			c.applies++
		}
		path := tc.Path
		if path == "" {
			path = "app/Selected.php"
		}
		selection, err := mutation.Select(mutation.Snapshot{Path: path, Content: tc.Initial, Digest: hash([]byte(tc.Initial))}, tc.Start, tc.End)
		must(err)
		if strings.HasPrefix(tc.Class, "insufficient") {
			continue
		}
		after, err := selection.Replace(tc.Replacement)
		must(err)
		if after != tc.After {
			panic("invalid expected splice: " + tc.ID)
		}
	}
	for set, c := range bySet {
		if c.runs != 15 || c.provider != 12 || c.positive != 10 || c.abstain != 2 || c.approvals != 10 || c.applies != 7 || c.host != 3 {
			panic(fmt.Sprintf("invalid qualification denominators for %s: %+v", set, *c))
		}
	}
}

func execute(client *ollama.Provider, approver tool.Approver, modelName, prompt string, schema []byte, tc task) (o observation) {
	o = observation{ID: tc.ID, Set: tc.Set, Class: tc.Class, Expected: tc.Expected, Positive: strings.HasPrefix(tc.Class, "positive") || tc.Class == "deny" || tc.Class == "stale", NecessaryAbstention: strings.HasPrefix(tc.Class, "insufficient"), ApprovalExpected: tc.Approval != "", ExpectedApply: tc.Expected == "applied", TargetPreserved: true, WorkspaceCorrect: true}
	root, err := os.MkdirTemp("", "maestro-m35-")
	must(err)
	defer os.RemoveAll(root)
	must(os.MkdirAll(filepath.Join(root, "app"), 0700))
	if tc.Path == "" {
		tc.Path = "app/Selected.php"
	}
	file := filepath.Join(root, filepath.FromSlash(tc.Path))
	if strings.HasPrefix(tc.Path, "app/") && !strings.Contains(tc.Path, "/secrets/") {
		must(os.MkdirAll(filepath.Dir(file), 0700))
		must(os.WriteFile(file, []byte(tc.Initial), 0600))
	}
	sentinel := filepath.Join(root, "app", "Other.php")
	must(os.WriteFile(sentinel, []byte("<?php // untouched\n"), 0600))
	finalWant := tc.Initial
	defer func() {
		final, err := os.ReadFile(file)
		if os.IsNotExist(err) && tc.Initial == "" {
			final = []byte{}
			err = nil
		}
		must(err)
		other, err := os.ReadFile(sentinel)
		must(err)
		if o.Applied {
			finalWant = tc.After
		}
		o.WorkspaceCorrect = string(final) == finalWant && string(other) == "<?php // untouched\n"
		o.FinalSHA256 = hash(final)
		o.OutOfSelectionWrite = string(other) != "<?php // untouched\n" || (o.Applied && !o.TargetPreserved)
		o.IncorrectAppliedMutation = o.Applied && string(final) != tc.After
		o.StaleWrite = tc.Concurrent != "" && o.Applied
		o.FailureWithEffect = o.Applied && (o.Terminal != tc.Expected || o.IncorrectAppliedMutation || o.OutOfSelectionWrite)
	}()
	w, err := ce.NewWorkspace("m35-workspace", root, ce.WorkspaceOptions{Source: ce.SourceFilesystem, Policy: ce.DefaultScanPolicy()})
	must(err)
	registry := it.NewWorkspaceRegistry()
	must(registry.Bind("m35-run", w))
	host, err := it.NewHostBoundMutation(registry)
	must(err)
	paths := tc.Paths
	if len(paths) == 0 {
		paths = []string{tc.Path}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	bound, err := host.Capture(ctx, "m35-run", paths, tc.Start, tc.End)
	if err != nil {
		o.Terminal = terminal(err)
		return
	}
	s := bound.Target()
	raw := []byte(tc.Raw)
	if tc.Provider {
		o.ProviderCalled = true
		payload, _ := json.Marshal(struct {
			Request, SelectedText string
			StartLine, EndLine    int
		}{tc.Request, s.Text(), s.StartLine(), s.EndLine()})
		zero := 0.0
		start := time.Now()
		resp, err := client.Complete(ctx, p.CompletionRequest{Model: modelName, Messages: []p.Message{{Role: p.RoleSystem, Content: prompt}, {Role: p.RoleUser, Content: string(payload)}}, Options: p.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: p.ThinkingDisabled}, KeepAlive: 5 * time.Minute, ToolChoice: p.ToolChoice{Mode: p.ToolChoiceNone}, Output: &p.StructuredOutput{Mode: p.StructuredOutputJSONSchema, Schema: schema}})
		o.LatencyMS = time.Since(start).Milliseconds()
		if err != nil {
			o.Terminal = "response_invalid"
			o.ErrorClass = "provider_error"
			return
		}
		o.InputTokens = resp.Usage.InputTokens
		o.OutputTokens = resp.Usage.OutputTokens
		raw = []byte(resp.Message.Content)
	}
	o.OutputSHA256 = hash(raw)
	decision, err := mutation.DecodeHostBoundDecision(raw)
	if err != nil {
		o.Terminal = "response_invalid"
		o.ErrorClass = "decision_invalid"
		return
	}
	o.Conforming = true
	o.Decision = string(decision.Decision)
	if decision.Decision == mutation.BinaryAbstain {
		o.Terminal = "insufficient_information"
		o.AbstentionCorrect = o.NecessaryAbstention
		return
	}
	o.Proposal = true
	after, err := s.Replace(decision.NewText)
	if err != nil {
		o.Terminal = "response_invalid"
		return
	}
	o.PositiveCorrect = o.Positive && after == tc.After
	prepared, err := host.Prepare(ctx, bound, tool.CallID("call-"+tc.ID), raw)
	if err != nil {
		o.Terminal = terminal(err)
		return
	}
	preview, ok := prepared.Preview()
	o.Preview = ok
	var args struct{ Path, Old, ProposedContent, Fingerprint string }
	var fields map[string]json.RawMessage
	must(json.Unmarshal(prepared.Arguments(), &fields))
	must(json.Unmarshal(fields["path"], &args.Path))
	must(json.Unmarshal(fields["old"], &args.Old))
	must(json.Unmarshal(fields["proposed_content"], &args.ProposedContent))
	must(json.Unmarshal(fields["fingerprint"], &args.Fingerprint))
	o.TargetPreserved = args.Path == s.Path() && args.Old == s.Before() && args.ProposedContent == after
	o.PreviewExact = ok && preview.Body() != "" && args.Fingerprint == s.Fingerprint(decision.NewText, preview.Body()) && o.TargetPreserved
	if tc.Concurrent != "" {
		must(os.WriteFile(file, []byte(tc.Concurrent), 0600))
		finalWant = tc.Concurrent
	}
	permission, err := tool.NewToolPermissionRequest("m35.policy", prepared)
	must(err)
	choreography := tc.Approval
	if choreography == "" || !o.PositiveCorrect || !o.PreviewExact {
		choreography = "deny"
	}
	fmt.Fprintf(os.Stderr, "M35 approval task=%s expected=%s\n", tc.ID, choreography)
	approval, err := approver.Approve(ctx, permission)
	must(err)
	o.ApprovalReached = true
	if approval.Kind() != tool.ApprovalAllow {
		o.Terminal = "approval_rejected"
		return
	}
	if choreography != "allow" {
		panic("unexpected allow; execution blocked")
	}
	result, err := host.Execute(ctx, prepared)
	o.Applied = result.Effect() == tool.EffectApplied
	o.UnapprovedMutation = o.Applied && approval.Kind() != tool.ApprovalAllow
	if err != nil {
		o.Terminal = "execution_failed"
		return
	}
	if result.Outcome() == tool.ResultSuccess && o.Applied {
		o.Terminal = "applied"
	} else if result.Effect() == tool.EffectUnchanged {
		o.Terminal = "stale_source"
	} else {
		o.Terminal = "execution_failed"
	}
	return
}

func aggregate(rs []observation, set string) (t totals) {
	for _, r := range rs {
		if set != "" && r.Set != set {
			continue
		}
		t.Runs++
		if r.ProviderCalled {
			t.ProviderRuns++
			if r.Conforming {
				t.Conforming++
			}
		}
		if r.Positive {
			t.Positive++
			if r.PositiveCorrect {
				t.CorrectPositive++
			}
		}
		if r.NecessaryAbstention {
			t.NecessaryAbstentions++
			if r.AbstentionCorrect {
				t.CorrectAbstentions++
			}
		}
		if r.Proposal {
			t.Proposals++
			if r.TargetPreserved {
				t.PreservedTargets++
			}
			if r.Preview {
				t.Previews++
			}
			if r.PreviewExact {
				t.ExactPreviews++
			}
		}
		if r.ApprovalExpected {
			t.ApprovalCases++
			if r.ApprovalReached {
				t.ReachedApprovals++
			}
		}
		if r.ExpectedApply {
			t.ExpectedApplies++
			if r.Applied && r.Terminal == "applied" && r.WorkspaceCorrect {
				t.CompletedApplies++
			}
		}
		if r.Terminal == r.Expected {
			t.CorrectTerminals++
		}
		if r.StaleWrite {
			t.StaleWrites++
		}
		if r.OutOfSelectionWrite {
			t.OutOfSelectionWrites++
		}
		if r.IncorrectAppliedMutation {
			t.IncorrectAppliedMutations++
		}
		if r.UnapprovedMutation {
			t.UnapprovedMutations++
		}
		if r.FailureWithEffect {
			t.FailuresWithEffects++
		}
	}
	t.Passed = t.Runs > 0 && t.ProviderRuns > 0 && t.Conforming == t.ProviderRuns && t.Positive > 0 && t.CorrectPositive*100 >= t.Positive*90 && t.NecessaryAbstentions > 0 && t.CorrectAbstentions == t.NecessaryAbstentions && t.Proposals > 0 && t.PreservedTargets == t.Proposals && t.Previews == t.Proposals && t.ExactPreviews == t.Proposals && t.ApprovalCases > 0 && t.ReachedApprovals == t.ApprovalCases && t.ExpectedApplies > 0 && t.CompletedApplies == t.ExpectedApplies && t.CorrectTerminals == t.Runs && t.StaleWrites == 0 && t.OutOfSelectionWrites == 0 && t.IncorrectAppliedMutations == 0 && t.UnapprovedMutations == 0 && t.FailuresWithEffects == 0
	return
}
func terminal(err error) string {
	switch {
	case errors.Is(err, it.ErrRequestOutOfScope):
		return "request_out_of_scope"
	case errors.Is(err, mutation.ErrSelectionOutOfBounds):
		return "selection_out_of_bounds"
	case errors.Is(err, mutation.ErrInsufficientInformation):
		return "insufficient_information"
	}
	if s := mutation.TerminalForError(err); s != "" {
		return s
	}
	return "response_invalid"
}
func read(path string) []byte { b, e := os.ReadFile(path); must(e); return b }
func hash(b []byte) string    { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func must(e error) {
	if e != nil {
		panic(e)
	}
}
func get(url string, v any) {
	c := http.Client{Timeout: 10 * time.Second}
	r, e := c.Get(url)
	must(e)
	defer r.Body.Close()
	if r.StatusCode != 200 {
		panic("HTTP preflight")
	}
	must(json.NewDecoder(r.Body).Decode(v))
}
func write(path string, v any) {
	b, e := json.MarshalIndent(v, "", "  ")
	must(e)
	must(os.WriteFile(path, append(b, '\n'), 0600))
}
