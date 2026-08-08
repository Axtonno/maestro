package llamacpp

import (
	"bytes"
	"encoding/json"
	"fmt"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func validateLlamaCPPCompletionRequest(
	request pkgProvider.CompletionRequest,
) error {
	if err := request.Validate(); err != nil {
		return err
	}
	for _, message := range request.Messages {
		if message.Role == pkgProvider.RoleTool && message.ToolCallID == "" {
			return fmt.Errorf(
				"llama.cpp tool results require a tool call ID: %w",
				pkgProvider.ErrInvalidRequest,
			)
		}
		for _, call := range message.ToolCalls {
			if call.ID == "" {
				return fmt.Errorf(
					"llama.cpp assistant tool calls require an ID: %w",
					pkgProvider.ErrInvalidRequest,
				)
			}
		}
	}

	return nil
}

func newChatRequest(
	model string,
	request pkgProvider.CompletionRequest,
	stream bool,
) chatRequest {
	messages := make([]chatMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		translated := chatMessage{
			Role: string(message.Role), Content: message.Content,
			ToolCallID: message.ToolCallID,
		}
		for index, call := range message.ToolCalls {
			translated.ToolCalls = append(translated.ToolCalls, toolCall{
				Index: index, ID: call.ID, Type: "function",
				Function: toolCallFunction{
					Name: call.Name, Arguments: string(call.Arguments),
				},
			})
		}
		messages = append(messages, translated)
	}

	result := chatRequest{
		Model: model, Messages: messages, Stream: stream, N: 1,
		MaxTokens:   request.Options.MaxTokens,
		Temperature: request.Options.Temperature,
		TopP:        request.Options.TopP,
		Stop:        request.Options.Stop,
	}
	if stream {
		result.StreamOptions = &streamOptions{IncludeUsage: true}
	}
	if request.Output != nil {
		result.ResponseFormat = &responseFormat{Type: "json_object"}
		if request.Output.Mode == pkgProvider.StructuredOutputJSONSchema {
			result.ResponseFormat.Schema = request.Output.Schema
		}
	}
	for _, tool := range request.Tools {
		result.Tools = append(result.Tools, chatTool{
			Type: "function",
			Function: toolFunction{
				Name: tool.Name, Description: tool.Description,
				Parameters: tool.Parameters,
			},
		})
	}
	if len(request.Tools) > 0 {
		switch request.ToolChoice.Mode {
		case "", pkgProvider.ToolChoiceAuto:
			result.ToolChoice = json.RawMessage(`"auto"`)
		case pkgProvider.ToolChoiceNone:
			result.ToolChoice = json.RawMessage(`"none"`)
		case pkgProvider.ToolChoiceRequired:
			result.ToolChoice = json.RawMessage(`"required"`)
		case pkgProvider.ToolChoiceNamed:
			value, _ := json.Marshal(map[string]any{
				"type":     "function",
				"function": map[string]string{"name": request.ToolChoice.Name},
			})
			result.ToolChoice = value
		}
	}

	return result
}

func translateLlamaCPPToolCalls(
	calls []toolCall,
) ([]pkgProvider.ToolCall, error) {
	translated := make([]pkgProvider.ToolCall, 0, len(calls))
	for _, call := range calls {
		arguments := json.RawMessage(call.Function.Arguments)
		if call.ID == "" || call.Function.Name == "" || !jsonObject(arguments) {
			return nil, fmt.Errorf(
				"invalid llama.cpp tool call: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}
		translated = append(translated, pkgProvider.ToolCall{
			ID: call.ID, Name: call.Function.Name,
			Arguments: append(json.RawMessage(nil), arguments...),
		})
	}

	return translated, nil
}

func llamaCPPToolCallDeltas(
	calls []toolCall,
) ([]pkgProvider.ToolCallDelta, error) {
	deltas := make([]pkgProvider.ToolCallDelta, 0, len(calls))
	for _, call := range calls {
		if call.Index < 0 {
			return nil, fmt.Errorf(
				"invalid llama.cpp tool call index: %w",
				pkgProvider.ErrInvalidResponse,
			)
		}
		deltas = append(deltas, pkgProvider.ToolCallDelta{
			Index: call.Index, ID: call.ID, Name: call.Function.Name,
			Arguments: call.Function.Arguments,
		})
	}

	return deltas, nil
}

func validateStructuredContent(
	output *pkgProvider.StructuredOutput,
	content []byte,
) error {
	if output == nil {
		return nil
	}
	if !json.Valid(bytes.TrimSpace(content)) {
		return fmt.Errorf(
			"llama.cpp structured output is not valid JSON: %w",
			pkgProvider.ErrInvalidResponse,
		)
	}

	return nil
}

func jsonObject(value json.RawMessage) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(value, &object) == nil && object != nil
}
