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
) (pkgProvider.CompletionResponse, error) {
	model, err := p.model(request.Model)
	if err != nil {
		return pkgProvider.CompletionResponse{}, err
	}

	response := chatResponse{}
	if err := p.doJSON(
		ctx,
		http.MethodPost,
		"/api/chat",
		newChatRequest(model, request.Messages, false),
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

	return pkgProvider.CompletionResponse{
		Model: response.Model,
		Message: pkgProvider.Message{
			Role:    pkgProvider.Role(response.Message.Role),
			Content: response.Message.Content,
		},
		FinishReason: response.DoneReason,
		Usage: pkgProvider.Usage{
			InputTokens:  response.PromptEvalCount,
			OutputTokens: response.EvalCount,
		},
	}, nil
}

func newChatRequest(
	model string,
	messages []pkgProvider.Message,
	stream bool,
) chatRequest {
	translated := make([]chatMessage, 0, len(messages))
	for _, message := range messages {
		translated = append(translated, chatMessage{
			Role:    string(message.Role),
			Content: message.Content,
		})
	}

	return chatRequest{
		Model:    model,
		Messages: translated,
		Stream:   stream,
	}
}
