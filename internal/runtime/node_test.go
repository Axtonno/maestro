package runtime

import "testing"

func TestNewNode(t *testing.T) {
	component := newTestComponent("config")
	currentNode := newNode(component)

	if currentNode.Component() != component {
		t.Fatal("node contains unexpected component")
	}
	if currentNode.ID() != "config" {
		t.Fatalf("expected node ID config, got %q", currentNode.ID())
	}

	if len(currentNode.Dependencies()) != 0 {
		t.Fatal("new node unexpectedly contains dependencies")
	}

	if len(currentNode.Dependents()) != 0 {
		t.Fatal("new node unexpectedly contains dependents")
	}
}

func TestNodeAddDependency(t *testing.T) {
	componentNode := newNode(
		newTestComponent("provider"),
	)

	dependencyNode := newNode(
		newTestComponent("config"),
	)

	componentNode.AddDependency(dependencyNode)

	dependencies := componentNode.Dependencies()

	if len(dependencies) != 1 {
		t.Fatalf(
			"expected 1 dependency, got %d",
			len(dependencies),
		)
	}

	if dependencies[0] != dependencyNode {
		t.Fatal("unexpected dependency node")
	}
}

func TestNodeAddDependent(t *testing.T) {
	dependencyNode := newNode(
		newTestComponent("config"),
	)

	dependentNode := newNode(
		newTestComponent("provider"),
	)

	dependencyNode.AddDependent(dependentNode)

	dependents := dependencyNode.Dependents()

	if len(dependents) != 1 {
		t.Fatalf(
			"expected 1 dependent, got %d",
			len(dependents),
		)
	}

	if dependents[0] != dependentNode {
		t.Fatal("unexpected dependent node")
	}
}
