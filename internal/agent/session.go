package agent

import (
	"errors"
	"fmt"
	"sync"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
)

type session struct {
	mu sync.RWMutex

	request  pkgAgent.RunRequest
	snapshot pkgAgent.SessionSnapshot
	history  []pkgAgent.Plan
	active   bool
}

func newSession(request pkgAgent.RunRequest) (*session, error) {
	snapshot, err := pkgAgent.NewSessionSnapshot(
		request.Run(), request.Agent(), request.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: 1, State: pkgAgent.SessionCreated,
			Counters: pkgAgent.Counters{SessionBytes: len(request.Instruction())},
		},
	)
	if err != nil {
		return nil, pkgAgent.NewRunError(pkgAgent.ErrorInvalid, request.Run(), request.Agent(), "session_create", err)
	}
	current := &session{request: request, snapshot: snapshot, active: true}
	if len(request.Instruction()) > request.Limits().MaxSessionBytes {
		return nil, pkgAgent.NewRunError(pkgAgent.ErrorLimit, request.Run(), request.Agent(), "session_byte_limit", pkgAgent.ErrLimitExceeded)
	}
	return current, nil
}

func (current *session) isActive() bool {
	current.mu.RLock()
	defer current.mu.RUnlock()
	return current.active
}

func (current *session) snapshotValue() pkgAgent.SessionSnapshot {
	current.mu.RLock()
	defer current.mu.RUnlock()
	return current.snapshot
}

func (current *session) transition(
	next pkgAgent.SessionState,
	plan *pkgAgent.Plan,
	workspaceGeneration uint64,
) error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() == pkgAgent.SessionTerminal {
		return pkgAgent.ErrSessionTerminal
	}
	if !pkgAgent.CanTransitionSession(current.snapshot.State(), next) {
		return fmt.Errorf("session cannot transition from %q to %q: %w", current.snapshot.State(), next, pkgAgent.ErrInvalidTransition)
	}
	if plan == nil {
		if existing, ok := current.snapshot.Plan(); ok {
			plan = &existing
		}
	}
	if workspaceGeneration == 0 {
		workspaceGeneration = current.snapshot.WorkspaceGeneration()
	}
	snapshot, err := pkgAgent.NewSessionSnapshot(
		current.snapshot.Run(), current.snapshot.Agent(), current.snapshot.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: current.snapshot.Generation() + 1, State: next,
			WorkspaceGeneration: workspaceGeneration, Plan: plan,
			Counters: current.snapshot.Counters(), ContextStale: current.snapshot.ContextStale(),
		},
	)
	if err != nil {
		return err
	}
	current.snapshot = snapshot
	if plan != nil {
		current.history = append(current.history, *plan)
	}
	return nil
}

func (current *session) consume(delta counterDelta) error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active {
		return pkgAgent.ErrSessionTerminal
	}
	counters := current.snapshot.Counters()
	counters.ModelTurns += delta.modelTurns
	counters.ToolCalls += delta.toolCalls
	counters.PlanRevisions += delta.planRevisions
	counters.InputTokens += delta.inputTokens
	counters.OutputTokens += delta.outputTokens
	counters.SessionBytes += delta.sessionBytes
	limits := current.request.Limits()
	if counters.ModelTurns > limits.MaxModelTurns || counters.ToolCalls > limits.MaxToolCalls ||
		counters.PlanRevisions > limits.MaxPlanRevisions || counters.InputTokens > limits.MaxInputTokens ||
		counters.OutputTokens > limits.MaxOutputTokens || counters.SessionBytes > limits.MaxSessionBytes {
		return pkgAgent.ErrLimitExceeded
	}
	plan, hasPlan := current.snapshot.Plan()
	var planPointer *pkgAgent.Plan
	if hasPlan {
		planPointer = &plan
	}
	snapshot, err := pkgAgent.NewSessionSnapshot(
		current.snapshot.Run(), current.snapshot.Agent(), current.snapshot.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: current.snapshot.Generation() + 1, State: current.snapshot.State(),
			WorkspaceGeneration: current.snapshot.WorkspaceGeneration(), Plan: planPointer,
			Counters: counters, ContextStale: current.snapshot.ContextStale(), Terminal: current.snapshot.Terminal(),
		},
	)
	if err != nil {
		return err
	}
	current.snapshot = snapshot
	return nil
}

func (current *session) transitionStep(id pkgAgent.StepID, status pkgAgent.StepStatus, reason string) error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() != pkgAgent.SessionRunning {
		return pkgAgent.ErrInvalidTransition
	}
	plan, ok := current.snapshot.Plan()
	if !ok {
		return pkgAgent.ErrInvalidPlan
	}
	updated, err := plan.TransitionStep(id, status, reason)
	if err != nil {
		return err
	}
	if err := current.replaceLocked(updated, current.snapshot.Counters()); err != nil {
		return err
	}
	if len(current.history) > 0 {
		current.history[len(current.history)-1] = updated
	}
	return nil
}

