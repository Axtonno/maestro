package tool

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestCatalogRegistrationOrderingAndCollisions(t *testing.T) {
	runtime := NewRuntime()
	second := newFixtureTool(t, "workspace.write", "workspace_write", pkgTool.EffectWorkspaceMutate)
	first := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	if err := runtime.Register(second); err != nil {
		t.Fatalf("register second: %v", err)
	}
	if err := runtime.Register(first); err != nil {
		t.Fatalf("register first: %v", err)
	}
	descriptors := runtime.Descriptors()
	if len(descriptors) != 2 || descriptors[0].ID() != "workspace.read" || descriptors[1].ID() != "workspace.write" {
		t.Fatalf("unexpected descriptors: %#v", descriptors)
	}
	if err := runtime.Register(first); !errors.Is(err, pkgTool.ErrAlreadyRegistered) {
		t.Fatalf("expected ID collision, got %v", err)
	}
	sameName := newFixtureTool(t, "workspace.other", "workspace_read", pkgTool.EffectWorkspaceInspect)
	if err := runtime.Register(sameName); !errors.Is(err, pkgTool.ErrAlreadyRegistered) {
		t.Fatalf("expected name collision, got %v", err)
	}
}

func TestInvokeRunsPrepareAuthorizeExecuteWithoutCatalogLock(t *testing.T) {
	authorizer := &fixtureAuthorizer{decision: allowDecision(t)}
	runtime := newRuntime(authorizer)
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	prepareEntered := make(chan struct{})
	releasePrepare := make(chan struct{})
	candidate.prepare = func(ctx context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
		close(prepareEntered)
		select {
		case <-releasePrepare:
		case <-ctx.Done():
			return pkgTool.PreparedInvocation{}, ctx.Err()
		}
		action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceInspect, "main.go", "workspace")
		return pkgTool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), []pkgTool.Action{action})
	}
	if err := runtime.Register(candidate); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024))
		done <- err
	}()
	<-prepareEntered
	registered := make(chan error, 1)
	go func() {
		registered <- runtime.Register(newFixtureTool(t, "workspace.write", "workspace_write", pkgTool.EffectWorkspaceMutate))
	}()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatalf("register while Prepare blocked: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("catalog lock held during Prepare")
	}
	close(releasePrepare)
	if err := <-done; err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if candidate.executions.Load() != 1 || authorizer.calls.Load() != 1 {
		t.Fatalf("unexpected calls: execute=%d authorize=%d", candidate.executions.Load(), authorizer.calls.Load())
	}
}

func TestInvokeDefaultDenyAndExplicitDenyNeverExecute(t *testing.T) {
	tests := []struct {
		name        string
		runtime     *Runtime
		disposition pkgTool.DenyDisposition
	}{
		{name: "default deny", runtime: NewRuntime(), disposition: pkgTool.DenyTerminal},
		{name: "recoverable deny", runtime: newRuntime(&fixtureAuthorizer{decision: denyDecision(t, pkgTool.DenyRecoverable)}), disposition: pkgTool.DenyRecoverable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
			if err := test.runtime.Register(candidate); err != nil {
				t.Fatal(err)
			}
			result, err := test.runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024))
			if err != nil {
				t.Fatalf("invoke denied: %v", err)
			}
			if result.Outcome() != pkgTool.ResultDenied || result.Disposition() != test.disposition || candidate.executions.Load() != 0 {
				t.Fatalf("unexpected denied result/execution: %#v executions=%d", result, candidate.executions.Load())
			}
		})
	}
}

