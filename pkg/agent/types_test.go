package agent_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	agent "github.com/antonio-cafeo/maestro/pkg/agent"
	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/runtime"
	"github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestPublicContractAssertions(t *testing.T) {
	var _ agent.Agent = agentStub{}
	var _ agent.PlanningAgent = planningAgentStub{}
	var _ agent.Runtime = runtimeStub{}
}

func TestAgentValidationRejectsTypedNil(t *testing.T) {
	var candidate *agentStub
	if !errors.Is(agent.ValidateAgent(candidate), agent.ErrInvalidAgent) {
		t.Fatalf("expected typed nil agent rejection, got %v", agent.ValidateAgent(candidate))
	}
}

func TestDescriptorValidationAndCopies(t *testing.T) {
	capabilities := []runtime.Capability{agent.CapabilityWorkspaceAware, agent.CapabilityRun}
	descriptor, err := agent.NewDescriptor("agent.general", "1.0.0", "General workspace agent.", capabilities)
	if err != nil {
		t.Fatalf("construct descriptor: %v", err)
	}
	capabilities[0] = "agent.changed"
	if got := descriptor.Capabilities(); !reflect.DeepEqual(got, []runtime.Capability{agent.CapabilityRun, agent.CapabilityWorkspaceAware}) {
		t.Fatalf("descriptor capabilities are not sorted/defensive: %#v", got)
	}
	returned := descriptor.Capabilities()
	returned[0] = "agent.changed"
	if descriptor.Capabilities()[0] != agent.CapabilityRun {
		t.Fatal("descriptor exposes capability storage")
	}
	if _, err := agent.NewDescriptor("general", "1", "description", []runtime.Capability{agent.CapabilityRun}); !errors.Is(err, agent.ErrInvalidAgentID) {
		t.Fatalf("expected invalid ID, got %v", err)
	}
	if _, err := agent.NewDescriptor("agent.general", "1", "description", []runtime.Capability{"run"}); !errors.Is(err, agent.ErrInvalidDescriptor) {
		t.Fatalf("expected non-namespaced capability rejection, got %v", err)
	}
}

func TestRunRequestRequiresExplicitTargetsLimitsContextAndTools(t *testing.T) {
	build := buildRequest(t, "workspace")
	toolIDs := []tool.ID{"workspace.write", "workspace.read"}
	request, err := agent.NewRunRequest(
		"run-1", "agent.general", "ollama", "model", "workspace", "policy.default-deny",
		"Update the workspace.", validLimits(),
		agent.RunRequestOptions{Context: build, Tools: toolIDs},
	)
	if err != nil {
		t.Fatalf("construct run request: %v", err)
	}
	toolIDs[0] = "workspace.changed"
	if got := request.Tools(); !reflect.DeepEqual(got, []tool.ID{"workspace.read", "workspace.write"}) {
		t.Fatalf("tool IDs are not ordered/defensive: %#v", got)
	}
	if request.Provider() != "ollama" || request.Model() != "model" || request.Workspace() != "workspace" {
		t.Fatalf("explicit targets were not preserved: %#v", request)
	}
	workspace, err := contextengine.NewWorkspace("workspace", t.TempDir(), contextengine.WorkspaceOptions{Source: contextengine.SourceFilesystem, Policy: contextengine.DefaultScanPolicy()})
	if err != nil {
		t.Fatal(err)
	}
	withTarget, err := agent.NewRunRequest(
		"run-target", "agent.general", "ollama", "model", "workspace", "policy.default-deny",
		"task", validLimits(), agent.RunRequestOptions{Context: build, Tools: []tool.ID{"workspace.read"}, Workspace: &workspace},
	)
	if target, ok := withTarget.WorkspaceTarget(); err != nil || !ok || target.Root() != workspace.Root() {
		t.Fatalf("workspace target was not preserved: target=%#v ok=%v err=%v", target, ok, err)
	}
	otherWorkspace, _ := contextengine.NewWorkspace("other", t.TempDir(), contextengine.WorkspaceOptions{Source: contextengine.SourceFilesystem, Policy: contextengine.DefaultScanPolicy()})
	if _, err := agent.NewRunRequest(
		"run-target", "agent.general", "ollama", "model", "workspace", "policy.default-deny",
		"task", validLimits(), agent.RunRequestOptions{Context: build, Tools: []tool.ID{"workspace.read"}, Workspace: &otherWorkspace},
	); !errors.Is(err, agent.ErrInvalidRequest) {
		t.Fatalf("expected mismatched workspace target rejection, got %v", err)
	}
	if _, err := agent.NewRunRequest(
		"run-1", "agent.general", "", "model", "workspace", "policy.default-deny",
		"task", validLimits(), agent.RunRequestOptions{Context: build, Tools: []tool.ID{"workspace.read"}},
	); !errors.Is(err, agent.ErrInvalidRequest) {
		t.Fatalf("expected empty provider rejection, got %v", err)
	}
	other := buildRequest(t, "other")
	if _, err := agent.NewRunRequest(
		"run-1", "agent.general", "ollama", "model", "workspace", "policy.default-deny",
		"task", validLimits(), agent.RunRequestOptions{Context: other, Tools: []tool.ID{"workspace.read"}},
	); !errors.Is(err, agent.ErrInvalidRequest) {
		t.Fatalf("expected mismatched workspace rejection, got %v", err)
	}
	var typedNil *approverStub
	if _, err := agent.NewRunRequest(
		"run-1", "agent.general", "ollama", "model", "workspace", "policy.default-deny",
		"task", validLimits(), agent.RunRequestOptions{Context: build, Tools: []tool.ID{"workspace.read"}, Approver: typedNil},
	); !errors.Is(err, agent.ErrInvalidRequest) {
		t.Fatalf("expected typed nil approver rejection, got %v", err)
	}
}

