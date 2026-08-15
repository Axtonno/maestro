package productconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadStrictVersionedConfigurationAndResolveRelativeWorkspace(t *testing.T) {
	directory := t.TempDir()
	workspace := filepath.Join(directory, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "config.yaml")
	writeConfig(t, path, validYAML("workspace"))

	config, err := Load(path)
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	if config.Version != Version || config.Path() != path || config.Workspace.Root != workspace {
		t.Fatalf("unexpected configuration: %#v", config)
	}
	if config.Provider.Timeout.Duration != 2*time.Minute || config.AgentLimits().MaxModelTurns != 5 || config.ContextBudget().EvidenceTokens() != 832 {
		t.Fatalf("unexpected derived configuration: %#v %#v", config.AgentLimits(), config.ContextBudget())
	}
	ids := config.ToolIDs()
	if string(ids[0]) != "workspace.patch" || string(ids[1]) != "workspace.read" {
		t.Fatalf("tool IDs were not normalized: %v", ids)
	}
}

func TestPublishedExampleMatchesCurrentSchema(t *testing.T) {
	config, err := Load("../../configs/maestro.example.yaml")
	if err != nil {
		t.Fatalf("published example is invalid: %v", err)
	}
	if config.Provider.ID != "ollama" || config.Models.Chat != "llama3.1:8b" || config.Workspace.ID != "laravel" {
		t.Fatalf("unexpected published example: %#v", config)
	}
	ids := config.ToolIDs()
	if len(ids) != 3 || string(ids[0]) != "workspace.list" || string(ids[1]) != "workspace.read" || string(ids[2]) != "workspace.search" {
		t.Fatalf("published example must expose only the supported read-only tools: %v", ids)
	}
	if config.Policy.WorkspaceMutation != "deny" {
		t.Fatalf("published example must deny workspace mutation: %q", config.Policy.WorkspaceMutation)
	}
}

func TestLoadRejectsUnknownDuplicateMultipleAndAliasYAML(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(string) string
	}{
		{name: "unknown", mutate: func(value string) string { return strings.Replace(value, "version: 1", "version: 1\nunknown: true", 1) }},
		{name: "duplicate", mutate: func(value string) string { return strings.Replace(value, "version: 1", "version: 1\nversion: 1", 1) }},
		{name: "multiple", mutate: func(value string) string { return value + "\n---\nversion: 1\n" }},
		{name: "alias", mutate: func(value string) string {
			return strings.Replace(value, "chat: fixture-model", "chat: &model fixture-model\n  embedding: *model", 1)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			writeConfig(t, path, testCase.mutate(validYAML("/tmp/workspace")))
			if _, err := Load(path); !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected ErrInvalid, got %v", err)
			}
		})
	}
}

func TestValidationRejectsUnsafeOrImplicitTargets(t *testing.T) {
	config := validConfig(t.TempDir())
	for _, testCase := range []struct {
		name   string
		mutate func(*Config)
		field  string
	}{
		{name: "version", mutate: func(c *Config) { c.Version = 2 }, field: "version"},
		{name: "provider", mutate: func(c *Config) { c.Provider.ID = "auto" }, field: "provider.id"},
		{name: "URL path", mutate: func(c *Config) { c.Provider.BaseURL = "http://localhost:11434/api" }, field: "provider.base_url"},
		{name: "workspace ID", mutate: func(c *Config) { c.Workspace.ID = "project" }, field: "workspace.id"},
		{name: "agent", mutate: func(c *Config) { c.Agent.ID = "agent.other" }, field: "agent.id"},
		{name: "tool", mutate: func(c *Config) { c.Agent.Tools = []string{"shell.run"} }, field: "agent.tools"},
		{name: "policy", mutate: func(c *Config) { c.Policy.WorkspaceMutation = "auto" }, field: "policy.workspace_mutate"},
		{name: "retrieval", mutate: func(c *Config) { c.Context.Retrieval = "semantic" }, field: "context.retrieval"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			candidate := config
			candidate.Agent.Tools = append([]string(nil), config.Agent.Tools...)
			testCase.mutate(&candidate)
			err := candidate.Validate()
			if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), testCase.field) {
				t.Fatalf("expected field %s error, got %v", testCase.field, err)
			}
		})
	}
}

func TestResolvePathPrecedenceAndSecretReference(t *testing.T) {
	values := map[string]string{
		"MAESTRO_CONFIG":  "/env/config.yaml",
		"XDG_CONFIG_HOME": "/xdg",
		"HOME":            "/home/fixture",
		"MAESTRO_KEY":     "secret-value",
	}
	getenv := func(name string) string { return values[name] }
	path, err := ResolvePath("/flag/config.yaml", getenv)
	if err != nil || path != "/flag/config.yaml" {
		t.Fatalf("explicit path: %q %v", path, err)
	}
	path, _ = ResolvePath("", getenv)
	if path != "/env/config.yaml" {
		t.Fatalf("environment path: %q", path)
	}
	delete(values, "MAESTRO_CONFIG")
	path, _ = ResolvePath("", getenv)
	if path != "/xdg/maestro/config.yaml" {
		t.Fatalf("XDG path: %q", path)
	}

	config := validConfig(t.TempDir())
	config.Provider.ID = "llama.cpp"
	config.Provider.APIKeyEnv = "MAESTRO_KEY"
	secret, err := config.Secret(getenv)
	if err != nil || secret != "secret-value" {
		t.Fatalf("secret lookup: %q %v", secret, err)
	}
	delete(values, "MAESTRO_KEY")
	if _, err := config.Secret(getenv); !errors.Is(err, ErrSecretMissing) || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("unexpected missing secret error: %v", err)
	}
}

func validConfig(root string) Config {
	return Config{
		Version:   Version,
		Provider:  ProviderConfig{ID: "ollama", BaseURL: "http://127.0.0.1:11434", Timeout: Duration{2 * time.Minute}},
		Models:    ModelsConfig{Chat: "fixture-model"},
		Workspace: WorkspaceConfig{ID: "laravel", Root: filepath.Clean(root), Framework: "laravel"},
		Agent:     AgentConfig{ID: "agent.reference", Tools: []string{"workspace.read", "workspace.patch"}},
		Policy:    PolicyConfig{ID: "policy.test", Model: "allow", WorkspaceInspect: "allow", WorkspaceMutation: "allow"},
		Limits:    LimitsConfig{Duration: Duration{time.Minute}, ModelTurns: 5, ToolCalls: 4, ToolCallsPerTurn: 2, PlanSteps: 3, PlanRevisions: 2, ToolResultBytes: 65536, SessionBytes: 1048576, InputTokens: 10000, OutputTokens: 10000},
		Context:   ContextConfig{Retrieval: "lexical", TopK: 5, MaxTokens: 1024, ReservedTokens: 128, SafetyTokens: 64},
	}
}

func validYAML(root string) string {
	return fmt.Sprintf(`version: 1
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 2m
  api_key_env: ""
models:
  chat: fixture-model
  embedding: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
agent:
  id: agent.reference
  streaming: false
  tools:
    - workspace.read
    - workspace.patch
policy:
  id: policy.test
  model: allow
  workspace_inspect: allow
  workspace_mutate: allow
limits:
  duration: 1m
  model_turns: 5
  tool_calls: 4
  tool_calls_per_turn: 2
  plan_steps: 3
  plan_revisions: 2
  tool_result_bytes: 65536
  session_bytes: 1048576
  input_tokens: 10000
  output_tokens: 10000
context:
  retrieval: lexical
  top_k: 5
  max_tokens: 1024
  reserved_tokens: 128
  safety_tokens: 64
`, root)
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
