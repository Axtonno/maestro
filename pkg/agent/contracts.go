package agent

import (
	"context"
	"fmt"
)

type Agent interface {
	Descriptor() Descriptor
}

func ValidateAgent(candidate Agent) error {
	if candidate == nil || nilInterface(candidate) {
		return ErrInvalidAgent
	}
	if err := candidate.Descriptor().Validate(); err != nil {
		return fmt.Errorf("agent descriptor: %w: %w", err, ErrInvalidAgent)
	}
	return nil
}

// A Planner is an optional capability of an Agent. Agent Runtime remains the
// owner of provider calls, permissions, tool execution, and session state.
type PlanningAgent interface {
	Agent
	Planner
}

type Runtime interface {
	Register(Agent) error
	Descriptors() []Descriptor
	Run(context.Context, RunRequest) (RunResult, error)
	Session(RunID) (SessionSnapshot, bool)
}
