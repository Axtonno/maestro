package tool

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type grantKey struct {
	policy      pkgTool.PolicyID
	run         pkgTool.RunID
	fingerprint pkgTool.Fingerprint
}

func (runtime *Runtime) Authorize(
	ctx context.Context,
	request pkgTool.PermissionRequest,
	approver pkgTool.Approver,
) (decision pkgTool.Decision, err error) {
	if ctx == nil || request.Validate() != nil {
		return pkgTool.Decision{}, pkgTool.ErrInvalidPermissionRequest
	}
	if approver != nil && typedNil(approver) {
		return pkgTool.Decision{}, pkgTool.ErrInvalidApprover
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Decision{}, err
	}
	key := grantKey{policy: request.Policy(), run: request.Run(), fingerprint: request.Fingerprint()}
	runtime.mu.RLock()
	_, granted := runtime.grants[key]
	policy, exists := runtime.policies[request.Policy()]
	runtime.mu.RUnlock()
	if granted {
		return pkgTool.NewDecision(pkgTool.DecisionAllow, "run_grant", "", pkgTool.GrantRun)
	}
	if !exists {
		return pkgTool.Decision{}, fmt.Errorf("authorize with policy %q: %w", request.Policy(), pkgTool.ErrPolicyNotFound)
	}
	decision, err = decide(ctx, policy, request)
	if err != nil {
		return pkgTool.Decision{}, err
	}
	if decision.Kind() == pkgTool.DecisionPrompt {
		if approver == nil {
			return pkgTool.NewDecision(pkgTool.DecisionDeny, "approver_unavailable", pkgTool.DenyTerminal, "")
		}
		approval, approvalErr := approve(ctx, approver, request)
		if approvalErr != nil {
			return pkgTool.Decision{}, approvalErr
		}
		if approval.Kind() == pkgTool.ApprovalAllow {
			decision, err = pkgTool.NewDecision(pkgTool.DecisionAllow, approval.Reason(), "", approval.Scope())
		} else {
			decision, err = pkgTool.NewDecision(pkgTool.DecisionDeny, approval.Reason(), approval.Disposition(), "")
		}
		if err != nil {
			return pkgTool.Decision{}, err
		}
	}
	if decision.Kind() == pkgTool.DecisionAllow && decision.Scope() == pkgTool.GrantRun {
		runtime.mu.Lock()
		runtime.grants[key] = struct{}{}
		runtime.mu.Unlock()
	}
	return decision, nil
}

func decide(ctx context.Context, policy pkgTool.Policy, request pkgTool.PermissionRequest) (decision pkgTool.Decision, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("policy panicked: %w", pkgTool.ErrInvalidPolicy)
		}
	}()
	decision, err = policy.Decide(ctx, request)
	if err != nil {
		return pkgTool.Decision{}, err
	}
	if err := decision.Validate(); err != nil {
		return pkgTool.Decision{}, errors.Join(pkgTool.ErrInvalidDecision, err)
	}
	return decision, nil
}

func approve(ctx context.Context, approver pkgTool.Approver, request pkgTool.PermissionRequest) (approval pkgTool.Approval, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("approver panicked: %w", pkgTool.ErrInvalidApprover)
		}
	}()
	approval, err = approver.Approve(ctx, request)
	if err != nil {
		return pkgTool.Approval{}, err
	}
	if err := approval.Validate(); err != nil {
		return pkgTool.Approval{}, errors.Join(pkgTool.ErrInvalidApproval, err)
	}
	return approval, nil
}

func typedNil(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