func TestPermitIsIssuerBoundFingerprintBoundAndOneShot(t *testing.T) {
	runtime := NewRuntime()
	other := NewRuntime()
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	invocation := invocation(t, "workspace.read")
	action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceInspect, "main.go", "workspace")
	prepared, _ := pkgTool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), []pkgTool.Action{action})
	permission, _ := pkgTool.NewToolPermissionRequest("policy.test", prepared)
	permit := runtime.issue(permission)
	if _, err := other.execute(context.Background(), candidate, prepared, permission.Fingerprint(), permit); !errors.Is(err, pkgTool.ErrPermissionDenied) {
		t.Fatalf("expected foreign issuer rejection, got %v", err)
	}
	if _, err := runtime.execute(context.Background(), candidate, prepared, permission.Fingerprint(), permit); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := runtime.execute(context.Background(), candidate, prepared, permission.Fingerprint(), permit); !errors.Is(err, pkgTool.ErrPermissionDenied) {
		t.Fatalf("expected replay rejection, got %v", err)
	}
	changedAction, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceInspect, "other.go", "workspace")
	changed, _ := pkgTool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), []pkgTool.Action{changedAction})
	if _, err := runtime.execute(context.Background(), candidate, changed, permission.Fingerprint(), runtime.issue(permission)); !errors.Is(err, pkgTool.ErrPermissionDenied) {
		t.Fatalf("expected fingerprint rejection, got %v", err)
	}
	otherPermission, _ := pkgTool.NewToolPermissionRequest("policy.other", prepared)
	if _, err := runtime.execute(context.Background(), candidate, prepared, otherPermission.Fingerprint(), runtime.issue(permission)); !errors.Is(err, pkgTool.ErrPermissionDenied) {
		t.Fatalf("expected permission fingerprint rejection, got %v", err)
	}
}

func TestInvokeValidatesPreparedOutputAndDeclaredEffects(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(context.Context, pkgTool.Invocation) (pkgTool.PreparedInvocation, error)
	}{
		{name: "changed identity", prepare: func(_ context.Context, source pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
			changed, _ := pkgTool.NewInvocation(source.Tool(), "changed-call", source.Run(), source.Arguments())
			action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceInspect, "main.go", "workspace")
			return pkgTool.NewPreparedInvocation(changed, "1", source.Arguments(), []pkgTool.Action{action})
		}},
		{name: "undeclared effect", prepare: func(_ context.Context, source pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
			action, _ := pkgTool.NewAction(pkgTool.EffectNetworkAccess, "example.invalid:443", "")
			return pkgTool.NewPreparedInvocation(source, "1", source.Arguments(), []pkgTool.Action{action})
		}},
		{name: "wrong version", prepare: func(_ context.Context, source pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
			action, _ := pkgTool.NewAction(pkgTool.EffectWorkspaceInspect, "main.go", "workspace")
			return pkgTool.NewPreparedInvocation(source, "2", source.Arguments(), []pkgTool.Action{action})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime := newRuntime(&fixtureAuthorizer{decision: allowDecision(t)})
			candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
			candidate.prepare = test.prepare
			_ = runtime.Register(candidate)
			_, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024))
			if !errors.Is(err, pkgTool.ErrInvalidInvocation) && !strings.Contains(err.Error(), "invalid_prepared_invocation") {
				t.Fatalf("expected prepared validation failure, got %v", err)
			}
			if candidate.executions.Load() != 0 {
				t.Fatal("invalid prepared invocation executed")
			}
		})
	}
}

func TestInvokeAppliesOutputAndItemLimits(t *testing.T) {
	runtime := newRuntime(&fixtureAuthorizer{decision: allowDecision(t)})
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	candidate.execute = func(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
		return pkgTool.NewResult(pkgTool.ResultSuccess, "abc€xyz", "text/plain", "completed", 1, false, "")
	}
	_ = runtime.Register(candidate)
	result, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 5))
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if result.Content() != "abc" || !result.Truncated() {
		t.Fatalf("UTF-8 output was not safely truncated: %q truncated=%v", result.Content(), result.Truncated())
	}

	candidate.execute = func(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
		return pkgTool.NewResult(pkgTool.ResultSuccess, "ok", "text/plain", "completed", 11, false, "")
	}
	_, err = runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024))
	if !errors.Is(err, pkgTool.ErrLimitExceeded) {
		t.Fatalf("expected item limit, got %v", err)
	}
}

