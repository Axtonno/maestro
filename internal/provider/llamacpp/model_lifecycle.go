package llamacpp

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) LoadModel(
	ctx context.Context,
	request pkgProvider.ModelLoadRequest,
) error {
	return p.changeModelLifecycle(ctx, request.Model, "/models/load", "load")
}

func (p *Provider) UnloadModel(
	ctx context.Context,
	request pkgProvider.ModelUnloadRequest,
) error {
	return p.changeModelLifecycle(
		ctx,
		request.Model,
		"/models/unload",
		"unload",
	)
}

func (p *Provider) changeModelLifecycle(
	ctx context.Context,
	requestModel string,
	path string,
	operation string,
) error {
	model, err := p.model(requestModel)
	if err != nil {
		return err
	}

	response := modelLifecycleResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		path,
		modelLifecycleRequest{Model: model},
		&response,
	); err != nil {
		return fmt.Errorf("%s llama.cpp model: %w", operation, err)
	}

	if message := errorMessage(response.Error); message != "" {
		return &apiError{message: message}
	}
	if !response.Success {
		return fmt.Errorf(
			"%s llama.cpp model: operation was not successful: %w",
			operation,
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}
