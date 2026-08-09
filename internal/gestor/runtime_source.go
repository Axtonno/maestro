package gestor

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

const runtimeComponentSourceID pkgGestor.SourceID = "runtime.components"

type componentCatalog interface {
	Components() []pkgRuntime.Component
}

type RuntimeComponentSource struct {
	catalog componentCatalog
}

func NewRuntimeComponentSource(catalog componentCatalog) (*RuntimeComponentSource, error) {
	if nilInterface(catalog) {
		return nil, fmt.Errorf("runtime component catalog is nil: %w", pkgGestor.ErrInvalidSource)
	}

	return &RuntimeComponentSource{catalog: catalog}, nil
}

func (source *RuntimeComponentSource) ID() pkgGestor.SourceID {
	return runtimeComponentSourceID
}

func (source *RuntimeComponentSource) Discover(
	ctx context.Context,
) ([]pkgGestor.Descriptor, error) {
	if ctx == nil {
		return nil, fmt.Errorf("discover runtime components with nil context: %w", pkgGestor.ErrInvalidSource)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	components := source.catalog.Components()
	metadata := make([]pkgRuntime.Metadata, 0, len(components))
	for index, component := range components {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if nilInterface(component) {
			return nil, fmt.Errorf("runtime component %d is nil: %w", index, pkgGestor.ErrInvalidDescriptor)
		}
		metadata = append(metadata, component.Metadata())
	}
	slices.SortFunc(metadata, func(left, right pkgRuntime.Metadata) int {
		if left.ID < right.ID {
			return -1
		}
		if left.ID > right.ID {
			return 1
		}

		return 0
	})

	descriptors := make([]pkgGestor.Descriptor, 0)
	seenComponents := make(map[pkgRuntime.ComponentID]struct{}, len(metadata))
	for _, component := range metadata {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		target := pkgGestor.Target{
			Kind:  pkgGestor.TargetKindComponent,
			ID:    string(component.ID),
			Scope: pkgGestor.ScopeComponent,
		}
		if err := target.Validate(); err != nil {
			return nil, fmt.Errorf("runtime component %q target: %w: %w", component.ID, err, pkgGestor.ErrInvalidDescriptor)
		}
		if _, exists := seenComponents[component.ID]; exists {
			return nil, fmt.Errorf("runtime component %q is duplicated: %w", component.ID, pkgGestor.ErrInvalidDescriptor)
		}
		seenComponents[component.ID] = struct{}{}

		seenCapabilities := make(map[pkgRuntime.Capability]struct{}, len(component.Capabilities))
		for _, capability := range component.Capabilities {
			if _, exists := seenCapabilities[capability]; exists {
				return nil, fmt.Errorf("runtime component %q capability %q is duplicated: %w", component.ID, capability, pkgGestor.ErrInvalidDescriptor)
			}
			seenCapabilities[capability] = struct{}{}
			capabilityID, err := runtimeCapabilityID(capability)
			if err != nil {
				return nil, fmt.Errorf("runtime component %q capability %q: %w: %w", component.ID, capability, err, pkgGestor.ErrInvalidDescriptor)
			}
			descriptors = append(descriptors, pkgGestor.Descriptor{
				Capability:   capabilityID,
				Target:       target,
				Availability: pkgGestor.AvailabilityUnknown,
				Source:       runtimeComponentSourceID,
			})
		}
	}

	slices.SortFunc(descriptors, func(left, right pkgGestor.Descriptor) int {
		return left.Compare(right)
	})

	return descriptors, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
