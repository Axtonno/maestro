package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	"github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestRuntimeRegistersAgentsAndOrdersDescriptors(t *testing.T) {
	runtime := NewRuntime(&contextStub{bundle: testBundle(t, "workspace")})
	second := newAgentFixture(t, "agent.second", pendingPlan(t, "second"))
	first := newAgentFixture(t, "agent.first", pendingPlan(t, "first"))
	if err := runtime.Register(second); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Register(first); err != nil {
		t.Fatal(err)
	}
	descriptors := runtime.Descriptors()
	if len(descriptors) != 2 || descriptors[0].ID() != "agent.first" || descriptors[1].ID() != "agent.second" {
		t.Fatalf("unexpected descriptors: %#v", descriptors)
	}
	if err := runtime.Register(first); !errors.Is(err, pkgAgent.ErrAlreadyRegistered) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}

func TestRunBuildsPlanPublishesSnapshotsAndStopsAtPhaseBoundary(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	candidate := newAgentFixture(t, "agent.general", pendingPlan(t, "inspect", "modify"))
	_ = runtime.Register(candidate)
	result, err := runtime.Run(context.Background(), runRequest(t, "run-1", "agent.general", "workspace", 10))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	snapshot := result.Session()
	if snapshot.State() != pkgAgent.SessionTerminal || snapshot.Terminal() != pkgAgent.TerminalBlocked || snapshot.WorkspaceGeneration() != bundle.Generation() {
		t.Fatalf("unexpected terminal snapshot: state=%q terminal=%q workspace_generation=%d", snapshot.State(), snapshot.Terminal(), snapshot.WorkspaceGeneration())
	}
	plan, ok := snapshot.Plan()
	if !ok || len(plan.Steps()) != 2 || snapshot.Generation() < 4 {
		t.Fatalf("plan/snapshot was not published: %#v generation=%d", plan, snapshot.Generation())
	}
	stored, ok := runtime.Session("run-1")
	if !ok || stored.Generation() != snapshot.Generation() {
		t.Fatalf("session lookup mismatch: %#v %v", stored, ok)
	}
	if candidate.calls.Load() != 1 {
		t.Fatalf("planner calls=%d", candidate.calls.Load())
	}
}

func TestOneCoordinatorAndNoRunReuse(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	entered := make(chan struct{})
	release := make(chan struct{})
	candidate := newAgentFixture(t, "agent.general", pendingPlan(t, "inspect"))
	candidate.plan = func(ctx context.Context, _ pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
		close(entered)
		select {
		case <-release:
			return pendingPlan(t, "inspect"), nil
		case <-ctx.Done():
			return pkgAgent.Plan{}, ctx.Err()
		}
	}
	_ = runtime.Register(candidate)
	request := runRequest(t, "run-1", "agent.general", "workspace", 10)
	done := make(chan error, 1)
	go func() {
		_, err := runtime.Run(context.Background(), request)
		done <- err
	}()
	<-entered
	if _, err := runtime.Run(context.Background(), request); !errors.Is(err, pkgAgent.ErrSessionActive) {
		t.Fatalf("expected active coordinator rejection, got %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Run(context.Background(), request); !errors.Is(err, pkgAgent.ErrSessionTerminal) {
		t.Fatalf("expected terminal reuse rejection, got %v", err)
	}
}

func TestRegistryIsBoundedWithoutEvictingTerminalRunIDs(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime, _ := NewRuntimeWithOptions(&contextStub{bundle: bundle}, Options{MaxSessions: 1})
	_ = runtime.Register(newAgentFixture(t, "agent.general", pendingPlan(t, "inspect")))
	if _, err := runtime.Run(context.Background(), runRequest(t, "run-1", "agent.general", "workspace", 10)); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Run(context.Background(), runRequest(t, "run-2", "agent.general", "workspace", 10)); !errors.Is(err, pkgAgent.ErrLimitExceeded) {
		t.Fatalf("expected registry limit, got %v", err)
	}
}

func TestPlannerRunsOutsideRegistryLockAndPanicIsContained(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	entered := make(chan struct{})
	release := make(chan struct{})
	blocking := newAgentFixture(t, "agent.blocking", pendingPlan(t, "inspect"))
	blocking.plan = func(ctx context.Context, _ pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
		close(entered)
		select {
		case <-release:
			return pendingPlan(t, "inspect"), nil
		case <-ctx.Done():
			return pkgAgent.Plan{}, ctx.Err()
		}
	}
	_ = runtime.Register(blocking)
	done := make(chan error, 1)
	go func() {
		_, err := runtime.Run(context.Background(), runRequest(t, "run-blocking", "agent.blocking", "workspace", 10))
		done <- err
	}()
	<-entered
	registered := make(chan error, 1)
	go func() { registered <- runtime.Register(newAgentFixture(t, "agent.other", pendingPlan(t, "other"))) }()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("agent registry lock held during planner callback")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}

	panicking := newAgentFixture(t, "agent.panic", pendingPlan(t, "inspect"))
	panicking.plan = func(context.Context, pkgAgent.PlanningRequest) (pkgAgent.Plan, error) { panic("boom") }
	_ = runtime.Register(panicking)
	_, err := runtime.Run(context.Background(), runRequest(t, "run-panic", "agent.panic", "workspace", 10))
	if !errors.Is(err, pkgAgent.ErrPlanningFailed) {
		t.Fatalf("expected planner panic containment, got %v", err)
	}
	snapshot, _ := runtime.Session("run-panic")
	if snapshot.Terminal() != pkgAgent.TerminalPlanningFailure {
		t.Fatalf("unexpected terminal after panic: %q", snapshot.Terminal())
	}
}

