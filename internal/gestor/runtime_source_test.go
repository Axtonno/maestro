package gestor

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	internalPlugin "github.com/antonio-cafeo/maestro/internal/plugin"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type sourceTestComponent struct {
	metadata pkgRuntime.Metadata
}

func (component *sourceTestComponent) Metadata() pkgRuntime.Metadata {
	metadata := component.metadata
	metadata.Capabilities = slices.Clone(component.metadata.Capabilities)
	metadata.Dependencies = slices.Clone(component.metadata.Dependencies)

	return metadata
}

type sourceTestPlugin struct {
	sourceTestComponent
}

func (plugin *sourceTestPlugin) Manifest() pkgPlugin.Manifest {
	return pkgPlugin.Manifest{RuntimeAPIVersion: pkgPlugin.RuntimeAPIVersion}
}

type componentMemoryCatalog struct {
	mu         sync.RWMutex
	components []pkgRuntime.Component
}

func (catalog *componentMemoryCatalog) Register(component pkgRuntime.Component) error {
	catalog.mu.Lock()
	defer catalog.mu.Unlock()

	for _, registered := range catalog.components {
		if registered.Metadata().ID == component.Metadata().ID {
			return pkgRuntime.ErrAlreadyRegistered
		}
	}
	catalog.components = append(catalog.components, component)

	return nil
}

func (catalog *componentMemoryCatalog) Components() []pkgRuntime.Component {
	catalog.mu.RLock()
	defer catalog.mu.RUnlock()

	return slices.Clone(catalog.components)
}

func TestRuntimeComponentSourceMapsZeroSingleAndMultipleCapabilities(t *testing.T) {
	catalog := &componentMemoryCatalog{components: []pkgRuntime.Component{
		&sourceTestComponent{metadata: pkgRuntime.Metadata{ID: "zero"}},
		&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID:           "single",
			Capabilities: []pkgRuntime.Capability{pkgRuntime.CapabilityStart},
		}},
		&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID: "multiple",
			Capabilities: []pkgRuntime.Capability{
				pkgRuntime.CapabilityStop,
				pkgRuntime.CapabilityConfigure,
				pkgRuntime.CapabilityHealth,
				pkgRuntime.CapabilityInitialize,
				pkgRuntime.CapabilityReload,
				pkgRuntime.CapabilityStart,
			},
		}},
	}}
	source, err := NewRuntimeComponentSource(catalog)
	if err != nil {
		t.Fatalf("new runtime component source: %v", err)
	}
	if source.ID() != runtimeComponentSourceID {
		t.Fatalf("expected source ID %q, got %q", runtimeComponentSourceID, source.ID())
	}

	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover runtime components: %v", err)
	}
	if len(descriptors) != 7 {
		t.Fatalf("expected 7 descriptors, got %d", len(descriptors))
	}
	if !slices.IsSortedFunc(descriptors, func(left, right pkgGestor.Descriptor) int {
		return left.Compare(right)
	}) {
		t.Fatal("runtime descriptors are not ordered")
	}

	wantCapabilities := map[pkgGestor.CapabilityID]bool{
		pkgGestor.CapabilityRuntimeConfigure:  false,
		pkgGestor.CapabilityRuntimeInitialize: false,
		pkgGestor.CapabilityRuntimeStart:      false,
		pkgGestor.CapabilityRuntimeStop:       false,
		pkgGestor.CapabilityRuntimeReload:     false,
		pkgGestor.CapabilityRuntimeHealth:     false,
	}
	for _, descriptor := range descriptors {
		if descriptor.Source != runtimeComponentSourceID ||
			descriptor.Target.Kind != pkgGestor.TargetKindComponent ||
			descriptor.Target.Scope != pkgGestor.ScopeComponent ||
			descriptor.Availability != pkgGestor.AvailabilityUnknown {
			t.Fatalf("unexpected runtime descriptor: %#v", descriptor)
		}
		if descriptor.Target.ID == "multiple" {
			wantCapabilities[descriptor.Capability] = true
		}
		if descriptor.Target.ID == "zero" {
			t.Fatal("zero-capability component produced a descriptor")
		}
	}
	for capability, found := range wantCapabilities {
		if !found {
			t.Errorf("missing mapped runtime capability %q", capability)
		}
	}
}

