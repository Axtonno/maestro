package tool

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

var _ pkgTool.Runtime = (*Runtime)(nil)

type authorizer interface {
	Authorize(context.Context, pkgTool.PermissionRequest, pkgTool.Approver) (pkgTool.Decision, error)
}

type denyAuthorizer struct{}

func (denyAuthorizer) Authorize(
	context.Context,
	pkgTool.PermissionRequest,
	pkgTool.Approver,
) (pkgTool.Decision, error) {
	return pkgTool.NewDecision(pkgTool.DecisionDeny, "no_policy_engine", pkgTool.DenyTerminal, "")
}

type Runtime struct {
	mu sync.RWMutex

	tools                   map[pkgTool.ID]pkgTool.Tool
	descriptors             map[pkgTool.ID]pkgTool.Descriptor
	names                   map[pkgTool.Name]pkgTool.ID
	policies                map[pkgTool.PolicyID]pkgTool.Policy
	grants                  map[grantKey]struct{}
	authorizer              authorizer
	executionObservers      []func(pkgTool.PreparedInvocation)
	registrationInvalidator func()
	events                  pkgRuntime.EventBus
}

func NewRuntime() *Runtime {
	runtime := newRuntime(nil)
	runtime.authorizer = runtime
	return runtime
}

func NewRuntimeWithEventBus(events pkgRuntime.EventBus) *Runtime {
	runtime := NewRuntime()
	runtime.events = events
	return runtime
}

func newRuntime(authorizer authorizer) *Runtime {
	if authorizer == nil {
		authorizer = denyAuthorizer{}
	}
	return &Runtime{
		tools: make(map[pkgTool.ID]pkgTool.Tool), descriptors: make(map[pkgTool.ID]pkgTool.Descriptor),
		names: make(map[pkgTool.Name]pkgTool.ID), policies: make(map[pkgTool.PolicyID]pkgTool.Policy),
		grants:     make(map[grantKey]struct{}),
		authorizer: authorizer,
	}
}

