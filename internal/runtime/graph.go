package runtime

import (
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
	nodes := make([]*node, 0, len(g.nodes))

	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}

	return nodes
}

func (g *graph) Len() int {
	return len(g.nodes)
}