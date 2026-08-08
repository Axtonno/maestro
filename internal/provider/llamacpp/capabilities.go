package llamacpp

import (
	"context"
	"fmt"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) InspectCapabilities(
	ctx context.Context,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	if err := request.Validate(); err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect llama.cpp capabilities: %w",
			err,
		)
	}

	report := newLlamaCPPCapabilityReport(request)
	if request.Target == pkgProvider.CapabilityTargetAdapter {
		return report, nil
	}

	models, err := p.modelCatalog(ctx)
	if err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect llama.cpp capabilities: probe instance: %w",
			err,
		)
	}
	infos, err := translateModelInfos(models)
	if err != nil {
		return pkgProvider.CapabilityReport{}, fmt.Errorf(
			"inspect llama.cpp capabilities: %w",
			err,
		)
	}

	routerAvailability := llamaCPPRouterAvailability(models)
	markLlamaCPPInstanceCapabilities(&report, routerAvailability)
	if request.Target == pkgProvider.CapabilityTargetInstance {
		return report, nil
	}

	modelIndex := -1
	for index := range infos {
		if infos[index].Model.ID == request.Model {
			modelIndex = index
			break
		}
	}
	if modelIndex < 0 {
		markMissingLlamaCPPModel(&report)

		return report, nil
	}

	model := models[modelIndex]
	state := infos[modelIndex].State
	if state == pkgProvider.ModelStateFailed ||
		state == pkgProvider.ModelStateDownloading ||
		state == pkgProvider.ModelStateLoading {
		markLlamaCPPInferenceAvailability(
			&report,
			pkgProvider.CapabilityAvailabilityUnavailable,
			pkgProvider.CapabilityAvailabilityUnavailable,
		)
	} else if embedding, known := llamaCPPEmbeddingMode(model.Status.Args); known {
		if embedding {
			markLlamaCPPInferenceAvailability(
				&report,
				pkgProvider.CapabilityAvailabilityUnavailable,
				pkgProvider.CapabilityAvailabilityAvailable,
			)
		} else {
			markLlamaCPPInferenceAvailability(
				&report,
				pkgProvider.CapabilityAvailabilityAvailable,
				pkgProvider.CapabilityAvailabilityUnavailable,
			)
		}
	} else if routerAvailability == pkgProvider.CapabilityAvailabilityUnavailable {
		// A single-model server that exposes its configured model can execute chat.
		// Embedding mode is not observable from /models and remains unknown.
		markLlamaCPPInferenceAvailability(
			&report,
			pkgProvider.CapabilityAvailabilityAvailable,
			pkgProvider.CapabilityAvailabilityUnknown,
		)
	}

	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityModelRemove,
	} {
		setLlamaCPPAvailability(&report, capability, routerAvailability)
	}

	return report, nil
}

func newLlamaCPPCapabilityReport(
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
		if capability == pkgProvider.CapabilityStructuredOutput ||
			capability == pkgProvider.CapabilityToolCalling {
			support = pkgProvider.CapabilityUnsupported
			availability = pkgProvider.CapabilityAvailabilityUnavailable
		}
		descriptors = append(descriptors, pkgProvider.CapabilityDescriptor{
			Capability: capability, Support: support, Availability: availability,
		})
	}

	return pkgProvider.CapabilityReport{
		Provider: providerID, Target: request.Target, Model: request.Model,
		Capabilities: descriptors,
	}
}

func llamaCPPRouterAvailability(
	models []modelData,
) pkgProvider.CapabilityAvailability {
	if len(models) == 0 {
		return pkgProvider.CapabilityAvailabilityUnknown
	}
	for _, model := range models {
		if model.Status.Value != "" {
			return pkgProvider.CapabilityAvailabilityAvailable
		}
	}

	return pkgProvider.CapabilityAvailabilityUnavailable
}

func markLlamaCPPInstanceCapabilities(
	report *pkgProvider.CapabilityReport,
	routerAvailability pkgProvider.CapabilityAvailability,
) {
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityModelListing,
		pkgProvider.CapabilityModelDiscovery,
	} {
		setLlamaCPPAvailability(
			report,
			capability,
			pkgProvider.CapabilityAvailabilityAvailable,
		)
	}
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityModelPull,
		pkgProvider.CapabilityModelRemove,
	} {
		setLlamaCPPAvailability(report, capability, routerAvailability)
	}
}

func markMissingLlamaCPPModel(report *pkgProvider.CapabilityReport) {
	for _, capability := range []pkgProvider.Capability{
		pkgProvider.CapabilityCompletion,
		pkgProvider.CapabilityStreaming,
		pkgProvider.CapabilityEmbedding,
		pkgProvider.CapabilityModelLoad,
		pkgProvider.CapabilityModelUnload,
		pkgProvider.CapabilityModelRemove,
	} {
		setLlamaCPPAvailability(
			report,
			capability,
			pkgProvider.CapabilityAvailabilityUnavailable,
		)
	}
}

func markLlamaCPPInferenceAvailability(
	report *pkgProvider.CapabilityReport,
	completion pkgProvider.CapabilityAvailability,
	embedding pkgProvider.CapabilityAvailability,
) {
	setLlamaCPPAvailability(report, pkgProvider.CapabilityCompletion, completion)
	setLlamaCPPAvailability(report, pkgProvider.CapabilityStreaming, completion)
	setLlamaCPPAvailability(report, pkgProvider.CapabilityEmbedding, embedding)
}

func llamaCPPEmbeddingMode(args []string) (bool, bool) {
	for _, argument := range args {
		switch argument {
		case "--embedding", "--embeddings":
			return true, true
		case "--no-embedding", "--no-embeddings":
			return false, true
		}
		if strings.HasPrefix(argument, "--embedding=") ||
			strings.HasPrefix(argument, "--embeddings=") {
			value := argument[strings.IndexByte(argument, '=')+1:]
			switch value {
			case "1", "true":
				return true, true
			case "0", "false":
				return false, true
			}
		}
	}

	return false, false
}

func setLlamaCPPAvailability(
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
