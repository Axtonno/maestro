package provider

import (
	"context"
	"fmt"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (r *runtime) Capabilities(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	if ctx == nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect provider capabilities: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}
	if err := request.Validate(); err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect provider capabilities: %w",
			err,
		)
	}

	selected, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect capabilities with provider %q: %w",
			providerID,
			err,
		)
	}

	inspector, ok := selected.(pkgProvider.CapabilityInspector)
	if !ok {
		return pkgProvider.CapabilityReport{}, unsupportedCapability(
			selected.ID(),
			"capability introspection",
		)
	}

	report, err := inspector.InspectCapabilities(ctx, request)
	if err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect capabilities with provider %q: %w",
			selected.ID(),
			err,
		)
	}
	if err := validateCapabilityReport(selected.ID(), request, report); err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect capabilities with provider %q: %w",
			selected.ID(),
			err,
		)
	}

	return report, nil
}

func validateCapabilityReport(
	providerID pkgProvider.ID,
	request pkgProvider.CapabilityRequest,
	report pkgProvider.CapabilityReport,
) error {
	if report.Provider != providerID || report.Target != request.Target ||
		report.Model != request.Model {
		return fmt.Errorf(
			"capability report identity does not match the request: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	known := pkgProvider.KnownCapabilities()
	if len(report.Capabilities) != len(known) {
		return fmt.Errorf(
			"capability report has %d descriptors, expected %d: %w",
			len(report.Capabilities),
			len(known),
			pkgProvider.ErrInvalidResponse,
		)
	}

	for index, capability := range known {
		descriptor := report.Capabilities[index]
		if descriptor.Capability != capability {
			return fmt.Errorf(
				"capability report descriptor %d is %q, expected %q: %w",
				index,
				descriptor.Capability,
				capability,
				pkgProvider.ErrInvalidResponse,
			)
		}
		if descriptor.Support != pkgProvider.CapabilitySupported &&
			descriptor.Support != pkgProvider.CapabilityUnsupported {
			return fmt.Errorf(
				"capability %q has invalid support %q: %w",
				capability,
				descriptor.Support,
				pkgProvider.ErrInvalidResponse,
			)
		}
		switch descriptor.Availability {
		case pkgProvider.CapabilityAvailabilityUnknown,
			pkgProvider.CapabilityAvailabilityAvailable,
			pkgProvider.CapabilityAvailabilityUnavailable:
		default:
			return fmt.Errorf(
				"capability %q has invalid availability %q: %w",
				capability,
				descriptor.Availability,
				pkgProvider.ErrInvalidResponse,
			)
		}
		if descriptor.Support == pkgProvider.CapabilityUnsupported &&
			descriptor.Availability != pkgProvider.CapabilityAvailabilityUnavailable {
			return fmt.Errorf(
				"unsupported capability %q must be unavailable: %w",
				capability,
				pkgProvider.ErrInvalidResponse,
			)
		}
	}

	return nil
}
