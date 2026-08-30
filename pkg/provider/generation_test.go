package provider

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestCompletionRequestValidationAcceptsBaselineAndAdvancedRequests(t *testing.T) {
	temperature := 0.2
	topP := 0.9
	requests := []CompletionRequest{
		{},
		{
			KeepAlive: 5 * time.Minute,
			Options: GenerationOptions{
				MaxTokens: 128, Temperature: &temperature, TopP: &topP,
				Stop: []string{"END"}, ContextWindow: 4096,
				Thinking: ThinkingDisabled,
			},
			Tools: []Tool{{
				Name: "weather", Parameters: json.RawMessage(`{"type":"object"}`),
			}},
			ToolChoice: ToolChoice{Mode: ToolChoiceNamed, Name: "weather"},
			Messages: []Message{
				{Role: RoleAssistant, ToolCalls: []ToolCall{{
					ID: "call-1", Name: "weather",
					Arguments: json.RawMessage(`{"city":"Rome"}`),
				}}},
				{Role: RoleTool, ToolCallID: "call-1", ToolName: "weather"},
			},
		},
		{
			Output: &StructuredOutput{
				Mode:   StructuredOutputJSONSchema,
				Schema: json.RawMessage(`{"type":"object"}`),
			},
		},
	}
	for index, request := range requests {
		if err := request.Validate(); err != nil {
			t.Fatalf("request %d should be valid: %v", index, err)
		}
	}
}

func TestCompletionRequestValidationRejectsInvalidAdvancedContracts(t *testing.T) {
	nan := math.NaN()
	tests := []CompletionRequest{
		{KeepAlive: -time.Second},
		{Options: GenerationOptions{MaxTokens: -1}},
		{Options: GenerationOptions{ContextWindow: -1}},
		{Options: GenerationOptions{ContextWindow: 1<<20 + 1}},
		{Options: GenerationOptions{Thinking: "sometimes"}},
		{Options: GenerationOptions{Temperature: &nan}},
		{Options: GenerationOptions{Stop: []string{""}}},
		{Output: &StructuredOutput{Mode: StructuredOutputJSONSchema, Schema: json.RawMessage(`[]`)}},
		{Output: &StructuredOutput{Mode: StructuredOutputJSON}, Tools: []Tool{{Name: "x", Parameters: json.RawMessage(`{}`)}}},
		{Tools: []Tool{{Name: "bad name", Parameters: json.RawMessage(`{}`)}}},
		{Tools: []Tool{{Name: "x", Parameters: json.RawMessage(`[]`)}}},
		{ToolChoice: ToolChoice{Mode: ToolChoiceRequired}},
		{Tools: []Tool{{Name: "x", Parameters: json.RawMessage(`{}`)}}, ToolChoice: ToolChoice{Mode: ToolChoiceNamed, Name: "y"}},
		{Messages: []Message{{Role: RoleUser, ToolCallID: "call"}}},
		{Messages: []Message{{Role: RoleTool}}},
		{Messages: []Message{{Role: RoleAssistant, ToolCalls: []ToolCall{{Name: "x", Arguments: json.RawMessage(`[]`)}}}}},
	}
	for index, request := range tests {
		if err := request.Validate(); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("request %d should be invalid, got %v", index, err)
		}
	}
}
