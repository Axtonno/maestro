package agent

import (
	"fmt"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

type SessionState string

const (
	SessionCreated  SessionState = "created"
	SessionPlanning SessionState = "planning"
	SessionRunning  SessionState = "running"
	SessionTerminal SessionState = "terminal"
)

func (state SessionState) Valid() bool {
	switch state {
	case SessionCreated, SessionPlanning, SessionRunning, SessionTerminal:
		return true
	default:
		return false
	}
}

func CanTransitionSession(from, to SessionState) bool {
	switch from {
	case SessionCreated:
		return to == SessionPlanning || to == SessionTerminal
	case SessionPlanning:
		return to == SessionRunning || to == SessionTerminal
	case SessionRunning:
		return to == SessionTerminal
	default:
		return false
	}
}

type TerminalReason string

const (
	TerminalCompleted        TerminalReason = "completed"
	TerminalDeadline         TerminalReason = "deadline_exceeded"
	TerminalCanceled         TerminalReason = "canceled"
	TerminalLimit            TerminalReason = "limit_exceeded"
	TerminalPermissionDenied TerminalReason = "permission_denied"
	TerminalBlocked          TerminalReason = "blocked"
	TerminalProviderFailure  TerminalReason = "provider_failure"
	TerminalToolFailure      TerminalReason = "tool_failure"
	TerminalPlanningFailure  TerminalReason = "planning_failure"
	TerminalInternalFailure  TerminalReason = "internal_failure"
)

func (reason TerminalReason) Valid() bool {
	return terminalPriority(reason) > 0
}

// SelectTerminal applies the deterministic pre-commit precedence defined by
// ADR-0025. Once a terminal snapshot is committed it is never replaced.
func SelectTerminal(candidates ...TerminalReason) (TerminalReason, error) {
	selected := TerminalReason("")
	priority := 0
	for _, candidate := range candidates {
		candidatePriority := terminalPriority(candidate)
		if candidatePriority == 0 {
			return "", fmt.Errorf("terminal reason %q is invalid: %w", candidate, ErrInvalidSession)
		}
		if candidatePriority > priority {
			selected = candidate
			priority = candidatePriority
		}
	}
	if selected == "" {
		return "", fmt.Errorf("at least one terminal candidate is required: %w", ErrInvalidSession)
	}
	return selected, nil
}

func terminalPriority(reason TerminalReason) int {
	switch reason {
	case TerminalDeadline:
		return 10
	case TerminalCanceled:
		return 9
	case TerminalLimit:
		return 8
	case TerminalPermissionDenied:
		return 7
	case TerminalBlocked:
		return 6
	case TerminalProviderFailure:
		return 5
	case TerminalToolFailure:
		return 4
	case TerminalPlanningFailure:
		return 3
	case TerminalInternalFailure:
		return 2
	case TerminalCompleted:
		return 1
	default:
		return 0
	}
}

type Counters struct {
	ModelTurns    int
	ToolCalls     int
	PlanRevisions int
	InputTokens   int
	OutputTokens  int
	SessionBytes  int
}

func (counters Counters) Validate() error {
	if counters.ModelTurns < 0 || counters.ToolCalls < 0 || counters.PlanRevisions < 0 ||
		counters.InputTokens < 0 || counters.OutputTokens < 0 || counters.SessionBytes < 0 {
		return fmt.Errorf("session counters cannot be negative: %w", ErrInvalidSession)
	}
	return nil
}

type SessionSnapshotOptions struct {
	Generation          uint64
	State               SessionState
	WorkspaceGeneration uint64
	Plan                *Plan
	Counters            Counters
	ContextStale        bool
	Terminal            TerminalReason
}

type SessionSnapshot struct {
	run                 RunID
	agent               ID
	workspace           contextengine.WorkspaceID
	generation          uint64
	state               SessionState
	workspaceGeneration uint64
	plan                *Plan
	counters            Counters
	contextStale        bool
	terminal            TerminalReason
}

func NewSessionSnapshot(
	run RunID,
	agent ID,
	workspace contextengine.WorkspaceID,
	options SessionSnapshotOptions,
) (SessionSnapshot, error) {
	if err := run.Validate(); err != nil || agent.Validate() != nil || workspace.Validate() != nil ||
		options.Generation == 0 || !options.State.Valid() || options.Counters.Validate() != nil {
		return SessionSnapshot{}, fmt.Errorf("session identity, generation, state, or counters are invalid: %w", ErrInvalidSession)
	}
	var plan *Plan
	if options.Plan != nil {
		if err := options.Plan.Validate(); err != nil {
			return SessionSnapshot{}, fmt.Errorf("session plan: %w: %w", err, ErrInvalidSession)
		}
		copyPlan := *options.Plan
		plan = &copyPlan
	}
	if options.State == SessionRunning && plan == nil {
		return SessionSnapshot{}, fmt.Errorf("running session requires a plan: %w", ErrInvalidSession)
	}
	if options.State == SessionTerminal {
		if !options.Terminal.Valid() {
			return SessionSnapshot{}, fmt.Errorf("terminal session requires a reason: %w", ErrInvalidSession)
		}
		if options.Terminal == TerminalCompleted && plan == nil {
			return SessionSnapshot{}, fmt.Errorf("completed session requires a plan: %w", ErrInvalidSession)
		}
		if options.Terminal == TerminalCompleted {
			for _, step := range plan.steps {
				if step.status != StepCompleted && step.status != StepSkipped {
					return SessionSnapshot{}, fmt.Errorf("completed session contains non-terminal step %q: %w", step.id, ErrInvalidSession)
				}
			}
		}
	} else if options.Terminal != "" {
		return SessionSnapshot{}, fmt.Errorf("non-terminal session cannot carry a terminal reason: %w", ErrInvalidSession)
	}
	return SessionSnapshot{
		run: run, agent: agent, workspace: workspace, generation: options.Generation,
		state: options.State, workspaceGeneration: options.WorkspaceGeneration,
		plan: plan, counters: options.Counters, contextStale: options.ContextStale,
		terminal: options.Terminal,
	}, nil
}

func (snapshot SessionSnapshot) Run() RunID                           { return snapshot.run }
func (snapshot SessionSnapshot) Agent() ID                            { return snapshot.agent }
func (snapshot SessionSnapshot) Workspace() contextengine.WorkspaceID { return snapshot.workspace }
func (snapshot SessionSnapshot) Generation() uint64                   { return snapshot.generation }
func (snapshot SessionSnapshot) State() SessionState                  { return snapshot.state }
func (snapshot SessionSnapshot) WorkspaceGeneration() uint64          { return snapshot.workspaceGeneration }
func (snapshot SessionSnapshot) Counters() Counters                   { return snapshot.counters }
func (snapshot SessionSnapshot) ContextStale() bool                   { return snapshot.contextStale }
func (snapshot SessionSnapshot) Terminal() TerminalReason             { return snapshot.terminal }
func (snapshot SessionSnapshot) Plan() (Plan, bool) {
	if snapshot.plan == nil {
		return Plan{}, false
	}
	return *snapshot.plan, true
}

func (snapshot SessionSnapshot) Validate() error {
	_, err := NewSessionSnapshot(snapshot.run, snapshot.agent, snapshot.workspace, SessionSnapshotOptions{
		Generation: snapshot.generation, State: snapshot.state,
		WorkspaceGeneration: snapshot.workspaceGeneration, Plan: snapshot.plan,
		Counters: snapshot.counters, ContextStale: snapshot.contextStale,
		Terminal: snapshot.terminal,
	})
	return err
}
