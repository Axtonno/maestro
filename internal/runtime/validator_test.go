package runtime

import (
	"errors"
	"strings"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestValidatorAcceptsValidGraph(t *testing.T) {
	dependencyGraph := newGraph()

	config := newTestComponent("config")
	provider := newTestComponent("provider")

	if err := dependencyGraph.Add(config); err != nil {
		t.Fatalf("add config: %v", err)
	}

	if err := dependencyGraph.Add(provider); err != nil {
		t.Fatalf("add provider: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"provider",
		"config",
	); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	graphValidator := newValidator()

	if err := graphValidator.Validate(dependencyGraph); err != nil {
		t.Fatalf("validate graph: %v", err)
	}
}

func TestValidatorRejectsNilGraph(t *testing.T) {
	graphValidator := newValidator()

	err := graphValidator.Validate(nil)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsEmptyComponentID(t *testing.T) {
	dependencyGraph := newGraph()

	component := &testComponent{
		metadata: pkgRuntime.Metadata{
			Name:    "invalid",
			Version: "1.0.0",
		},
	}

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsEmptyName(t *testing.T) {
	dependencyGraph := newGraph()

	component := newTestComponent("component")
	component.metadata.Name = "   "

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsEmptyVersion(t *testing.T) {
	dependencyGraph := newGraph()

	component := newTestComponent("component")
	component.metadata.Version = ""

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsDuplicateDependencies(t *testing.T) {
	dependencyGraph := newGraph()

	component := newTestComponent(
		"provider",
		pkgRuntime.Dependency{
			ID:       "config",
			Required: true,
		},
		pkgRuntime.Dependency{
			ID:       "config",
			Required: false,
		},
	)

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add provider: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsDuplicateCapabilities(t *testing.T) {
	dependencyGraph := newGraph()

	component := newTestComponent("provider")
	component.metadata.Capabilities = []pkgRuntime.Capability{
		pkgRuntime.CapabilityStart,
		pkgRuntime.CapabilityStart,
	}

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add component: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrInvalidMetadata) {
		t.Fatalf(
			"expected ErrInvalidMetadata, got %v",
			err,
		)
	}
}

func TestValidatorRejectsSelfDependency(t *testing.T) {
	dependencyGraph := newGraph()

	component := newTestComponent("provider")

	if err := dependencyGraph.Add(component); err != nil {
		t.Fatalf("add provider: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"provider",
		"provider",
	); err != nil {
		t.Fatalf("add self dependency: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrCyclicDependency) {
		t.Fatalf(
			"expected ErrCyclicDependency, got %v",
			err,
		)
	}
}

func TestValidatorRejectsDependencyCycle(t *testing.T) {
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

	if err := dependencyGraph.AddDependency(
		"first",
		"second",
	); err != nil {
		t.Fatalf("add first dependency: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"second",
		"third",
	); err != nil {
		t.Fatalf("add second dependency: %v", err)
	}

	if err := dependencyGraph.AddDependency(
		"third",
		"first",
	); err != nil {
		t.Fatalf("add third dependency: %v", err)
	}

	graphValidator := newValidator()

	err := graphValidator.Validate(dependencyGraph)
	if !errors.Is(err, pkgRuntime.ErrCyclicDependency) {
		t.Fatalf(
			"expected ErrCyclicDependency, got %v",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"dependency cycle detected",
	) {
		t.Fatalf(
			"expected cycle description, got %q",
			err.Error(),
		)
	}
}
