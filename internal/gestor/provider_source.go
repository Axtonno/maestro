package gestor

import (
	"context"
	"fmt"
	"slices"
	"strings"

	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const providerCapabilitySourceID pkgGestor.SourceID = "provider.capabilities"

type providerCapabilityRuntime interface {
	Registered() []pkgProvider.ID
	Capabilities(
		context.Context,
		pkgProvider.ID,
		pkgProvider.CapabilityRequest,
	) (pkgProvider.CapabilityReport, error)
}

type ProviderCapabilitySource struct {
	runtime providerCapabilityRuntime
	models  map[pkgProvider.ID][]string
}

func NewProviderCapabilitySource(
	runtime providerCapabilityRuntime,
	models map[pkgProvider.ID][]string,
) (*ProviderCapabilitySource, error) {
	if nilInterface(runtime) {
		return nil, fmt.Errorf("provider capability runtime is nil: %w", pkgGestor.ErrInvalidSource)
	}

	copiedModels := make(map[pkgProvider.ID][]string, len(models))
	for providerID, modelIDs := range models {
		if !exactProviderValue(string(providerID)) {
			return nil, fmt.Errorf("provider model target has invalid provider ID %q: %w", providerID, pkgGestor.ErrInvalidSource)
		}
		copied := slices.Clone(modelIDs)
		slices.Sort(copied)
		for index, modelID := range copied {
			if !exactProviderValue(modelID) {
				return nil, fmt.Errorf("provider %q model target %q is not exact: %w", providerID, modelID, pkgGestor.ErrInvalidSource)
			}
			if index > 0 && copied[index-1] == modelID {
				return nil, fmt.Errorf("provider %q model target %q is duplicated: %w", providerID, modelID, pkgGestor.ErrInvalidSource)
			}
		}
		copiedModels[providerID] = copied
	}

	return &ProviderCapabilitySource{runtime: runtime, models: copiedModels}, nil
}

func (source *ProviderCapabilitySource) ID() pkgGestor.SourceID {
	return providerCapabilitySourceID
}

func (source *ProviderCapabilitySource) Discover(
	ctx context.Context,
) ([]pkgGestor.Descriptor, error) {
	if ctx == nil {
		return nil, fmt.Errorf("discover provider capabilities with nil context: %w", pkgGestor.ErrInvalidSource)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	providerIDs := source.runtime.Registered()
	slices.Sort(providerIDs)
	for index, providerID := range providerIDs {
		if !exactProviderValue(string(providerID)) {
			return nil, fmt.Errorf("registered provider ID %q is invalid: %w", providerID, pkgGestor.ErrInvalidDescriptor)
		}
		if index > 0 && providerIDs[index-1] == providerID {
			return nil, fmt.Errorf("registered provider ID %q is duplicated: %w", providerID, pkgGestor.ErrInvalidDescriptor)
		}
	}

	descriptors := make([]pkgGestor.Descriptor, 0)
	for _, providerID := range providerIDs {
		requests := []pkgProvider.CapabilityRequest{
			{Target: pkgProvider.CapabilityTargetAdapter},
			{Target: pkgProvider.CapabilityTargetInstance},
		}
		for _, modelID := range source.models[providerID] {
			requests = append(requests, pkgProvider.CapabilityRequest{
				Target: pkgProvider.CapabilityTargetModel,
				Model:  modelID,
			})
		}

		for _, request := range requests {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			report, err := source.runtime.Capabilities(ctx, providerID, request)
			if err != nil {
				return nil, fmt.Errorf("inspect provider %q target %q model %q: %w", providerID, request.Target, request.Model, err)
			}
			mapped, err := mapCapabilityReport(providerID, request, report)
			if err != nil {
				return nil, err
			}
			descriptors = append(descriptors, mapped...)
		}
	}

	slices.SortFunc(descriptors, func(left, right pkgGestor.Descriptor) int {
		return left.Compare(right)
	})

	return descriptors, nil
}

func mapCapabilityReport(
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
	report pkgProvider.CapabilityReport,
) ([]pkgGestor.Descriptor, error) {
	if report.Provider != providerID || report.Target != request.Target || report.Model != request.Model {
		return nil, fmt.Errorf("provider %q capability report identity mismatch: %w", providerID, pkgGestor.ErrInvalidDescriptor)
	}
	if len(report.Capabilities) != len(pkgProvider.KnownCapabilities()) {
		return nil, fmt.Errorf("provider %q capability report has %d descriptors: %w", providerID, len(report.Capabilities), pkgGestor.ErrInvalidDescriptor)
	}

	target, err := providerTarget(providerID, request)
	if err != nil {
		return nil, err
	}
	descriptors := make([]pkgGestor.Descriptor, 0, len(report.Capabilities))
	known := pkgProvider.KnownCapabilities()
	for index, capability := range known {
		reported := report.Capabilities[index]
		if reported.Capability != capability {
			return nil, fmt.Errorf("provider %q capability descriptor %d is %q, expected %q: %w", providerID, index, reported.Capability, capability, pkgGestor.ErrInvalidDescriptor)
		}
		capabilityID, err := providerCapabilityID(reported.Capability)
		if err != nil {
			return nil, fmt.Errorf("provider %q capability %q: %w: %w", providerID, reported.Capability, err, pkgGestor.ErrInvalidDescriptor)
		}

		switch reported.Support {
		case pkgProvider.CapabilityUnsupported:
			if reported.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
				return nil, fmt.Errorf("provider %q unsupported capability %q is %q: %w", providerID, reported.Capability, reported.Availability, pkgGestor.ErrInvalidDescriptor)
			}
			continue
		case pkgProvider.CapabilitySupported:
		default:
			return nil, fmt.Errorf("provider %q capability %q has support %q: %w", providerID, reported.Capability, reported.Support, pkgGestor.ErrInvalidDescriptor)
		}

		availability, err := providerAvailability(reported.Availability)
		if err != nil {
			return nil, fmt.Errorf("provider %q capability %q: %w", providerID, reported.Capability, err)
		}
		descriptors = append(descriptors, pkgGestor.Descriptor{
			Capability:   capabilityID,
			Target:       target,
			Availability: availability,
			Source:       providerCapabilitySourceID,
		})
	}

	return descriptors, nil
}

func providerTarget(
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
) (pkgGestor.Target, error) {
	target := pkgGestor.Target{Kind: pkgGestor.TargetKindProvider, ID: string(providerID)}
	switch request.Target {
	case pkgProvider.CapabilityTargetAdapter:
		target.Scope = pkgGestor.ScopeAdapter
	case pkgProvider.CapabilityTargetInstance:
		target.Scope = pkgGestor.ScopeInstance
	case pkgProvider.CapabilityTargetModel:
		target.Scope = pkgGestor.ScopeModel
		target.Model = request.Model
	default:
		return pkgGestor.Target{}, fmt.Errorf("provider %q capability target %q is unknown: %w", providerID, request.Target, pkgGestor.ErrInvalidDescriptor)
	}
	if err := target.Validate(); err != nil {
		return pkgGestor.Target{}, fmt.Errorf("provider %q capability target: %w: %w", providerID, err, pkgGestor.ErrInvalidDescriptor)
	}

	return target, nil
}

func providerAvailability(
	availability pkgProvider.CapabilityAvailability,
) (pkgGestor.Availability, error) {
	switch availability {
	case pkgProvider.CapabilityAvailabilityUnknown:
		return pkgGestor.AvailabilityUnknown, nil
	case pkgProvider.CapabilityAvailabilityAvailable:
		return pkgGestor.AvailabilityAvailable, nil
	case pkgProvider.CapabilityAvailabilityUnavailable:
		return pkgGestor.AvailabilityUnavailable, nil
	default:
		return "", fmt.Errorf("provider availability %q is unknown: %w", availability, pkgGestor.ErrInvalidDescriptor)
	}
}

func exactProviderValue(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}
