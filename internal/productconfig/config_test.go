package productconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

func TestInteractionExampleLoadsStrictProfiles(t *testing.T) {
	config, err := Load("../../configs/maestro.interaction.example.yaml")
	if err != nil {
		t.Fatalf("interaction example is invalid: %v", err)
	}
	chat, ok := config.ChatProfile()
	agent := config.AgentProfile()
	if !ok || config.Version != CandidateVersion || chat.Model != "qwen2.5-coder:7b" ||
		chat.NumCtx != 4096 || chat.Thinking != ThinkingDisabled || chat.Streaming ||
		chat.MaxFileBytes != 1<<20 || chat.MaxOutputBytes != 1<<20 {
		t.Fatalf("unexpected chat profile: %#v %#v", chat, config)
	}
	if agent.Model != "qwen3.5:9b" || agent.NumCtx != 8192 ||
		agent.Thinking != ThinkingDefault || !agent.Streaming ||
		config.Models.Chat != agent.Model || !config.Agent.Streaming {
		t.Fatalf("unexpected normalized agent profile: %#v %#v", agent, config)
	}
	options := chat.GenerationOptions()
	if options.ContextWindow != 4096 || options.Thinking != "false" {
		t.Fatalf("unexpected generation options: %#v", options)
	}
}

func TestChatExampleLoadsWithoutAgentConfiguration(t *testing.T) {
	config, err := LoadChat("../../configs/maestro.chat.example.yaml")
	if err != nil {
		t.Fatalf("chat example is invalid: %v", err)
	}
	chat, ok := config.ChatProfile()
	if !ok || chat.Model != "qwen3.5:9b" || chat.NumCtx != 4096 ||
		chat.Thinking != ThinkingDisabled || !chat.Streaming ||
		chat.MaxFileBytes != 1<<20 || chat.MaxOutputBytes != 1<<20 {
		t.Fatalf("unexpected chat profile: %#v", chat)
	}
	if config.Agent.ID != "" || len(config.Agent.Tools) != 0 || config.Context.Retrieval != "" {
		t.Fatalf("chat-only example acquired agent configuration: %#v", config)
	}
	if config.Policy.WorkspaceMutation != "deny" {
		t.Fatalf("chat-only example does not deny mutation: %#v", config.Policy)
	}
	if _, err := Load("../../configs/maestro.chat.example.yaml"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("agent loader accepted chat-only configuration: %v", err)
	}
}

func TestLoadChatValidatesOnlyChatAuthorityAndStrictSyntax(t *testing.T) {
	valid := validChatYAML("workspace")
	directory := t.TempDir()
	for _, testCase := range []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{name: "chat only", encoded: valid},
		{name: "irrelevant invalid agent", encoded: valid + "agent:\n  id: agent.unsupported\n  tools: [shell.run]\n"},
		{name: "mutation prompt", encoded: strings.Replace(valid, "workspace_mutate: deny", "workspace_mutate: prompt", 1), wantErr: true},
		{name: "missing model", encoded: strings.Replace(valid, "model: chat-model", "model: \"\"", 1), wantErr: true},
		{name: "unknown", encoded: strings.Replace(valid, "version: 2", "version: 2\nunknown: true", 1), wantErr: true},
		{name: "duplicate", encoded: strings.Replace(valid, "version: 2", "version: 2\nversion: 2", 1), wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(directory, strings.ReplaceAll(testCase.name, " ", "-")+".yaml")
			writeConfig(t, path, testCase.encoded)
			config, err := LoadChat(path)
			if testCase.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("invalid chat config accepted: %#v %v", config, err)
				}
				return
			}
			if err != nil || filepath.Base(config.Workspace.Root) != "workspace" {
				t.Fatalf("load chat config: %#v %v", config, err)
			}
		})
	}
}

