package ollama

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
	operationModel := request.Model
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationModelLoad,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return err
	}
	operationModel = model

	return p.changeModelLifecycle(ctx, model, -1, "load")
}

func (p *Provider) UnloadModel(
	ctx context.Context,
	request pkgProvider.ModelUnloadRequest,
) (operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationModelUnload,
			operationModel,
			operationError,
		)
	}()

	model, err := p.model(request.Model)
	if err != nil {
		return err
	}
	operationModel = model

	return p.changeModelLifecycle(ctx, model, 0, "unload")
}

func (p *Provider) changeModelLifecycle(
	ctx context.Context,
	requestModel string,
	keepAlive int,
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
		"/api/generate",
		modelLifecycleRequest{
			Model:     model,
			Stream:    false,
			KeepAlive: keepAlive,
		},
		&response,
	); err != nil {
		return fmt.Errorf("%s Ollama model: %w", operation, err)
	}

	if response.Error != "" {
		return &apiError{message: response.Error}
	}
	if !response.Done {
		return fmt.Errorf(
			"%s Ollama model: response is not final: %w",
			operation,
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}
