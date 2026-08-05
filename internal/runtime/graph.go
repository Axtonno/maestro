package runtime

import (
	"fmt"
	"sort"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type graph struct {
	nodes map[pkgRuntime.ComponentID]*node
}

func newGraph() *graph {
	return &graph{
		nodes: make(map[pkgRuntime.ComponentID]*node),
	}
}

func (g *graph) Add(component pkgRuntime.Component) error {
	id := component.Metadata().ID

	if _, exists := g.nodes[id]; exists {
		return pkgRuntime.ErrAlreadyRegistered
	}

	g.nodes[id] = newNode(component)

	return nil
}

func (g *graph) Node(id pkgRuntime.ComponentID) (*node, bool) {
	n, exists := g.nodes[id]

	return n, exists
}

func (g *graph) Has(id pkgRuntime.ComponentID) bool {
	_, exists := g.nodes[id]

	return exists
}

func (g *graph) AddDependency(
	componentID pkgRuntime.ComponentID,
	dependencyID pkgRuntime.ComponentID,
) error {
	componentNode, exists := g.nodes[componentID]
	if !exists {
		return pkgRuntime.ErrNotFound
	}

	dependencyNode, exists := g.nodes[dependencyID]
	if !exists {
		return pkgRuntime.ErrNotFound
	}

	componentNode.AddDependency(dependencyNode)
	dependencyNode.AddDependent(componentNode)

	return nil
}

func (g *graph) Nodes() []*node {
	ids := g.sortedIDs()

	nodes := make([]*node, 0, len(ids))

	for _, id := range ids {
		nodes = append(nodes, g.nodes[id])
	}

	return nodes
}

func (g *graph) Len() int {
	return len(g.nodes)
}

func (g *graph) TopologicalOrder() ([]*node, error) {
	ordered := make([]*node, 0, len(g.nodes))
	states := make(map[pkgRuntime.ComponentID]visitState)
	path := make([]pkgRuntime.ComponentID, 0)

	for _, id := range g.sortedIDs() {
		if states[id] != visitStateUnvisited {
			continue
		}

		if err := g.visitTopological(
			g.nodes[id],
			states,
			&path,
			&ordered,
		); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}

func (g *graph) ReverseTopologicalOrder() ([]*node, error) {
	ordered, err := g.TopologicalOrder()
	if err != nil {
		return nil, err
	}

	for left, right := 0, len(ordered)-1; left < right; left, right = left+1, right-1 {
		ordered[left], ordered[right] = ordered[right], ordered[left]
	}

	return ordered, nil
}

func (g *graph) visitTopological(
	currentNode *node,
	states map[pkgRuntime.ComponentID]visitState,
	path *[]pkgRuntime.ComponentID,
	ordered *[]*node,
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
				"dependency cycle detected while sorting graph: %s: %w",
				formatComponentPath(cycle),
				pkgRuntime.ErrCyclicDependency,
			)

		case visitStateUnvisited:
			if err := g.visitTopological(
				dependencyNode,
				states,
				path,
				ordered,
			); err != nil {
				return err
			}
		}
	}

	*path = (*path)[:len(*path)-1]
	states[componentID] = visitStateVisited
	*ordered = append(*ordered, currentNode)

	return nil
}

func (g *graph) sortedIDs() []pkgRuntime.ComponentID {
	ids := make([]pkgRuntime.ComponentID, 0, len(g.nodes))

	for id := range g.nodes {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(left, right int) bool {
		return ids[left] < ids[right]
	})

	return ids
}
