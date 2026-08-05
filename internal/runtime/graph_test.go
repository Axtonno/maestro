package runtime

import (
	"errors"
	"reflect"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestGraphAddAndRetrieveNode(t *testing.T) {
	dependencyGraph := newGraph()
	component := newTestComponent("config")

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	currentNode, exists := dependencyGraph.Node("config")
	if !exists {
		t.Fatal("graph does not contain added component")
	}

	if currentNode.Component() != component {
		t.Fatal("node contains unexpected component")
	}
}

func TestGraphRejectsDuplicateComponent(t *testing.T) {
	dependencyGraph := newGraph()
	component := newTestComponent("config")

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add first component: %v", err)
	}

	err := dependencyGraph.Add(component)
	if !errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
		t.Fatalf(
			"expected ErrAlreadyRegistered, got %v",
			err,
		)
	}
}

func TestGraphAddDependency(t *testing.T) {
	dependencyGraph := newGraph()

	provider := newTestComponent("provider")
	config := newTestComponent("config")

	if err := dependencyGraph.Add(provider); err != nil {
		t.Fatalf("add provider: %v", err)
	}

	if err := dependencyGraph.Add(config); err != nil {
		t.Fatalf("add config: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"provider",
		"config",
	); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	providerNode, _ := dependencyGraph.Node("provider")
	configNode, _ := dependencyGraph.Node("config")

	if len(providerNode.Dependencies()) != 1 {
		t.Fatalf(
			"expected provider to have 1 dependency, got %d",
			len(providerNode.Dependencies()),
		)
	}

	if providerNode.Dependencies()[0] != configNode {
		t.Fatal("provider has unexpected dependency")
	}

	if len(configNode.Dependents()) != 1 {
		t.Fatalf(
			"expected config to have 1 dependent, got %d",
			len(configNode.Dependents()),
		)
	}

	if configNode.Dependents()[0] != providerNode {
		t.Fatal("config has unexpected dependent")
	}
}

func TestGraphAddDependencyWithMissingComponent(t *testing.T) {
	dependencyGraph := newGraph()

	if err := dependencyGraph.Add(
		newTestComponent("config"),
	); err != nil {
		t.Fatalf("add config: %v", err)
	}

	err := dependencyGraph.AddDependency(
		"provider",
		"config",
	)

	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestGraphAddDependencyWithMissingDependency(t *testing.T) {
	dependencyGraph := newGraph()

	if err := dependencyGraph.Add(
		newTestComponent("provider"),
	); err != nil {
		t.Fatalf("add provider: %v", err)
	}

	err := dependencyGraph.AddDependency(
		"provider",
		"config",
	)

	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestGraphLen(t *testing.T) {
	dependencyGraph := newGraph()

	components := []*testComponent{
		newTestComponent("first"),
		newTestComponent("second"),
		newTestComponent("third"),
	}

	for _, component := range components {
		if err := dependencyGraph.Add(component); err != nil {
			t.Fatalf("add component: %v", err)
		}
	}

	if dependencyGraph.Len() != 3 {
		t.Fatalf(
			"expected graph length 3, got %d",
			dependencyGraph.Len(),
		)
	}
}

func TestGraphTopologicalOrder(t *testing.T) {
	dependencyGraph := newGraph()

	config := newTestComponent("config")
	provider := newTestComponent("provider")
	agent := newTestComponent("agent")

	for _, component := range []pkgRuntime.Component{
		agent,
		provider,
		config,
	} {
		if err := dependencyGraph.Add(component); err != nil {
			t.Fatalf("add component: %v", err)
		}
	}

	if err := dependencyGraph.AddDependency(
		"provider",
		"config",
	); err != nil {
		t.Fatalf("add provider dependency: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"agent",
		"provider",
	); err != nil {
		t.Fatalf("add agent dependency: %v", err)
	}

	orderedNodes, err := dependencyGraph.TopologicalOrder()
	if err != nil {
		t.Fatalf("topological order: %v", err)
	}

	if got, want := componentIDs(orderedNodes), []pkgRuntime.ComponentID{
		"config",
		"provider",
		"agent",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected order %v, got %v", want, got)
	}
}

func TestGraphReverseTopologicalOrder(t *testing.T) {
	dependencyGraph := newGraph()

	config := newTestComponent("config")
	provider := newTestComponent("provider")

	for _, component := range []pkgRuntime.Component{
		provider,
		config,
	} {
		if err := dependencyGraph.Add(component); err != nil {
			t.Fatalf("add component: %v", err)
		}
	}

	if err := dependencyGraph.AddDependency(
		"provider",
		"config",
	); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	orderedNodes, err := dependencyGraph.ReverseTopologicalOrder()
	if err != nil {
		t.Fatalf("reverse topological order: %v", err)
	}

	if got, want := componentIDs(orderedNodes), []pkgRuntime.ComponentID{
		"provider",
		"config",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected order %v, got %v", want, got)
	}
}

func componentIDs(nodes []*node) []pkgRuntime.ComponentID {
	ids := make([]pkgRuntime.ComponentID, 0, len(nodes))

	for _, currentNode := range nodes {
		ids = append(ids, currentNode.Component().Metadata().ID)
	}

	return ids
}
