package runtime

import (
	"fmt"
	"strings"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type validator struct{}

func newValidator() *validator {
	return &validator{}
}

func (v *validator) Validate(dependencyGraph *graph) error {
	if dependencyGraph == nil {
		return fmt.Errorf(
			"validate dependency graph: graph is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	if err := v.validateMetadata(dependencyGraph); err != nil {
		return err
	}

	if err := v.validateSelfDependencies(dependencyGraph); err != nil {
		return err
	}

	if err := v.validateCycles(dependencyGraph); err != nil {
		return err
	}

	return nil
}

func (v *validator) validateMetadata(dependencyGraph *graph) error {
	for _, currentNode := range dependencyGraph.Nodes() {
		component := currentNode.Component()
		metadata := component.Metadata()

		if metadata.ID == "" {
			return fmt.Errorf(
				"component has an empty ID: %w",
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		if strings.TrimSpace(metadata.Name) == "" {
			return fmt.Errorf(
				"component %q has an empty name: %w",
				metadata.ID,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		if strings.TrimSpace(metadata.Version) == "" {
			return fmt.Errorf(
				"component %q has an empty version: %w",
				metadata.ID,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		if err := v.validateDependencies(metadata); err != nil {
			return err
		}

		if err := v.validateCapabilities(metadata); err != nil {
			return err
		}
	}

	return nil
}

func (v *validator) validateDependencies(
	metadata pkgRuntime.Metadata,
) error {
	seen := make(map[pkgRuntime.ComponentID]struct{})

	for _, dependency := range metadata.Dependencies {
		if dependency.ID == "" {
			return fmt.Errorf(
				"component %q declares a dependency with an empty ID: %w",
				metadata.ID,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		if _, exists := seen[dependency.ID]; exists {
			return fmt.Errorf(
				"component %q declares dependency %q more than once: %w",
				metadata.ID,
				dependency.ID,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		seen[dependency.ID] = struct{}{}
	}

	return nil
}

func (v *validator) validateCapabilities(
	metadata pkgRuntime.Metadata,
) error {
	seen := make(map[pkgRuntime.Capability]struct{})

	for _, capability := range metadata.Capabilities {
		if capability == "" {
			return fmt.Errorf(
				"component %q declares an empty capability: %w",
				metadata.ID,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		if _, exists := seen[capability]; exists {
			return fmt.Errorf(
				"component %q declares capability %q more than once: %w",
				metadata.ID,
				capability,
				pkgRuntime.ErrInvalidMetadata,
			)
		}

		seen[capability] = struct{}{}
	}

	return nil
}

func (v *validator) validateSelfDependencies(
	dependencyGraph *graph,
) error {
	for _, currentNode := range dependencyGraph.Nodes() {
		componentID := currentNode.Component().Metadata().ID

		for _, dependencyNode := range currentNode.Dependencies() {
			dependencyID := dependencyNode.Component().Metadata().ID

			if componentID == dependencyID {
				return fmt.Errorf(
					"component %q depends on itself: %w",
					componentID,
					pkgRuntime.ErrCyclicDependency,
				)
			}
		}
	}

	return nil
}

func (v *validator) validateCycles(dependencyGraph *graph) error {
	states := make(map[pkgRuntime.ComponentID]visitState)
	path := make([]pkgRuntime.ComponentID, 0)

	for _, currentNode := range dependencyGraph.Nodes() {
		componentID := currentNode.Component().Metadata().ID

		if states[componentID] != visitStateUnvisited {
			continue
		}

		if err := v.visit(currentNode, states, &path); err != nil {
			return err
		}
	}

	return nil
}

func (v *validator) visit(
	currentNode *node,
	states map[pkgRuntime.ComponentID]visitState,
	path *[]pkgRuntime.ComponentID,
) error {
	componentID := currentNode.Component().Metadata().ID

	states[componentID] = visitStateVisiting
	*path = append(*path, componentID)

	for _, dependencyNode := range currentNode.Dependencies() {
		dependencyID := dependencyNode.Component().Metadata().ID

		switch states[dependencyID] {
		case visitStateVisiting:
			cycle := cyclePath(*path, dependencyID)

			return fmt.Errorf(
				"dependency cycle detected: %s: %w",
				formatComponentPath(cycle),
				pkgRuntime.ErrCyclicDependency,
			)

		case visitStateUnvisited:
			if err := v.visit(dependencyNode, states, path); err != nil {
				return err
			}
		}
	}

	*path = (*path)[:len(*path)-1]
	states[componentID] = visitStateVisited

	return nil
}

func cyclePath(
	path []pkgRuntime.ComponentID,
	repeatedID pkgRuntime.ComponentID,
) []pkgRuntime.ComponentID {
	start := 0

	for index, componentID := range path {
		if componentID == repeatedID {
			start = index
			break
		}
	}

	cycle := append(
		[]pkgRuntime.ComponentID(nil),
		path[start:]...,
	)

	return append(cycle, repeatedID)
}

func formatComponentPath(
	path []pkgRuntime.ComponentID,
) string {
	values := make([]string, 0, len(path))

	for _, componentID := range path {
		values = append(values, string(componentID))
	}

	return strings.Join(values, " -> ")
}

type visitState uint8

const (
	visitStateUnvisited visitState = iota
	visitStateVisiting
	visitStateVisited
)
