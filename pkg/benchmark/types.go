package benchmark

import (
	"context"
	"time"
)

const (
	ManifestSchemaVersion = 1
	ReportSchemaVersion   = "1.1.0"
)

type ResultState string

const (
	ResultPassed      ResultState = "passed"
	ResultFailed      ResultState = "failed"
	ResultSkipped     ResultState = "skipped"
	ResultUnsupported ResultState = "unsupported"
)

type RedactionPolicy struct {
	Exclude []string `json:"exclude" yaml:"exclude"`
}

type ProviderManifest struct {
	RequiredEnvironment    []string          `json:"required_environment,omitempty" yaml:"required_environment"`
	ModelEnvironment       map[string]string `json:"model_environment,omitempty" yaml:"model_environment"`
	OptionalEnvironment    []string          `json:"optional_environment,omitempty" yaml:"optional_environment"`
	VersionSensitiveLimits []string          `json:"version_sensitive_limits,omitempty" yaml:"version_sensitive_limits"`
}

type ScenarioDefinition struct {
	ID            string `json:"id" yaml:"id"`
	Capability    string `json:"capability" yaml:"capability"`
	ModelRole     string `json:"model_role" yaml:"model_role"`
	Cleanup       string `json:"cleanup" yaml:"cleanup"`
	MutationGuard string `json:"mutation_guard,omitempty" yaml:"mutation_guard,omitempty"`
}

type Manifest struct {
	Version         int                         `json:"version" yaml:"version"`
	Owner           string                      `json:"owner" yaml:"owner"`
	SourceMilestone string                      `json:"source_milestone" yaml:"source_milestone"`
	ReportRedaction RedactionPolicy             `json:"report_redaction" yaml:"report_redaction"`
	Providers       map[string]ProviderManifest `json:"providers" yaml:"providers"`
	Scenarios       []ScenarioDefinition        `json:"scenarios" yaml:"scenarios"`
	ResultStates    []ResultState               `json:"result_states" yaml:"result_states"`
}

type Iteration struct {
	Index  int  `json:"index"`
	Warmup bool `json:"warmup"`
}

type Measurement struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Scope  string  `json:"scope,omitempty"`
	Method string  `json:"method,omitempty"`
}

type ErrorRecord struct {
	Kind       string `json:"kind"`
	Code       string `json:"code"`
	Operation  string `json:"operation,omitempty"`
	Provider   string `json:"provider,omitempty"`
	Model      string `json:"model,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Retryable  bool   `json:"retryable,omitempty"`
}

type IterationResult struct {
	State        ResultState
	ReasonCode   string
	Measurements []Measurement
	Error        *ErrorRecord
}

// Scenario is the executable unit consumed by the benchmark runner. Cleanup
// is invoked after every warmup and measured iteration, including failures and
// panics.
type Scenario interface {
	Definition() ScenarioDefinition
	Run(context.Context, Iteration) (IterationResult, error)
	Cleanup(context.Context, Iteration) error
}

// ScenarioFuncs adapts functions to Scenario and is useful for small scenarios
// and deterministic tests. A nil CleanupFunc is treated as a no-op.
type ScenarioFuncs struct {
	DefinitionValue ScenarioDefinition
	RunFunc         func(context.Context, Iteration) (IterationResult, error)
	CleanupFunc     func(context.Context, Iteration) error
}

func (s ScenarioFuncs) Definition() ScenarioDefinition {
	return s.DefinitionValue
}

func (s ScenarioFuncs) Run(
	ctx context.Context,
	iteration Iteration,
) (IterationResult, error) {
	if s.RunFunc == nil {
		return IterationResult{
			State: ResultFailed,
			Error: &ErrorRecord{Kind: "runner", Code: "missing_run_function"},
		}, nil
	}

	return s.RunFunc(ctx, iteration)
}

func (s ScenarioFuncs) Cleanup(
	ctx context.Context,
	iteration Iteration,
) error {
	if s.CleanupFunc == nil {
		return nil
	}

	return s.CleanupFunc(ctx, iteration)
}

type Aggregate struct {
	Name   string   `json:"name"`
	Unit   string   `json:"unit"`
	Scope  string   `json:"scope,omitempty"`
	Method string   `json:"method,omitempty"`
	Count  int      `json:"count"`
	Min    float64  `json:"min"`
	Median float64  `json:"median"`
	P95    *float64 `json:"p95,omitempty"`
	Max    float64  `json:"max"`
}

type Sample struct {
	Iteration    Iteration     `json:"iteration"`
	StartedAt    time.Time     `json:"started_at"`
	DurationMS   float64       `json:"duration_ms"`
	State        ResultState   `json:"state"`
	ReasonCode   string        `json:"reason_code,omitempty"`
	Measurements []Measurement `json:"measurements,omitempty"`
	Error        *ErrorRecord  `json:"error,omitempty"`
	CleanupError *ErrorRecord  `json:"cleanup_error,omitempty"`
}

type ScenarioReport struct {
	Scenario   ScenarioDefinition `json:"scenario"`
	State      ResultState        `json:"state"`
	Samples    []Sample           `json:"samples"`
	Aggregates []Aggregate        `json:"aggregates,omitempty"`
}

type HardwareProfile struct {
	OS           string `json:"os,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	CPU          string `json:"cpu,omitempty"`
	LogicalCPUs  int    `json:"logical_cpus,omitempty"`
	MemoryMB     int64  `json:"memory_mb,omitempty"`
	GPU          string `json:"gpu,omitempty"`
	Backend      string `json:"backend,omitempty"`
	VRAMMB       int64  `json:"vram_mb,omitempty"`
}

