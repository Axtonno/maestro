package agent

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

var _ pkgAgent.Runtime = (*Runtime)(nil)

type loop interface {
	Run(context.Context, *session, pkgAgent.RunRequest, pkgContext.ContextBundle) (string, pkgAgent.TerminalReason, error)
}

type generationRuntime interface {
	Complete(context.Context, provider.ID, provider.CompletionRequest) (provider.CompletionResponse, error)
	Stream(context.Context, provider.ID, provider.CompletionRequest) (provider.Stream, error)
}

type workspaceBinder interface {
	Bind(pkgTool.RunID, pkgContext.Workspace) error
	Unbind(pkgTool.RunID)
}

type executionStartObserver interface {
	ObserveExecutionStart(func(pkgTool.PreparedInvocation)) error
}

type Options struct {
	MaxSessions int
	Permissions pkgTool.Runtime
	Providers   generationRuntime
	Tools       pkgTool.Runtime
	Workspaces  workspaceBinder
	EventBus    pkgRuntime.EventBus
}

func DefaultOptions() Options { return Options{MaxSessions: 1024} }

type Runtime struct {
	mu sync.RWMutex

	context                 pkgContext.Engine
	permissions             pkgTool.Runtime
	options                 Options
	agents                  map[pkgAgent.ID]pkgAgent.Agent
	sessions                map[pkgAgent.RunID]*session
	loop                    loop
	registrationInvalidator func()
	events                  pkgRuntime.EventBus
}

func NewRuntime(contextEngine pkgContext.Engine) *Runtime {
	runtime, err := NewRuntimeWithOptions(contextEngine, DefaultOptions())
	if err != nil {
		panic(err)
	}
	return runtime
}

func NewRuntimeWithOptions(contextEngine pkgContext.Engine, options Options) (*Runtime, error) {
	if contextEngine == nil || typedNil(contextEngine) || options.MaxSessions <= 0 || options.MaxSessions > 1_000_000 {
		return nil, fmt.Errorf("agent runtime dependencies or limits are invalid: %w", pkgAgent.ErrInvalidRequest)
	}
	permissions := options.Permissions
	if permissions == nil {
		permissions = options.Tools
	}
	runtime := &Runtime{
		context: contextEngine, permissions: permissions, options: options, events: options.EventBus,
		agents: make(map[pkgAgent.ID]pkgAgent.Agent), sessions: make(map[pkgAgent.RunID]*session),
	}
	if options.Providers != nil && !typedNil(options.Providers) && options.Tools != nil && !typedNil(options.Tools) {
		runtime.loop = &agentLoop{providers: options.Providers, tools: options.Tools, permissions: permissions, context: contextEngine, events: options.EventBus}
	}
	if observable, ok := options.Tools.(executionStartObserver); ok {
		if err := observable.ObserveExecutionStart(runtime.executionStarted); err != nil {
			return nil, err
		}
	}
	return runtime, nil
}

func (runtime *Runtime) executionStarted(prepared pkgTool.PreparedInvocation) {
	mutating := false
	for _, action := range prepared.Actions() {
		if action.Effect() == pkgTool.EffectWorkspaceMutate {
			mutating = true
			break
		}
	}
	if !mutating {
		return
	}
	runtime.mu.RLock()
	current := runtime.sessions[pkgAgent.RunID(prepared.Invocation().Run())]
	runtime.mu.RUnlock()
	if current != nil {
		_ = current.markStale()
	}
}

func (runtime *Runtime) Register(candidate pkgAgent.Agent) error {
	if err := pkgAgent.ValidateAgent(candidate); err != nil {
		return fmt.Errorf("register agent: %w", err)
	}
	descriptor := candidate.Descriptor()
	runtime.mu.Lock()
	if _, exists := runtime.agents[descriptor.ID()]; exists {
		runtime.mu.Unlock()
		return fmt.Errorf("register agent %q: %w", descriptor.ID(), pkgAgent.ErrAlreadyRegistered)
	}
	runtime.agents[descriptor.ID()] = candidate
	invalidator := runtime.registrationInvalidator
	runtime.mu.Unlock()
	if invalidator != nil {
		invalidator()
	}
	return nil
}

func (runtime *Runtime) SetRegistrationInvalidator(invalidator func()) {
	runtime.mu.Lock()
	runtime.registrationInvalidator = invalidator
	runtime.mu.Unlock()
}

