package gestor

import (
	"context"
	"fmt"
	"slices"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
)

const agentCatalogSourceID pkgGestor.SourceID = "agent.catalog"

type agentCatalog interface {
	Descriptors() []pkgAgent.Descriptor
}

type AgentCatalogSource struct{ catalog agentCatalog }

func NewAgentCatalogSource(catalog agentCatalog) (*AgentCatalogSource, error) {
	if nilInterface(catalog) {
		return nil, fmt.Errorf("agent catalog is nil: %w", pkgGestor.ErrInvalidSource)
	}
	return &AgentCatalogSource{catalog: catalog}, nil
}

func (*AgentCatalogSource) ID() pkgGestor.SourceID { return agentCatalogSourceID }

func (source *AgentCatalogSource) Discover(ctx context.Context) ([]pkgGestor.Descriptor, error) {
	if ctx == nil {
		return nil, fmt.Errorf("discover agent catalog with nil context: %w", pkgGestor.ErrInvalidSource)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	descriptors := make([]pkgGestor.Descriptor, 0)
	for _, agentDescriptor := range source.catalog.Descriptors() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := agentDescriptor.Validate(); err != nil {
			return nil, fmt.Errorf("agent descriptor %q: %w: %w", agentDescriptor.ID(), err, pkgGestor.ErrInvalidDescriptor)
		}
		target := pkgGestor.Target{Kind: pkgGestor.TargetKindAgent, ID: string(agentDescriptor.ID()), Scope: pkgGestor.ScopeAgent}
		for _, capability := range agentDescriptor.Capabilities() {
			descriptor := pkgGestor.Descriptor{
				Capability: pkgGestor.CapabilityID(capability), Target: target,
				Availability: pkgGestor.AvailabilityAvailable, Source: agentCatalogSourceID,
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
