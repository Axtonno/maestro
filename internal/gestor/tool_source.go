package gestor

import (
	"context"
	"fmt"
	"slices"
	"strings"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const toolCatalogSourceID pkgGestor.SourceID = "tool.catalog"

type toolCatalog interface {
	Descriptors() []pkgTool.Descriptor
}

type ToolCatalogSource struct{ catalog toolCatalog }

func NewToolCatalogSource(catalog toolCatalog) (*ToolCatalogSource, error) {
	if nilInterface(catalog) {
		return nil, fmt.Errorf("tool catalog is nil: %w", pkgGestor.ErrInvalidSource)
	}
	return &ToolCatalogSource{catalog: catalog}, nil
}

func (*ToolCatalogSource) ID() pkgGestor.SourceID { return toolCatalogSourceID }

func (source *ToolCatalogSource) Discover(ctx context.Context) ([]pkgGestor.Descriptor, error) {
	if ctx == nil {
		return nil, fmt.Errorf("discover tool catalog with nil context: %w", pkgGestor.ErrInvalidSource)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	descriptors := make([]pkgGestor.Descriptor, 0)
	for _, toolDescriptor := range source.catalog.Descriptors() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := toolDescriptor.Validate(); err != nil {
			return nil, fmt.Errorf("tool descriptor %q: %w: %w", toolDescriptor.ID(), err, pkgGestor.ErrInvalidDescriptor)
		}
		target := pkgGestor.Target{Kind: pkgGestor.TargetKindTool, ID: string(toolDescriptor.ID()), Scope: pkgGestor.ScopeTool}
		capabilities := []pkgGestor.CapabilityID{pkgGestor.CapabilityToolInvoke}
		for _, effect := range toolDescriptor.Effects() {
			capabilities = append(capabilities, pkgGestor.CapabilityID("tool.effect."+strings.ReplaceAll(string(effect), ".", "_")))
		}
		for _, capability := range capabilities {
			descriptor := pkgGestor.Descriptor{
				Capability: capability, Target: target,
				Availability: pkgGestor.AvailabilityAvailable, Source: toolCatalogSourceID,
			}
			if err := descriptor.Validate(); err != nil {
				return nil, err
			}
			descriptors = append(descriptors, descriptor)
		}
	}
	slices.SortFunc(descriptors, func(left, right pkgGestor.Descriptor) int { return left.Compare(right) })
	return descriptors, nil
}
