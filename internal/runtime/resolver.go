package runtime

import (
	"fmt"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type resolver struct {
	registry *registry
}

func newResolver(registry *registry) *resolver {
	return &resolver{
		registry: registry,
	}
}

func (r *resolver) Resolve() (*graph, error) {
	dependencyGraph := newGraph()

	components := r.registry.Components()

	if err := r.addComponents(dependencyGraph, components); err != nil {
		return nil, err
	}

	if err := r.addDependencies(dependencyGraph, components); err != nil {
		return nil, err
	}

	return dependencyGraph, nil
}

func (r *resolver) addComponents(
	dependencyGraph *graph,
	components []pkgRuntime.Component,
) error {
	for _, component := range components {
		if err := dependencyGraph.Add(component); err != nil {
			return fmt.Errorf(
				"add component %q to dependency graph: %w",
				component.Metadata().ID,
				err,
			)
		}
	}

	return nil
}

func (r *resolver) addDependencies(
	dependencyGraph *graph,
	components []pkgRuntime.Component,
) error {
	for _, component := range components {
		metadata := component.Metadata()

		for _, dependency := range metadata.Dependencies {
			if !dependencyGraph.Has(dependency.ID) {
				if dependency.Required {
					return fmt.Errorf(
						"component %q requires missing dependency %q: %w",
						metadata.ID,
						dependency.ID,
						pkgRuntime.ErrNotFound,
					)
				}

				continue
			}

			if err := dependencyGraph.AddDependency(
				metadata.ID,
				dependency.ID,
			); err != nil {
				return fmt.Errorf(
					"resolve dependency %q for component %q: %w",
					dependency.ID,
					metadata.ID,
					err,
				)
			}
		}
	}

	return nil
}