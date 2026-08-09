package ollama

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Complete(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (result pkgProvider.CompletionResponse, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyOllamaError(
			pkgProvider.OperationCompletion,
			operationModel,
			operationError,
		)
	}()
	if err := validateOllamaCompletionRequest(request); err != nil {
		return pkgProvider.CompletionResponse{}, err
	}

	model, err := p.model(request.Model)
	if err != nil {
		return pkgProvider.CompletionResponse{}, err
	}
	operationModel = model

	response := chatResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/api/chat",
		newChatRequest(model, request, false),
		&response,
	); err != nil {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with Ollama: %w",
			err,
		)
	}

	if response.Error != "" {
		return pkgProvider.CompletionResponse{}, &apiError{
			message: response.Error,
		}
	}

	if !response.Done {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with Ollama: response is not final: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}
	toolCalls, err := translateOllamaToolCalls(response.Message.ToolCalls)
	if err != nil {
		return pkgProvider.CompletionResponse{}, err
	}
	if err := validateStructuredContent(
		request.Output,
		[]byte(response.Message.Content),
	); err != nil {
		return pkgProvider.CompletionResponse{}, err
	}

	return pkgProvider.CompletionResponse{
		Model: response.Model,
		Message: pkgProvider.Message{
			Role:      pkgProvider.Role(response.Message.Role),
			Content:   response.Message.Content,
			ToolCalls: toolCalls,
		},
		FinishReason: normalizeOllamaFinishReason(
			response.Done,
			response.DoneReason,
			len(toolCalls) > 0,
		),
		Usage: pkgProvider.Usage{
			InputTokens:  response.PromptEvalCount,
			OutputTokens: response.EvalCount,
		},
	}, nil
}
