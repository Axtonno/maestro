package llamacpp

import (
	"context"
	"fmt"
	"net/http"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func (p *Provider) Complete(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (responseValue pkgProvider.CompletionResponse, operationError error) {
	operationModel := request.Model
	defer func() {
		operationError = classifyLlamaCPPError(
			pkgProvider.OperationCompletion,
			operationModel,
			operationError,
		)
	}()
	if err := validateLlamaCPPCompletionRequest(request); err != nil {
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
		"/v1/chat/completions",
		newChatRequest(model, request, false),
		&response,
	); err != nil {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with llama.cpp: %w",
			err,
		)
	}

	if hasLlamaCPPAPIError(response.Error) {
		return pkgProvider.CompletionResponse{}, newLlamaCPPAPIError(0, response.Error)
	}

	if len(response.Choices) != 1 {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with llama.cpp: expected one choice, received %d: %w",
			len(response.Choices),
			pkgProvider.ErrInvalidResponse,
		)
	}

	choice := response.Choices[0]
	if choice.Index != 0 || choice.FinishReason == nil {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with llama.cpp: invalid final choice: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}
	toolCalls, err := translateLlamaCPPToolCalls(choice.Message.ToolCalls)
	if err != nil {
		return pkgProvider.CompletionResponse{}, err
	}
	if err := validateStructuredContent(
		request.Output,
		[]byte(choice.Message.Content),
	); err != nil {
		return pkgProvider.CompletionResponse{}, err
	}

	return pkgProvider.CompletionResponse{
		Model: response.Model,
		Message: pkgProvider.Message{
			Role:      pkgProvider.Role(choice.Message.Role),
			Content:   choice.Message.Content,
			ToolCalls: toolCalls,
		},
		FinishReason: *choice.FinishReason,
		Usage:        providerUsage(response.Usage),
	}, nil
}

func providerUsage(value *usage) pkgProvider.Usage {
	if value == nil {
		return pkgProvider.Usage{}
	}

	return pkgProvider.Usage{
		InputTokens:  value.PromptTokens,
		OutputTokens: value.CompletionTokens,
	}
}
