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
) (operationError error) {
	model, err := p.model(request.Model)
	if err != nil {
		return classifyLlamaCPPError(
			pkgProvider.OperationModelLoad,
			request.Model,
			err,
		)
	}
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationModelLoad,
			model,
			operationError,
		)
	}()

	return p.changeModelLifecycle(ctx, model, "/models/load", "load")
}

func (p *Provider) UnloadModel(
	ctx context.Context,
	request pkgProvider.ModelUnloadRequest,
) (operationError error) {
	model, err := p.model(request.Model)
	if err != nil {
		return classifyLlamaCPPError(
			pkgProvider.OperationModelUnload,
			request.Model,
			err,
		)
	}
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationModelUnload,
			model,
			operationError,
		)
	}()

	return p.changeModelLifecycle(
		ctx,
		model,
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

	if hasLlamaCPPAPIError(response.Error) {
		return newLlamaCPPAPIError(0, response.Error)
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
