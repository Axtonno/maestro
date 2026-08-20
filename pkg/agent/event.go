package agent

const (
	EventSessionStarted       = "agent.session.started"
	EventSessionCompleted     = "agent.session.completed"
	EventSessionFailed        = "agent.session.failed"
	EventPlanCreated          = "agent.plan.created"
	EventPlanRevised          = "agent.plan.revised"
	EventStepTransitioned     = "agent.step.transitioned"
	EventTurnStarted          = "agent.turn.started"
	EventTurnCompleted        = "agent.turn.completed"
	EventLimitReached         = "agent.limit.reached"
	EventMutationTransitioned = "agent.mutation.transitioned"
)

type MutationStage string

const (
	MutationStageApply   MutationStage = "apply"
	MutationStageReindex MutationStage = "reindex"
)

type MutationStatus string

const (
	MutationStarted   MutationStatus = "started"
	MutationSucceeded MutationStatus = "succeeded"
	MutationFailed    MutationStatus = "failed"
	MutationCanceled  MutationStatus = "canceled"
)

type MutationEffect string

const (
	MutationEffectUnchanged MutationEffect = "unchanged"
	MutationEffectApplied   MutationEffect = "applied"
)

type EventFailure string

const (
	EventFailureNone       EventFailure = ""
	EventFailureInvalid    EventFailure = "invalid"
	EventFailurePlanning   EventFailure = "planning_failure"
	EventFailureProvider   EventFailure = "provider_failure"
	EventFailureTool       EventFailure = "tool_failure"
	EventFailurePermission EventFailure = "permission_denied"
	EventFailureLimit      EventFailure = "limit_exceeded"
	EventFailureCanceled   EventFailure = "canceled"
	EventFailureDeadline   EventFailure = "deadline_exceeded"
	EventFailureInternal   EventFailure = "internal"
)

// EventPayload is an exact telemetry allowlist. It contains no provider,
// model, policy, workspace, instruction, prompt, plan objective, content,
// tool arguments/output, path, or external error text.
type EventPayload struct {
	Run                 RunID
	Agent               ID
	Step                StepID
	State               SessionState
	StepState           StepStatus
	Terminal            TerminalReason
	PlanVersion         uint64
	ModelTurns          int
	ToolCalls           int
	InputTokens         int
	OutputTokens        int
	DurationMillis      int64
	Failure             EventFailure
	MutationStage       MutationStage
	MutationStatus      MutationStatus
	MutationEffect      MutationEffect
	Durable             bool
	WorkspaceGeneration uint64
}

type Event struct {
	Topic string
	Data  EventPayload
}

func (event Event) Name() string { return event.Topic }
func (event Event) Payload() any { return event.Data }
