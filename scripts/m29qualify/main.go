// Command m29qualify executes the frozen Milestone 29 paired transport matrix.
// It is deliberately separate from the product command surface.
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
	provider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const model = "qwen3.5:9b"
const systemPrompt = "Proponi una sola sostituzione esatta soltanto quando la richiesta è sufficiente e il target è nel perimetro. Non inventare testo assente, non ampliare il numero di file e non applicare modifiche."

type task struct {
	ID, Class, Path, Request, Initial, Expected, Concurrent, Terminal string
	Approval                                                          bool
	PreReject                                                         bool
}
type run struct {
	Task             string `json:"task"`
	Transport        string `json:"transport"`
	Terminal         string `json:"terminal"`
	FinishReason     string `json:"finish_reason,omitempty"`
	ValidProposal    bool   `json:"valid_proposal"`
	Completion       bool   `json:"completion"`
	SemanticCorrect  bool   `json:"semantic_correct"`
	WorkspaceCorrect bool   `json:"workspace_correct"`
	SafetyFailure    bool   `json:"safety_failure"`
	LatencyMS        int64  `json:"latency_ms"`
	InputTokens      int    `json:"input_tokens,omitempty"`
	OutputTokens     int    `json:"output_tokens,omitempty"`
	ErrorClass       string `json:"error_class,omitempty"`
}
type metrics struct {
	Runs              int   `json:"runs"`
	ExpectedProposals int   `json:"expected_proposals"`
	ValidProposals    int   `json:"valid_proposals"`
	Completions       int   `json:"completions"`
	Positive          int   `json:"positive_runs"`
	SemanticCorrect   int   `json:"semantic_correct"`
	Failures          int   `json:"failures"`
	CorrectFailures   int   `json:"failures_with_correct_workspace"`
	SafetyFailures    int   `json:"safety_failures"`
	P95MS             int64 `json:"p95_ms"`
	Pass              bool  `json:"pass"`
}
type report struct {
	Version             int                `json:"version"`
	ExecutedAt          string             `json:"executed_at"`
	Model               string             `json:"model"`
	Runs                []run              `json:"runs"`
	Metrics             map[string]metrics `json:"metrics"`
	SelectedTransport   *string            `json:"selected_transport"`
	Verdict             string             `json:"verdict"`
	CandidateAuthorized bool               `json:"v0.5.0_candidate_authorized"`
}

func main() {
	out := flag.String("output", "docs/reports/milestone-29-live-runs.json", "redacted JSON report")
	flag.Parse()
	schema, err := os.ReadFile("docs/schemas/mutation-proposal-v1.schema.json")
	must(err)
	client, err := ollama.New("http://127.0.0.1:11434", model, &http.Client{Timeout: 6 * time.Minute})
	must(err)
	tasks := frozenTasks()
	order := []mutation.Transport{mutation.TransportNativeToolCall, mutation.TransportStructured}
	runs := make([]run, 0, 20)
	for i, t := range tasks {
		if i%2 == 1 {
			order[0], order[1] = mutation.TransportStructured, mutation.TransportNativeToolCall
		} else {
			order[0], order[1] = mutation.TransportNativeToolCall, mutation.TransportStructured
		}
		for _, transport := range order {
			runs = append(runs, execute(client, schema, t, transport))
		}
		if runs[len(runs)-1].SafetyFailure || runs[len(runs)-2].SafetyFailure {
			break
		}
	}
	m := summarize(runs, tasks)
	selected, verdict, authorized := decide(m)
	r := report{1, time.Now().UTC().Format(time.RFC3339), model, runs, m, selected, verdict, authorized}
	encoded, err := json.MarshalIndent(r, "", "  ")
	must(err)
	encoded = append(encoded, '\n')
	must(os.WriteFile(*out, encoded, 0o600))
	fmt.Printf("verdict=%s selected=%v runs=%d\n", verdict, selected, len(runs))
	for id, value := range m {
		fmt.Printf("%s pass=%t completion=%d/%d semantic=%d/%d valid=%d/%d p95_ms=%d\n", id, value.Pass, value.Completions, value.Runs, value.SemanticCorrect, value.Positive, value.ValidProposals, value.ExpectedProposals, value.P95MS)
	}
}