func TestPlanAndSessionBudgetsAreEnforced(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	_ = runtime.Register(newAgentFixture(t, "agent.general", pendingPlan(t, "one", "two")))
	_, err := runtime.Run(context.Background(), runRequest(t, "run-steps", "agent.general", "workspace", 1))
	if !errors.Is(err, pkgAgent.ErrLimitExceeded) {
		t.Fatalf("expected plan step limit, got %v", err)
	}
	snapshot, _ := runtime.Session("run-steps")
	if snapshot.Terminal() != pkgAgent.TerminalLimit {
		t.Fatalf("unexpected plan limit terminal: %q", snapshot.Terminal())
	}
}

func TestSessionRevisionsAreBoundedAndPreserveHistory(t *testing.T) {
	request := runRequest(t, "run-revisions", "agent.general", "workspace", 10)
	current, err := newSession(request)
	if err != nil {
		t.Fatal(err)
	}
	initial := pendingPlan(t, "inspect", "modify")
	if err := current.transition(pkgAgent.SessionPlanning, nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := current.transition(pkgAgent.SessionRunning, &initial, 1); err != nil {
		t.Fatal(err)
	}
	for version := uint64(2); version <= 3; version++ {
		steps := initial.Steps()
		revised, planErr := pkgAgent.NewPlan(version, steps)
		if planErr != nil {
			t.Fatal(planErr)
		}
		if err := current.revise(revised); err != nil {
			t.Fatalf("revision %d: %v", version, err)
		}
	}
	overLimit, _ := pkgAgent.NewPlan(4, initial.Steps())
	if err := current.revise(overLimit); !errors.Is(err, pkgAgent.ErrLimitExceeded) {
		t.Fatalf("expected revision limit, got %v", err)
	}
	history := current.planHistory()
	if len(history) != 3 || history[0].Version() != 1 || history[2].Version() != 3 {
		t.Fatalf("unexpected bounded history: %#v", history)
	}
}

func TestDeadlineDuringPlanningCommitsDeadlineTerminal(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	candidate := newAgentFixture(t, "agent.slow", pendingPlan(t, "inspect"))
	candidate.plan = func(ctx context.Context, _ pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
		<-ctx.Done()
		return pkgAgent.Plan{}, ctx.Err()
	}
	_ = runtime.Register(candidate)
	request := runRequest(t, "run-deadline", "agent.slow", "workspace", 10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := runtime.Run(ctx, request)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
	snapshot, ok := runtime.Session("run-deadline")
	if !ok || snapshot.Terminal() != pkgAgent.TerminalDeadline {
		t.Fatalf("unexpected deadline snapshot: %#v", snapshot)
	}
}

func TestTerminalSelectionIsCommittedOnce(t *testing.T) {
	request := runRequest(t, "run-1", "agent.general", "workspace", 10)
	current, err := newSession(request)
	if err != nil {
		t.Fatal(err)
	}
	result, err := current.finish("", pkgAgent.TerminalCompleted, pkgAgent.TerminalCanceled, pkgAgent.TerminalDeadline)
	if err != nil || result.Session().Terminal() != pkgAgent.TerminalDeadline {
		t.Fatalf("unexpected selected terminal: %#v %v", result, err)
	}
	if _, err := current.finish("", pkgAgent.TerminalCompleted); !errors.Is(err, pkgAgent.ErrSessionTerminal) {
		t.Fatalf("expected immutable terminal, got %v", err)
	}
}

func TestConcurrentSessionSnapshotsAreComplete(t *testing.T) {
	bundle := testBundle(t, "workspace")
	runtime := NewRuntime(&contextStub{bundle: bundle})
	_ = runtime.Register(newAgentFixture(t, "agent.general", pendingPlan(t, "inspect")))
	request := runRequest(t, "run-1", "agent.general", "workspace", 10)
	var wait sync.WaitGroup
	for range 20 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				if snapshot, ok := runtime.Session("run-1"); ok && snapshot.Validate() != nil {
					t.Errorf("observed invalid snapshot: %#v", snapshot)
				}
			}
		}()
	}
	_, err := runtime.Run(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	wait.Wait()
}

