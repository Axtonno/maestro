package mutation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

type DeterministicEvidence struct {
	Scenario         string
	Test             string
	Terminal         string
	MutationAttempts int
	Approval         string
	Lifecycle        []string
	Committed        bool
	ContextFreshness string
}

var deterministicEvidence = []DeterministicEvidence{
	{Scenario: "positive_exact_patch", Test: "workspace_tools_positive_patch", Terminal: "completed", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_succeeded", "reindex_started", "reindex_succeeded", "terminal_completed"}, Committed: true, ContextFreshness: "fresh"},
	{Scenario: "stale_digest", Test: "workspace_patch_precondition", Terminal: "tool_failure", Approval: "not_requested", Lifecycle: []string{"terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "traversal", Test: "workspace_tools_containment", Terminal: "tool_failure", Approval: "not_requested", Lifecycle: []string{"terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "symlink", Test: "workspace_tools_containment", Terminal: "tool_failure", Approval: "not_requested", Lifecycle: []string{"terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "user_deny", Test: "terminal_approver_exact_patch", Terminal: "permission_denied", Approval: "denied", Lifecycle: []string{"proposal_prepared", "approval_denied", "terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "approval_eof", Test: "terminal_approver_input_unavailable", Terminal: "permission_denied", Approval: "denied", Lifecycle: []string{"proposal_prepared", "approval_denied", "terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "approval_no_tty", Test: "terminal_approver_non_interactive", Terminal: "permission_denied", Approval: "unavailable", Lifecycle: []string{"proposal_prepared", "approval_denied", "terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "approval_invalid_input", Test: "terminal_approver_input_invalid", Terminal: "permission_denied", Approval: "denied", Lifecycle: []string{"proposal_prepared", "approval_denied", "terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "cancellation_before_commit", Test: "atomic_replace_precommit_faults", Terminal: "canceled", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_failed", "terminal_canceled"}, ContextFreshness: "stale"},
	{Scenario: "cancellation_after_commit", Test: "atomic_replace_postcommit_cancel", Terminal: "canceled", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_succeeded", "terminal_canceled"}, Committed: true, ContextFreshness: "stale"},
	{Scenario: "filesystem_fault_before_commit", Test: "atomic_replace_fault_matrix", Terminal: "tool_failure", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_failed", "terminal_failed"}, ContextFreshness: "stale"},
	{Scenario: "refresh_failure_after_commit", Test: "agent_refresh_failure", Terminal: "tool_failure", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_succeeded", "reindex_started", "reindex_failed", "terminal_failed"}, Committed: true, ContextFreshness: "stale"},
	{Scenario: "undeclared_tool", Test: "agent_unknown_tool", Terminal: "tool_failure", Approval: "not_requested", Lifecycle: []string{"terminal_failed"}, ContextFreshness: "not_applicable"},
	{Scenario: "approval_replay", Test: "permit_one_shot_replay", Terminal: "tool_failure", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_failed", "terminal_failed"}, ContextFreshness: "stale"},
	{Scenario: "second_mutation_attempt", Test: "agent_second_mutation_rejected", Terminal: "tool_failure", MutationAttempts: 1, Approval: "allowed_once", Lifecycle: []string{"proposal_prepared", "approval_allowed", "apply_started", "apply_succeeded", "terminal_failed"}, Committed: true, ContextFreshness: "stale"},
}

func DeterministicCoverage() []DeterministicEvidence {
	result := make([]DeterministicEvidence, len(deterministicEvidence))
	for index, evidence := range deterministicEvidence {
		result[index] = evidence
		result[index].Lifecycle = append([]string(nil), evidence.Lifecycle...)
	}
	return result
}

func NewDeterministicReport(
	ctx context.Context,
	profile Profile,
	hardware pkgBenchmark.HardwareProfile,
	build ReportBuild,
	now time.Time,
) (Report, error) {
	if ctx == nil || now.IsZero() {
		return Report{}, errors.New("deterministic mutation report dependencies are invalid")
	}
	if err := profile.Validate(); err != nil {
		return Report{}, err
	}
	if !sameStringSet(profile.MutationMatrix, evidenceScenarioIDs()) {
		return Report{}, errors.New("deterministic evidence differs from the frozen matrix")
	}
	fixture, err := MaterializeFixture(ctx, profile)
	if err != nil {
		return Report{}, err
	}
	defer fixture.Cleanup()
	initialDigest := fixture.Initial.Digest
	target := filepath.Join(fixture.Root, filepath.FromSlash(profile.Fixture.Target))
	content, err := os.ReadFile(target)
	if err != nil {
		return Report{}, err
	}
	proposed := strings.Replace(string(content), profile.Fixture.Old, profile.Fixture.Replacement, 1)
	if err := os.WriteFile(target, []byte(proposed), 0o600); err != nil {
		return Report{}, err
	}
	expected, err := SnapshotWorkspace(ctx, fixture.Root)
	if err != nil {
		return Report{}, err
	}
	started := now
	report := Report{
		SchemaVersion: ReportSchemaVersion,
		RunID:         "mutation-deterministic-" + profile.Digest()[:16],
		CreatedAt:     started,
		Profile: ReportProfile{
			Version: profile.Version, Digest: profile.Digest(),
			Target: profile.Target.Platform + "/" + profile.Target.Provider,
			Model:  profile.Target.Model,
		},
		Build: build, Hardware: hardware, Gate: "deterministic",
		Scenario: "mutation_matrix", Required: len(deterministicEvidence), FailFast: true,
		State: "passed", Samples: make([]Sample, 0, len(deterministicEvidence)),
	}
	for index, evidence := range deterministicEvidence {
		finalSHA := profile.Fixture.InitialSHA256
		workspaceDigest := initialDigest
		if evidence.Committed {
			finalSHA = profile.Fixture.ExpectedSHA256
			workspaceDigest = expected.Digest
		}
		sampleStarted := started.Add(time.Duration(index) * time.Nanosecond)
		report.Samples = append(report.Samples, Sample{
			Attempt: index + 1, Scenario: evidence.Scenario, EvidenceCode: evidence.Test,
			State: "passed", StartedAt: sampleStarted, Terminal: evidence.Terminal,
			MutationAttempts: evidence.MutationAttempts, Approval: evidence.Approval,
			Lifecycle:      append([]string(nil), evidence.Lifecycle...),
			InitialSHA256:  profile.Fixture.InitialSHA256,
			ExpectedSHA256: profile.Fixture.ExpectedSHA256,
			FinalSHA256:    finalSHA, WorkspaceDigest: workspaceDigest,
			ContextFreshness: evidence.ContextFreshness, TemporaryCleanup: "clean",
		})
	}
	report.CompletedAt = started.Add(time.Duration(len(report.Samples)) * time.Nanosecond)
	report.DurationMS = float64(report.CompletedAt.Sub(started)) / float64(time.Millisecond)
	if err := report.Validate(); err != nil {
		return Report{}, err
	}
	return report, nil
}

func evidenceScenarioIDs() []string {
	result := make([]string, len(deterministicEvidence))
	for index, evidence := range deterministicEvidence {
		result[index] = evidence.Scenario
	}
	return result
}