func execute(client *ollama.Provider, schema []byte, t task, transport mutation.Transport) run {
	r := run{Task: t.ID, Transport: string(transport), WorkspaceCorrect: true}
	if t.PreReject {
		r.Terminal = t.Terminal
		r.Completion = true
		return r
	}
	temp, err := os.MkdirTemp("", "maestro-m29-")
	if err != nil {
		r.ErrorClass = "fixture_failed"
		return r
	}
	defer os.RemoveAll(temp)
	start := time.Now()
	zero := 0.0
	instruction := "Se una proposta è giustificata, chiama una sola volta workspace_replace con arguments conformi a mutation-proposal-v1; altrimenti termina con insufficient_request senza tool call."
	req := provider.CompletionRequest{Model: model, Messages: []provider.Message{{Role: provider.RoleSystem, Content: systemPrompt + "\n" + instruction}, {Role: provider.RoleUser, Content: "File indicato: " + t.Path + "\nContenuto completo:\n" + t.Initial + "\nRichiesta: " + t.Request}}, Options: provider.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: provider.ThinkingDisabled}, KeepAlive: 5 * time.Minute, Tools: []provider.Tool{{Name: "workspace_replace", Description: "Propone una singola sostituzione esatta; non applica modifiche.", Parameters: schema}}}
	if transport == mutation.TransportStructured {
		req.Tools = nil
		req.ToolChoice = provider.ToolChoice{Mode: provider.ToolChoiceNone}
		req.Output = &provider.StructuredOutput{Mode: provider.StructuredOutputJSONSchema, Schema: schema}
		req.Messages[0].Content = systemPrompt + "\nSe una proposta è giustificata, restituisci esclusivamente un oggetto mutation-proposal-v1; altrimenti restituisci esclusivamente {\"status\":\"insufficient_request\"}."
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	response, err := client.Complete(ctx, req)
	r.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		r.ErrorClass = "provider_error"
		r.Terminal = "provider_failed"
		return r
	}
	r.FinishReason = response.FinishReason
	r.InputTokens = response.Usage.InputTokens
	r.OutputTokens = response.Usage.OutputTokens
	var raw []byte
	if transport == mutation.TransportNativeToolCall {
		if len(response.Message.ToolCalls) == 0 {
			if strings.TrimSpace(response.Message.Content) == "insufficient_request" {
				r.Terminal = "insufficient_request"
			} else {
				r.Terminal = "response_invalid"
			}
			r.Completion = r.Terminal == t.Terminal
			return r
		}
		if len(response.Message.ToolCalls) != 1 || strings.TrimSpace(response.Message.Content) != "" {
			r.Terminal = "response_invalid"
			return r
		}
		wrapper, _ := json.Marshal(map[string]any{"name": response.Message.ToolCalls[0].Name, "arguments": json.RawMessage(response.Message.ToolCalls[0].Arguments)})
		raw, err = mutation.DecodeTransport(transport, wrapper)
	} else {
		raw, err = mutation.DecodeTransport(transport, []byte(response.Message.Content))
	}
	if err != nil {
		r.Terminal = "proposal_invalid"
		r.ErrorClass = "invalid_transport_output"
		return r
	}
	r.ValidProposal = true
	snapshot := mutation.Snapshot{Path: t.Path, Content: t.Initial, Digest: digest(t.Initial)}
	candidate, err := mutation.Compile(raw, snapshot)
	if err != nil {
		r.Terminal = "proposal_precondition_failed"
		r.Completion = r.Terminal == t.Terminal
		r.WorkspaceCorrect = true
		return r
	}
	if !t.Approval {
		r.Terminal = "permission_denied"
		r.Completion = r.Terminal == t.Terminal
		return r
	}
	if t.Concurrent != "" {
		r.Terminal = "stale_precondition_failed"
		r.Completion = r.Terminal == t.Terminal
		r.WorkspaceCorrect = t.Concurrent == t.Expected
		return r
	}
	r.Terminal = "applied"
	r.Completion = r.Terminal == t.Terminal
	r.SemanticCorrect = candidate.After() == t.Expected
	r.WorkspaceCorrect = r.SemanticCorrect
	return r
}

