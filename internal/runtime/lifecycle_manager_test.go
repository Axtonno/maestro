package runtime

import (
	"context"
	"errors"
	"reflect"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestLifecycleManagerStartsComponent(t *testing.T) {
	calls := make([]string, 0)
	component := newLifecycleTestComponent("config", &calls)
	componentStateManager := newStateManager()

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	lifecycle := newTestLifecycleManager(componentStateManager)

	if err := lifecycle.Start(context.Background(), component); err != nil {
		t.Fatalf("start component: %v", err)
	}

	if got, want := calls, []string{
		"config:configure",
		"config:initialize",
		"config:start",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateRunning {
		t.Fatalf(
			"expected state running, got %d",
			componentState.State,
		)
	}
}

func TestLifecycleManagerStopsComponent(t *testing.T) {
	calls := make([]string, 0)
	component := newLifecycleTestComponent("config", &calls)
	componentStateManager := newStateManager()

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	lifecycle := newTestLifecycleManager(componentStateManager)

	if err := lifecycle.Start(context.Background(), component); err != nil {
		t.Fatalf("start component: %v", err)
	}

	calls = calls[:0]

	if err := lifecycle.Stop(context.Background(), component); err != nil {
		t.Fatalf("stop component: %v", err)
	}

	if got, want := calls, []string{
		"config:stop",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected calls %v, got %v", want, got)
	}

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateStopped {
		t.Fatalf(
			"expected state stopped, got %d",
			componentState.State,
		)
	}
}

func TestLifecycleManagerMarksFailedComponent(t *testing.T) {
	calls := make([]string, 0)
	component := newLifecycleTestComponent("config", &calls)
	component.failOn = "initialize"
	componentStateManager := newStateManager()

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	lifecycle := newTestLifecycleManager(componentStateManager)

	err := lifecycle.Start(context.Background(), component)
	if err == nil {
		t.Fatal("expected start error")
	}

	componentState := componentStateManager.Get(component)
	if componentState.State != pkgRuntime.StateFailed {
		t.Fatalf(
			"expected state failed, got %d",
			componentState.State,
		)
	}
}

func TestLifecycleManagerRejectsStoppingNonRunningComponent(t *testing.T) {
	calls := make([]string, 0)
	component := newLifecycleTestComponent("config", &calls)
	componentStateManager := newStateManager()

	if err := componentStateManager.create(component); err != nil {
		t.Fatalf("create component state: %v", err)
	}

	lifecycle := newTestLifecycleManager(componentStateManager)

	err := lifecycle.Stop(context.Background(), component)
	if !errors.Is(err, pkgRuntime.ErrInvalidState) {
		t.Fatalf(
			"expected ErrInvalidState, got %v",
			err,
		)
	}
}

func newTestLifecycleManager(
	componentStateManager *stateManager,
) *lifecycleManager {
	componentRegistry := newRegistry()
	componentEventBus := newEventBus()
	runtimeContext := newRuntimeContext(
		newEmptyConfig(),
		newNoopLogger(),
		componentEventBus,
		componentRegistry,
	)

	return newLifecycleManager(
		componentStateManager,
		runtimeContext,
	)
}
