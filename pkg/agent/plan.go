package agent

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
	StepBlocked   StepStatus = "blocked"
	StepSkipped   StepStatus = "skipped"
)

func (status StepStatus) Valid() bool {
	switch status {
	case StepPending, StepRunning, StepCompleted, StepFailed, StepBlocked, StepSkipped:
		return true
	default:
		return false
	}
}

func CanTransitionStep(from, to StepStatus) bool {
	switch from {
	case StepPending:
		return to == StepRunning || to == StepBlocked || to == StepSkipped
	case StepRunning:
		return to == StepCompleted || to == StepFailed || to == StepBlocked
	default:
		return false
	}
}

type PlanStep struct {
	id           StepID
	objective    string
	dependencies []StepID
	status       StepStatus
	reason       string
}

func NewPlanStep(
	id StepID,
	objective string,
	dependencies []StepID,
	status StepStatus,
	reason string,
) (PlanStep, error) {
	if err := id.Validate(); err != nil {
		return PlanStep{}, err
	}
	if strings.TrimSpace(objective) == "" || len(objective) > 16<<10 ||
		!utf8.ValidString(objective) || strings.ContainsRune(objective, 0) {
		return PlanStep{}, fmt.Errorf("plan step objective is blank, oversized, or unsafe: %w", ErrInvalidStep)
	}
	if !status.Valid() {
		return PlanStep{}, fmt.Errorf("plan step status %q is invalid: %w", status, ErrInvalidStep)
	}
	if status == StepPending || status == StepRunning || status == StepCompleted {
		if reason != "" {
			return PlanStep{}, fmt.Errorf("step status %q cannot carry a reason: %w", status, ErrInvalidStep)
		}
	} else if !safeCode(reason) {
		return PlanStep{}, fmt.Errorf("terminal step status %q requires a safe reason: %w", status, ErrInvalidStep)
	}
	cloned := slices.Clone(dependencies)
	seen := make(map[StepID]struct{}, len(cloned))
	for _, dependency := range cloned {
		if err := dependency.Validate(); err != nil || dependency == id {
			return PlanStep{}, fmt.Errorf("plan step dependency %q is invalid: %w", dependency, ErrInvalidStep)
		}
		if _, exists := seen[dependency]; exists {
			return PlanStep{}, fmt.Errorf("plan step dependency %q is duplicated: %w", dependency, ErrInvalidStep)
		}
		seen[dependency] = struct{}{}
	}
	slices.Sort(cloned)
	return PlanStep{id: id, objective: objective, dependencies: cloned, status: status, reason: reason}, nil
}

func (step PlanStep) ID() StepID             { return step.id }
func (step PlanStep) Objective() string      { return step.objective }
func (step PlanStep) Dependencies() []StepID { return slices.Clone(step.dependencies) }
func (step PlanStep) Status() StepStatus     { return step.status }
func (step PlanStep) Reason() string         { return step.reason }

func (step PlanStep) Validate() error {
	_, err := NewPlanStep(step.id, step.objective, step.dependencies, step.status, step.reason)
	return err
}

type Plan struct {
	version uint64
	steps   []PlanStep
}

func NewPlan(version uint64, steps []PlanStep) (Plan, error) {
	if version == 0 || len(steps) == 0 || len(steps) > 1_000 {
		return Plan{}, fmt.Errorf("plan version and cardinality are invalid: %w", ErrInvalidPlan)
	}
	cloned := slices.Clone(steps)
	byID := make(map[StepID]PlanStep, len(cloned))
	for index, step := range cloned {
		if err := step.Validate(); err != nil {
			return Plan{}, fmt.Errorf("plan step %d: %w: %w", index, err, ErrInvalidPlan)
		}
		if _, exists := byID[step.id]; exists {
			return Plan{}, fmt.Errorf("plan step %q is duplicated: %w", step.id, ErrInvalidPlan)
		}
		byID[step.id] = step
	}
	for _, step := range cloned {
		for _, dependency := range step.dependencies {
			if _, exists := byID[dependency]; !exists {
				return Plan{}, fmt.Errorf("step %q dependency %q is absent: %w", step.id, dependency, ErrInvalidPlan)
			}
		}
	}
	visiting := make(map[StepID]bool, len(cloned))
	visited := make(map[StepID]bool, len(cloned))
	var visit func(StepID) bool
	visit = func(id StepID) bool {
		if visiting[id] {
			return false
		}
		if visited[id] {
			return true
		}
		visiting[id] = true
		for _, dependency := range byID[id].dependencies {
			if !visit(dependency) {
				return false
			}
		}
		visiting[id] = false
		visited[id] = true
		return true
	}
	for id := range byID {
		if !visit(id) {
			return Plan{}, fmt.Errorf("plan contains a dependency cycle: %w", ErrInvalidPlan)
		}
	}
	return Plan{version: version, steps: cloned}, nil
}

