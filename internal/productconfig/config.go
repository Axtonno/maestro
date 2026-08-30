// Package productconfig owns Maestro's versioned user-facing configuration.
// It is separate from pkg/runtime.Config, which remains the generic component
// configuration contract used by the Runtime Core.
package productconfig

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const (
	Version          = 1
	CandidateVersion = 2
)

var (
	ErrInvalid       = errors.New("invalid product configuration")
	ErrNotFound      = errors.New("product configuration not found")
	ErrSecretMissing = errors.New("configured secret environment is missing")
)

var environmentName = regexp.MustCompile(`^[A-Z_][A-Z0-9_]{0,127}$`)

type Duration struct{ time.Duration }

func (duration *Duration) UnmarshalYAML(unmarshal func(any) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value, err)
	}
	duration.Duration = parsed
	return nil
}

type Config struct {
	Version     int               `yaml:"version"`
	Provider    ProviderConfig    `yaml:"provider"`
	Models      ModelsConfig      `yaml:"models"`
	Workspace   WorkspaceConfig   `yaml:"workspace"`
	Interaction InteractionConfig `yaml:"interaction,omitempty"`
	Agent       AgentConfig       `yaml:"agent"`
	Policy      PolicyConfig      `yaml:"policy"`
	Limits      LimitsConfig      `yaml:"limits"`
	Context     ContextConfig     `yaml:"context"`

	path string
}

type ProviderConfig struct {
	ID        string   `yaml:"id"`
	BaseURL   string   `yaml:"base_url"`
	Timeout   Duration `yaml:"timeout"`
	APIKeyEnv string   `yaml:"api_key_env"`
}

type ModelsConfig struct {
	Chat      string `yaml:"chat"`
	Embedding string `yaml:"embedding"`
}

type ThinkingMode string

const (
	ThinkingDefault  ThinkingMode = "default"
	ThinkingEnabled  ThinkingMode = "true"
	ThinkingDisabled ThinkingMode = "false"
)

type InteractionConfig struct {
	Chat  ChatProfileConfig  `yaml:"chat"`
	Agent AgentProfileConfig `yaml:"agent"`
}

type ProfileConfig struct {
	Model     string       `yaml:"model"`
	Timeout   Duration     `yaml:"timeout"`
	Streaming bool         `yaml:"streaming"`
	NumCtx    int          `yaml:"num_ctx"`
	Thinking  ThinkingMode `yaml:"thinking"`
}

type ChatProfileConfig struct {
	ProfileConfig  `yaml:",inline"`
	MaxFileBytes   int `yaml:"max_file_bytes"`
	MaxOutputBytes int `yaml:"max_output_bytes"`
}

type AgentProfileConfig struct {
	ProfileConfig `yaml:",inline"`
}

type WorkspaceConfig struct {
	ID        string `yaml:"id"`
	Root      string `yaml:"root"`
	Framework string `yaml:"framework"`
}

type AgentConfig struct {
	ID        string   `yaml:"id"`
	Streaming bool     `yaml:"streaming"`
	Tools     []string `yaml:"tools"`
}

type PolicyConfig struct {
	ID                string `yaml:"id"`
	Model             string `yaml:"model"`
	WorkspaceInspect  string `yaml:"workspace_inspect"`
	WorkspaceMutation string `yaml:"workspace_mutate"`
}

type LimitsConfig struct {
	Duration         Duration `yaml:"duration"`
	ModelTurns       int      `yaml:"model_turns"`
	ToolCalls        int      `yaml:"tool_calls"`
	ToolCallsPerTurn int      `yaml:"tool_calls_per_turn"`
	PlanSteps        int      `yaml:"plan_steps"`
	PlanRevisions    int      `yaml:"plan_revisions"`
	ToolResultBytes  int      `yaml:"tool_result_bytes"`
	SessionBytes     int      `yaml:"session_bytes"`
	InputTokens      int      `yaml:"input_tokens"`
	OutputTokens     int      `yaml:"output_tokens"`
}

type ContextConfig struct {
	Retrieval      string `yaml:"retrieval"`
	TopK           int    `yaml:"top_k"`
	MaxTokens      int    `yaml:"max_tokens"`
	ReservedTokens int    `yaml:"reserved_tokens"`
	SafetyTokens   int    `yaml:"safety_tokens"`
}

