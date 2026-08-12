package tool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestStaticPolicyIsExactDefaultDenyAndAtomic(t *testing.T) {
	allow := policyDecision(t, pkgTool.DecisionAllow, "inspect_allowed", "", pkgTool.GrantRun)
	deny := policyDecision(t, pkgTool.DecisionDeny, "network_denied", pkgTool.DenyRecoverable, "")
	inspectRule := policyRule(t, pkgTool.EffectWorkspaceInspect, "main.go", "workspace", allow)
	networkRule := policyRule(t, pkgTool.EffectNetworkAccess, "example.invalid:443", "", deny)
	policy, err := NewStaticPolicy("policy.exact", []pkgTool.Rule{inspectRule, networkRule})
	if err != nil {
		t.Fatal(err)
	}

	request := permissionWithActions(t, "policy.exact", []pkgTool.Action{
		policyAction(t, pkgTool.EffectWorkspaceInspect, "main.go", "workspace"),
		policyAction(t, pkgTool.EffectNetworkAccess, "example.invalid:443", ""),
	})
	decision, err := policy.Decide(context.Background(), request)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny || decision.Disposition() != pkgTool.DenyRecoverable {
		t.Fatalf("deny did not atomically dominate: %#v %v", decision, err)
	}

	unmatched := permissionWithActions(t, "policy.exact", []pkgTool.Action{
		policyAction(t, pkgTool.EffectWorkspaceInspect, "main.go/child", "workspace"),
	})
	decision, err = policy.Decide(context.Background(), unmatched)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny || decision.Reason() != "no_matching_rule" {
		t.Fatalf("unmatched action was not default-denied: %#v %v", decision, err)
	}
}

func TestStaticPolicyRejectsDuplicateMatchers(t *testing.T) {
	allow := policyDecision(t, pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantOneShot)
	rule := policyRule(t, pkgTool.EffectWorkspaceInspect, "main.go", "workspace", allow)
	if _, err := NewStaticPolicy("policy.duplicate", []pkgTool.Rule{rule, rule}); !errors.Is(err, pkgTool.ErrInvalidPolicy) {
		t.Fatalf("expected duplicate rule rejection, got %v", err)
	}
}

func TestRuntimeUsesExactPolicyAndNoImplicitDefault(t *testing.T) {
	runtime := NewRuntime()
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	if err := runtime.Register(candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024)); !errors.Is(err, pkgTool.ErrPolicyNotFound) {
		t.Fatalf("expected exact policy lookup failure, got %v", err)
	}
	allow := policyDecision(t, pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantOneShot)
	policy, _ := NewStaticPolicy("policy.test", []pkgTool.Rule{
		policyRule(t, pkgTool.EffectWorkspaceInspect, "main.go", "workspace", allow),
	})
	if err := runtime.RegisterPolicy(policy); err != nil {
		t.Fatal(err)
	}
	result, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024))
	if err != nil || result.Outcome() != pkgTool.ResultSuccess || candidate.executions.Load() != 1 {
		t.Fatalf("authorized execution failed: %#v %v executions=%d", result, err, candidate.executions.Load())
	}
}

func TestPromptRequiresApproverAndHonorsDisposition(t *testing.T) {
	prompt := policyDecision(t, pkgTool.DecisionPrompt, "approval_required", "", "")
	policy, _ := NewStaticPolicy("policy.prompt", []pkgTool.Rule{
		policyRule(t, pkgTool.EffectModelInvoke, "ollama/model", "", prompt),
	})
	runtime := NewRuntime()
	_ = runtime.RegisterPolicy(policy)
	request := modelPermission(t, "policy.prompt")

	decision, err := runtime.AuthorizeModel(context.Background(), request, nil)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny || decision.Disposition() != pkgTool.DenyTerminal {
		t.Fatalf("prompt without approver was not terminal deny: %#v %v", decision, err)
	}
	approver := &fixtureApprover{approval: policyApproval(t, pkgTool.ApprovalDeny, "user_denied", pkgTool.DenyRecoverable, "")}
	decision, err = runtime.AuthorizeModel(context.Background(), request, approver)
	if err != nil || decision.Kind() != pkgTool.DecisionDeny || decision.Disposition() != pkgTool.DenyRecoverable {
		t.Fatalf("approver deny was not preserved: %#v %v", decision, err)
	}
	approver.approval = policyApproval(t, pkgTool.ApprovalAllow, "user_allowed", "", pkgTool.GrantOneShot)
	decision, err = runtime.AuthorizeModel(context.Background(), request, approver)
	if err != nil || decision.Kind() != pkgTool.DecisionAllow || decision.Scope() != pkgTool.GrantOneShot {
		t.Fatalf("approver allow was not preserved: %#v %v", decision, err)
	}
}

