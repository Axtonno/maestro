package runtime

import (
	"errors"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestRegistryRegisterAndResolve(t *testing.T) {
	componentRegistry := newRegistry()
	component := newTestComponent("config")

	if err := componentRegistry.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	resolved, err := componentRegistry.Resolve("config")
	if err != nil {
		t.Fatalf("resolve component: %v", err)
	}

	if resolved != component {
		t.Fatal("resolved component does not match registered component")
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	componentRegistry := newRegistry()
	component := newTestComponent("config")

	if err := componentRegistry.Register(component); err != nil {
		t.Fatalf("register first component: %v", err)
	}

	err := componentRegistry.Register(component)
	if !errors.Is(err, pkgRuntime.ErrAlreadyRegistered) {
		t.Fatalf(
			"expected ErrAlreadyRegistered, got %v",
			err,
		)
	}
}

func TestRegistryResolveMissingComponent(t *testing.T) {
	componentRegistry := newRegistry()

	_, err := componentRegistry.Resolve("missing")
	if !errors.Is(err, pkgRuntime.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound, got %v",
			err,
		)
	}
}

func TestRegistryHas(t *testing.T) {
	componentRegistry := newRegistry()
	component := newTestComponent("config")

	if componentRegistry.Has("config") {
		t.Fatal("registry unexpectedly contains component")
	}

	if err := componentRegistry.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	if !componentRegistry.Has("config") {
		t.Fatal("registry does not contain registered component")
	}
}

func TestRegistryComponents(t *testing.T) {
	componentRegistry := newRegistry()

	first := newTestComponent("first")
	second := newTestComponent("second")

	if err := componentRegistry.Register(first); err != nil {
		t.Fatalf("register first component: %v", err)
	}

	if err := componentRegistry.Register(second); err != nil {
		t.Fatalf("register second component: %v", err)
	}

	components := componentRegistry.Components()

	if len(components) != 2 {
		t.Fatalf(
			"expected 2 components, got %d",
			len(components),
		)
	}
}

func TestRegistryComponentsReturnsIndependentSlice(t *testing.T) {
	componentRegistry := newRegistry()
	component := newTestComponent("config")

	if err := componentRegistry.Register(component); err != nil {
		t.Fatalf("register component: %v", err)
	}

	components := componentRegistry.Components()
	components = append(
		components,
		newTestComponent("external"),
	)

	if componentRegistry.Has("external") {
		t.Fatal("modifying returned slice changed registry")
	}

	if len(componentRegistry.Components()) != 1 {
		t.Fatal("registry contents changed unexpectedly")
	}
}
