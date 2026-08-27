package ollama

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) InspectCapabilities(
	ctx context.Context,
	request pkgProvider.CapabilityRequest,
) (result pkgProvider.CapabilityReport, operationError error) {
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationCapabilityIntrospection,
			request.Model,
			operationError,
		)
	}()

	if err := request.Validate(); err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect Ollama capabilities: %w",
			err,
		)
	}

	report := newOllamaCapabilityReport(request)
	if request.Target == pkgProvider.CapabilityTargetAdapter {
		return report, nil
	}

	models, err := p.Models(ctx)
	if err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect Ollama capabilities: probe instance: %w",
			err,
		)
	}
	markOllamaInstanceCapabilities(&report)
	if request.Target == pkgProvider.CapabilityTargetInstance {
		return report, nil
	}

	found := false
	for _, model := range models {
		if model.ID == request.Model {
			found = true
			break
		}
	}
	if !found {
		markMissingOllamaModel(&report)

		return report, nil
	}

	response := modelShowResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/api/show",
		modelShowRequest{Model: request.Model},
		&response,
	); err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect Ollama capabilities: show model %q: %w",
			request.Model,
			err,
		)
	}
	if response.Error != "" {
		return pkgProvider.CapabilityReport{}, &apiError{message: response.Error}
	}

	observed := make(map[string]struct{}, len(response.Capabilities))
	for _, capability := range response.Capabilities {
		if capability == "" {
			return pkgProvider.CapabilityReport{}, fmt.Errorf(
				"inspect Ollama capabilities: model %q has an empty capability: %w",
				request.Model,
				pkgProvider.ErrInvalidResponse,
			)
		}
		if _, exists := observed[capability]; exists {
			return pkgProvider.CapabilityReport{}, fmt.Errorf(
				"inspect Ollama capabilities: model %q repeats capability %q: %w",
				request.Model,
				capability,
				pkgProvider.ErrInvalidResponse,
			)
		}
		observed[capability] = struct{}{}
	}

	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityCompletion,
		availabilityFromOllamaCapability(observed, "completion"),
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityStreaming,
		availabilityFromOllamaCapability(observed, "completion"),
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityEmbedding,
		availabilityFromOllamaCapability(observed, "embedding"),
	)
	completionAvailability := availabilityFromOllamaCapability(
		observed,
		"completion",
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityStructuredOutput,
		completionAvailability,
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityToolCalling,
		availabilityFromOllamaCapability(observed, "tools"),
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityContextWindowControl,
		completionAvailability,
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityThinkingControl,
		completionAvailability,
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityThinking,
		availabilityFromOllamaCapability(observed, "thinking"),
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityAvailabilityAvailable,
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityAvailabilityAvailable,
	)
	setOllamaAvailability(
		&report,
		pkgProvider.CapabilityModelRemove,
		pkgProvider.CapabilityAvailabilityAvailable,
	)

	return report, nil
}

func newOllamaCapabilityReport(
	request pkgProvider.CapabilityRequest,
) pkgProvider.CapabilityReport {
	descriptors := make(
		[]pkgProvider.CapabilityDescriptor,
		0,
		len(pkgProvider.KnownCapabilities()),
	)
	for _, capability := range pkgProvider.KnownCapabilities() {
		support := pkgProvider.CapabilitySupported
		availability := pkgProvider.CapabilityAvailabilityUnknown
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{
			Capability: capability, Support: support, Availability: availability,
		})
	}

	return pkgProvider.CapabilityReport{
		Provider: providerID, Target: request.Target, Model: request.Model,
		Capabilities: descriptors,
	}
}

func markOllamaInstanceCapabilities(report *pkgProvider.CapabilityReport) {
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityModelListing,
		pkgProvider.CapabilityModelDiscovery,
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityModelPull,
		pkgProvider.CapabilityModelRemove,
	} {
		setOllamaAvailability(
			report,
			capability,
			pkgProvider.CapabilityAvailabilityAvailable,
		)
	}
}

func markMissingOllamaModel(report *pkgProvider.CapabilityReport) {
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityCompletion,
		pkgProvider.CapabilityStreaming,
		pkgProvider.CapabilityEmbedding,
		pkgProvider.CapabilityStructuredOutput,
		pkgProvider.CapabilityToolCalling,
		pkgProvider.CapabilityContextWindowControl,
		pkgProvider.CapabilityThinkingControl,
		pkgProvider.CapabilityThinking,
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityModelRemove,
	} {
		setOllamaAvailability(
			report,
			capability,
			pkgProvider.CapabilityAvailabilityUnavailable,
		)
	}
}

func availabilityFromOllamaCapability(
	capabilities map[string]struct{},
	wanted string,
) pkgProvider.CapabilityAvailability {
	if _, exists := capabilities[wanted]; exists {
		return pkgProvider.CapabilityAvailabilityAvailable
	}

	return pkgProvider.CapabilityAvailabilityUnavailable
}

func setOllamaAvailability(
	report *pkgProvider.CapabilityReport,
	capability pkgProvider.Capability,
	availability pkgProvider.CapabilityAvailability,
) {
	for index := range report.Capabilities {
		if report.Capabilities[index].Capability == capability {
			report.Capabilities[index].Availability = availability

			return
		}
	}
}