func (runtime *Runtime) Register(candidate pkgTool.Tool) error {
	if err := pkgTool.ValidateTool(candidate); err != nil {
		return fmt.Errorf("register tool: %w", err)
	}
	descriptor := candidate.Descriptor()

	runtime.mu.Lock()
	if _, exists := runtime.tools[descriptor.ID()]; exists {
		runtime.mu.Unlock()
		return fmt.Errorf("register tool %q: %w", descriptor.ID(), pkgTool.ErrAlreadyRegistered)
	}
	if owner, exists := runtime.names[descriptor.Name()]; exists {
		runtime.mu.Unlock()
		return fmt.Errorf(
			"register tool %q: provider name %q belongs to %q: %w",
			descriptor.ID(), descriptor.Name(), owner, pkgTool.ErrAlreadyRegistered,
		)
	}
	runtime.tools[descriptor.ID()] = candidate
	runtime.descriptors[descriptor.ID()] = descriptor
	runtime.names[descriptor.Name()] = descriptor.ID()
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

func (runtime *Runtime) Descriptors() []pkgTool.Descriptor {
	runtime.mu.RLock()
	descriptors := make([]pkgTool.Descriptor, 0, len(runtime.descriptors))
	for _, descriptor := range runtime.descriptors {
		descriptors = append(descriptors, descriptor)
	}
	runtime.mu.RUnlock()
	slices.SortFunc(descriptors, func(left, right pkgTool.Descriptor) int {
		return left.ID().Compare(right.ID())
	})
	return descriptors
}

func (runtime *Runtime) RegisterPolicy(policy pkgTool.Policy) error {
	if err := pkgTool.ValidatePolicy(policy); err != nil {
		return fmt.Errorf("register policy: %w", err)
	}
	id := policy.ID()
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if _, exists := runtime.policies[id]; exists {
		return fmt.Errorf("register policy %q: %w", id, pkgTool.ErrPolicyAlreadyRegistered)
	}
	runtime.policies[id] = policy
	return nil
}

func (runtime *Runtime) Policies() []pkgTool.PolicyID {
	runtime.mu.RLock()
	ids := make([]pkgTool.PolicyID, 0, len(runtime.policies))
	for id := range runtime.policies {
		ids = append(ids, id)
	}
	runtime.mu.RUnlock()
	slices.Sort(ids)
	return ids
}

// ObserveExecutionStart registers a best-effort callback invoked after allow
// and immediately before Execute. Callbacks run outside runtime locks and do
// not receive tool output.
func (runtime *Runtime) ObserveExecutionStart(observer func(pkgTool.PreparedInvocation)) error {
	if observer == nil {
		return fmt.Errorf("execution observer is nil: %w", pkgTool.ErrInvalidTool)
	}
	runtime.mu.Lock()
	runtime.executionObservers = append(runtime.executionObservers, observer)
	runtime.mu.Unlock()
	return nil
}

func (runtime *Runtime) Invoke(ctx context.Context, request pkgTool.ExecutionRequest) (result pkgTool.Result, err error) {
	started := time.Now()
	observedActions := 0
	observedDecision := pkgTool.DecisionKind("")
	defer func() {
		if err != nil {
			runtime.publish(pkgTool.EventInvocationFailed, toolEventPayload(request, result, observedActions, observedDecision, time.Since(started), classifyToolEventFailure(err)))
		}
	}()
	if ctx == nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "nil_context", pkgTool.ErrInvalidExecutionRequest)
	}
	if err := request.Validate(); err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "invalid_request", err)
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Result{}, contextExecutionError(request, err)
	}

	invocation := request.Invocation()
	runtime.mu.RLock()
	candidate, exists := runtime.tools[invocation.Tool()]
	descriptor := runtime.descriptors[invocation.Tool()]
	runtime.mu.RUnlock()
	if !exists {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorNotFound, "tool_not_found", pkgTool.ErrNotFound)
	}

	executionContext, cancel := context.WithTimeout(ctx, request.Limits().MaxDuration)
	defer cancel()
	prepared, err := prepare(executionContext, candidate, invocation)
	if err != nil {
		return pkgTool.Result{}, classifyExecutionError(request, "prepare_failed", err)
	}
	if err := validatePrepared(descriptor, invocation, prepared); err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "invalid_prepared_invocation", err)
	}
	observedActions = len(prepared.Actions())
	runtime.publish(pkgTool.EventInvocationPrepared, toolEventPayload(request, pkgTool.Result{}, len(prepared.Actions()), "", time.Since(started), pkgTool.EventFailureNone))
	permission, err := pkgTool.NewToolPermissionRequest(request.Policy(), prepared)
	if err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "invalid_permission_request", err)
	}
	decision, err := runtime.authorizer.Authorize(executionContext, permission, request.Approver())
	if err != nil {
		return pkgTool.Result{}, classifyExecutionError(request, "authorization_failed", err)
	}
	if err := decision.Validate(); err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "invalid_decision", err)
	}
	observedDecision = decision.Kind()
	runtime.publish(pkgTool.EventInvocationAuthorized, toolEventPayload(request, pkgTool.Result{}, len(prepared.Actions()), decision.Kind(), time.Since(started), pkgTool.EventFailureNone))
	permissionFailure := pkgTool.EventFailureNone
	if decision.Kind() == pkgTool.DecisionDeny {
		permissionFailure = pkgTool.EventFailureDenied
	}
	runtime.publish(pkgTool.EventPermissionDecided, toolEventPayload(request, pkgTool.Result{}, len(prepared.Actions()), decision.Kind(), time.Since(started), permissionFailure))
	if decision.Kind() != pkgTool.DecisionAllow {
		disposition := decision.Disposition()
		if decision.Kind() == pkgTool.DecisionPrompt {
			disposition = pkgTool.DenyTerminal
		}
		denied, deniedErr := pkgTool.NewResult(
			pkgTool.ResultDenied, "", "", decision.Reason(), 0, false, disposition,
		)
		if deniedErr != nil {
			return pkgTool.Result{}, executionError(request, pkgTool.ErrorInternal, "invalid_denied_result", deniedErr)
		}
		runtime.publish(pkgTool.EventInvocationCompleted, toolEventPayload(request, denied, len(prepared.Actions()), decision.Kind(), time.Since(started), pkgTool.EventFailureDenied))
		return denied, nil
	}

	permit := runtime.issue(permission)
	runtime.notifyExecutionStart(prepared)
	result, err = runtime.execute(executionContext, candidate, prepared, permission.Fingerprint(), permit)
	if err != nil {
		return pkgTool.Result{}, classifyExecutionError(request, "execute_failed", err)
	}
	result, err = limitResult(request, result)
	if err != nil {
		return pkgTool.Result{}, err
	}
	runtime.publish(pkgTool.EventInvocationCompleted, toolEventPayload(request, result, len(prepared.Actions()), decision.Kind(), time.Since(started), pkgTool.EventFailureNone))
	return result, nil
}

