package application

import (
	"context"
	"fmt"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type ProductPolicy struct {
	id        pkgTool.PolicyID
	provider  pkgProvider.ID
	model     string
	workspace pkgContext.WorkspaceID
	tools     map[pkgTool.ID]struct{}
	modes     map[pkgTool.Effect]string
}

func NewProductPolicy(config productconfig.Config) (*ProductPolicy, error) {
	if err := config.ValidateExecutionProfile(); err != nil {
		return nil, err
	}
	tools := make(map[pkgTool.ID]struct{}, len(config.Agent.Tools))
	for _, id := range config.ToolIDs() {
		tools[id] = struct{}{}
	}
	return &ProductPolicy{
		id:        pkgTool.PolicyID(config.Policy.ID),
		provider:  pkgProvider.ID(config.Provider.ID),
		model:     config.Models.Chat,
		workspace: pkgContext.WorkspaceID(config.Workspace.ID),
		tools:     tools,
		modes: map[pkgTool.Effect]string{
			pkgTool.EffectModelInvoke:      config.Policy.Model,
			pkgTool.EffectModelDisclose:    config.Policy.Model,
			pkgTool.EffectWorkspaceInspect: config.Policy.WorkspaceInspect,
			pkgTool.EffectWorkspaceMutate:  config.Policy.WorkspaceMutation,
		},
	}, nil
}

func (policy *ProductPolicy) ID() pkgTool.PolicyID { return policy.id }

func (policy *ProductPolicy) Decide(ctx context.Context, request pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	if policy == nil || ctx == nil || request.Validate() != nil || request.Policy() != policy.id {
		return pkgTool.Decision{}, pkgTool.ErrInvalidPermissionRequest
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Decision{}, err
	}
	if !policy.matchesSubject(request) {
		return deny("target_not_allowed")
	}
	mode := "allow"
	for _, action := range request.Actions() {
		configured, exists := policy.modes[action.Effect()]
		if !exists || !policy.matchesAction(action) {
			return deny("action_not_allowed")
		}
		if configured == "deny" {
			return deny("configured_deny")
		}
		if action.Effect() == pkgTool.EffectWorkspaceMutate && configured != "prompt" {
			return deny("mutation_requires_prompt")
		}
		if configured == "prompt" {
			mode = "prompt"
		}
	}
	if mode == "prompt" {
		return pkgTool.NewDecision(pkgTool.DecisionPrompt, "approval_required", "", "")
	}
	return pkgTool.NewDecision(pkgTool.DecisionAllow, "configured_allow", "", pkgTool.GrantRun)
}

func (policy *ProductPolicy) matchesSubject(request pkgTool.PermissionRequest) bool {
	switch request.Subject() {
	case pkgTool.PermissionSubjectModel:
		target, exists := request.ModelTarget()
		if !exists || target.Provider() != policy.provider || target.Model() != policy.model {
			return false
		}
		if disclosure, exists := request.Disclosure(); exists && disclosure.Workspace() != policy.workspace {
			return false
		}
		return true
	case pkgTool.PermissionSubjectTool:
		prepared, exists := request.Prepared()
		if !exists {
			return false
		}
		_, allowed := policy.tools[prepared.Invocation().Tool()]
		return allowed
	default:
		return false
	}
}

func (policy *ProductPolicy) matchesAction(action pkgTool.Action) bool {
	switch action.Effect() {
	case pkgTool.EffectModelInvoke:
		return action.Workspace() == "" && action.Resource() == string(policy.provider)+"/"+policy.model
	case pkgTool.EffectModelDisclose, pkgTool.EffectWorkspaceInspect, pkgTool.EffectWorkspaceMutate:
		return action.Workspace() == policy.workspace
	default:
		return false
	}
}

func deny(reason string) (pkgTool.Decision, error) {
	decision, err := pkgTool.NewDecision(pkgTool.DecisionDeny, reason, pkgTool.DenyTerminal, "")
	if err != nil {
		return pkgTool.Decision{}, fmt.Errorf("construct product policy deny: %w", err)
	}
	return decision, nil
}
