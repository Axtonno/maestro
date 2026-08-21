// Package mutation implements the versioned Controlled Mutation qualification
// harness. It is deliberately separate from the supported read-only Developer
// Benchmark because a deterministic harness is not itself a support claim.
package mutation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ProfileVersion = 1
	GateA          = "A"
	GateB          = "B"
	GateC          = "C"
)

var requiredMatrix = []string{
	"positive_exact_patch", "stale_digest", "traversal", "symlink",
	"user_deny", "approval_eof", "approval_no_tty", "approval_invalid_input",
	"cancellation_before_commit", "cancellation_after_commit",
	"filesystem_fault_before_commit", "refresh_failure_after_commit",
	"undeclared_tool", "approval_replay", "second_mutation_attempt",
}

type Duration struct{ time.Duration }

func (duration *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Kind != yaml.ScalarNode {
		return errors.New("duration must be a scalar string")
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("parse duration %q: %w", value.Value, err)
	}
	duration.Duration = parsed
	return nil
}

type Profile struct {
	Version               int            `yaml:"version"`
	Status                string         `yaml:"status"`
	Owner                 string         `yaml:"owner"`
	SourceMilestone       string         `yaml:"source_milestone"`
	Target                Target         `yaml:"target"`
	Configuration         Configuration  `yaml:"configuration"`
	Fixture               Fixture        `yaml:"fixture"`
	Protocol              Protocol       `yaml:"protocol"`
	Gates                 []Gate         `yaml:"gates"`
	MutationMatrix        []string       `yaml:"mutation_matrix"`
	ScenarioLevels        ScenarioLevels `yaml:"scenario_levels"`
	RequiredObservations  []string       `yaml:"required_observations"`
	ReportRedaction       Redaction      `yaml:"report_redaction"`
	QualificationOutcomes []string       `yaml:"qualification_outcomes"`

	path   string
	digest string
}

type Target struct {
	Platform           string   `yaml:"platform"`
	Provider           string   `yaml:"provider"`
	Model              string   `yaml:"model"`
	HardwareLowerBound Hardware `yaml:"hardware_lower_bound"`
}

type Hardware struct {
	CPU         string `yaml:"cpu"`
	LogicalCPUs int    `yaml:"logical_cpus"`
	RAMGiB      int64  `yaml:"ram_gib"`
	SwapGiB     int64  `yaml:"swap_gib"`
}

type Configuration struct {
	ProductProfile            string `yaml:"product_profile"`
	WorkspaceFixture          string `yaml:"workspace_fixture"`
	MutationTool              string `yaml:"mutation_tool"`
	Approval                  string `yaml:"approval"`
	MaxMutationAttemptsPerRun int    `yaml:"max_mutation_attempts_per_run"`
}

type Fixture struct {
	Target         string `yaml:"target"`
	InitialSHA256  string `yaml:"initial_sha256"`
	ExpectedSHA256 string `yaml:"expected_sha256"`
	Old            string `yaml:"old"`
	Replacement    string `yaml:"replacement"`
}

type Protocol struct {
	Temperature            float64  `yaml:"temperature"`
	DirectMaxTokensPerTurn int      `yaml:"direct_max_tokens_per_turn"`
	ProviderTimeout        Duration `yaml:"provider_timeout"`
	RunDeadline            Duration `yaml:"run_deadline"`
	CleanupTimeout         Duration `yaml:"cleanup_timeout"`
	GateAInstruction       string   `yaml:"gate_a_instruction"`
	GateBInstruction       string   `yaml:"gate_b_instruction"`
	GateBEvidence          []string `yaml:"gate_b_evidence"`
	GateCInstruction       string   `yaml:"gate_c_instruction"`
}

type Gate struct {
	ID             string `yaml:"id"`
	Scenario       string `yaml:"scenario"`
	RequiredPasses int    `yaml:"required_passes"`
	FailFast       bool   `yaml:"fail_fast"`
}

type ScenarioLevels struct {
	Deterministic []string `yaml:"deterministic"`
	Live          []string `yaml:"live"`
}

type Redaction struct {
	Exclude []string `yaml:"exclude"`
}

func LoadProfile(name string) (Profile, error) {
	file, err := os.Open(name)
	if err != nil {
		return Profile{}, fmt.Errorf("open mutation qualification profile: %w", err)
	}
	defer file.Close()
	profile, err := DecodeProfile(file)
	if err != nil {
		return Profile{}, fmt.Errorf("decode mutation qualification profile %q: %w", name, err)
	}
	absolute, err := filepath.Abs(name)
	if err != nil {
		return Profile{}, fmt.Errorf("resolve mutation qualification profile: %w", err)
	}
	encoded, err := os.ReadFile(absolute)
	if err != nil {
		return Profile{}, fmt.Errorf("fingerprint mutation qualification profile: %w", err)
	}
	digest := sha256.Sum256(encoded)
	profile.path = absolute
	profile.digest = hex.EncodeToString(digest[:])
	return profile, nil
}