func (config Config) Path() string { return config.path }

func (config Config) HasChatProfile() bool { return config.Version == CandidateVersion }

func (config Config) ChatProfile() (ChatProfileConfig, bool) {
	if !config.HasChatProfile() {
		return ChatProfileConfig{}, false
	}
	return config.Interaction.Chat, true
}

func (config Config) AgentProfile() AgentProfileConfig {
	if config.Version == CandidateVersion {
		return config.Interaction.Agent
	}
	return AgentProfileConfig{ProfileConfig: ProfileConfig{
		Model: config.Models.Chat, Timeout: config.Provider.Timeout,
		Streaming: config.Agent.Streaming,
	}}
}

func (config ProfileConfig) GenerationOptions() pkgProvider.GenerationOptions {
	return pkgProvider.GenerationOptions{
		ContextWindow: config.NumCtx,
		Thinking:      pkgProvider.ThinkingMode(config.Thinking),
	}
}

func (config Config) AgentLimits() pkgAgent.Limits {
	return pkgAgent.Limits{
		MaxDuration:         config.Limits.Duration.Duration,
		MaxModelTurns:       config.Limits.ModelTurns,
		MaxToolCalls:        config.Limits.ToolCalls,
		MaxToolCallsPerTurn: config.Limits.ToolCallsPerTurn,
		MaxPlanSteps:        config.Limits.PlanSteps,
		MaxPlanRevisions:    config.Limits.PlanRevisions,
		MaxToolResultBytes:  config.Limits.ToolResultBytes,
		MaxSessionBytes:     config.Limits.SessionBytes,
		MaxInputTokens:      config.Limits.InputTokens,
		MaxOutputTokens:     config.Limits.OutputTokens,
	}
}

func (config Config) ContextBudget() pkgContext.Budget {
	return pkgContext.Budget{
		MaxTokens:      config.Context.MaxTokens,
		ReservedTokens: config.Context.ReservedTokens,
		SafetyTokens:   config.Context.SafetyTokens,
	}
}

func (config Config) ToolIDs() []pkgTool.ID {
	ids := make([]pkgTool.ID, len(config.Agent.Tools))
	for index, id := range config.Agent.Tools {
		ids[index] = pkgTool.ID(id)
	}
	slices.Sort(ids)
	return ids
}

func (config Config) Validate() error {
	if config.Version != Version && config.Version != CandidateVersion {
		return fieldError("version", fmt.Sprintf("must equal %d or %d", Version, CandidateVersion))
	}
	if err := validateProvider(config.Provider); err != nil {
		return err
	}
	if config.Version == Version {
		if !exact(config.Models.Chat, 512) {
			return fieldError("models.chat", "must be explicit and exact")
		}
	} else if err := validateInteraction(config.Interaction, config.Provider.Timeout.Duration); err != nil {
		return err
	}
	if config.Models.Embedding != "" && !exact(config.Models.Embedding, 512) {
		return fieldError("models.embedding", "must be exact when set")
	}
	if err := validateWorkspace(config.Workspace); err != nil {
		return err
	}
	if err := pkgAgent.ID(config.Agent.ID).Validate(); err != nil {
		return fieldWrap("agent.id", err)
	}
	if !supportedAgentID(config.Agent.ID) {
		return fieldError("agent.id", "agent is not supported by this build")
	}
	if err := validateTools(config.Agent.Tools); err != nil {
		return err
	}
	if err := pkgTool.PolicyID(config.Policy.ID).Validate(); err != nil {
		return fieldWrap("policy.id", err)
	}
	for name, value := range map[string]string{
		"policy.model":             config.Policy.Model,
		"policy.workspace_inspect": config.Policy.WorkspaceInspect,
		"policy.workspace_mutate":  config.Policy.WorkspaceMutation,
	} {
		if !validDecision(value) {
			return fieldError(name, "must be allow, deny, or prompt")
		}
	}
	if err := config.AgentLimits().Validate(); err != nil {
		return fieldWrap("limits", err)
	}
	if config.Context.Retrieval != "lexical" {
		return fieldError("context.retrieval", "v0.1.0 requires lexical")
	}
	if config.Context.TopK < 1 || config.Context.TopK > 100 {
		return fieldError("context.top_k", "must be between 1 and 100")
	}
	if err := config.ContextBudget().Validate(); err != nil {
		return fieldWrap("context", err)
	}
	return nil
}

