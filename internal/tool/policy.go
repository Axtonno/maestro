package tool

import (
	"context"
	"fmt"
	"slices"

	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

var _ pkgTool.Policy = (*StaticPolicy)(nil)

type StaticPolicy struct {
	id    pkgTool.PolicyID
	rules []pkgTool.Rule
}

func NewStaticPolicy(id pkgTool.PolicyID, rules []pkgTool.Rule) (*StaticPolicy, error) {
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("static policy identity: %w: %w", err, pkgTool.ErrInvalidPolicy)
	}
	cloned := slices.Clone(rules)
	type ruleKey struct {
		effect    pkgTool.Effect
		resource  string
		workspace string
	}
	seen := make(map[ruleKey]struct{}, len(cloned))
	for index, rule := range cloned {
		if err := rule.Validate(); err != nil {
			return nil, fmt.Errorf("static policy rule %d: %w: %w", index, err, pkgTool.ErrInvalidPolicy)
		}
		key := ruleKey{effect: rule.Effect(), resource: rule.Resource(), workspace: string(rule.Workspace())}
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("static policy rule %d duplicates an exact matcher: %w", index, pkgTool.ErrInvalidPolicy)
		}
		seen[key] = struct{}{}
	}
	return &StaticPolicy{id: id, rules: cloned}, nil
}

func (policy *StaticPolicy) ID() pkgTool.PolicyID { return policy.id }

func (policy *StaticPolicy) Decide(ctx context.Context, request pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	if ctx == nil || request.Validate() != nil {
		return pkgTool.Decision{}, pkgTool.ErrInvalidPermissionRequest
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Decision{}, err
	}
	allRunScoped := true
	needsPrompt := false
	for _, action := range request.Actions() {
		decision, matched := policy.match(action)
		if !matched {
			return pkgTool.NewDecision(pkgTool.DecisionDeny, "no_matching_rule", pkgTool.DenyTerminal, "")
		}
		switch decision.Kind() {
		case pkgTool.DecisionDeny:
			return decision, nil
		case pkgTool.DecisionPrompt:
			needsPrompt = true
		case pkgTool.DecisionAllow:
			allRunScoped = allRunScoped && decision.Scope() == pkgTool.GrantRun
		default:
			return pkgTool.Decision{}, pkgTool.ErrInvalidDecision
		}
	}
	if needsPrompt {
		return pkgTool.NewDecision(pkgTool.DecisionPrompt, "approval_required", "", "")
	}
	scope := pkgTool.GrantOneShot
	if allRunScoped {
		scope = pkgTool.GrantRun
	}
	return pkgTool.NewDecision(pkgTool.DecisionAllow, "all_actions_allowed", "", scope)
}

func (policy *StaticPolicy) match(action pkgTool.Action) (pkgTool.Decision, bool) {
	for _, rule := range policy.rules {
		if rule.Matches(action) {
			return rule.Decision(), true
		}
	}
	return pkgTool.Decision{}, false
}