func (runtime *Runtime) Descriptors() []pkgAgent.Descriptor {
	runtime.mu.RLock()
	descriptors := make([]pkgAgent.Descriptor, 0, len(runtime.agents))
	for _, candidate := range runtime.agents {
		descriptors = append(descriptors, candidate.Descriptor())
	}
	runtime.mu.RUnlock()
	slices.SortFunc(descriptors, func(left, right pkgAgent.Descriptor) int {
		return left.ID().Compare(right.ID())
	})
	return descriptors
}

func (runtime *Runtime) Run(ctx context.Context, request pkgAgent.RunRequest) (result pkgAgent.RunResult, err error) {
	if ctx == nil {
		return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorInvalid, request.Run(), request.Agent(), "nil_context", pkgAgent.ErrInvalidRequest)
	}
	if err := request.Validate(); err != nil {
		return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorInvalid, request.Run(), request.Agent(), "invalid_request", err)
	}
	if err := ctx.Err(); err != nil {
		return pkgAgent.RunResult{}, runContextError(request, err)
	}
	runtime.mu.Lock()
	candidate, exists := runtime.agents[request.Agent()]
	if !exists {
		runtime.mu.Unlock()
		return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorNotFound, request.Run(), request.Agent(), "agent_not_found", pkgAgent.ErrNotFound)
	}
	if current, exists := runtime.sessions[request.Run()]; exists {
		runtime.mu.Unlock()
		if current.isActive() {
			return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorInvalid, request.Run(), request.Agent(), "session_active", pkgAgent.ErrSessionActive)
		}
		return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorInvalid, request.Run(), request.Agent(), "session_terminal", pkgAgent.ErrSessionTerminal)
	}
	if len(runtime.sessions) >= runtime.options.MaxSessions {
		runtime.mu.Unlock()
		return pkgAgent.RunResult{}, pkgAgent.NewRunError(pkgAgent.ErrorLimit, request.Run(), request.Agent(), "session_registry_limit", pkgAgent.ErrLimitExceeded)
	}
	current, err := newSession(request)
	if err != nil {
		runtime.mu.Unlock()
		return pkgAgent.RunResult{}, err
	}
	runtime.sessions[request.Run()] = current
	runtime.mu.Unlock()
	started := time.Now()
	publishAgentEvent(runtime.events, pkgAgent.EventSessionStarted, sessionEventPayload(current.snapshotValue(), 0, pkgAgent.EventFailureNone))
	defer func() {
		snapshot := current.snapshotValue()
		failure := classifyAgentEventFailure(err, snapshot.Terminal())
		topic := pkgAgent.EventSessionCompleted
		if err != nil {
			topic = pkgAgent.EventSessionFailed
		}
		if snapshot.Terminal() == pkgAgent.TerminalLimit {
			publishAgentEvent(runtime.events, pkgAgent.EventLimitReached, sessionEventPayload(snapshot, time.Since(started), failure))
		}
		publishAgentEvent(runtime.events, topic, sessionEventPayload(snapshot, time.Since(started), failure))
	}()
	if workspace, ok := request.WorkspaceTarget(); ok && runtime.options.Workspaces != nil {
		toolRun := pkgTool.RunID(request.Run())
		if err := runtime.options.Workspaces.Bind(toolRun, workspace); err != nil {
			return current.fail(pkgAgent.TerminalInternalFailure, pkgAgent.ErrorInternal, "workspace_bind", err)
		}
		defer runtime.options.Workspaces.Unbind(toolRun)
	}

	runContext, cancel := context.WithTimeout(ctx, request.Limits().MaxDuration)
	defer cancel()
	return runtime.coordinate(runContext, candidate, current, request)
}

