package tool

import (
	"fmt"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

// Rule is an exact action matcher. Empty resources, prefix matching and
// wildcards are intentionally unsupported by the baseline permission model.
type Rule struct {
	effect    Effect
	resource  string
	workspace contextengine.WorkspaceID
	decision  Decision
}

func NewRule(
	effect Effect,
	resource string,
	workspace contextengine.WorkspaceID,
	decision Decision,
) (Rule, error) {
	action, err := NewAction(effect, resource, workspace)
	if err != nil {
		return Rule{}, fmt.Errorf("rule action: %w: %w", err, ErrInvalidPolicy)
	}
	if err := decision.Validate(); err != nil {
		return Rule{}, fmt.Errorf("rule decision: %w: %w", err, ErrInvalidPolicy)
	}
	return Rule{effect: action.Effect(), resource: action.Resource(), workspace: action.Workspace(), decision: decision}, nil
}

func (rule Rule) Effect() Effect                       { return rule.effect }
func (rule Rule) Resource() string                     { return rule.resource }
func (rule Rule) Workspace() contextengine.WorkspaceID { return rule.workspace }
func (rule Rule) Decision() Decision                   { return rule.decision }

func (rule Rule) Validate() error {
	_, err := NewRule(rule.effect, rule.resource, rule.workspace, rule.decision)
	return err
}

func (rule Rule) Matches(action Action) bool {
	return rule.effect == action.Effect() && rule.resource == action.Resource() && rule.workspace == action.Workspace()
}
