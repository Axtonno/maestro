package runtime

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