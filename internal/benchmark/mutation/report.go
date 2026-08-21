package mutation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

const ReportSchemaVersion = "mutation-qualification-report/1.0.0"

type Report struct {
	SchemaVersion string                       `json:"schema_version"`
	RunID         string                       `json:"run_id"`
	CreatedAt     time.Time                    `json:"created_at"`
	CompletedAt   time.Time                    `json:"completed_at"`
	DurationMS    float64                      `json:"duration_ms"`
	Profile       ReportProfile                `json:"profile"`
	Build         ReportBuild                  `json:"build"`
	Hardware      pkgBenchmark.HardwareProfile `json:"hardware"`
	Gate          string                       `json:"gate"`
	Scenario      string                       `json:"scenario"`
	Required      int                          `json:"required_passes"`
	FailFast      bool                         `json:"fail_fast"`
	State         string                       `json:"state"`
	Samples       []Sample                     `json:"samples"`
}

type ReportProfile struct {
	Version int    `json:"version"`
	Digest  string `json:"digest"`
	Target  string `json:"target"`
	Model   string `json:"model"`
}

type ReportBuild struct {
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
	Digest  string `json:"digest,omitempty"`
}

type Sample struct {
	Attempt          int       `json:"attempt"`
	State            string    `json:"state"`
	ReasonCode       string    `json:"reason_code,omitempty"`
	FailureClass     string    `json:"failure_class,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	DurationMS       float64   `json:"duration_ms"`
	Terminal         string    `json:"terminal"`
	ModelTurns       int       `json:"model_turns"`
	ToolCalls        int       `json:"tool_calls"`
	MutationAttempts int       `json:"mutation_attempts"`
	Approval         string    `json:"approval"`
	Lifecycle        []string  `json:"lifecycle"`
	InitialSHA256    string    `json:"initial_sha256"`
	ExpectedSHA256   string    `json:"expected_sha256"`
	FinalSHA256      string    `json:"final_sha256"`
	WorkspaceDigest  string    `json:"workspace_digest"`
	ContextFreshness string    `json:"context_freshness"`
	TemporaryCleanup string    `json:"temporary_cleanup"`
}

func (report Report) Validate() error {
	if report.SchemaVersion != ReportSchemaVersion || !safeID(report.RunID, 128) ||
		report.CreatedAt.IsZero() || report.CompletedAt.Before(report.CreatedAt) ||
		report.DurationMS < 0 || math.IsNaN(report.DurationMS) || math.IsInf(report.DurationMS, 0) {
		return errors.New("mutation qualification report metadata is invalid")
	}
	if report.Profile.Version != ProfileVersion || !validSHA256(report.Profile.Digest) ||
		report.Profile.Target != "linux_amd64/ollama" || !safeText(report.Profile.Model, 512) {
		return errors.New("mutation qualification report profile is invalid")
	}
	if !safeOptionalID(report.Build.Version, 128) || !safeOptionalID(report.Build.Commit, 128) ||
		(report.Build.Digest != "" && !validSHA256(report.Build.Digest)) {
		return errors.New("mutation qualification report build identity is invalid")
	}
	if report.Hardware.LogicalCPUs < 0 || report.Hardware.MemoryMB < 0 || report.Hardware.VRAMMB < 0 {
		return errors.New("mutation qualification report hardware is invalid")
	}
	if report.Gate != "deterministic" && report.Gate != GateA && report.Gate != GateB && report.Gate != GateC {
		return errors.New("mutation qualification report gate is invalid")
	}
	if !safeID(report.Scenario, 128) || report.Required < 1 || !report.FailFast ||
		!validState(report.State) || len(report.Samples) == 0 || len(report.Samples) > report.Required {
		return errors.New("mutation qualification report gate result is invalid")
	}
	expectedState := "passed"
	for index, sample := range report.Samples {
		if err := sample.Validate(report, index+1); err != nil {
			return fmt.Errorf("mutation qualification report sample: %w", err)
		}
		if expectedState == "passed" && sample.State != "passed" {
			expectedState = sample.State
		}
	}
	if expectedState == "passed" && len(report.Samples) != report.Required {
		expectedState = "failed"
	}
	if report.State != expectedState {
		return errors.New("mutation qualification report state differs from samples")
	}
	return nil
}

func (sample Sample) Validate(report Report, expectedAttempt int) error {
	if sample.Attempt != expectedAttempt || !validState(sample.State) || sample.StartedAt.IsZero() ||
		sample.StartedAt.Before(report.CreatedAt) || sample.StartedAt.After(report.CompletedAt) ||
		sample.DurationMS < 0 || math.IsNaN(sample.DurationMS) || math.IsInf(sample.DurationMS, 0) ||
		sample.ModelTurns < 0 || sample.ToolCalls < 0 || sample.MutationAttempts < 0 ||
		sample.MutationAttempts > 1 {
		return errors.New("sample metadata is invalid")
	}
	if sample.State == "failed" {
		if !safeID(sample.ReasonCode, 128) || !validFailureClass(sample.FailureClass) {
			return errors.New("failed sample classification is invalid")
		}
	} else if sample.FailureClass != "" || (sample.State == "passed" && sample.ReasonCode != "") {
		return errors.New("sample classification contradicts state")
	} else if (sample.State == "skipped" || sample.State == "unsupported") && !safeID(sample.ReasonCode, 128) {
		return errors.New("non-passed sample reason is invalid")
	}
	if !safeID(sample.Terminal, 128) || !validApproval(sample.Approval) ||
		!validFreshness(sample.ContextFreshness) || !validCleanup(sample.TemporaryCleanup) ||
		!validSHA256(sample.InitialSHA256) || !validSHA256(sample.ExpectedSHA256) ||
		!validSHA256(sample.FinalSHA256) || !validSHA256(sample.WorkspaceDigest) {
		return errors.New("sample observations are invalid")
	}
	if !slices.Equal(sample.Lifecycle, normalizedLifecycle(sample.Lifecycle)) {
		return errors.New("sample lifecycle is invalid")
	}
	return nil
}

func EncodeReportJSON(writer io.Writer, report Report) error {
	if writer == nil {
		return errors.New("mutation qualification JSON writer is nil")
	}
	if err := report.Validate(); err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func DecodeReportJSON(reader io.Reader) (Report, error) {
	if reader == nil {
		return Report{}, errors.New("mutation qualification JSON reader is nil")
	}
	decoder := json.NewDecoder(io.LimitReader(reader, 16<<20))
	decoder.DisallowUnknownFields()
	var report Report
	if err := decoder.Decode(&report); err != nil {
		return Report{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Report{}, errors.New("mutation qualification report contains multiple JSON documents")
		}
		return Report{}, err
	}
	if err := report.Validate(); err != nil {
		return Report{}, err
	}
	return report, nil
}

func EncodeReportMarkdown(writer io.Writer, report Report) error {
	if writer == nil {
		return errors.New("mutation qualification Markdown writer is nil")
	}
	if err := report.Validate(); err != nil {
		return err
	}
	var rendered strings.Builder
	rendered.WriteString("# Maestro Mutation Qualification Report\n\n")
	fmt.Fprintf(&rendered, "- Schema: `%s`\n", report.SchemaVersion)
	fmt.Fprintf(&rendered, "- Run: `%s`\n", report.RunID)
	fmt.Fprintf(&rendered, "- Gate: `%s`\n", report.Gate)
	fmt.Fprintf(&rendered, "- Scenario: `%s`\n", report.Scenario)
	fmt.Fprintf(&rendered, "- State: **%s**\n", strings.ToUpper(report.State))
	fmt.Fprintf(&rendered, "- Candidate: `%s`, `%s`\n", report.Profile.Target, report.Profile.Model)
	fmt.Fprintf(&rendered, "- Profile SHA-256: `%s`\n\n", report.Profile.Digest)
	rendered.WriteString("| Attempt | State | Terminal | Turns | Calls | Mutations | Approval | Freshness | Cleanup |\n")
	rendered.WriteString("|---:|---|---|---:|---:|---:|---|---|---|\n")
	for _, sample := range report.Samples {
		fmt.Fprintf(&rendered, "| %d | %s | %s | %d | %d | %d | %s | %s | %s |\n",
			sample.Attempt, sample.State, sample.Terminal, sample.ModelTurns, sample.ToolCalls,
			sample.MutationAttempts, sample.Approval, sample.ContextFreshness, sample.TemporaryCleanup)
	}
	rendered.WriteString("\n## Workspace evidence\n\n")
	for _, sample := range report.Samples {
		fmt.Fprintf(&rendered, "- Attempt %d: initial `%s`, final `%s`, workspace `%s`.\n",
			sample.Attempt, sample.InitialSHA256, sample.FinalSHA256, sample.WorkspaceDigest)
	}
	_, err := io.WriteString(writer, rendered.String())
	return err
}

func WriteReport(path string, encode func(io.Writer) error) error {
	if path == "" || path == "-" || encode == nil {
		return errors.New("mutation qualification report path is invalid")
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".maestro-mutation-report-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(name)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	if err := encode(temporary); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	committed = true
	return nil
}

func JSONBytes(report Report) ([]byte, error) {
	var buffer bytes.Buffer
	if err := EncodeReportJSON(&buffer, report); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func validState(value string) bool {
	return value == "passed" || value == "failed" || value == "skipped" || value == "unsupported"
}

func validFailureClass(value string) bool {
	return value == "product" || value == "model" || value == "environment" || value == "operator" || value == "harness"
}

func validApproval(value string) bool {
	return value == "not_requested" || value == "allowed_once" || value == "denied" || value == "unavailable"
}

func validFreshness(value string) bool {
	return value == "fresh" || value == "stale" || value == "not_applicable"
}

func validCleanup(value string) bool {
	return value == "clean" || value == "failed" || value == "not_applicable"
}

func normalizedLifecycle(values []string) []string {
	stage := map[string]int{
		"proposal_prepared": 1, "approval_allowed": 2, "approval_denied": 2,
		"apply_started": 3, "apply_succeeded": 3, "apply_failed": 3,
		"reindex_started": 4, "reindex_succeeded": 4, "reindex_failed": 4,
		"terminal_completed": 5, "terminal_failed": 5, "terminal_canceled": 5,
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	previous := 0
	for _, value := range values {
		current, exists := stage[value]
		if !exists || current < previous {
			return nil
		}
		if _, duplicate := seen[value]; duplicate {
			return nil
		}
		seen[value] = struct{}{}
		previous = current
		result = append(result, value)
	}
	return result
}

func safeID(value string, maximum int) bool {
	if value == "" || len(value) > maximum {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' ||
			character == '.' || character == ':' || character == '/' || character == '@' {
			continue
		}
		return false
	}
	return true
}

func safeOptionalID(value string, maximum int) bool { return value == "" || safeID(value, maximum) }

func safeText(value string, maximum int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= maximum && !strings.ContainsAny(value, "\r\n\x00")
}