func TestLimitsArePositiveAndCoherent(t *testing.T) {
	limits := validLimits()
	if err := limits.Validate(); err != nil {
		t.Fatalf("valid limits: %v", err)
	}
	limits.MaxToolCallsPerTurn = limits.MaxToolCalls + 1
	if !errors.Is(limits.Validate(), agent.ErrInvalidLimits) {
		t.Fatalf("expected incoherent tool limit rejection, got %v", limits.Validate())
	}
	limits = validLimits()
	limits.MaxSessionBytes = limits.MaxToolResultBytes - 1
	if !errors.Is(limits.Validate(), agent.ErrInvalidLimits) {
		t.Fatalf("expected byte limit rejection, got %v", limits.Validate())
	}
}

func TestPlanValidationTransitionsAndDefensiveCopies(t *testing.T) {
	first := planStep(t, "inspect", nil, agent.StepPending, "")
	second := planStep(t, "modify", []agent.StepID{"inspect"}, agent.StepPending, "")
	steps := []agent.PlanStep{first, second}
	plan, err := agent.NewPlan(1, steps)
	if err != nil {
		t.Fatalf("construct plan: %v", err)
	}
	steps[0] = second
	returned := plan.Steps()
	returned[0] = second
	if plan.Steps()[0].ID() != "inspect" {
		t.Fatal("plan exposes step storage")
	}
	if !agent.CanTransitionStep(agent.StepPending, agent.StepRunning) || agent.CanTransitionStep(agent.StepCompleted, agent.StepRunning) {
		t.Fatal("unexpected step transition semantics")
	}
	left := planStep(t, "left", []agent.StepID{"right"}, agent.StepPending, "")
	right := planStep(t, "right", []agent.StepID{"left"}, agent.StepPending, "")
	if _, err := agent.NewPlan(1, []agent.PlanStep{left, right}); !errors.Is(err, agent.ErrInvalidPlan) {
		t.Fatalf("expected cycle rejection, got %v", err)
	}
	if _, err := agent.NewPlanStep("step", "objective", nil, agent.StepFailed, ""); !errors.Is(err, agent.ErrInvalidStep) {
		t.Fatalf("expected missing terminal reason rejection, got %v", err)
	}
	if _, err := plan.TransitionStep("modify", agent.StepRunning, ""); !errors.Is(err, agent.ErrInvalidTransition) {
		t.Fatalf("expected incomplete dependency rejection, got %v", err)
	}
	plan, err = plan.TransitionStep("inspect", agent.StepRunning, "")
	if err != nil {
		t.Fatal(err)
	}
	plan, err = plan.TransitionStep("inspect", agent.StepCompleted, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.TransitionStep("modify", agent.StepRunning, ""); err != nil {
		t.Fatalf("completed dependency did not unlock step: %v", err)
	}
}

func TestTerminalPrecedenceAndSessionInvariants(t *testing.T) {
	selected, err := agent.SelectTerminal(
		agent.TerminalCompleted,
		agent.TerminalToolFailure,
		agent.TerminalCanceled,
		agent.TerminalDeadline,
	)
	if err != nil || selected != agent.TerminalDeadline {
		t.Fatalf("unexpected terminal selection: %q %v", selected, err)
	}
	if !agent.CanTransitionSession(agent.SessionPlanning, agent.SessionRunning) ||
		agent.CanTransitionSession(agent.SessionTerminal, agent.SessionRunning) {
		t.Fatal("unexpected session transition semantics")
	}
	pendingPlan, _ := agent.NewPlan(1, []agent.PlanStep{planStep(t, "inspect", nil, agent.StepPending, "")})
	_, err = agent.NewSessionSnapshot("run-1", "agent.general", "workspace", agent.SessionSnapshotOptions{
		Generation: 1, State: agent.SessionTerminal, Plan: &pendingPlan, Terminal: agent.TerminalCompleted,
	})
	if !errors.Is(err, agent.ErrInvalidSession) {
		t.Fatalf("expected incomplete plan rejection, got %v", err)
	}
	completedPlan, _ := agent.NewPlan(1, []agent.PlanStep{planStep(t, "inspect", nil, agent.StepCompleted, "")})
	snapshot, err := agent.NewSessionSnapshot("run-1", "agent.general", "workspace", agent.SessionSnapshotOptions{
		Generation: 3, State: agent.SessionTerminal, WorkspaceGeneration: 2,
		Plan: &completedPlan, Counters: agent.Counters{ModelTurns: 2, ToolCalls: 1},
		Terminal: agent.TerminalCompleted,
	})
	if err != nil {
		t.Fatalf("construct terminal session: %v", err)
	}
	if got, ok := snapshot.Plan(); !ok || got.Version() != 1 || snapshot.Generation() != 3 {
		t.Fatalf("unexpected session snapshot: %#v %v", got, ok)
	}
	result, err := agent.NewRunResult(snapshot, "Completed.")
	if err != nil || result.Content() != "Completed." {
		t.Fatalf("construct result: %#v %v", result, err)
	}
}

func TestRunErrorPreservesKindAndCause(t *testing.T) {
	cause := errors.New("cause")
	runErr := agent.NewRunError(agent.ErrorPermission, "run-1", "agent.general", "permission_denied", cause)
	if !errors.Is(runErr, agent.ErrPermissionDenied) || !errors.Is(runErr, cause) {
		t.Fatalf("run error does not preserve kind/cause: %v", runErr)
	}
}

func validLimits() agent.Limits {
	return agent.Limits{
		MaxDuration: time.Minute, MaxModelTurns: 10, MaxToolCalls: 20,
		MaxToolCallsPerTurn: 4, MaxPlanSteps: 10, MaxPlanRevisions: 2,
		MaxToolResultBytes: 1 << 20, MaxSessionBytes: 8 << 20,
		MaxInputTokens: 10_000, MaxOutputTokens: 2_000,
	}
}

func buildRequest(t *testing.T, workspace contextengine.WorkspaceID) contextengine.BuildRequest {
	t.Helper()
	query, err := contextengine.NewRetrievalQuery(workspace, "task", contextengine.RetrievalQueryOptions{
		Methods: []contextengine.RetrievalMethod{contextengine.RetrievalLexical}, TopK: 5,
	})
	if err != nil {
		t.Fatalf("construct retrieval query: %v", err)
	}
	return contextengine.BuildRequest{
		Query:     query,
		Budget:    contextengine.Budget{MaxTokens: 1000, ReservedTokens: 100, SafetyTokens: 100},
		Estimator: "context.test-estimator",
	}
}

func planStep(t *testing.T, id agent.StepID, dependencies []agent.StepID, status agent.StepStatus, reason string) agent.PlanStep {
	t.Helper()
	step, err := agent.NewPlanStep(id, "Objective for "+string(id), dependencies, status, reason)
	if err != nil {
		t.Fatalf("construct step %q: %v", id, err)
	}
	return step
}

type agentStub struct{}

func (agentStub) Descriptor() agent.Descriptor { return agent.Descriptor{} }

type planningAgentStub struct{ agentStub }

func (planningAgentStub) Plan(context.Context, agent.PlanningRequest) (agent.Plan, error) {
	return agent.Plan{}, nil
}

type runtimeStub struct{}

func (runtimeStub) Register(agent.Agent) error      { return nil }
func (runtimeStub) Descriptors() []agent.Descriptor { return nil }
func (runtimeStub) Run(context.Context, agent.RunRequest) (agent.RunResult, error) {
	return agent.RunResult{}, nil
}
func (runtimeStub) Session(agent.RunID) (agent.SessionSnapshot, bool) {
	return agent.SessionSnapshot{}, false
}

type approverStub struct{}

func (*approverStub) Approve(context.Context, tool.PermissionRequest) (tool.Approval, error) {
	return tool.Approval{}, nil
}
