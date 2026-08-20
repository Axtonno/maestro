package application

import (
	"fmt"
	"io"
	"sync"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

// ProgressRenderer subscribes only to the public redacted event contracts.
// It never receives prompts, workspace content, invocation arguments or tool
// output.
type ProgressRenderer struct {
	output io.Writer
	mu     sync.Mutex
}

func NewProgressRenderer(output io.Writer) *ProgressRenderer {
	if output == nil {
		output = io.Discard
	}
	return &ProgressRenderer{output: output}
}

func (renderer *ProgressRenderer) RenderLimits(limits pkgAgent.Limits) {
	if renderer == nil || limits.Validate() != nil {
		return
	}
	renderer.write("limits\tduration=%s model_turns=%d tool_calls=%d tool_calls_per_turn=%d plan_steps=%d plan_revisions=%d input_tokens=%d output_tokens=%d tool_result_bytes=%d session_bytes=%d\n",
		limits.MaxDuration, limits.MaxModelTurns, limits.MaxToolCalls, limits.MaxToolCallsPerTurn,
		limits.MaxPlanSteps, limits.MaxPlanRevisions, limits.MaxInputTokens, limits.MaxOutputTokens,
		limits.MaxToolResultBytes, limits.MaxSessionBytes)
}

func (renderer *ProgressRenderer) Subscribe(events pkgRuntime.EventBus) error {
	if renderer == nil || events == nil {
		return pkgRuntime.ErrInvalidSubscription
	}
	for _, topic := range []string{
		pkgAgent.EventSessionStarted, pkgAgent.EventPlanCreated, pkgAgent.EventPlanRevised,
		pkgAgent.EventStepTransitioned, pkgAgent.EventTurnCompleted, pkgAgent.EventLimitReached,
		pkgAgent.EventMutationTransitioned,
		pkgAgent.EventSessionCompleted, pkgAgent.EventSessionFailed,
	} {
		if err := events.Subscribe(topic, renderer.agentEvent); err != nil {
			return fmt.Errorf("subscribe progress topic %q: %w", topic, err)
		}
	}
	for _, topic := range []string{pkgTool.EventInvocationPrepared, pkgTool.EventPermissionDecided} {
		if err := events.Subscribe(topic, renderer.toolEvent); err != nil {
			return fmt.Errorf("subscribe mutation progress topic %q: %w", topic, err)
		}
	}
	return nil
}

func (renderer *ProgressRenderer) agentEvent(event pkgRuntime.Event) {
	defer func() { _ = recover() }()
	payload, ok := event.Payload().(pkgAgent.EventPayload)
	if !ok {
		return
	}
	switch event.Name() {
	case pkgAgent.EventSessionStarted:
		renderer.write("progress\trun=%s state=%s\n", payload.Run, payload.State)
	case pkgAgent.EventPlanCreated, pkgAgent.EventPlanRevised:
		renderer.write("plan\trun=%s version=%d\n", payload.Run, payload.PlanVersion)
	case pkgAgent.EventStepTransitioned:
		renderer.write("step\trun=%s id=%s state=%s\n", payload.Run, payload.Step, payload.StepState)
	case pkgAgent.EventTurnCompleted:
		renderer.write("progress\trun=%s model_turns=%d tool_calls=%d tokens=%d/%d\n",
			payload.Run, payload.ModelTurns, payload.ToolCalls, payload.InputTokens, payload.OutputTokens)
	case pkgAgent.EventLimitReached:
		renderer.write("limit\trun=%s terminal=%s failure=%s\n", payload.Run, payload.Terminal, payload.Failure)
	case pkgAgent.EventMutationTransitioned:
		renderer.write("mutation\trun=%s stage=%s status=%s effect=%s durable=%t generation=%d\n",
			payload.Run, payload.MutationStage, payload.MutationStatus, payload.MutationEffect,
			payload.Durable, payload.WorkspaceGeneration)
	case pkgAgent.EventSessionCompleted, pkgAgent.EventSessionFailed:
		renderer.write("terminal\trun=%s reason=%s model_turns=%d tool_calls=%d duration_ms=%d\n",
			payload.Run, payload.Terminal, payload.ModelTurns, payload.ToolCalls, payload.DurationMillis)
	}
}

func (renderer *ProgressRenderer) toolEvent(event pkgRuntime.Event) {
	defer func() { _ = recover() }()
	payload, ok := event.Payload().(pkgTool.EventPayload)
	if !ok {
		return
	}
	if payload.Tool != "workspace.patch" {
		if event.Name() == pkgTool.EventPermissionDecided {
			renderer.write("permission\trun=%s decision=%s actions=%d failure=%s\n",
				payload.Run, payload.Decision, payload.ActionCount, payload.Failure)
		}
		return
	}
	switch event.Name() {
	case pkgTool.EventInvocationPrepared:
		renderer.write("mutation\trun=%s stage=proposal status=prepared\n", payload.Run)
	case pkgTool.EventPermissionDecided:
		status := "denied"
		if payload.Decision == pkgTool.DecisionAllow {
			status = "allowed"
		}
		renderer.write("mutation\trun=%s stage=approval status=%s\n", payload.Run, status)
	}
}

func (renderer *ProgressRenderer) write(format string, values ...any) {
	renderer.mu.Lock()
	defer renderer.mu.Unlock()
	defer func() { _ = recover() }()
	_, _ = fmt.Fprintf(renderer.output, format, values...)
}
