package gestor

import (
	"cmp"
	"fmt"
	"strings"
)

type TargetKind string

const (
	TargetKindComponent TargetKind = "component"
	TargetKindProvider  TargetKind = "provider"
)

func (kind TargetKind) Valid() bool {
	switch kind {
	case TargetKindComponent, TargetKindProvider:
		return true
	default:
		return false
	}
}

type Scope string

const (
	ScopeComponent Scope = "component"
	ScopeAdapter   Scope = "adapter"
	ScopeInstance  Scope = "instance"
	ScopeModel     Scope = "model"
)

func (scope Scope) Valid() bool {
	switch scope {
	case ScopeComponent, ScopeAdapter, ScopeInstance, ScopeModel:
		return true
	default:
		return false
	}
}

type Target struct {
	Kind  TargetKind
	ID    string
	Scope Scope
	Model string
}

func (target Target) Validate() error {
	if !target.Kind.Valid() {
		return fmt.Errorf("target kind %q is unknown: %w", target.Kind, ErrInvalidTarget)
	}
	if !exactID(target.ID) {
		return fmt.Errorf("target ID %q is not exact: %w", target.ID, ErrInvalidTarget)
	}
	if !target.Scope.Valid() {
		return fmt.Errorf("target scope %q is unknown: %w", target.Scope, ErrInvalidTarget)
	}

	switch target.Kind {
	case TargetKindComponent:
		if target.Scope != ScopeComponent || target.Model != "" {
			return fmt.Errorf("component target requires component scope and no model: %w", ErrInvalidTarget)
		}
	case TargetKindProvider:
		if target.Scope == ScopeComponent {
			return fmt.Errorf("provider target cannot use component scope: %w", ErrInvalidTarget)
		}
		if target.Scope == ScopeModel {
			if !exactID(target.Model) {
				return fmt.Errorf("model target requires an exact model ID: %w", ErrInvalidTarget)
			}
		} else if target.Model != "" {
			return fmt.Errorf("non-model target cannot include a model ID: %w", ErrInvalidTarget)
		}
	}

	return nil
}

func (target Target) Compare(other Target) int {
	if result := cmp.Compare(target.Kind, other.Kind); result != 0 {
		return result
	}
	if result := cmp.Compare(target.ID, other.ID); result != 0 {
		return result
	}
	if result := cmp.Compare(target.Scope, other.Scope); result != 0 {
		return result
	}

	return cmp.Compare(target.Model, other.Model)
}

func exactID(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
