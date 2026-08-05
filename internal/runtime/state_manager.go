package runtime

import (
	"fmt"
	"sync"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type stateManager struct {
	mu sync.RWMutex

	states map[pkgRuntime.ComponentID]pkgRuntime.ComponentState
}

func newStateManager() *stateManager {
	return &stateManager{
		states: make(map[pkgRuntime.ComponentID]pkgRuntime.ComponentState),
	}
}

func (m *stateManager) Get(
	component pkgRuntime.Component,
) pkgRuntime.ComponentState {
	if component == nil {
		return pkgRuntime.ComponentState{
			State: pkgRuntime.StateUnknown,
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	componentID := component.Metadata().ID

	componentState, exists := m.states[componentID]
	if !exists {
		return pkgRuntime.ComponentState{
			Component: component,
			State:     pkgRuntime.StateUnknown,
		}
	}

	return componentState
}

func (m *stateManager) Set(
	component pkgRuntime.Component,
	state pkgRuntime.State,
) {
	if component == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[component.Metadata().ID] = pkgRuntime.ComponentState{
		Component: component,
		State:     state,
	}
}

func (m *stateManager) Fail(
	component pkgRuntime.Component,
	err error,
) {
	if component == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[component.Metadata().ID] = pkgRuntime.ComponentState{
		Component: component,
		State:     pkgRuntime.StateFailed,
		Error:     err,
	}
}

func (m *stateManager) create(
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"create component state: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	componentID := component.Metadata().ID
	if _, exists := m.states[componentID]; exists {
		return nil
	}

	m.states[componentID] = pkgRuntime.ComponentState{
		Component: component,
		State:     pkgRuntime.StateCreated,
	}

	return nil
}

func (m *stateManager) transition(
	component pkgRuntime.Component,
	to pkgRuntime.State,
) error {
	if component == nil {
		return fmt.Errorf(
			"transition component state: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	componentID := component.Metadata().ID

	current, exists := m.states[componentID]
	if !exists {
		return fmt.Errorf(
			"transition component %q to state %d: current state is unknown: %w",
			componentID,
			to,
			pkgRuntime.ErrInvalidState,
		)
	}

	if !isValidTransition(current.State, to) {
		return fmt.Errorf(
			"transition component %q from state %d to state %d: %w",
			componentID,
			current.State,
			to,
			pkgRuntime.ErrInvalidState,
		)
	}

	m.states[componentID] = pkgRuntime.ComponentState{
		Component: component,
		State:     to,
	}

	return nil
}

func (m *stateManager) reset(
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"reset component state: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[component.Metadata().ID] = pkgRuntime.ComponentState{
		Component: component,
		State:     pkgRuntime.StateCreated,
	}

	return nil
}

func isValidTransition(
	from pkgRuntime.State,
	to pkgRuntime.State,
) bool {
	for _, transition := range pkgRuntime.ValidTransitions {
		if transition.From == from && transition.To == to {
			return true
		}
	}

	return false
}