func TestInvokeContainsPanicAndHonorsCancellation(t *testing.T) {
	runtime := newRuntime(&fixtureAuthorizer{decision: allowDecision(t)})
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	candidate.execute = func(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) { panic("boom") }
	_ = runtime.Register(candidate)
	if _, err := runtime.Invoke(context.Background(), executionRequest(t, "workspace.read", 1024)); !errors.Is(err, pkgTool.ErrExecutionFailed) {
		t.Fatalf("expected contained panic, got %v", err)
	}

	candidate.execute = func(ctx context.Context, _ pkgTool.PreparedInvocation) (pkgTool.Result, error) {
		<-ctx.Done()
		return pkgTool.Result{}, ctx.Err()
	}
	request := executionRequest(t, "workspace.read", 1024)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runtime.Invoke(ctx, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestCatalogConcurrentRegistrationHasOneWinner(t *testing.T) {
	runtime := NewRuntime()
	candidate := newFixtureTool(t, "workspace.read", "workspace_read", pkgTool.EffectWorkspaceInspect)
	const contenders = 32
	var successes atomic.Int32
	var wait sync.WaitGroup
	for range contenders {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if runtime.Register(candidate) == nil {
				successes.Add(1)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || len(runtime.Descriptors()) != 1 {
		t.Fatalf("unexpected concurrent registration: successes=%d descriptors=%d", successes.Load(), len(runtime.Descriptors()))
	}
}

type fixtureTool struct {
	descriptor pkgTool.Descriptor
	prepare    func(context.Context, pkgTool.Invocation) (pkgTool.PreparedInvocation, error)
	execute    func(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error)
	executions atomic.Int32
}

func newFixtureTool(t *testing.T, id pkgTool.ID, name pkgTool.Name, effect pkgTool.Effect) *fixtureTool {
	t.Helper()
	descriptor, err := pkgTool.NewDescriptor(id, name, "1", "Fixture tool.", json.RawMessage(`{"type":"object"}`), []pkgTool.Effect{effect})
	if err != nil {
		t.Fatal(err)
	}
	candidate := &fixtureTool{descriptor: descriptor}
	candidate.prepare = func(_ context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
		workspace := ""
		resource := "local"
		if effect == pkgTool.EffectWorkspaceInspect || effect == pkgTool.EffectWorkspaceMutate {
			workspace = "workspace"
			resource = "main.go"
		}
		action, _ := pkgTool.NewAction(effect, resource, contextengine.WorkspaceID(workspace))
		return pkgTool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), []pkgTool.Action{action})
	}
	candidate.execute = func(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
		return pkgTool.NewResult(pkgTool.ResultSuccess, "ok", "text/plain", "completed", 1, false, "")
	}
	return candidate
}

func (candidate *fixtureTool) Descriptor() pkgTool.Descriptor { return candidate.descriptor }
func (candidate *fixtureTool) Prepare(ctx context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
	return candidate.prepare(ctx, invocation)
}
func (candidate *fixtureTool) Execute(ctx context.Context, invocation pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	candidate.executions.Add(1)
	return candidate.execute(ctx, invocation)
}

type fixtureAuthorizer struct {
	decision pkgTool.Decision
	err      error
	calls    atomic.Int32
}

func (authorizer *fixtureAuthorizer) Authorize(context.Context, pkgTool.PermissionRequest, pkgTool.Approver) (pkgTool.Decision, error) {
	authorizer.calls.Add(1)
	return authorizer.decision, authorizer.err
}

func allowDecision(t *testing.T) pkgTool.Decision {
	t.Helper()
	decision, err := pkgTool.NewDecision(pkgTool.DecisionAllow, "fixture_allow", "", pkgTool.GrantOneShot)
	if err != nil {
		t.Fatal(err)
	}
	return decision
}

func denyDecision(t *testing.T, disposition pkgTool.DenyDisposition) pkgTool.Decision {
	t.Helper()
	decision, err := pkgTool.NewDecision(pkgTool.DecisionDeny, "fixture_deny", disposition, "")
	if err != nil {
		t.Fatal(err)
	}
	return decision
}

func executionRequest(t *testing.T, id pkgTool.ID, maxBytes int) pkgTool.ExecutionRequest {
	t.Helper()
	request, err := pkgTool.NewExecutionRequest(
		invocation(t, id), "policy.test", nil,
		pkgTool.ExecutionLimits{MaxDuration: time.Second, MaxOutputBytes: maxBytes, MaxItems: 10},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func invocation(t *testing.T, id pkgTool.ID) pkgTool.Invocation {
	t.Helper()
	invocation, err := pkgTool.NewInvocation(id, "call-1", "run-1", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	return invocation
}