type agentFixture struct {
	descriptor pkgAgent.Descriptor
	plan       func(context.Context, pkgAgent.PlanningRequest) (pkgAgent.Plan, error)
	calls      atomic.Int32
}

func newAgentFixture(t *testing.T, id pkgAgent.ID, plan pkgAgent.Plan) *agentFixture {
	t.Helper()
	descriptor, err := pkgAgent.NewDescriptor(id, "1", "Fixture planning agent.", []pkgRuntime.Capability{pkgAgent.CapabilityRun, pkgAgent.CapabilityPlanning})
	if err != nil {
		t.Fatal(err)
	}
	return &agentFixture{descriptor: descriptor, plan: func(context.Context, pkgAgent.PlanningRequest) (pkgAgent.Plan, error) { return plan, nil }}
}

func (candidate *agentFixture) Descriptor() pkgAgent.Descriptor { return candidate.descriptor }
func (candidate *agentFixture) Plan(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
	candidate.calls.Add(1)
	return candidate.plan(ctx, request)
}

type contextStub struct {
	bundle pkgContext.ContextBundle
	err    error
}

func (*contextStub) RegisterSource(pkgContext.Source) error            { return nil }
func (*contextStub) RegisterAnalyzer(pkgContext.Analyzer) error        { return nil }
func (*contextStub) RegisterEstimator(pkgContext.TokenEstimator) error { return nil }
func (*contextStub) Index(context.Context, pkgContext.Workspace) (pkgContext.Snapshot, error) {
	return pkgContext.Snapshot{}, nil
}
func (*contextStub) Snapshot(pkgContext.WorkspaceID) (pkgContext.Snapshot, bool) {
	return pkgContext.Snapshot{}, false
}
func (*contextStub) Retrieve(context.Context, pkgContext.RetrievalQuery) ([]pkgContext.RetrievalResult, error) {
	return nil, nil
}
func (stub *contextStub) Build(context.Context, pkgContext.BuildRequest) (pkgContext.ContextBundle, error) {
	return stub.bundle, stub.err
}
func (*contextStub) CacheStats() pkgContext.CacheStats { return pkgContext.CacheStats{} }

func pendingPlan(t *testing.T, ids ...pkgAgent.StepID) pkgAgent.Plan {
	t.Helper()
	steps := make([]pkgAgent.PlanStep, 0, len(ids))
	for _, id := range ids {
		step, err := pkgAgent.NewPlanStep(id, "Objective "+string(id), nil, pkgAgent.StepPending, "")
		if err != nil {
			t.Fatal(err)
		}
		steps = append(steps, step)
	}
	plan, err := pkgAgent.NewPlan(1, steps)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func runRequest(t *testing.T, run pkgAgent.RunID, agentID pkgAgent.ID, workspace pkgContext.WorkspaceID, maxSteps int) pkgAgent.RunRequest {
	t.Helper()
	query, _ := pkgContext.NewRetrievalQuery(workspace, "task", pkgContext.RetrievalQueryOptions{Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1})
	request, err := pkgAgent.NewRunRequest(
		run, agentID, "ollama", "model", workspace, "policy.test", "Inspect the workspace.",
		pkgAgent.Limits{
			MaxDuration: time.Second, MaxModelTurns: 5, MaxToolCalls: 5, MaxToolCallsPerTurn: 2,
			MaxPlanSteps: maxSteps, MaxPlanRevisions: 2, MaxToolResultBytes: 1024,
			MaxSessionBytes: 1 << 20, MaxInputTokens: 10_000, MaxOutputTokens: 10_000,
		},
		pkgAgent.RunRequestOptions{
			Context: pkgContext.BuildRequest{Query: query, Budget: pkgContext.Budget{MaxTokens: 100, ReservedTokens: 10, SafetyTokens: 10}, Estimator: "context.test"},
			Tools:   []tool.ID{"workspace.read"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func testBundle(t *testing.T, workspaceID pkgContext.WorkspaceID) pkgContext.ContextBundle {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace(workspaceID, t.TempDir(), pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	if err != nil {
		t.Fatal(err)
	}
	document, _ := pkgContext.NewDocument("main.go", "text/x-go", "go", "package main\n")
	snapshot, _ := pkgContext.NewSnapshot(workspace, 1, []pkgContext.Document{document}, nil, nil)
	return mustBundle(t, snapshot, document)
}

func mustBundle(t *testing.T, snapshot pkgContext.Snapshot, document pkgContext.Document) pkgContext.ContextBundle {
	t.Helper()
	section := pkgContext.ContextSection{
		Path: document.Path(), Range: pkgContext.SourceRange{Start: 0, End: document.SizeBytes()},
		Role: "evidence", Method: pkgContext.RetrievalLexical, ReasonCode: "term_match",
		Text: document.Content(), Tokens: 3,
	}
	bundle, err := pkgContext.NewContextBundle(snapshot, "context.test", "1", pkgContext.Budget{MaxTokens: 100, ReservedTokens: 10, SafetyTokens: 10}, []pkgContext.ContextSection{section})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}
