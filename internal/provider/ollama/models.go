package ollama

import (
	"context"
	"fmt"
	"net/http"

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
		modelID := model.Model
		if modelID == "" {
			modelID = model.Name
		}

		modelName := model.Name
		if modelName == "" {
			modelName = model.Model
		}

		if modelID == "" {
			return nil, fmt.Errorf(
				"list Ollama models: model %d has no identity: %w",
				index,
				pkgProvider.ErrInvalidResponse,
			)
		}

		models = append(models, pkgProvider.Model{
			ID:   modelID,
			Name: modelName,
		})
	}

	return models, nil
}