func (runtime *Runtime) publish(topic string, payload pkgTool.EventPayload) {
	if runtime.events == nil {
		return
	}
	func() {
		defer func() { _ = recover() }()
		_ = runtime.events.Publish(pkgTool.Event{Topic: topic, Data: payload})
	}()
}

func toolEventPayload(request pkgTool.ExecutionRequest, result pkgTool.Result, actionCount int, decision pkgTool.DecisionKind, duration time.Duration, failure pkgTool.EventFailure) pkgTool.EventPayload {
	invocation := request.Invocation()
	return pkgTool.EventPayload{
		Run: invocation.Run(), Tool: invocation.Tool(), Call: invocation.Call(),
		ActionCount: actionCount, Decision: decision, Outcome: result.Outcome(),
		Disposition: result.Disposition(), Truncated: result.Truncated(),
		DurationMillis: duration.Milliseconds(), Failure: failure,
	}
}

func classifyToolEventFailure(err error) pkgTool.EventFailure {
	switch {
	case err == nil:
		return pkgTool.EventFailureNone
	case errors.Is(err, context.DeadlineExceeded):
		return pkgTool.EventFailureDeadline
	case errors.Is(err, context.Canceled):
		return pkgTool.EventFailureCanceled
	case errors.Is(err, pkgTool.ErrPermissionDenied):
		return pkgTool.EventFailureDenied
	case errors.Is(err, pkgTool.ErrLimitExceeded):
		return pkgTool.EventFailureLimit
	case errors.Is(err, pkgTool.ErrInvalidInvocation), errors.Is(err, pkgTool.ErrInvalidExecutionRequest):
		return pkgTool.EventFailureInvalid
	case errors.Is(err, pkgTool.ErrExecutionFailed):
		return pkgTool.EventFailureTool
	default:
		return pkgTool.EventFailureInternal
	}
}

func (runtime *Runtime) notifyExecutionStart(prepared pkgTool.PreparedInvocation) {
	runtime.mu.RLock()
	observers := slices.Clone(runtime.executionObservers)
	runtime.mu.RUnlock()
	for _, observer := range observers {
		func() {
			defer func() { _ = recover() }()
			observer(prepared)
		}()
	}
}

func (runtime *Runtime) AuthorizeModel(
	ctx context.Context,
	request pkgTool.PermissionRequest,
	approver pkgTool.Approver,
) (decision pkgTool.Decision, err error) {
	started := time.Now()
	defer func() {
		failure := classifyToolEventFailure(err)
		if err == nil && decision.Kind() == pkgTool.DecisionDeny {
			failure = pkgTool.EventFailureDenied
		}
		runtime.publish(pkgTool.EventPermissionDecided, pkgTool.EventPayload{
			Run: request.Run(), ActionCount: len(request.Actions()), Decision: decision.Kind(),
			Disposition: decision.Disposition(), DurationMillis: time.Since(started).Milliseconds(), Failure: failure,
		})
	}()
	if ctx == nil || request.Validate() != nil || request.Subject() != pkgTool.PermissionSubjectModel {
		return pkgTool.Decision{}, pkgTool.ErrInvalidPermissionRequest
	}
	decision, err = runtime.authorizer.Authorize(ctx, request, approver)
	if err != nil {
		return pkgTool.Decision{}, err
	}
	if err := decision.Validate(); err != nil {
		return pkgTool.Decision{}, err
	}
	return decision, nil
}

