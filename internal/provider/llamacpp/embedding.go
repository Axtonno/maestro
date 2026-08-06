package llamacpp

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Embed(
	ctx context.Context,
	request pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	model, err := p.model(request.Model)
	if err != nil {
		return pkgProvider.EmbeddingResponse{}, err
	}

	if len(request.Inputs) == 0 {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with llama.cpp: input is empty: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}

	response := embeddingResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/v1/embeddings",
		embeddingRequest{
			Model:          model,
			Input:          request.Inputs,
			EncodingFormat: "float",
		},
		&response,
	); err != nil {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with llama.cpp: %w",
			err,
		)
	}

	if message := errorMessage(response.Error); message != "" {
		return pkgProvider.EmbeddingResponse{}, &apiError{message: message}
	}

	if len(response.Data) != len(request.Inputs) {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with llama.cpp: expected %d embeddings, received %d: %w",
			len(request.Inputs),
			len(response.Data),
			pkgProvider.ErrInvalidResponse,
		)
	}

	embeddings := make([][]float32, len(response.Data))
	seen := make([]bool, len(response.Data))
	for _, item := range response.Data {
		if item.Index < 0 || item.Index >= len(embeddings) || seen[item.Index] {
			return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
				"embed with llama.cpp: invalid embedding index %d: %w",
				item.Index,
				pkgProvider.ErrInvalidResponse,
			)
		}
		if len(item.Embedding) == 0 {
			return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
				"embed with llama.cpp: embedding %d is empty: %w",
				item.Index,
				pkgProvider.ErrInvalidResponse,
			)
		}

		seen[item.Index] = true
		embeddings[item.Index] = item.Embedding
	}

	return pkgProvider.EmbeddingResponse{
		Model:      response.Model,
		Embeddings: embeddings,
		Usage: pkgProvider.Usage{
			InputTokens: response.Usage.PromptTokens,
		},
	}, nil
}
