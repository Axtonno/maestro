package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func validateOllamaCompletionRequest(request pkgProvider.CompletionRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if request.ToolChoice.Mode == pkgProvider.ToolChoiceRequired ||
		request.ToolChoice.Mode == pkgProvider.ToolChoiceNamed {
		return fmt.Errorf(
			"Ollama native chat API supports only auto or none tool choice: %w",
			pkgProvider.ErrUnsupportedCapability,
		)
	}
	for _, message := range request.Messages {
		if message.Role == pkgProvider.RoleTool && message.ToolName == "" {
			return fmt.Errorf(
				"Ollama tool results require a tool name: %w",
				pkgProvider.ErrInvalidRequest,
			)
		}
	}

	return nil
}

func newChatRequest(
	model string,
	request pkgProvider.CompletionRequest,
	stream bool,
) chatRequest {
	translatedMessages := make([]chatMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		translated := chatMessage{
			Role:       string(message.Role),
			Content:    message.Content,
			ToolCallID: message.ToolCallID,
			ToolName:   message.ToolName,
		}
		for index, call := range message.ToolCalls {
			translated.ToolCalls = append(translated.ToolCalls, toolCall{
				ID: call.ID, Type: "function",
				Function: toolCallFunction{
					Index: index, Name: call.Name, Arguments: call.Arguments,
				},
			})
		}
		translatedMessages = append(translatedMessages, translated)
	}

	result := chatRequest{
		Model: model, Messages: translatedMessages, Stream: stream,
	}
	if request.KeepAlive > 0 {
		result.KeepAlive = request.KeepAlive.String()
	}
	switch request.Options.Thinking {
	case pkgProvider.ThinkingEnabled:
		enabled := true
		result.Think = &enabled
	case pkgProvider.ThinkingDisabled:
		enabled := false
		result.Think = &enabled
	}
	if request.Options.MaxTokens != 0 || request.Options.Temperature != nil ||
		request.Options.TopP != nil || len(request.Options.Stop) != 0 ||
		request.Options.ContextWindow != 0 {
		result.Options = &chatOptions{
			NumPredict:  request.Options.MaxTokens,
			NumCtx:      request.Options.ContextWindow,
			Temperature: request.Options.Temperature,
			TopP:        request.Options.TopP,
			Stop:        request.Options.Stop,
		}
	}
	if request.Output != nil {
		if request.Output.Mode == pkgProvider.StructuredOutputJSON {
			result.Format = json.RawMessage(`"json"`)
		} else {
			result.Format = request.Output.Schema
		}
	}
	if request.ToolChoice.Mode != pkgProvider.ToolChoiceNone {
		for _, tool := range request.Tools {
			result.Tools = append(result.Tools, chatTool{
				Type: "function",
				Function: toolFunction{
					Name: tool.Name, Description: tool.Description,
					Parameters: tool.Parameters,
				},
			})
		}
	}

	return result
}

func translateOllamaToolCalls(
	calls []toolCall,
) ([]pkgProvider.ToolCall, error) {
	translated := make([]pkgProvider.ToolCall, 0, len(calls))
	for _, call := range calls {
		if call.Function.Name == "" || !jsonObject(call.Function.Arguments) {
			return nil, fmt.Errorf(
				"invalid Ollama tool call: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}
		translated = append(translated, pkgProvider.ToolCall{
			ID: call.ID, Name: call.Function.Name,
			Arguments: append(json.RawMessage(nil), call.Function.Arguments...),
		})
	}

	return translated, nil
}

func ollamaToolCallDeltas(
	calls []toolCall,
) ([]pkgProvider.ToolCallDelta, error) {
	deltas := make([]pkgProvider.ToolCallDelta, 0, len(calls))
	for position, call := range calls {
		if call.Function.Name == "" || !jsonObject(call.Function.Arguments) {
			return nil, fmt.Errorf(
				"invalid Ollama streaming tool call: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}
		index := call.Function.Index
		if index == 0 && position > 0 {
			index = position
		}
		deltas = append(deltas, pkgProvider.ToolCallDelta{
			Index: index, ID: call.ID, Name: call.Function.Name,
			Arguments: string(call.Function.Arguments),
		})
	}

	return deltas, nil
}

func normalizeOllamaFinishReason(
	done bool,
	finishReason string,
	toolCallSeen bool,
) string {
	if done && finishReason == pkgProvider.FinishReasonStop && toolCallSeen {
		return pkgProvider.FinishReasonToolCalls
	}

	return finishReason
}

func validateStructuredContent(
	output *pkgProvider.StructuredOutput,
	content []byte,
) error {
	if output == nil {
		return nil
	}
	content = bytes.TrimSpace(content)
	if !json.Valid(content) {
		return fmt.Errorf(
			"Ollama structured output is not valid JSON: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}

func jsonObject(value json.RawMessage) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(value, &object) == nil && object != nil
}
