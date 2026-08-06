package llamacpp

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Models(
	ctx context.Context,
) ([]pkgProvider.Model, error) {
	response := modelsResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodGet,
		"/v1/models",
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("list llama.cpp models: %w", err)
	}

	if message := errorMessage(response.Error); message != "" {
		return nil, &apiError{message: message}
	}

	models := make([]pkgProvider.Model, 0, len(response.Data))
	for index, model := range response.Data {
		if strings.TrimSpace(model.ID) == "" ||
			strings.TrimSpace(model.ID) != model.ID {
			return nil, fmt.Errorf(
				"list llama.cpp models: model %d has invalid identity: %w",
				index,
				pkgProvider.ErrInvalidResponse,
			)
		}

		models = append(models, pkgProvider.Model{
			ID:   model.ID,
			Name: model.ID,
		})
	}

	return models, nil
}

func (p *Provider) DiscoverModels(
	ctx context.Context,
) ([]pkgProvider.ModelInfo, error) {
	response := modelsResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodGet,
		"/models",
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("discover llama.cpp models: %w", err)
	}

	if message := errorMessage(response.Error); message != "" {
		return nil, &apiError{message: message}
	}

	infos := make([]pkgProvider.ModelInfo, 0, len(response.Data))
	seen := make(map[string]struct{}, len(response.Data))
	for index, model := range response.Data {
		if strings.TrimSpace(model.ID) == "" ||
			strings.TrimSpace(model.ID) != model.ID {
			return nil, fmt.Errorf(
				"discover llama.cpp models: model %d has invalid identity: %w",
				index,
				pkgProvider.ErrInvalidResponse,
			)
		}
		if _, exists := seen[model.ID]; exists {
			return nil, fmt.Errorf(
				"discover llama.cpp models: duplicate model %q: %w",
				model.ID,
				pkgProvider.ErrInvalidResponse,
			)
		}
		seen[model.ID] = struct{}{}
		if model.Meta.Size < 0 || model.Meta.ContextLen < 0 {
			return nil, fmt.Errorf(
				"discover llama.cpp models: model %q has negative resource metadata: %w",
				model.ID,
				pkgProvider.ErrInvalidResponse,
			)
		}

		infos = append(infos, pkgProvider.ModelInfo{
			Model: pkgProvider.Model{
				ID:   model.ID,
				Name: model.ID,
			},
			State:         translateModelState(model.Status),
			SizeBytes:     model.Meta.Size,
			ContextLength: model.Meta.ContextLen,
			Format:        modelFormat(model.Path),
		})
	}

	return infos, nil
}

func translateModelState(status modelStatusData) pkgProvider.ModelState {
	if status.Failed {
		return pkgProvider.ModelStateFailed
	}

	switch status.Value {
	case "unloaded":
		return pkgProvider.ModelStateAvailable
	case "downloading":
		return pkgProvider.ModelStateDownloading
	case "loading":
		return pkgProvider.ModelStateLoading
	case "loaded":
		return pkgProvider.ModelStateLoaded
	case "sleeping":
		return pkgProvider.ModelStateSleeping
	default:
		return pkgProvider.ModelStateUnknown
	}
}

func modelFormat(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".gguf") {
		return "gguf"
	}

	return ""
}
