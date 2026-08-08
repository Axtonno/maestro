package ollama

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Embed(
	ctx context.Context,
	request pkgProvider.EmbeddingRequest,
) (result pkgProvider.EmbeddingResponse, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationEmbedding,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return pkgProvider.EmbeddingResponse{}, err
	}
	operationModel = model

	if len(request.Inputs) == 0 {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with Ollama: input is empty: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}

	response := embedResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/api/embed",
		embedRequest{Model: model, Input: request.Inputs},
		&response,
	); err != nil {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with Ollama: %w",
			err,
		)
	}

	if response.Error != "" {
		return pkgProvider.EmbeddingResponse{}, &apiError{
			message: response.Error,
		}
	}

	if len(response.Embeddings) != len(request.Inputs) {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with Ollama: expected %d embeddings, received %d: %w",
			len(request.Inputs),
			len(response.Embeddings),
			pkgProvider.ErrInvalidResponse,
		)
	}

	for index, embedding := range response.Embeddings {
		if len(embedding) == 0 {
			return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
				"embed with Ollama: embedding %d is empty: %w",
				index,
				pkgProvider.ErrInvalidResponse,
			)
		}
	}

	return pkgProvider.EmbeddingResponse{
		Model:      response.Model,
		Embeddings: response.Embeddings,
		Usage: pkgProvider.Usage{
			InputTokens: response.PromptEvalCount,
		},
	}, nil
}