func (runtime *Runtime) coordinate(
	ctx context.Context,
	candidate pkgAgent.Agent,
	current *session,
	request pkgAgent.RunRequest,
) (pkgAgent.RunResult, error) {
	if err := current.transition(pkgAgent.SessionPlanning, nil, 0); err != nil {
		return current.fail(pkgAgent.TerminalInternalFailure, pkgAgent.ErrorInternal, "planning_transition", err)
	}
	bundle, err := runtime.context.Build(ctx, request.Context())
	if err != nil {
		return current.fail(reasonForError(ctx, pkgAgent.TerminalPlanningFailure), kindForError(ctx, pkgAgent.ErrorPlanning), "context_build", err)
	}
	if err := current.consume(counterDelta{sessionBytes: bundleBytes(bundle)}); err != nil {
		return current.fail(pkgAgent.TerminalLimit, pkgAgent.ErrorLimit, "session_byte_limit", err)
	}
	planner, ok := candidate.(pkgAgent.PlanningAgent)
	if !ok {
		return current.fail(pkgAgent.TerminalPlanningFailure, pkgAgent.ErrorPlanning, "planner_unsupported", pkgAgent.ErrPlanningFailed)
	}
	planningRequest, err := pkgAgent.NewPlanningRequest(request.Run(), request.Instruction(), bundle, request.Limits().MaxPlanSteps)
	if err != nil {
		return current.fail(pkgAgent.TerminalPlanningFailure, pkgAgent.ErrorPlanning, "invalid_planning_request", err)
	}
	plan, usage, modelBacked, err := runtime.callPlanner(ctx, current, planner, planningRequest, request, bundle)
	if err != nil {
		reason := reasonForError(ctx, pkgAgent.TerminalPlanningFailure)
		kind := kindForError(ctx, pkgAgent.ErrorPlanning)
		if errors.Is(err, pkgAgent.ErrPermissionDenied) {
			reason, kind = pkgAgent.TerminalPermissionDenied, pkgAgent.ErrorPermission
		} else if errors.Is(err, pkgAgent.ErrLimitExceeded) {
			reason, kind = pkgAgent.TerminalLimit, pkgAgent.ErrorLimit
		}
		return current.fail(reason, kind, "planning_failed", err)
	}
	if modelBacked {
		if err := current.consume(counterDelta{
			inputTokens: usage.InputTokens, outputTokens: usage.OutputTokens,
			sessionBytes: planBytes(plan),
		}); err != nil {
			return current.fail(pkgAgent.TerminalLimit, pkgAgent.ErrorLimit, "planning_budget", err)
		}
	}
	if err := validateInitialPlan(plan, request.Limits()); err != nil {
		reason := pkgAgent.TerminalPlanningFailure
		kind := pkgAgent.ErrorPlanning
		if errors.Is(err, pkgAgent.ErrLimitExceeded) {
			reason = pkgAgent.TerminalLimit
			kind = pkgAgent.ErrorLimit
		}
		return current.fail(reason, kind, "invalid_plan", err)
	}
	if err := current.transition(pkgAgent.SessionRunning, &plan, bundle.Generation()); err != nil {
		return current.fail(pkgAgent.TerminalInternalFailure, pkgAgent.ErrorInternal, "running_transition", err)
	}
	publishAgentEvent(runtime.events, pkgAgent.EventPlanCreated, sessionEventPayload(current.snapshotValue(), 0, pkgAgent.EventFailureNone))
	if runtime.loop == nil {
		return current.finish("", pkgAgent.TerminalBlocked)
	}
	content, terminal, err := runtime.loop.Run(ctx, current, request, bundle)
	if err != nil {
		return current.fail(terminal, kindForTerminal(terminal), "loop_failed", err)
	}
	return current.finish(content, terminal)
}

func (runtime *Runtime) Session(run pkgAgent.RunID) (pkgAgent.SessionSnapshot, bool) {
	runtime.mu.RLock()
	current, exists := runtime.sessions[run]
	runtime.mu.RUnlock()
	if !exists {
		return pkgAgent.SessionSnapshot{}, false
	}
	return current.snapshotValue(), true
}

type measuredPlanner interface {
	pkgAgent.Planner
	Target() (provider.ID, string)
	PlanMeasured(context.Context, pkgAgent.PlanningRequest) (pkgAgent.Plan, provider.Usage, error)
}

