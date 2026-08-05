package runtime

import (
	"errors"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestStateManagerCreatesComponentState(t *testing.T) {
	componentStateManager := newStateManager()
	component := newTestComponent("config")

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateCreated {
		t.Fatalf(
			"expected state created, got %d",
			componentState.State,
		)
	}
}

func TestStateManagerTransitionsComponentState(t *testing.T) {
	componentStateManager := newStateManager()
	component := newTestComponent("config")

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	if err := componentStateManager.transition(
		component,
		pkgRuntime.StateConfigured,
	); err != nil {
		t.Fatalf("transition component: %v", err)
	}

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateConfigured {
		t.Fatalf(
			"expected state configured, got %d",
			componentState.State,
		)
	}
}

func TestStateManagerRejectsInvalidTransition(t *testing.T) {
	componentStateManager := newStateManager()
	component := newTestComponent("config")

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	err := componentStateManager.transition(
		component,
		pkgRuntime.StateRunning,
	)
	if !errors.Is(err, pkgRuntime.ErrInvalidState) {
		t.Fatalf(
			"expected ErrInvalidState, got %v",
			err,
		)
	}
}

func TestStateManagerMarksComponentAsFailed(t *testing.T) {
	componentStateManager := newStateManager()
	component := newTestComponent("config")
	cause := errors.New("boom")

	componentStateManager.Fail(component, cause)

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateFailed {
		t.Fatalf(
			"expected state failed, got %d",
			componentState.State,
		)
	}

	if !errors.Is(componentState.Error, cause) {
		t.Fatalf(
			"expected failure cause %v, got %v",
			cause,
			componentState.Error,
		)
	}
}
