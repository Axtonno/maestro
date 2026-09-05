// m35select compares frozen mutation-specific profiles without applying files.
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
	"sort"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	p "github.com/antonio-cafeo/maestro/pkg/provider"
	"gopkg.in/yaml.v3"
)

type candidate struct{ Name, Digest string }
type testCase struct{ ID, Class, Request, Selected, Replacement, Decision string }
type matrix struct {
	Version    int
	Status     string
	Candidates []candidate
	Cases      []testCase
}
type observation struct {
	Model, Digest, Case, Class, Decision, OutputSHA256, ErrorClass string
	Conforming, Correct                                            bool
	LatencyMS                                                      int64
	InputTokens, OutputTokens                                      int
}
type totals struct {
	Runs, Conforming, Positive, CorrectPositive, Abstain, CorrectAbstain int
	LatencyMS                                                            int64
	Eligible                                                             bool
}
type report struct {
	Version                                                                                                       int
	ExecutedAt, ProviderVersion, MatrixSHA256, SchemaSHA256, PromptSHA256, Verdict, SelectedModel, SelectedDigest string
	Runs                                                                                                          []observation
	Totals                                                                                                        map[string]totals
}

func main() {
	matrixPath := flag.String("matrix", "docs/milestone-35-selection-cases.yaml", "frozen matrix")
	out := flag.String("output", "docs/reports/milestone-35-selection-runs.json", "exclusive report")
	preflight := flag.Bool("preflight", false, "check without generations")
	flag.Parse()
	data := read(*matrixPath)
	var m matrix
	must(yaml.Unmarshal(data, &m))
	if m.Status != "frozen_not_run" || len(m.Candidates) < 2 || len(m.Cases) < 1 {
		panic("selection not frozen")
	}
	schema := read("docs/schemas/host-bound-mutation-decision-v1.schema.json")
	systemPrompt := string(read("docs/prompts/mutation-host-bound-model-selection-v1.txt"))
	var version struct{ Version string }
	get("http://127.0.0.1:11434/api/version", &version)
	var tags struct {
		Models []struct{ Name, Digest string }
	}
	get("http://127.0.0.1:11434/api/tags", &tags)
	available := map[string]string{}
	for _, x := range tags.Models {
		available[x.Name] = x.Digest
	}
	for _, c := range m.Candidates {
		if available[c.Name] != c.Digest {
			panic("candidate identity mismatch: " + c.Name)
		}
	}
	if *preflight {
		fmt.Printf("PASS candidates=%d cases=%d matrix=%s schema=%s prompt=%s\n", len(m.Candidates), len(m.Cases), hash(data), hash(schema), hash([]byte(systemPrompt)))
		return
	}
	f, err := os.OpenFile(*out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	must(err)
	must(f.Close())
	r := report{Version: 1, ExecutedAt: time.Now().UTC().Format(time.RFC3339), ProviderVersion: version.Version, MatrixSHA256: hash(data), SchemaSHA256: hash(schema), PromptSHA256: hash([]byte(systemPrompt)), Verdict: "in_progress", Totals: map[string]totals{}}
	write(*out, r)
	for _, c := range m.Candidates {
		client, err := ollama.New("http://127.0.0.1:11434", c.Name, &http.Client{Timeout: 6 * time.Minute})
		must(err)
		for _, tc := range m.Cases {
			r.Runs = append(r.Runs, run(client, c, tc, schema, systemPrompt))
			write(*out, r)
		}
	}
	for _, c := range m.Candidates {
		r.Totals[c.Name] = aggregate(r.Runs, c.Name)
	}
	eligible := []candidate{}
	for _, c := range m.Candidates {
		if r.Totals[c.Name].Eligible {
			eligible = append(eligible, c)
		}
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		return r.Totals[eligible[i].Name].LatencyMS < r.Totals[eligible[j].Name].LatencyMS
	})
	r.Verdict = "mutation_specific_model_selection_rejected"
	if len(eligible) > 0 {
		r.Verdict = "mutation_specific_model_selected"
		r.SelectedModel = eligible[0].Name
		r.SelectedDigest = eligible[0].Digest
	}
	write(*out, r)
	fmt.Printf("verdict=%s selected=%s\n", r.Verdict, r.SelectedModel)
}
func run(client *ollama.Provider, c candidate, tc testCase, schema []byte, systemPrompt string) (o observation) {
	o = observation{Model: c.Name, Digest: c.Digest, Case: tc.ID, Class: tc.Class}
	payload, _ := json.Marshal(struct{ Request, SelectedText string }{tc.Request, tc.Selected})
	zero := 0.0
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	start := time.Now()
	resp, err := client.Complete(ctx, p.CompletionRequest{Model: c.Name, Messages: []p.Message{{Role: p.RoleSystem, Content: systemPrompt}, {Role: p.RoleUser, Content: string(payload)}}, Options: p.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: p.ThinkingDisabled}, KeepAlive: 5 * time.Minute, ToolChoice: p.ToolChoice{Mode: p.ToolChoiceNone}, Output: &p.StructuredOutput{Mode: p.StructuredOutputJSONSchema, Schema: schema}})
	o.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		o.ErrorClass = "provider_error"
		return
	}
	o.InputTokens = resp.Usage.InputTokens
	o.OutputTokens = resp.Usage.OutputTokens
	o.OutputSHA256 = hash([]byte(resp.Message.Content))
	d, err := mutation.DecodeHostBoundDecision([]byte(resp.Message.Content))
	if err != nil {
		o.ErrorClass = "decision_invalid"
		return
	}
	o.Conforming = true
	o.Decision = string(d.Decision)
	if tc.Decision == "abstain" {
		o.Correct = d.Decision == mutation.BinaryAbstain
	} else {
		o.Correct = d.Decision == mutation.BinaryPropose && d.NewText == tc.Replacement
	}
	return
}
func aggregate(rs []observation, model string) (t totals) {
	for _, r := range rs {
		if r.Model != model {
			continue
		}
		t.Runs++
		t.LatencyMS += r.LatencyMS
		if r.Conforming {
			t.Conforming++
		}
		if strings.HasPrefix(r.Class, "positive") {
			t.Positive++
			if r.Correct {
				t.CorrectPositive++
			}
		} else {
			t.Abstain++
			if r.Correct {
				t.CorrectAbstain++
			}
		}
	}
	t.Eligible = t.Runs > 0 && t.Conforming == t.Runs && t.Positive > 0 && t.CorrectPositive*100 >= t.Positive*90 && t.Abstain > 0 && t.CorrectAbstain == t.Abstain
	return
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
