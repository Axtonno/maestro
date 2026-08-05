package runtime

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

type node struct {
	component pkgRuntime.Component

	dependencies []*node

	dependents []*node
}

func newNode(component pkgRuntime.Component) *node {
	return &node{
		component: component,
		dependencies: make([]*node, 0),
		dependents: make([]*node, 0),
	}
}

func (n *node) Component() pkgRuntime.Component {
	return n.component
}

func (n *node) Dependencies() []*node {
	return n.dependencies
}

func (n *node) Dependents() []*node {
	return n.dependents
}

func (n *node) AddDependency(dependency *node) {
	n.dependencies = append(n.dependencies, dependency)
}

func (n *node) AddDependent(dependent *node) {
	n.dependents = append(n.dependents, dependent)
}