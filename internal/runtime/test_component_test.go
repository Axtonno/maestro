package runtime

import "fmt"

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

type testComponent struct {
	metadata pkgRuntime.Metadata
}

func newTestComponent(
	id pkgRuntime.ComponentID,
	dependencies ...pkgRuntime.Dependency,
) *testComponent {
	return &testComponent{
		metadata: pkgRuntime.Metadata{
			ID:           id,
			Name:         string(id),
			Version:      "1.0.0",
			Description:  "Test component",
			Dependencies: dependencies,
		},
	}
}

func (c *testComponent) Metadata() pkgRuntime.Metadata {
	return c.metadata
}

type lifecycleTestComponent struct {
	metadata pkgRuntime.Metadata
	calls    *[]string
	failOn   string
}

func newLifecycleTestComponent(
	id pkgRuntime.ComponentID,
	calls *[]string,
	dependencies ...pkgRuntime.Dependency,
) *lifecycleTestComponent {
	return &lifecycleTestComponent{
		metadata: pkgRuntime.Metadata{
			ID:           id,
			Name:         string(id),
			Version:      "1.0.0",
			Description:  "Lifecycle test component",
			Dependencies: dependencies,
			Capabilities: []pkgRuntime.Capability{
				pkgRuntime.CapabilityConfigure,
				pkgRuntime.CapabilityInitialize,
				pkgRuntime.CapabilityStart,
				pkgRuntime.CapabilityStop,
				pkgRuntime.CapabilityReload,
				pkgRuntime.CapabilityHealth,
			},
		},
		calls: calls,
	}
}

func (c *lifecycleTestComponent) Metadata() pkgRuntime.Metadata {
	return c.metadata
}

func (c *lifecycleTestComponent) Configure(
	_ pkgRuntime.Context,
) error {
	return c.record("configure")
}

func (c *lifecycleTestComponent) Initialize(
	_ pkgRuntime.Context,
) error {
	return c.record("initialize")
}

func (c *lifecycleTestComponent) Start(
	_ pkgRuntime.Context,
) error {
	return c.record("start")
}

func (c *lifecycleTestComponent) Stop(
	_ pkgRuntime.Context,
) error {
	return c.record("stop")
}

func (c *lifecycleTestComponent) Reload(
	_ pkgRuntime.Context,
) error {
	return c.record("reload")
}

func (c *lifecycleTestComponent) Health(
	_ pkgRuntime.Context,
) error {
	return c.record("health")
}

func (c *lifecycleTestComponent) record(phase string) error {
	call := fmt.Sprintf("%s:%s", c.metadata.ID, phase)
	*c.calls = append(*c.calls, call)

	if c.failOn == phase {
		return fmt.Errorf("%s failed", phase)
	}

	return nil
}