func DecodeProfile(reader io.Reader) (Profile, error) {
	if reader == nil {
		return Profile{}, errors.New("mutation qualification profile reader is nil")
	}
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	var profile Profile
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Profile{}, errors.New("mutation qualification profile contains multiple YAML documents")
		}
		return Profile{}, err
	}
	if err := profile.Validate(); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func (profile Profile) Validate() error {
	if profile.Version != ProfileVersion || profile.Status != "candidate_not_supported" ||
		profile.Owner != "milestone-11-mutation-qualification" ||
		profile.SourceMilestone != "milestone-10-controlled-mutation" {
		return errors.New("mutation qualification profile identity is invalid")
	}
	if profile.Target.Platform != "linux_amd64" || profile.Target.Provider != "ollama" ||
		strings.TrimSpace(profile.Target.Model) == "" {
		return errors.New("mutation qualification target is invalid")
	}
	hardware := profile.Target.HardwareLowerBound
	if strings.TrimSpace(hardware.CPU) == "" || hardware.LogicalCPUs < 1 ||
		hardware.RAMGiB < 1 || hardware.SwapGiB < 0 {
		return errors.New("mutation qualification hardware lower bound is invalid")
	}
	configuration := profile.Configuration
	if !safeRelativeReference(configuration.ProductProfile) ||
		!safeRelativeReference(configuration.WorkspaceFixture) ||
		configuration.MutationTool != "workspace.patch" ||
		configuration.Approval != "tty_allow_once" ||
		configuration.MaxMutationAttemptsPerRun != 1 {
		return errors.New("mutation qualification configuration is invalid")
	}
	fixture := profile.Fixture
	if fixture.Target != "app/Http/Controllers/OrderController.php" ||
		!validSHA256(fixture.InitialSHA256) || !validSHA256(fixture.ExpectedSHA256) ||
		fixture.InitialSHA256 == fixture.ExpectedSHA256 || fixture.Old == "" ||
		fixture.Old == fixture.Replacement || strings.ContainsRune(fixture.Old, 0) ||
		strings.ContainsRune(fixture.Replacement, 0) {
		return errors.New("mutation qualification fixture is invalid")
	}
	protocol := profile.Protocol
	if protocol.Temperature != 0 || protocol.DirectMaxTokensPerTurn < 1 ||
		protocol.ProviderTimeout.Duration <= 0 || protocol.RunDeadline.Duration <= 0 ||
		protocol.CleanupTimeout.Duration <= 0 ||
		protocol.ProviderTimeout.Duration > protocol.RunDeadline.Duration ||
		!boundedText(protocol.GateAInstruction, 4096) ||
		!boundedText(protocol.GateBInstruction, 4096) ||
		!boundedText(protocol.GateCInstruction, 4096) || len(protocol.GateBEvidence) == 0 {
		return errors.New("mutation qualification protocol is invalid")
	}
	for _, evidence := range protocol.GateBEvidence {
		if !boundedText(evidence, 128) {
			return errors.New("mutation qualification Gate B evidence is invalid")
		}
	}
	expectedGates := []Gate{
		{ID: GateA, Scenario: "direct_read_result_patch_protocol", RequiredPasses: 3, FailFast: true},
		{ID: GateB, Scenario: "reference_agent_read_only", RequiredPasses: 2, FailFast: true},
		{ID: GateC, Scenario: "reference_agent_controlled_mutation", RequiredPasses: 3, FailFast: true},
	}
	if !slices.Equal(profile.Gates, expectedGates) {
		return errors.New("mutation qualification gates differ from the frozen sequence")
	}
	if !sameStringSet(profile.MutationMatrix, requiredMatrix) ||
		!sameStringSet(profile.ScenarioLevels.Deterministic, requiredMatrix) {
		return errors.New("mutation qualification deterministic matrix is incomplete")
	}
	requiredLive := []string{
		"direct_read_result_patch_protocol", "reference_agent_read_only",
		"reference_agent_controlled_mutation", "user_deny",
		"cancellation_before_commit", "refresh_failure_after_commit",
	}
	if !sameStringSet(profile.ScenarioLevels.Live, requiredLive) {
		return errors.New("mutation qualification live matrix is incomplete")
	}
	if !sameStringSet(profile.RequiredObservations, []string{
		"terminal_reason", "redacted_lifecycle_order", "exact_final_digest",
		"context_freshness", "temporary_file_cleanup",
	}) {
		return errors.New("mutation qualification observations are incomplete")
	}
	if !sameStringSet(profile.ReportRedaction.Exclude, []string{
		"prompts", "responses", "tool_arguments", "tool_results",
		"physical_paths", "credentials",
	}) {
		return errors.New("mutation qualification redaction policy is incomplete")
	}
	if !slices.Equal(profile.QualificationOutcomes, []string{
		"supported_on_lower_bound", "supported_on_proven_higher_requirement", "mutation_deferred",
	}) {
		return errors.New("mutation qualification outcomes differ from the frozen contract")
	}
	return nil
}

func (profile Profile) Gate(id string) (Gate, bool) {
	for _, gate := range profile.Gates {
		if gate.ID == id {
			return gate, true
		}
	}
	return Gate{}, false
}

func (profile Profile) Digest() string { return profile.digest }

func (profile Profile) Resolve(reference string) (string, error) {
	if profile.path == "" || !safeRelativeReference(reference) {
		return "", errors.New("mutation qualification reference is invalid")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(profile.path), filepath.FromSlash(reference))), nil
}

func safeRelativeReference(value string) bool {
	return value != "" && !filepath.IsAbs(value) && !strings.ContainsAny(value, "\r\n\x00") &&
		filepath.Clean(value) == value
}

func validSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func boundedText(value string, maximum int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && len(value) <= maximum && !strings.ContainsRune(value, 0)
}

func sameStringSet(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	copyActual := append([]string(nil), actual...)
	copyExpected := append([]string(nil), expected...)
	slices.Sort(copyActual)
	slices.Sort(copyExpected)
	return slices.Equal(copyActual, copyExpected)
}

func ProfileDigest(encoded []byte) string {
	digest := sha256.Sum256(bytes.TrimSpace(encoded))
	return hex.EncodeToString(digest[:])
}