type ProviderProfile struct {
	ID            string `json:"id,omitempty"`
	ServerVersion string `json:"server_version,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
}

type ModelProfile struct {
	ID            string `json:"id,omitempty"`
	Digest        string `json:"digest,omitempty"`
	Format        string `json:"format,omitempty"`
	Quantization  string `json:"quantization,omitempty"`
	ContextLength int    `json:"context_length,omitempty"`
}

type PluginProfile struct {
	ID      string `json:"id"`
	Version string `json:"version,omitempty"`
}

type GenerationProfile struct {
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	StopCount   int      `json:"stop_count,omitempty"`
}

type ExecutionProfile struct {
	Cold             bool    `json:"cold"`
	Warmup           int     `json:"warmup"`
	Runs             int     `json:"runs"`
	TimeoutMS        float64 `json:"timeout_ms,omitempty"`
	CleanupTimeoutMS float64 `json:"cleanup_timeout_ms"`
}

type DatasetProfile struct {
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
}

type ConfigurationProfile struct {
	Hardware   HardwareProfile         `json:"hardware"`
	Provider   ProviderProfile         `json:"provider"`
	Model      ModelProfile            `json:"model"`
	Models     map[string]ModelProfile `json:"models,omitempty"`
	Plugins    []PluginProfile         `json:"plugins,omitempty"`
	Generation GenerationProfile       `json:"generation"`
	Execution  ExecutionProfile        `json:"execution"`
	Dataset    DatasetProfile          `json:"dataset"`
}

type RunMetadata struct {
	Command         string `json:"command,omitempty"`
	MaestroVersion  string `json:"maestro_version,omitempty"`
	MaestroCommit   string `json:"maestro_commit,omitempty"`
	ManifestVersion int    `json:"manifest_version"`
	ManifestOwner   string `json:"manifest_owner"`
	SourceMilestone string `json:"source_milestone,omitempty"`
}

type Report struct {
	SchemaVersion string               `json:"schema_version"`
	RunID         string               `json:"run_id"`
	CreatedAt     time.Time            `json:"created_at"`
	CompletedAt   time.Time            `json:"completed_at"`
	DurationMS    float64              `json:"duration_ms"`
	Metadata      RunMetadata          `json:"metadata"`
	Configuration ConfigurationProfile `json:"configuration"`
	Scenarios     []ScenarioReport     `json:"scenarios"`
}