func TestRuntimeComponentSourceDiscoversPluginOnceThroughGlobalCatalog(t *testing.T) {
	catalog := &componentMemoryCatalog{}
	pluginRuntime := internalPlugin.NewRuntime(catalog)
	plugin := &sourceTestPlugin{sourceTestComponent: sourceTestComponent{
		metadata: pkgRuntime.Metadata{
			ID:           "laravel",
			Capabilities: []pkgRuntime.Capability{"plugin.workspace_detection"},
		},
	}}
	if err := pluginRuntime.Register(plugin); err != nil {
		t.Fatalf("register plugin: %v", err)
	}

	source, err := NewRuntimeComponentSource(catalog)
	if err != nil {
		t.Fatalf("new runtime component source: %v", err)
	}
	descriptors, err := source.Discover(context.Background())
	if err != nil {
		t.Fatalf("discover plugin component: %v", err)
	}
	if len(descriptors) != 1 {
		t.Fatalf("expected plugin exactly once, got %d descriptors", len(descriptors))
	}
	if descriptor := descriptors[0]; descriptor.Capability != "plugin.workspace_detection" || descriptor.Target.ID != "laravel" {
		t.Fatalf("unexpected plugin descriptor: %#v", descriptor)
	}
}

func TestRuntimeComponentSourceRejectsInvalidCatalogData(t *testing.T) {
	if _, err := NewRuntimeComponentSource(nil); !errors.Is(err, pkgGestor.ErrInvalidSource) {
		t.Fatalf("nil catalog: expected ErrInvalidSource, got %v", err)
	}
	var typedNil *componentMemoryCatalog
	if _, err := NewRuntimeComponentSource(typedNil); !errors.Is(err, pkgGestor.ErrInvalidSource) {
		t.Fatalf("typed nil catalog: expected ErrInvalidSource, got %v", err)
	}

	tests := []struct {
		name       string
		components []pkgRuntime.Component
	}{
		{"nil component", []pkgRuntime.Component{(*sourceTestComponent)(nil)}},
		{"invalid ID", []pkgRuntime.Component{&sourceTestComponent{metadata: pkgRuntime.Metadata{ID: " bad"}}}},
		{"duplicate ID", []pkgRuntime.Component{
			&sourceTestComponent{metadata: pkgRuntime.Metadata{ID: "same"}},
			&sourceTestComponent{metadata: pkgRuntime.Metadata{ID: "same"}},
		}},
		{"duplicate capability", []pkgRuntime.Component{&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID: "same",
			Capabilities: []pkgRuntime.Capability{
				pkgRuntime.CapabilityStart,
				pkgRuntime.CapabilityStart,
			},
		}}}},
		{"unnamespaced capability", []pkgRuntime.Component{&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID:           "same",
			Capabilities: []pkgRuntime.Capability{"future"},
		}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, err := NewRuntimeComponentSource(&componentMemoryCatalog{components: test.components})
			if err != nil {
				t.Fatalf("new source: %v", err)
			}
			_, err = source.Discover(context.Background())
			if !errors.Is(err, pkgGestor.ErrInvalidDescriptor) {
				t.Fatalf("expected ErrInvalidDescriptor, got %v", err)
			}
		})
	}
}

func TestRuntimeComponentSourcePropagatesCancellation(t *testing.T) {
	source, err := NewRuntimeComponentSource(&componentMemoryCatalog{})
	if err != nil {
		t.Fatalf("new source: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.Discover(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestRuntimeComponentSourceConcurrentDiscovery(t *testing.T) {
	catalog := &componentMemoryCatalog{components: []pkgRuntime.Component{
		&sourceTestComponent{metadata: pkgRuntime.Metadata{
			ID:           "component",
			Capabilities: []pkgRuntime.Capability{pkgRuntime.CapabilityHealth},
		}},
	}}
	source, err := NewRuntimeComponentSource(catalog)
	if err != nil {
		t.Fatalf("new source: %v", err)
	}

	const workers = 16
	var wait sync.WaitGroup
	results := make(chan error, workers)
	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			descriptors, err := source.Discover(context.Background())
			if err == nil && len(descriptors) != 1 {
				err = errors.New("unexpected descriptor count")
			}
			results <- err
		}()
	}
	wait.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Errorf("concurrent discovery: %v", err)
		}
	}
}
