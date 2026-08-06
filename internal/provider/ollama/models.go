package ollama

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Models(
	ctx context.Context,
) ([]pkgProvider.Model, error) {
	response := tagsResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodGet,
		"/api/tags",
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("list Ollama models: %w", err)
	}

	if response.Error != "" {
		return nil, &apiError{message: response.Error}
	}

	models := make([]pkgProvider.Model, 0, len(response.Models))
	for index, model := range response.Models {
		translated, err := translateModel(model)
		if err != nil {
			return nil, fmt.Errorf("list Ollama models: model %d: %w", index, err)
		}

		models = append(models, translated)
	}

	return models, nil
}

func (p *Provider) DiscoverModels(
	ctx context.Context,
) ([]pkgProvider.ModelInfo, error) {
	available := tagsResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodGet,
		"/api/tags",
		nil,
		&available,
	); err != nil {
		return nil, fmt.Errorf("discover Ollama models: list available: %w", err)
	}
	if available.Error != "" {
		return nil, &apiError{message: available.Error}
	}

	running := tagsResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodGet,
		"/api/ps",
		nil,
		&running,
	); err != nil {
		return nil, fmt.Errorf("discover Ollama models: list running: %w", err)
	}
	if running.Error != "" {
		return nil, &apiError{message: running.Error}
	}

	infos := make([]pkgProvider.ModelInfo, 0, len(available.Models))
	indexes := make(map[string]int, len(available.Models))
	for index, model := range available.Models {
		info, err := translateModelInfo(model, pkgProvider.ModelStateAvailable)
		if err != nil {
			return nil, fmt.Errorf(
				"discover Ollama models: available model %d: %w",
				index,
				err,
			)
		}
		if _, exists := indexes[info.Model.ID]; exists {
			return nil, fmt.Errorf(
				"discover Ollama models: duplicate model %q: %w",
				info.Model.ID,
				pkgProvider.ErrInvalidResponse,
			)
		}

		indexes[info.Model.ID] = len(infos)
		infos = append(infos, info)
	}

	seenRunning := make(map[string]struct{}, len(running.Models))
	for index, model := range running.Models {
		info, err := translateModelInfo(model, pkgProvider.ModelStateLoaded)
		if err != nil {
			return nil, fmt.Errorf(
				"discover Ollama models: running model %d: %w",
				index,
				err,
			)
		}
		if _, exists := seenRunning[info.Model.ID]; exists {
			return nil, fmt.Errorf(
				"discover Ollama models: duplicate running model %q: %w",
				info.Model.ID,
				pkgProvider.ErrInvalidResponse,
			)
		}
		seenRunning[info.Model.ID] = struct{}{}

		if existingIndex, exists := indexes[info.Model.ID]; exists {
			infos[existingIndex] = mergeModelInfo(infos[existingIndex], info)
			continue
		}

		indexes[info.Model.ID] = len(infos)
		infos = append(infos, info)
	}

	return infos, nil
}

func translateModel(model modelResponse) (pkgProvider.Model, error) {
	modelID := model.Model
	if modelID == "" {
		modelID = model.Name
	}

	modelName := model.Name
	if modelName == "" {
		modelName = model.Model
	}

	if strings.TrimSpace(modelID) == "" || strings.TrimSpace(modelID) != modelID {
		return pkgProvider.Model{}, fmt.Errorf(
			"model has invalid identity: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return pkgProvider.Model{ID: modelID, Name: modelName}, nil
}

func translateModelInfo(
	model modelResponse,
	state pkgProvider.ModelState,
) (pkgProvider.ModelInfo, error) {
	translated, err := translateModel(model)
	if err != nil {
		return pkgProvider.ModelInfo{}, err
	}
	if model.Size < 0 || model.SizeVRAM < 0 || model.ContextLength < 0 {
		return pkgProvider.ModelInfo{}, fmt.Errorf(
			"model has negative resource metadata: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return pkgProvider.ModelInfo{
		Model:         translated,
		State:         state,
		Digest:        model.Digest,
		SizeBytes:     model.Size,
		VRAMBytes:     model.SizeVRAM,
		ContextLength: model.ContextLength,
		Format:        model.Details.Format,
		Family:        model.Details.Family,
		ParameterSize: model.Details.ParameterSize,
		Quantization:  model.Details.Quantization,
		ModifiedAt:    model.ModifiedAt,
		ExpiresAt:     model.ExpiresAt,
	}, nil
}

func mergeModelInfo(
	available pkgProvider.ModelInfo,
	running pkgProvider.ModelInfo,
) pkgProvider.ModelInfo {
	running.Model = available.Model
	if running.Digest == "" {
		running.Digest = available.Digest
	}
	if running.SizeBytes == 0 {
		running.SizeBytes = available.SizeBytes
	}
	if running.Format == "" {
		running.Format = available.Format
	}
	if running.Family == "" {
		running.Family = available.Family
	}
	if running.ParameterSize == "" {
		running.ParameterSize = available.ParameterSize
	}
	if running.Quantization == "" {
		running.Quantization = available.Quantization
	}
	if running.ModifiedAt.IsZero() {
		running.ModifiedAt = available.ModifiedAt
	}

	return running
}