func (plan Plan) Version() uint64   { return plan.version }
func (plan Plan) Steps() []PlanStep { return slices.Clone(plan.steps) }
func (plan Plan) Validate() error {
	_, err := NewPlan(plan.version, plan.steps)
	return err
}

// TransitionStep returns a new immutable plan with one validated state
// transition. A step can start only after all dependencies are completed or
// skipped.
func (plan Plan) TransitionStep(id StepID, status StepStatus, reason string) (Plan, error) {
	if err := plan.Validate(); err != nil {
		return Plan{}, err
	}
	steps := plan.Steps()
	index := -1
	states := make(map[StepID]StepStatus, len(steps))
	for candidateIndex, step := range steps {
		states[step.ID()] = step.Status()
		if step.ID() == id {
			index = candidateIndex
		}
	}
	if index < 0 {
		return Plan{}, fmt.Errorf("plan step %q is absent: %w", id, ErrInvalidTransition)
	}
	current := steps[index]
	if !CanTransitionStep(current.Status(), status) {
		return Plan{}, fmt.Errorf("plan step %q cannot transition from %q to %q: %w", id, current.Status(), status, ErrInvalidTransition)
	}
	if status == StepRunning {
		for _, dependency := range current.Dependencies() {
			if states[dependency] != StepCompleted && states[dependency] != StepSkipped {
				return Plan{}, fmt.Errorf("plan step %q dependency %q is not complete: %w", id, dependency, ErrInvalidTransition)
			}
		}
	}
	updated, err := NewPlanStep(current.ID(), current.Objective(), current.Dependencies(), status, reason)
	if err != nil {
		return Plan{}, err
	}
	steps[index] = updated
	return NewPlan(plan.version, steps)
}

type PlanningRequest struct {
	run         RunID
	instruction string
	bundle      contextengine.ContextBundle
	maxSteps    int
}

func NewPlanningRequest(run RunID, instruction string, bundle contextengine.ContextBundle, maxSteps int) (PlanningRequest, error) {
	if err := run.Validate(); err != nil || strings.TrimSpace(instruction) == "" || len(instruction) > 1<<20 ||
		!utf8.ValidString(instruction) || strings.ContainsRune(instruction, 0) ||
		bundle.Generation() == 0 || bundle.Workspace() == "" || maxSteps <= 0 || maxSteps > 1_000 {
		return PlanningRequest{}, fmt.Errorf("planning input is invalid: %w", ErrInvalidRequest)
	}
	if err := bundle.Workspace().Validate(); err != nil || bundle.Budget().Validate() != nil {
		return PlanningRequest{}, fmt.Errorf("planning context bundle is invalid: %w", ErrInvalidRequest)
	}
	return PlanningRequest{run: run, instruction: instruction, bundle: bundle, maxSteps: maxSteps}, nil
}

func (request PlanningRequest) Run() RunID                          { return request.run }
func (request PlanningRequest) Instruction() string                 { return request.instruction }
func (request PlanningRequest) Bundle() contextengine.ContextBundle { return request.bundle }
func (request PlanningRequest) MaxSteps() int                       { return request.maxSteps }

type Planner interface {
	Plan(context.Context, PlanningRequest) (Plan, error)
}
