package gestor

import (
	"fmt"
	"slices"
	"strings"
)

type QueryOptions struct {
	TargetKind       TargetKind
	Scope            Scope
	Model            string
	RequireAvailable bool
	PreferredTargets []Target
}

// Query is immutable after construction. PreferredTargets returns a copy.
type Query struct {
	capability       CapabilityID
	targetKind       TargetKind
	scope            Scope
	model            string
	requireAvailable bool
	preferredTargets []Target
}

func NewQuery(capability CapabilityID, options QueryOptions) (Query, error) {
	query := Query{
		capability:       capability,
		targetKind:       options.TargetKind,
		scope:            options.Scope,
		model:            options.Model,
		requireAvailable: options.RequireAvailable,
		preferredTargets: slices.Clone(options.PreferredTargets),
	}
	if err := query.Validate(); err != nil {
		return Query{}, err
	}

	return query, nil
}

func (query Query) Capability() CapabilityID { return query.capability }
func (query Query) TargetKind() TargetKind   { return query.targetKind }
func (query Query) Scope() Scope             { return query.scope }
func (query Query) Model() string            { return query.model }
func (query Query) RequireAvailable() bool   { return query.requireAvailable }

func (query Query) PreferredTargets() []Target {
	return slices.Clone(query.preferredTargets)
}

func (query Query) Validate() error {
	if err := query.capability.Validate(); err != nil {
		return fmt.Errorf("query capability: %w: %w", err, ErrInvalidQuery)
	}
	if query.targetKind != "" && !query.targetKind.Valid() {
		return fmt.Errorf("query target kind %q is unknown: %w", query.targetKind, ErrInvalidQuery)
	}
	if query.scope != "" && !query.scope.Valid() {
		return fmt.Errorf("query scope %q is unknown: %w", query.scope, ErrInvalidQuery)
	}
	if query.model != "" && strings.TrimSpace(query.model) != query.model {
		return fmt.Errorf("query model ID %q is not exact: %w", query.model, ErrInvalidQuery)
	}
	if query.scope == ScopeModel && !exactID(query.model) {
		return fmt.Errorf("model scope requires an exact model ID: %w", ErrInvalidQuery)
	}
	if query.scope != ScopeModel && query.model != "" {
		return fmt.Errorf("model ID requires model scope: %w", ErrInvalidQuery)
	}
	if query.targetKind == TargetKindComponent && query.scope != "" && query.scope != ScopeComponent {
		return fmt.Errorf("component query cannot use scope %q: %w", query.scope, ErrInvalidQuery)
	}
	if query.targetKind == TargetKindProvider && query.scope == ScopeComponent {
		return fmt.Errorf("provider query cannot use component scope: %w", ErrInvalidQuery)
	}

	seen := make(map[Target]struct{}, len(query.preferredTargets))
	for index, target := range query.preferredTargets {
		if err := target.Validate(); err != nil {
			return fmt.Errorf("preferred target %d: %w: %w", index, err, ErrInvalidQuery)
		}
		if _, exists := seen[target]; exists {
			return fmt.Errorf("preferred target %d is duplicated: %w", index, ErrInvalidQuery)
		}
		seen[target] = struct{}{}
		if query.targetKind != "" && target.Kind != query.targetKind {
			return fmt.Errorf("preferred target %d does not match target kind: %w", index, ErrInvalidQuery)
		}
		if query.scope != "" && target.Scope != query.scope {
			return fmt.Errorf("preferred target %d does not match scope: %w", index, ErrInvalidQuery)
		}
		if query.model != "" && target.Model != query.model {
			return fmt.Errorf("preferred target %d does not match model ID: %w", index, ErrInvalidQuery)
		}
	}

	return nil
}
