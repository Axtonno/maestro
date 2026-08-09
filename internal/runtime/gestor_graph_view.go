package runtime

import (
	"fmt"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

// gestorGraphView exposes only generation, eligibility and dependency order.
// It reads the Runtime-owned graph and never adds nodes, edges or copied graph
// state.
type gestorGraphView struct {
	runtime *runtime
}

func newGestorGraphView(runtime *runtime) *gestorGraphView {
	return &gestorGraphView{runtime: runtime}
}

func (view *gestorGraphView) State() (uint64, bool) {
	view.runtime.mu.RLock()
	defer view.runtime.mu.RUnlock()

	generation := view.runtime.graphGeneration
	current := view.runtime.dependencyGraph != nil &&
		view.runtime.graphComponentGeneration == view.runtime.componentGeneration

	return generation, current
}

func (view *gestorGraphView) DependencyPlan(
	componentID pkgRuntime.ComponentID,
) (uint64, []pkgRuntime.ComponentID, error) {
	view.runtime.mu.RLock()
	generation := view.runtime.graphGeneration
	dependencyGraph := view.runtime.dependencyGraph
	current := dependencyGraph != nil &&
		view.runtime.graphComponentGeneration == view.runtime.componentGeneration
	view.runtime.mu.RUnlock()

	if !current {
		return generation, nil, fmt.Errorf(
			"dependency graph generation %d is stale: %w",
			generation,
			pkgRuntime.ErrInvalidState,
		)
	}
	selected, exists := dependencyGraph.Node(componentID)
	if !exists {
		return generation, nil, fmt.Errorf(
			"component %q is absent from dependency graph: %w",
			componentID,
			pkgRuntime.ErrNotFound,
		)
	}

	required := make(map[pkgRuntime.ComponentID]struct{})
	var collect func(*node)
	collect = func(current *node) {
		for _, dependency := range current.Dependencies() {
			dependencyID := dependency.ID()
			if _, visited := required[dependencyID]; visited {
				continue
			}
			required[dependencyID] = struct{}{}
			collect(dependency)
		}
	}
	collect(selected)

	ordered, err := dependencyGraph.TopologicalOrder()
	if err != nil {
		return generation, nil, err
	}
	plan := make([]pkgRuntime.ComponentID, 0, len(required))
	for _, current := range ordered {
		currentID := current.ID()
		if _, included := required[currentID]; included {
			plan = append(plan, currentID)
		}
	}

	return generation, plan, nil
}