func (current *session) revise(plan pkgAgent.Plan) error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() == pkgAgent.SessionTerminal {
		return pkgAgent.ErrSessionTerminal
	}
	previous, ok := current.snapshot.Plan()
	if !ok || plan.Validate() != nil || plan.Version() != previous.Version()+1 ||
		len(plan.Steps()) > current.request.Limits().MaxPlanSteps {
		return pkgAgent.ErrInvalidPlan
	}
	counters := current.snapshot.Counters()
	counters.PlanRevisions++
	if counters.PlanRevisions > current.request.Limits().MaxPlanRevisions {
		return pkgAgent.ErrLimitExceeded
	}
	if err := current.replaceLocked(plan, counters); err != nil {
		return err
	}
	current.history = append(current.history, plan)
	return nil
}

func (current *session) planHistory() []pkgAgent.Plan {
	current.mu.RLock()
	defer current.mu.RUnlock()
	return append([]pkgAgent.Plan(nil), current.history...)
}

func (current *session) markStale() error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() != pkgAgent.SessionRunning {
		return pkgAgent.ErrInvalidTransition
	}
	return current.replaceFreshnessLocked(true, current.snapshot.WorkspaceGeneration())
}

func (current *session) markFresh(generation uint64) error {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() != pkgAgent.SessionRunning || generation == 0 ||
		generation < current.snapshot.WorkspaceGeneration() || (current.snapshot.ContextStale() && generation <= current.snapshot.WorkspaceGeneration()) {
		return pkgAgent.ErrInvalidTransition
	}
	return current.replaceFreshnessLocked(false, generation)
}

func (current *session) replaceFreshnessLocked(stale bool, generation uint64) error {
	plan, ok := current.snapshot.Plan()
	if !ok {
		return pkgAgent.ErrInvalidPlan
	}
	snapshot, err := pkgAgent.NewSessionSnapshot(
		current.snapshot.Run(), current.snapshot.Agent(), current.snapshot.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: current.snapshot.Generation() + 1, State: current.snapshot.State(),
			WorkspaceGeneration: generation, Plan: &plan, Counters: current.snapshot.Counters(),
			ContextStale: stale, Terminal: current.snapshot.Terminal(),
		},
	)
	if err != nil {
		return err
	}
	current.snapshot = snapshot
	return nil
}

func (current *session) replaceLocked(plan pkgAgent.Plan, counters pkgAgent.Counters) error {
	snapshot, err := pkgAgent.NewSessionSnapshot(
		current.snapshot.Run(), current.snapshot.Agent(), current.snapshot.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: current.snapshot.Generation() + 1, State: current.snapshot.State(),
			WorkspaceGeneration: current.snapshot.WorkspaceGeneration(), Plan: &plan,
			Counters: counters, ContextStale: current.snapshot.ContextStale(), Terminal: current.snapshot.Terminal(),
		},
	)
	if err != nil {
		return err
	}
	current.snapshot = snapshot
	return nil
}

func (current *session) finish(content string, candidates ...pkgAgent.TerminalReason) (pkgAgent.RunResult, error) {
	current.mu.Lock()
	defer current.mu.Unlock()
	if !current.active || current.snapshot.State() == pkgAgent.SessionTerminal {
		return pkgAgent.RunResult{}, pkgAgent.ErrSessionTerminal
	}
	reason, err := pkgAgent.SelectTerminal(candidates...)
	if err != nil {
		return pkgAgent.RunResult{}, err
	}
	plan, hasPlan := current.snapshot.Plan()
	var planPointer *pkgAgent.Plan
	if hasPlan {
		planPointer = &plan
	}
	snapshot, err := pkgAgent.NewSessionSnapshot(
		current.snapshot.Run(), current.snapshot.Agent(), current.snapshot.Workspace(),
		pkgAgent.SessionSnapshotOptions{
			Generation: current.snapshot.Generation() + 1, State: pkgAgent.SessionTerminal,
			WorkspaceGeneration: current.snapshot.WorkspaceGeneration(), Plan: planPointer,
			Counters: current.snapshot.Counters(), ContextStale: current.snapshot.ContextStale(), Terminal: reason,
		},
	)
	if err != nil {
		return pkgAgent.RunResult{}, err
	}
	result, err := pkgAgent.NewRunResult(snapshot, content)
	if err != nil {
		return pkgAgent.RunResult{}, err
	}
	current.snapshot = snapshot
	current.active = false
	return result, nil
}

func (current *session) fail(
	reason pkgAgent.TerminalReason,
	kind pkgAgent.ErrorKind,
	code string,
	cause error,
) (pkgAgent.RunResult, error) {
	result, terminalErr := current.finish("", reason)
	if terminalErr != nil && !errors.Is(terminalErr, pkgAgent.ErrSessionTerminal) {
		cause = errors.Join(cause, terminalErr)
	}
	return result, pkgAgent.NewRunError(kind, current.request.Run(), current.request.Agent(), code, cause)
}

type counterDelta struct {
	modelTurns    int
	toolCalls     int
	planRevisions int
	inputTokens   int
	outputTokens  int
	sessionBytes  int
}