// ValidateChatExecutionProfile validates the authority and inputs used by the
// direct-chat composition root. Agent identity, tools, limits and retrieval
// are intentionally outside this mode and remain subject to Validate and
// ValidateExecutionProfile when an agent command is requested.
func (config Config) ValidateChatExecutionProfile() error {
	if config.Version != CandidateVersion {
		return fieldError("version", fmt.Sprintf("direct chat requires %d", CandidateVersion))
	}
	if err := validateProvider(config.Provider); err != nil {
		return err
	}
	if config.Models.Embedding != "" && !exact(config.Models.Embedding, 512) {
		return fieldError("models.embedding", "must be exact when set")
	}
	if err := validateProfile("interaction.chat", config.Interaction.Chat.ProfileConfig, config.Provider.Timeout.Duration); err != nil {
		return err
	}
	if config.Interaction.Chat.MaxFileBytes < 1 || config.Interaction.Chat.MaxFileBytes > 16<<20 {
		return fieldError("interaction.chat.max_file_bytes", "must be between 1 and 16777216")
	}
	if config.Interaction.Chat.MaxOutputBytes < 1 || config.Interaction.Chat.MaxOutputBytes > 16<<20 {
		return fieldError("interaction.chat.max_output_bytes", "must be between 1 and 16777216")
	}
	if err := validateWorkspace(config.Workspace); err != nil {
		return err
	}
	if config.Policy.WorkspaceMutation != "deny" {
		return fieldError("policy.workspace_mutate", "direct chat requires deny")
	}
	return nil
}

func validateWorkspace(config WorkspaceConfig) error {
	workspaceID := pkgContext.WorkspaceID(config.ID)
	if err := workspaceID.Validate(); err != nil {
		return fieldWrap("workspace.id", err)
	}
	if config.Framework != "laravel" {
		return fieldError("workspace.framework", "requires laravel")
	}
	if workspaceID != "laravel" {
		return fieldError("workspace.id", "must equal the authoritative Laravel workspace ID \"laravel\"")
	}
	if config.Root == "" || strings.ContainsAny(config.Root, "\r\n\x00") ||
		!filepath.IsAbs(config.Root) || filepath.Clean(config.Root) != config.Root {
		return fieldError("workspace.root", "must be absolute and normalized")
	}
	return nil
}

func validateInteraction(config InteractionConfig, transportTimeout time.Duration) error {
	if err := validateProfile("interaction.chat", config.Chat.ProfileConfig, transportTimeout); err != nil {
		return err
	}
	if err := validateProfile("interaction.agent", config.Agent.ProfileConfig, transportTimeout); err != nil {
		return err
	}
	if config.Chat.MaxFileBytes < 1 || config.Chat.MaxFileBytes > 16<<20 {
		return fieldError("interaction.chat.max_file_bytes", "must be between 1 and 16777216")
	}
	if config.Chat.MaxOutputBytes < 1 || config.Chat.MaxOutputBytes > 16<<20 {
		return fieldError("interaction.chat.max_output_bytes", "must be between 1 and 16777216")
	}
	return nil
}

func validateProfile(field string, config ProfileConfig, transportTimeout time.Duration) error {
	if !exact(config.Model, 512) {
		return fieldError(field+".model", "must be explicit and exact")
	}
	if config.Timeout.Duration <= 0 || config.Timeout.Duration > transportTimeout {
		return fieldError(field+".timeout", "must be positive and at most provider.timeout")
	}
	if config.NumCtx < 128 || config.NumCtx > 1<<20 {
		return fieldError(field+".num_ctx", "must be between 128 and 1048576")
	}
	switch config.Thinking {
	case ThinkingDefault, ThinkingEnabled, ThinkingDisabled:
	default:
		return fieldError(field+".thinking", "must be default, true, or false")
	}
	return nil
}

