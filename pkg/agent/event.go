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
	EventEvidenceTransitioned = "agent.evidence.transitioned"
)

type EvidenceStage string

const (
	EvidenceStageRoute              EvidenceStage = "route"
	EvidenceStageControllerAction   EvidenceStage = "controller_action"
	EvidenceStageReferencedSymbols  EvidenceStage = "referenced_symbols"
	EvidenceStageEventsJobsServices EvidenceStage = "events_jobs_services"
)

type EvidenceStatus string

const (
	EvidencePending     EvidenceStatus = "pending"
	EvidenceCovered     EvidenceStatus = "covered"
	EvidenceUnavailable EvidenceStatus = "unavailable"
)

type EvidenceDecision string

const (
	EvidenceDecisionBootstrap EvidenceDecision = "bootstrap_read"
	EvidenceDecisionRead      EvidenceDecision = "read"
	EvidenceDecisionSearch    EvidenceDecision = "search"
	EvidenceDecisionDeclare   EvidenceDecision = "declare"
	EvidenceDecisionFinalize  EvidenceDecision = "finalize"
)

type EvidenceResult string

const (
	EvidenceResultPending  EvidenceResult = "pending"
	EvidenceResultSuccess  EvidenceResult = "success"
	EvidenceResultEmpty    EvidenceResult = "empty"
	EvidenceResultFailed   EvidenceResult = "failed"
	EvidenceResultAccepted EvidenceResult = "accepted"
	EvidenceResultRejected EvidenceResult = "rejected"
)

type EvidenceStopReason string

const (
	EvidenceStopStageComplete       EvidenceStopReason = "stage_complete"
	EvidenceStopDeclaredUnavailable EvidenceStopReason = "declared_unavailable"
	EvidenceStopIncomplete          EvidenceStopReason = "incomplete_evidence"
	EvidenceStopComplete            EvidenceStopReason = "complete"
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
	Run              RunID
	Agent            ID
	Step             StepID
	State            SessionState
	StepState        StepStatus
	Terminal         TerminalReason
	PlanVersion      uint64
	ModelTurns       int
	ToolCalls        int
	InputTokens      int
	OutputTokens     int
	DurationMillis   int64
	Failure          EventFailure
	EvidenceStage    EvidenceStage
	EvidenceStatus   EvidenceStatus
	EvidenceDecision EvidenceDecision
	EvidenceTool     string
	EvidenceResult   EvidenceResult
	EvidenceStop     EvidenceStopReason
}

// MutationEventPayload is a dedicated redacted allowlist for the mutating
// lifecycle. It contains no path, proposal, arguments, output, or error text.
type MutationEventPayload struct {
	Run                 RunID
	Agent               ID
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

type MutationEvent struct {
	Topic string
	Data  MutationEventPayload
}

func (event MutationEvent) Name() string { return event.Topic }
func (event MutationEvent) Payload() any { return event.Data }