func TestMilestone14CandidateProfileIsFrozenAndValid(t *testing.T) {
	config, err := Load("../../configs/maestro.milestone-14-candidate.yaml")
	if err != nil {
		t.Fatalf("candidate profile is invalid: %v", err)
	}
	chat, ok := config.ChatProfile()
	if !ok || chat.Model != "qwen2.5-coder:7b" || !chat.Streaming ||
		chat.NumCtx != 4096 || chat.Thinking != ThinkingDisabled ||
		chat.Timeout.Duration != 5*time.Minute || chat.MaxFileBytes != 1<<20 ||
		chat.MaxOutputBytes != 1<<20 ||
		filepath.Base(config.Workspace.Root) != "laravel-v1" {
		t.Fatalf("candidate profile drifted: %#v %#v", chat, config.Workspace)
	}
}

func TestV2RejectsLegacyAndInvalidProfileFields(t *testing.T) {
	encoded, err := os.ReadFile("../../configs/maestro.interaction.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(string) string
	}{
		{name: "legacy models chat", mutate: func(value string) string {
			return strings.Replace(value, "models:\n", "models:\n  chat: legacy\n", 1)
		}},
		{name: "legacy agent streaming", mutate: func(value string) string {
			return strings.Replace(value, "agent:\n  id:", "agent:\n  streaming: false\n  id:", 1)
		}},
		{name: "invalid thinking", mutate: func(value string) string {
			return strings.Replace(value, "thinking: \"false\"", "thinking: sometimes", 1)
		}},
		{name: "timeout above transport", mutate: func(value string) string {
			return strings.Replace(value, "timeout: 5m", "timeout: 11m", 1)
		}},
		{name: "small num ctx", mutate: func(value string) string {
			return strings.Replace(value, "num_ctx: 4096", "num_ctx: 64", 1)
		}},
		{name: "unknown field", mutate: func(value string) string {
			return strings.Replace(value, "version: 2", "version: 2\nunknown: true", 1)
		}},
		{name: "duplicate field", mutate: func(value string) string {
			return strings.Replace(value, "version: 2", "version: 2\nversion: 2", 1)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			writeConfig(t, path, testCase.mutate(string(encoded)))
			if _, err := Load(path); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid v2 config accepted: %v", err)
			}
		})
	}
}

func TestPublishedMutatingExampleIsSeparateExplicitAndPrompted(t *testing.T) {
	config, err := Load("../../configs/maestro.mutating.example.yaml")
	if err != nil {
		t.Fatalf("published mutating example is invalid: %v", err)
	}
	if err := config.ValidateExecutionProfile(); err != nil {
		t.Fatalf("published mutating execution profile is invalid: %v", err)
	}
	ids := config.ToolIDs()
	if !slices.Contains(ids, "workspace.read") || !slices.Contains(ids, "workspace.patch") || slices.Contains(ids, "workspace.write") {
		t.Fatalf("unexpected mutating tool profile: %v", ids)
	}
	if config.Policy.WorkspaceMutation != "prompt" || config.Models.Chat != "ibm/granite4.1:8b" {
		t.Fatalf("unexpected mutating policy/model: %#v", config)
	}
}

func TestExecutionProfileRejectsImplicitOrUnsupportedMutation(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "patch allow", mutate: func(config *Config) { config.Policy.WorkspaceMutation = "allow" }},
		{name: "patch without read", mutate: func(config *Config) { config.Agent.Tools = []string{"workspace.patch"} }},
		{name: "write", mutate: func(config *Config) { config.Agent.Tools = []string{"workspace.read", "workspace.write"} }},
		{name: "read-only prompt", mutate: func(config *Config) {
			config.Agent.Tools = []string{"workspace.read"}
			config.Policy.WorkspaceMutation = "prompt"
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			config := validConfig(t.TempDir())
			testCase.mutate(&config)
			if err := config.ValidateExecutionProfile(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unsupported execution profile accepted: %v", err)
			}
		})
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
		{name: "version", mutate: func(c *Config) { c.Version = 3 }, field: "version"},
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
		Policy:    PolicyConfig{ID: "policy.test", Model: "allow", WorkspaceInspect: "allow", WorkspaceMutation: "prompt"},
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
  workspace_mutate: prompt
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

func validChatYAML(root string) string {
	return fmt.Sprintf(`version: 2
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 2m
  api_key_env: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
interaction:
  chat:
    model: chat-model
    timeout: 1m
    streaming: false
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576
policy:
  workspace_mutate: deny
`, root)
}

func writeConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