func TestRunGrantIsFingerprintScopedAndOneShotIsNotCached(t *testing.T) {
	runtime := NewRuntime()
	runPolicy := &countingPolicy{id: "policy.run", decision: policyDecision(t, pkgTool.DecisionAllow, "run_allowed", "", pkgTool.GrantRun)}
	onePolicy := &countingPolicy{id: "policy.one", decision: policyDecision(t, pkgTool.DecisionAllow, "one_allowed", "", pkgTool.GrantOneShot)}
	_ = runtime.RegisterPolicy(runPolicy)
	_ = runtime.RegisterPolicy(onePolicy)
	runRequest := modelPermission(t, "policy.run")
	for range 2 {
		if _, err := runtime.AuthorizeModel(context.Background(), runRequest, nil); err != nil {
			t.Fatal(err)
		}
	}
	if runPolicy.calls.Load() != 1 {
		t.Fatalf("run grant was not reused: calls=%d", runPolicy.calls.Load())
	}
	oneRequest := modelPermission(t, "policy.one")
	for range 2 {
		if _, err := runtime.AuthorizeModel(context.Background(), oneRequest, nil); err != nil {
			t.Fatal(err)
		}
	}
	if onePolicy.calls.Load() != 2 {
		t.Fatalf("one-shot decision was cached: calls=%d", onePolicy.calls.Load())
	}
}

func TestPolicyAndApproverCallbacksRunOutsideLockAndPanicIsContained(t *testing.T) {
	runtime := NewRuntime()
	entered := make(chan struct{})
	release := make(chan struct{})
	policy := &blockingPolicy{id: "policy.blocking", entered: entered, release: release, decision: policyDecision(t, pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantOneShot)}
	_ = runtime.RegisterPolicy(policy)
	done := make(chan error, 1)
	go func() {
		_, err := runtime.AuthorizeModel(context.Background(), modelPermission(t, "policy.blocking"), nil)
		done <- err
	}()
	<-entered
	registered := make(chan error, 1)
	go func() {
		other, _ := NewStaticPolicy("policy.other", nil)
		registered <- runtime.RegisterPolicy(other)
	}()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("policy callback executed under registry lock")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}

	panicPolicy := &panicPolicy{id: "policy.panic"}
	_ = runtime.RegisterPolicy(panicPolicy)
	if _, err := runtime.AuthorizeModel(context.Background(), modelPermission(t, "policy.panic"), nil); !errors.Is(err, pkgTool.ErrInvalidPolicy) {
		t.Fatalf("expected policy panic containment, got %v", err)
	}
}

func policyDecision(t *testing.T, kind pkgTool.DecisionKind, reason string, disposition pkgTool.DenyDisposition, scope pkgTool.GrantScope) pkgTool.Decision {
	t.Helper()
	decision, err := pkgTool.NewDecision(kind, reason, disposition, scope)
	if err != nil {
		t.Fatal(err)
	}
	return decision
}

func policyApproval(t *testing.T, kind pkgTool.ApprovalKind, reason string, disposition pkgTool.DenyDisposition, scope pkgTool.GrantScope) pkgTool.Approval {
	t.Helper()
	approval, err := pkgTool.NewApproval(kind, reason, disposition, scope)
	if err != nil {
		t.Fatal(err)
	}
	return approval
}

func policyRule(t *testing.T, effect pkgTool.Effect, resource string, workspace contextengine.WorkspaceID, decision pkgTool.Decision) pkgTool.Rule {
	t.Helper()
	rule, err := pkgTool.NewRule(effect, resource, workspace, decision)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func policyAction(t *testing.T, effect pkgTool.Effect, resource string, workspace contextengine.WorkspaceID) pkgTool.Action {
	t.Helper()
	action, err := pkgTool.NewAction(effect, resource, workspace)
	if err != nil {
		t.Fatal(err)
	}
	return action
}

func permissionWithActions(t *testing.T, policy pkgTool.PolicyID, actions []pkgTool.Action) pkgTool.PermissionRequest {
	t.Helper()
	invocation := invocation(t, "workspace.composite")
	prepared, err := pkgTool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), actions)
	if err != nil {
		t.Fatal(err)
	}
	request, err := pkgTool.NewToolPermissionRequest(policy, prepared)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func modelPermission(t *testing.T, policy pkgTool.PolicyID) pkgTool.PermissionRequest {
	t.Helper()
	target, _ := pkgTool.NewModelTarget("ollama", "model")
	request, err := pkgTool.NewModelPermissionRequest(policy, "run-1", target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

type fixtureApprover struct {
	approval pkgTool.Approval
	err      error
	calls    atomic.Int32
}

func (approver *fixtureApprover) Approve(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
	approver.calls.Add(1)
	return approver.approval, approver.err
}

type countingPolicy struct {
	id       pkgTool.PolicyID
	decision pkgTool.Decision
	calls    atomic.Int32
}

func (policy *countingPolicy) ID() pkgTool.PolicyID { return policy.id }
func (policy *countingPolicy) Decide(context.Context, pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	policy.calls.Add(1)
	return policy.decision, nil
}

type blockingPolicy struct {
	id       pkgTool.PolicyID
	entered  chan struct{}
	release  chan struct{}
	decision pkgTool.Decision
}

func (policy *blockingPolicy) ID() pkgTool.PolicyID { return policy.id }
func (policy *blockingPolicy) Decide(ctx context.Context, _ pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	close(policy.entered)
	select {
	case <-policy.release:
		return policy.decision, nil
	case <-ctx.Done():
		return pkgTool.Decision{}, ctx.Err()
	}
}

type panicPolicy struct{ id pkgTool.PolicyID }

func (policy *panicPolicy) ID() pkgTool.PolicyID { return policy.id }
func (*panicPolicy) Decide(context.Context, pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	panic("boom")
}
