package llamacpp

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