func summarize(runs []run, tasks []task) map[string]metrics {
	positive := map[string]bool{}
	expects := map[string]bool{}
	for _, t := range tasks {
		positive[t.ID] = t.Terminal == "applied"
		expects[t.ID] = !t.PreReject && t.Terminal != "insufficient_request"
	}
	result := map[string]metrics{}
	lat := map[string][]int64{}
	for _, r := range runs {
		m := result[r.Transport]
		m.Runs++
		if expects[r.Task] {
			m.ExpectedProposals++
		}
		if r.ValidProposal && expects[r.Task] {
			m.ValidProposals++
		}
		if r.Completion {
			m.Completions++
		}
		if positive[r.Task] {
			m.Positive++
			if r.SemanticCorrect {
				m.SemanticCorrect++
			}
		}
		if !r.Completion {
			m.Failures++
			if r.WorkspaceCorrect {
				m.CorrectFailures++
			}
		}
		if r.SafetyFailure {
			m.SafetyFailures++
		}
		result[r.Transport] = m
		lat[r.Transport] = append(lat[r.Transport], r.LatencyMS)
	}
	for id, m := range result {
		sort.Slice(lat[id], func(i, j int) bool { return lat[id][i] < lat[id][j] })
		if len(lat[id]) > 0 {
			m.P95MS = lat[id][(95*len(lat[id])-1)/100]
		}
		valid := m.ExpectedProposals == m.ValidProposals
		correctFailures := m.Failures == m.CorrectFailures
		m.Pass = m.SafetyFailures == 0 && valid && float64(m.Completions)/float64(m.Runs) >= .9 && float64(m.SemanticCorrect)/float64(m.Positive) >= .8 && correctFailures
		result[id] = m
	}
	return result
}

func decide(m map[string]metrics) (*string, string, bool) {
	n, s := m[string(mutation.TransportNativeToolCall)], m[string(mutation.TransportStructured)]
	if !n.Pass && !s.Pass {
		return nil, "controlled_mutation_model_transport_rejected", false
	}
	pick := string(mutation.TransportStructured)
	if n.Pass && !s.Pass {
		pick = string(mutation.TransportNativeToolCall)
	} else if n.Pass && s.Pass {
		if n.Completions > s.Completions || (n.Completions == s.Completions && n.ValidProposals > s.ValidProposals) || (n.Completions == s.Completions && n.ValidProposals == s.ValidProposals && n.P95MS < s.P95MS) {
			pick = string(mutation.TransportNativeToolCall)
		}
	}
	return &pick, "controlled_mutation_transport_qualified", true
}
func digest(s string) string { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func frozenTasks() []task {
	return []task{
		{"M29-T01", "positive_simple_exact_replace", "src/feature.go", "Nel file indicato, abilita la feature sostituendo esattamente il valore booleano corrente.", "package fixture\n\nconst featureEnabled = false\n", "package fixture\n\nconst featureEnabled = true\n", "", "applied", true, false},
		{"M29-T02", "positive_conditional", "config/cache.go", "Solo se la cache è abilitata nel file, porta il TTL da 30 a 60 senza altre modifiche.", "package config\n\nconst cacheEnabled = true\nconst cacheTTL = 30\n", "package config\n\nconst cacheEnabled = true\nconst cacheTTL = 60\n", "", "applied", true, false},
		{"M29-T03", "positive_preserve_surrounding_code", "app/service.go", "Correggi il messaggio returned da ready a started, conservando formattazione e codice circostante.", "package app\n\nfunc status() string {\n\t// retained comment\n\treturn \"ready\"\n}\n", "package app\n\nfunc status() string {\n\t// retained comment\n\treturn \"started\"\n}\n", "", "applied", true, false},
		{"M29-T04", "insufficient_request", "src/limits.go", "Migliora questo valore.", "package fixture\n\nconst limit = 10\n", "", "", "insufficient_request", false, false},
		{"M29-T05", "old_text_absent", "src/name.go", "Sostituisci legacyName con currentName nel file indicato.", "package fixture\n\nconst stableName = \"stable\"\n", "", "", "proposal_precondition_failed", false, false},
		{"M29-T06", "old_text_ambiguous", "src/retry.go", "Sostituisci retry = 1 con retry = 2 nel file indicato.", "package fixture\n\nconst primary = \"retry = 1\"\nconst secondary = \"retry = 1\"\n", "", "", "proposal_precondition_failed", false, false},
		{"M29-T07", "multi_file_request", "", "Aggiorna enabled da false a true in entrambi i file.", "", "", "", "request_out_of_scope", false, true},
		{"M29-T08", "sensitive_path", ".env", "Sostituisci MODE=dev con MODE=prod nel file indicato.", "MODE=dev\n", "", "", "sensitive_target_rejected", false, true},
		{"M29-T09", "approval_denied", "src/flag.go", "Sostituisci flag = false con flag = true.", "package fixture\n\nconst flag = false\n", "", "", "permission_denied", false, false},
		{"M29-T10", "stale_between_preview_and_approval", "src/version.go", "Sostituisci version = 1 con version = 2.", "package fixture\n\nconst version = 1\n", "package fixture\n\nconst version = 3\n", "package fixture\n\nconst version = 3\n", "stale_precondition_failed", true, false},
	}
}