func prepare(ctx context.Context, candidate pkgTool.Tool, invocation pkgTool.Invocation) (prepared pkgTool.PreparedInvocation, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("tool prepare panicked: %w", pkgTool.ErrExecutionFailed)
		}
	}()
	return candidate.Prepare(ctx, invocation)
}

func validatePrepared(
	descriptor pkgTool.Descriptor,
	invocation pkgTool.Invocation,
	prepared pkgTool.PreparedInvocation,
) error {
	if err := prepared.Validate(); err != nil {
		return err
	}
	source := prepared.Invocation()
	if source.Tool() != invocation.Tool() || source.Call() != invocation.Call() || source.Run() != invocation.Run() {
		return fmt.Errorf("prepared invocation changed tool, call, or run: %w", pkgTool.ErrInvalidPreparedInvocation)
	}
	if prepared.Version() != descriptor.Version() {
		return fmt.Errorf("prepared version %q differs from descriptor %q: %w", prepared.Version(), descriptor.Version(), pkgTool.ErrInvalidPreparedInvocation)
	}
	declared := make(map[pkgTool.Effect]struct{}, len(descriptor.Effects()))
	for _, effect := range descriptor.Effects() {
		declared[effect] = struct{}{}
	}
	for _, action := range prepared.Actions() {
		if _, exists := declared[action.Effect()]; !exists {
			return fmt.Errorf("prepared effect %q was not declared: %w", action.Effect(), pkgTool.ErrInvalidPreparedInvocation)
		}
	}
	return nil
}

func limitResult(request pkgTool.ExecutionRequest, result pkgTool.Result) (pkgTool.Result, error) {
	if err := result.Validate(); err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInvalid, "invalid_tool_result", err)
	}
	limits := request.Limits()
	if result.ItemCount() > limits.MaxItems {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorLimit, "item_limit", pkgTool.ErrLimitExceeded)
	}
	content := result.Content()
	if len(content) <= limits.MaxOutputBytes {
		return result, nil
	}
	content = content[:limits.MaxOutputBytes]
	for !utf8.ValidString(content) {
		content = content[:len(content)-1]
	}
	limited, err := pkgTool.NewResult(
		result.Outcome(), content, result.MediaType(), result.Reason(),
		result.ItemCount(), true, result.Disposition(),
	)
	if err != nil {
		return pkgTool.Result{}, executionError(request, pkgTool.ErrorInternal, "truncate_failed", err)
	}
	return limited, nil
}

func executionError(request pkgTool.ExecutionRequest, kind pkgTool.ErrorKind, reason string, cause error) error {
	invocation := request.Invocation()
	return pkgTool.NewExecutionError(kind, invocation.Run(), invocation.Tool(), invocation.Call(), reason, cause)
}

func contextExecutionError(request pkgTool.ExecutionRequest, err error) error {
	kind := pkgTool.ErrorCanceled
	reason := "canceled"
	if errors.Is(err, context.DeadlineExceeded) {
		kind = pkgTool.ErrorDeadline
		reason = "deadline_exceeded"
	}
	return executionError(request, kind, reason, err)
}

func classifyExecutionError(request pkgTool.ExecutionRequest, reason string, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return contextExecutionError(request, err)
	}
	return executionError(request, pkgTool.ErrorExecution, reason, errors.Join(pkgTool.ErrExecutionFailed, err))
}
