package tool

import (
	"fmt"
	"strings"
	"unicode/utf8"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

type Effect string

const (
	EffectLocalCompute     Effect = "local.compute"
	EffectWorkspaceInspect Effect = "workspace.inspect"
	EffectWorkspaceMutate  Effect = "workspace.mutate"
	EffectProcessExecute   Effect = "process.execute"
	EffectNetworkAccess    Effect = "network.access"
	EffectModelInvoke      Effect = "model.invoke"
	EffectModelDisclose    Effect = "model.disclose"
)

func (effect Effect) Valid() bool {
	switch effect {
	case EffectLocalCompute, EffectWorkspaceInspect, EffectWorkspaceMutate,
		EffectProcessExecute, EffectNetworkAccess, EffectModelInvoke,
		EffectModelDisclose:
		return true
	default:
		return false
	}
}

func (effect Effect) ValidForTool() bool {
	return effect.Valid() && effect != EffectModelInvoke && effect != EffectModelDisclose
}

type Action struct {
	effect    Effect
	resource  string
	workspace contextengine.WorkspaceID
}

func NewAction(effect Effect, resource string, workspace contextengine.WorkspaceID) (Action, error) {
	if !effect.Valid() {
		return Action{}, fmt.Errorf("effect %q is invalid: %w", effect, ErrInvalidAction)
	}
	if !exactValue(resource, 4096) || !utf8.ValidString(resource) || strings.ContainsAny(resource, "\r\n") {
		return Action{}, fmt.Errorf("action resource is blank, oversized, or unsafe: %w", ErrInvalidAction)
	}
	workspaceEffect := effect == EffectWorkspaceInspect || effect == EffectWorkspaceMutate || effect == EffectModelDisclose
	if workspaceEffect {
		if err := workspace.Validate(); err != nil {
			return Action{}, fmt.Errorf("action workspace: %w: %w", err, ErrInvalidAction)
		}
	} else if workspace != "" {
		return Action{}, fmt.Errorf("effect %q cannot carry a workspace: %w", effect, ErrInvalidAction)
	}
	return Action{effect: effect, resource: resource, workspace: workspace}, nil
}

func (action Action) Effect() Effect                       { return action.effect }
func (action Action) Resource() string                     { return action.resource }
func (action Action) Workspace() contextengine.WorkspaceID { return action.workspace }

func (action Action) Validate() error {
	_, err := NewAction(action.effect, action.resource, action.workspace)
	return err
}
