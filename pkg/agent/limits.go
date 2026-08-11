package agent

import (
	"fmt"
	"time"
)

type Limits struct {
	MaxDuration         time.Duration
	MaxModelTurns       int
	MaxToolCalls        int
	MaxToolCallsPerTurn int
	MaxPlanSteps        int
	MaxPlanRevisions    int
	MaxToolResultBytes  int
	MaxSessionBytes     int
	MaxInputTokens      int
	MaxOutputTokens     int
}

func (limits Limits) Validate() error {
	if limits.MaxDuration <= 0 || limits.MaxDuration > 24*time.Hour ||
		limits.MaxModelTurns <= 0 || limits.MaxModelTurns > 10_000 ||
		limits.MaxToolCalls <= 0 || limits.MaxToolCalls > 100_000 ||
		limits.MaxToolCallsPerTurn <= 0 || limits.MaxToolCallsPerTurn > limits.MaxToolCalls ||
		limits.MaxPlanSteps <= 0 || limits.MaxPlanSteps > 1_000 ||
		limits.MaxPlanRevisions <= 0 || limits.MaxPlanRevisions > 1_000 ||
		limits.MaxToolResultBytes <= 0 || limits.MaxToolResultBytes > 64<<20 ||
		limits.MaxSessionBytes < limits.MaxToolResultBytes || limits.MaxSessionBytes > 256<<20 ||
		limits.MaxInputTokens <= 0 || limits.MaxInputTokens > 10_000_000 ||
		limits.MaxOutputTokens <= 0 || limits.MaxOutputTokens > 10_000_000 {
		return fmt.Errorf("agent limits must be positive, coherent, and bounded: %w", ErrInvalidLimits)
	}
	return nil
}
