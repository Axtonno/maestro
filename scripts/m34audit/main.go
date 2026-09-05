// m34audit reconstructs M33 requests against an in-memory HTTP transport.
// It never contacts Ollama and never generates model output.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/provider/ollama"
	p "github.com/antonio-cafeo/maestro/pkg/provider"
	"gopkg.in/yaml.v3"
)

type capture struct{ body []byte }

func (c *capture) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != "POST" || r.URL.Path != "/api/chat" {
		panic("unexpected request")
	}
	var err error
	c.body, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"model":"qwen3.5:9b","message":{"role":"assistant","content":"{\"decision\":\"abstain\"}"},"done":true,"done_reason":"stop"}`))}, nil
}
func main() {
	source := read("scripts/m33qualify/main.go")
	matches := regexp.MustCompile("(?s)const systemPrompt = `([^`]+)`").FindSubmatch(source)
	if len(matches) != 2 {
		panic("prompt not found")
	}
	prompt := string(matches[1])
	schema := read("docs/schemas/host-bound-mutation-decision-v1.schema.json")
	cases := read("docs/milestone-33-cases.yaml")
	var prior struct{ PromptSHA256, SchemaSHA256, MatrixSHA256, ModelDigest string }
	must(json.Unmarshal(read("docs/reports/milestone-33-live-runs.json"), &prior))
	if hash([]byte(prompt)) != prior.PromptSHA256 || hash(schema) != prior.SchemaSHA256 || hash(cases) != prior.MatrixSHA256 {
		panic("M33 freeze mismatch")
	}
	var matrix struct {
		Cases []struct {
			ID, Request, Initial, Replacement, After string
			Start, End                               int
		}
	}
	must(yaml.Unmarshal(cases, &matrix))
	expected := map[string]string{"M33-D03": "// pending", "M33-H01": "// dormant", "M33-H02": "// 日本語\n$retries = 6;"}
	type entry struct {
		ID                    string
		SelectedBytes         int
		SelectedSHA256        string
		ExpectedSpliceCorrect bool
		ReconstructedHTTP     json.RawMessage
	}
	entries := []entry{}
	for _, t := range matrix.Cases {
		want, ok := expected[t.ID]
		if !ok {
			continue
		}
		s, err := mutation.Select(mutation.Snapshot{Path: "app/Selected.php", Content: t.Initial, Digest: hash([]byte(t.Initial))}, t.Start, t.End)
		must(err)
		if s.Text() != want {
			panic("selected bytes mismatch")
		}
		after, err := s.Replace(t.Replacement)
		must(err)
		if after != t.After {
			panic("expected splice mismatch")
		}
		payload, err := json.Marshal(struct {
			Request, SelectedText string
			StartLine, EndLine    int
		}{t.Request, s.Text(), s.StartLine(), s.EndLine()})
		must(err)
		transport := &capture{}
		client, err := ollama.New("http://audit.invalid", "qwen3.5:9b", &http.Client{Transport: transport})
		must(err)
		zero := 0.0
		_, err = client.Complete(context.Background(), p.CompletionRequest{Model: "qwen3.5:9b", Messages: []p.Message{{Role: p.RoleSystem, Content: prompt}, {Role: p.RoleUser, Content: string(payload)}}, Options: p.GenerationOptions{MaxTokens: 1024, ContextWindow: 4096, Temperature: &zero, Thinking: p.ThinkingDisabled}, KeepAlive: 5 * time.Minute, ToolChoice: p.ToolChoice{Mode: p.ToolChoiceNone}, Output: &p.StructuredOutput{Mode: p.StructuredOutputJSONSchema, Schema: schema}})
		must(err)
		var wire struct {
			Messages []struct{ Role, Content string }
			Think    *bool
			Stream   bool
			Tools    json.RawMessage
			Options  struct {
				NumCtx      int `json:"num_ctx"`
				NumPredict  int `json:"num_predict"`
				Temperature *float64
			}
			Format json.RawMessage
		}
		must(json.Unmarshal(transport.body, &wire))
		if len(wire.Messages) != 2 || wire.Messages[0].Role != "system" || wire.Messages[0].Content != prompt || wire.Messages[1].Role != "user" || wire.Messages[1].Content != string(payload) || wire.Think == nil || *wire.Think || wire.Stream || len(wire.Tools) != 0 || wire.Options.NumCtx != 4096 || wire.Options.NumPredict != 1024 || wire.Options.Temperature == nil || *wire.Options.Temperature != 0 {
			panic("wire contract mismatch")
		}
		// Compare schema semantics, independently of serialization whitespace.
		var a, b any
		must(json.Unmarshal(schema, &a))
		must(json.Unmarshal(wire.Format, &b))
		ae, _ := json.Marshal(a)
		be, _ := json.Marshal(b)
		if string(ae) != string(be) {
			panic("schema changed")
		}
		entries = append(entries, entry{t.ID, len(s.Text()), hash([]byte(s.Text())), true, transport.body})
	}
	if len(entries) != 3 {
		panic("missing failure cases")
	}
	report := struct {
		Kind, ModelDigest, PromptSHA256, SchemaSHA256, MatrixSHA256 string
		ProviderGenerations                                         int
		Cases                                                       []entry
	}{"offline_reconstruction_not_historical_wire_capture", prior.ModelDigest, prior.PromptSHA256, prior.SchemaSHA256, prior.MatrixSHA256, 0, entries}
	encoded, err := json.MarshalIndent(report, "", "  ")
	must(err)
	must(os.WriteFile("docs/reports/milestone-34-offline-reconstruction.json", append(encoded, '\n'), 0600))
	fmt.Println("PASS: M33 freeze hashes; 3 exact selections and expected splices; 3 serialized requests; zero model generations")
}
func read(path string) []byte {
	b, e := os.ReadFile(path)
	must(e)
	return []byte(strings.ReplaceAll(string(b), "\r\n", "\n"))
}
func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func must(err error) {
	if err != nil {
		panic(err)
	}
}
