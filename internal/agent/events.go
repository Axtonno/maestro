package agent

import (
	"context"
	"errors"
	"time"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func publishAgentEvent(events pkgRuntime.EventBus, topic string, payload pkgAgent.EventPayload) {
	if events == nil {
		return
	}
	func() {
		defer func() { _ = recover() }()
		_ = events.Publish(pkgAgent.Event{Topic: topic, Data: payload})
	}()
}

func publishMutationEvent(events pkgRuntime.EventBus, payload pkgAgent.MutationEventPayload) {
	if events == nil {
		return
	}
	func() {
		defer func() { _ = recover() }()
		_ = events.Publish(pkgAgent.MutationEvent{Topic: pkgAgent.EventMutationTransitioned, Data: payload})
	}()
}

func publishEvidenceEvent(
	events pkgRuntime.EventBus,
	run pkgAgent.RunID,
	agent pkgAgent.ID,
	stage pkgAgent.EvidenceStage,
	status pkgAgent.EvidenceStatus,
	decision pkgAgent.EvidenceDecision,
	tool string,
	result pkgAgent.EvidenceResult,
	stop pkgAgent.EvidenceStopReason,
) {
	if events == nil {
		return
	}
	publishAgentEvent(events, pkgAgent.EventEvidenceTransitioned, pkgAgent.EventPayload{
		Run: run, Agent: agent, EvidenceStage: stage, EvidenceStatus: status,
		EvidenceDecision: decision, EvidenceTool: tool, EvidenceResult: result,
		EvidenceStop: stop,
	})
}

func sessionEventPayload(snapshot pkgAgent.SessionSnapshot, duration time.Duration, failure pkgAgent.EventFailure) pkgAgent.EventPayload {
	counters := snapshot.Counters()
	payload := pkgAgent.EventPayload{
		Run: snapshot.Run(), Agent: snapshot.Agent(), State: snapshot.State(), Terminal: snapshot.Terminal(),
		ModelTurns: counters.ModelTurns, ToolCalls: counters.ToolCalls,
		InputTokens: counters.InputTokens, OutputTokens: counters.OutputTokens,
		DurationMillis: duration.Milliseconds(), Failure: failure,
	}
	if plan, ok := snapshot.Plan(); ok {
		payload.PlanVersion = plan.Version()
	}
	return payload
}

func classifyAgentEventFailure(err error, terminal pkgAgent.TerminalReason) pkgAgent.EventFailure {
	switch {
	case err == nil:
		return pkgAgent.EventFailureNone
	case errors.Is(err, context.DeadlineExceeded) || terminal == pkgAgent.TerminalDeadline:
		return pkgAgent.EventFailureDeadline
	case errors.Is(err, context.Canceled) || terminal == pkgAgent.TerminalCanceled:
		return pkgAgent.EventFailureCanceled
	case terminal == pkgAgent.TerminalPlanningFailure:
		return pkgAgent.EventFailurePlanning
	case terminal == pkgAgent.TerminalProviderFailure:
		return pkgAgent.EventFailureProvider
	case terminal == pkgAgent.TerminalToolFailure:
		return pkgAgent.EventFailureTool
	case terminal == pkgAgent.TerminalPermissionDenied:
		return pkgAgent.EventFailurePermission
	case terminal == pkgAgent.TerminalLimit:
		return pkgAgent.EventFailureLimit
	case terminal == pkgAgent.TerminalInternalFailure:
		return pkgAgent.EventFailureInternal
	default:
		return pkgAgent.EventFailureInternal
	}
}