func (runtime *Runtime) callPlanner(
	ctx context.Context,
	current *session,
	planner pkgAgent.Planner,
	planningRequest pkgAgent.PlanningRequest,
	runRequest pkgAgent.RunRequest,
	bundle pkgContext.ContextBundle,
) (plan pkgAgent.Plan, usage provider.Usage, modelBacked bool, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("planner panicked: %w", pkgAgent.ErrPlanningFailed)
		}
	}()
	measured, ok := planner.(measuredPlanner)
	if !ok {
		plan, err = planner.Plan(ctx, planningRequest)
		return plan, provider.Usage{}, false, err
	}
	providerID, model := measured.Target()
	if providerID != runRequest.Provider() || model != runRequest.Model() {
		return pkgAgent.Plan{}, provider.Usage{}, true, fmt.Errorf("planner target differs from run target: %w", pkgAgent.ErrInvalidRequest)
	}
	if runtime.permissions == nil || typedNil(runtime.permissions) {
		return pkgAgent.Plan{}, provider.Usage{}, true, fmt.Errorf("model planner requires permission runtime: %w", pkgAgent.ErrPermissionDenied)
	}
	permission, permissionErr := planningPermission(runRequest, bundle)
	if permissionErr != nil {
		return pkgAgent.Plan{}, provider.Usage{}, true, permissionErr
	}
	decision, permissionErr := runtime.permissions.AuthorizeModel(ctx, permission, runRequest.Approver())
	if permissionErr != nil {
		return pkgAgent.Plan{}, provider.Usage{}, true, permissionErr
	}
	if decision.Kind() != pkgTool.DecisionAllow {
		return pkgAgent.Plan{}, provider.Usage{}, true, pkgAgent.ErrPermissionDenied
	}
	if err := current.consume(counterDelta{modelTurns: 1}); err != nil {
		return pkgAgent.Plan{}, provider.Usage{}, true, err
	}
	plan, usage, err = measured.PlanMeasured(ctx, planningRequest)
	return plan, usage, true, err
}

func validateInitialPlan(plan pkgAgent.Plan, limits pkgAgent.Limits) error {
	if err := plan.Validate(); err != nil || plan.Version() != 1 {
		return fmt.Errorf("initial plan must be valid at version 1: %w", pkgAgent.ErrInvalidPlan)
	}
	if len(plan.Steps()) > limits.MaxPlanSteps {
		return fmt.Errorf("plan contains %d steps with limit %d: %w", len(plan.Steps()), limits.MaxPlanSteps, pkgAgent.ErrLimitExceeded)
	}
	for _, step := range plan.Steps() {
		if step.Status() != pkgAgent.StepPending {
			return fmt.Errorf("initial plan step %q is not pending: %w", step.ID(), pkgAgent.ErrInvalidPlan)
		}
	}
	return nil
}

func bundleBytes(bundle pkgContext.ContextBundle) int {
	total := 0
	for _, section := range bundle.Sections() {
		total += len(section.Text)
	}
	return total
}

func planBytes(plan pkgAgent.Plan) int {
	total := 0
	for _, step := range plan.Steps() {
		total += len(step.ID()) + len(step.Objective()) + len(step.Reason())
		for _, dependency := range step.Dependencies() {
			total += len(dependency)
		}
	}
	return total
}

func reasonForError(ctx context.Context, fallback pkgAgent.TerminalReason) pkgAgent.TerminalReason {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return pkgAgent.TerminalDeadline
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return pkgAgent.TerminalCanceled
	}
	return fallback
}

func kindForError(ctx context.Context, fallback pkgAgent.ErrorKind) pkgAgent.ErrorKind {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return pkgAgent.ErrorDeadline
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return pkgAgent.ErrorCanceled
	}
	return fallback
}

func kindForTerminal(reason pkgAgent.TerminalReason) pkgAgent.ErrorKind {
	switch reason {
	case pkgAgent.TerminalDeadline:
		return pkgAgent.ErrorDeadline
	case pkgAgent.TerminalCanceled:
		return pkgAgent.ErrorCanceled
	case pkgAgent.TerminalLimit:
		return pkgAgent.ErrorLimit
	case pkgAgent.TerminalPermissionDenied:
		return pkgAgent.ErrorPermission
	case pkgAgent.TerminalProviderFailure:
		return pkgAgent.ErrorProvider
	case pkgAgent.TerminalToolFailure:
		return pkgAgent.ErrorTool
	case pkgAgent.TerminalPlanningFailure:
		return pkgAgent.ErrorPlanning
	default:
		return pkgAgent.ErrorInternal
	}
}

func runContextError(request pkgAgent.RunRequest, err error) error {
	kind := pkgAgent.ErrorCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		kind = pkgAgent.ErrorDeadline
	}
	return pkgAgent.NewRunError(kind, request.Run(), request.Agent(), "context_done", err)
}