// ValidateExecutionProfile narrows the versioned schema to one of the product
// profiles supported by the application composition root. The generic 0.x
// schema continues to parse experimental tool IDs, but cannot silently grant
// them product authority.
func (config Config) ValidateExecutionProfile() error {
	if err := config.Validate(); err != nil {
		return err
	}
	tools := make(map[string]struct{}, len(config.Agent.Tools))
	for _, id := range config.Agent.Tools {
		tools[id] = struct{}{}
	}
	_, patch := tools["workspace.patch"]
	_, write := tools["workspace.write"]
	_, read := tools["workspace.read"]
	if write {
		return fieldError("agent.tools", "workspace.write is outside the supported execution profiles")
	}
	if patch {
		if !read {
			return fieldError("agent.tools", "controlled mutation requires workspace.read before workspace.patch")
		}
		if config.Policy.WorkspaceMutation != "prompt" {
			return fieldError("policy.workspace_mutate", "controlled mutation requires prompt")
		}
		return nil
	}
	if config.Policy.WorkspaceMutation != "deny" {
		return fieldError("policy.workspace_mutate", "the read-only profile requires deny")
	}
	return nil
}

func validateProvider(config ProviderConfig) error {
	id := pkgProvider.ID(config.ID)
	if id != "ollama" && id != "llama.cpp" {
		return fieldError("provider.id", "must be ollama or llama.cpp")
	}
	parsed, err := url.Parse(config.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return fieldError("provider.base_url", "must be an HTTP(S) origin without credentials, query, fragment, or API path")
	}
	if config.Timeout.Duration <= 0 || config.Timeout.Duration > 10*time.Minute {
		return fieldError("provider.timeout", "must be positive and at most 10m")
	}
	if config.APIKeyEnv != "" && !environmentName.MatchString(config.APIKeyEnv) {
		return fieldError("provider.api_key_env", "must be an uppercase environment variable name")
	}
	if id == "ollama" && config.APIKeyEnv != "" {
		return fieldError("provider.api_key_env", "is supported only by llama.cpp")
	}
	return nil
}

func validateTools(values []string) error {
	if len(values) == 0 {
		return fieldError("agent.tools", "requires at least one built-in workspace tool")
	}
	allowed := map[string]struct{}{
		"workspace.list": {}, "workspace.read": {}, "workspace.search": {},
		"workspace.write": {}, "workspace.patch": {},
	}
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if _, exists := allowed[value]; !exists {
			return fieldError(fmt.Sprintf("agent.tools[%d]", index), "is not a v0.1.0 built-in workspace tool")
		}
		if _, exists := seen[value]; exists {
			return fieldError(fmt.Sprintf("agent.tools[%d]", index), "is duplicated")
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validDecision(value string) bool {
	return value == "allow" || value == "deny" || value == "prompt"
}

func exact(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\r\n\x00")
}

func fieldError(field, message string) error {
	return &validationError{field: field, message: message, cause: ErrInvalid}
}

func fieldWrap(field string, err error) error {
	return &validationError{field: field, message: err.Error(), cause: ErrInvalid}
}

func ResolvePath(explicit string, getenv func(string) string) (string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	path := strings.TrimSpace(explicit)
	if path == "" {
		path = strings.TrimSpace(getenv("MAESTRO_CONFIG"))
	}
	if path == "" {
		base := strings.TrimSpace(getenv("XDG_CONFIG_HOME"))
		if base == "" {
			home := strings.TrimSpace(getenv("HOME"))
			if home == "" {
				return "", fmt.Errorf("resolve default configuration: HOME and XDG_CONFIG_HOME are empty: %w", ErrNotFound)
			}
			base = filepath.Join(home, ".config")
		}
		path = filepath.Join(base, "maestro", "config.yaml")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve configuration path: %w: %w", err, ErrInvalid)
	}
	return filepath.Clean(absolute), nil
}

func (config Config) Secret(getenv func(string) string) (string, error) {
	if config.Provider.APIKeyEnv == "" {
		return "", nil
	}
	if getenv == nil {
		getenv = os.Getenv
	}
	value := getenv(config.Provider.APIKeyEnv)
	if value == "" {
		return "", fmt.Errorf("secret environment %s is empty: %w", config.Provider.APIKeyEnv, ErrSecretMissing)
	}
	return value, nil
}
