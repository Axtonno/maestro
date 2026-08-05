package runtime

import (
	"context"
	"fmt"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type lifecycleManager struct {
	stateManager *stateManager
	context      pkgRuntime.Context
}

func newLifecycleManager(
	stateManager *stateManager,
	context pkgRuntime.Context,
) *lifecycleManager {
	return &lifecycleManager{
		stateManager: stateManager,
		context:      context,
	}
}

func (m *lifecycleManager) Start(
	ctx context.Context,
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"start component: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	componentID := component.Metadata().ID
	componentState := m.stateManager.Get(component)

	switch componentState.State {
	case pkgRuntime.StateUnknown:
		if err := m.stateManager.create(component); err != nil {
			return err
		}
	case pkgRuntime.StateCreated:
	case pkgRuntime.StateRunning:
		return fmt.Errorf(
			"start component %q: %w",
			componentID,
			pkgRuntime.ErrAlreadyStarted,
		)
	default:
		return fmt.Errorf(
			"start component %q from state %d: %w",
			componentID,
			componentState.State,
			pkgRuntime.ErrInvalidState,
		)
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	if configurer, ok := component.(pkgRuntime.Configurer); ok {
		if err := configurer.Configure(m.context); err != nil {
			m.stateManager.Fail(component, err)

			return fmt.Errorf(
				"configure component %q: %w",
				componentID,
				err,
			)
		}
	}

	if err := m.stateManager.transition(
		component,
		pkgRuntime.StateConfigured,
	); err != nil {
		return err
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	if initializer, ok := component.(pkgRuntime.Initializer); ok {
		if err := initializer.Initialize(m.context); err != nil {
			m.stateManager.Fail(component, err)

			return fmt.Errorf(
				"initialize component %q: %w",
				componentID,
				err,
			)
		}
	}

	if err := m.stateManager.transition(
		component,
		pkgRuntime.StateInitialized,
	); err != nil {
		return err
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	if starter, ok := component.(pkgRuntime.Starter); ok {
		if err := starter.Start(m.context); err != nil {
			m.stateManager.Fail(component, err)

			return fmt.Errorf(
				"start component %q: %w",
				componentID,
				err,
			)
		}
	}

	if err := m.stateManager.transition(
		component,
		pkgRuntime.StateRunning,
	); err != nil {
		return err
	}

	return nil
}

func (m *lifecycleManager) Stop(
	ctx context.Context,
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"stop component: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	componentID := component.Metadata().ID
	componentState := m.stateManager.Get(component)

	switch componentState.State {
	case pkgRuntime.StateRunning:
	case pkgRuntime.StateStopped:
		return fmt.Errorf(
			"stop component %q: %w",
			componentID,
			pkgRuntime.ErrAlreadyStopped,
		)
	default:
		return fmt.Errorf(
			"stop component %q from state %d: %w",
			componentID,
			componentState.State,
			pkgRuntime.ErrInvalidState,
		)
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	if err := m.stateManager.transition(
		component,
		pkgRuntime.StateStopping,
	); err != nil {
		return err
	}

	if stopper, ok := component.(pkgRuntime.Stopper); ok {
		if err := stopper.Stop(m.context); err != nil {
			m.stateManager.Fail(component, err)

			return fmt.Errorf(
				"stop component %q: %w",
				componentID,
				err,
			)
		}
	}

	if err := m.stateManager.transition(
		component,
		pkgRuntime.StateStopped,
	); err != nil {
		return err
	}

	return nil
}

func (m *lifecycleManager) Restart(
	ctx context.Context,
	component pkgRuntime.Component,
) error {
	componentState := m.stateManager.Get(component)

	if componentState.State == pkgRuntime.StateRunning {
		if err := m.Stop(ctx, component); err != nil {
			return err
		}
	}

	if err := m.stateManager.reset(component); err != nil {
		return err
	}

	return m.Start(ctx, component)
}

func (m *lifecycleManager) Reload(
	ctx context.Context,
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"reload component: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	componentID := component.Metadata().ID

	if err := m.requireRunning(component, "reload"); err != nil {
		return err
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	reloader, ok := component.(pkgRuntime.Reloader)
	if !ok {
		return nil
	}

	if err := reloader.Reload(m.context); err != nil {
		m.stateManager.Fail(component, err)

		return fmt.Errorf(
			"reload component %q: %w",
			componentID,
			err,
		)
	}

	return nil
}

func (m *lifecycleManager) Health(
	ctx context.Context,
	component pkgRuntime.Component,
) error {
	if component == nil {
		return fmt.Errorf(
			"check component health: component is nil: %w",
			pkgRuntime.ErrInvalidMetadata,
		)
	}

	componentID := component.Metadata().ID

	if err := m.requireRunning(component, "check health"); err != nil {
		return err
	}

	if err := checkContext(ctx); err != nil {
		return err
	}

	healthChecker, ok := component.(pkgRuntime.HealthChecker)
	if !ok {
		return nil
	}

	if err := healthChecker.Health(m.context); err != nil {
		m.stateManager.Fail(component, err)

		return fmt.Errorf(
			"check component %q health: %w",
			componentID,
			err,
		)
	}

	return nil
}

func (m *lifecycleManager) requireRunning(
	component pkgRuntime.Component,
	operation string,
) error {
	componentState := m.stateManager.Get(component)

	if componentState.State != pkgRuntime.StateRunning {
		return fmt.Errorf(
			"%s component %q from state %d: %w",
			operation,
			component.Metadata().ID,
			componentState.State,
			pkgRuntime.ErrInvalidState,
		)
	}

	return nil
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
