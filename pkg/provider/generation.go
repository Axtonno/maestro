package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const (
	FinishReasonStop      = "stop"
	FinishReasonLength    = "length"
	FinishReasonToolCalls = "tool_calls"
)

type StructuredOutputMode string

const (
	StructuredOutputJSON       StructuredOutputMode = "json"
	StructuredOutputJSONSchema StructuredOutputMode = "json_schema"
)

// StructuredOutput requests syntactically valid JSON, optionally constrained
// by a JSON Schema object. Maestro validates syntax but not schema semantics.
type StructuredOutput struct {
	Mode   StructuredOutputMode
	Schema json.RawMessage
}

// GenerationOptions contains the provider-neutral sampling baseline. Nil
// floating-point values preserve provider defaults; MaxTokens zero is unset.
type GenerationOptions struct {
	MaxTokens   int
	Temperature *float64
	TopP        *float64
	Stop        []string
}

// Tool defines one function available to a completion. Parameters must be a
// JSON Schema object.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

type ToolChoiceMode string

const (
	ToolChoiceAuto     ToolChoiceMode = "auto"
	ToolChoiceNone     ToolChoiceMode = "none"
	ToolChoiceRequired ToolChoiceMode = "required"
	ToolChoiceNamed    ToolChoiceMode = "named"
)

// ToolChoice selects automatic, disabled, required, or one named function.
// The zero value is equivalent to auto when tools are present.
type ToolChoice struct {
	Mode ToolChoiceMode
	Name string
}

// ToolCall is a complete provider-neutral function invocation. Arguments must
// contain one JSON object. ID may be empty when a provider does not assign one.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolCallDelta carries incremental tool-call data. Arguments is a JSON text
// fragment and becomes a complete object after concatenating equal Index values.
type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}

func (r CompletionRequest) Validate() error {
	if r.Options.MaxTokens < 0 {
		return invalidCompletionRequest("max tokens cannot be negative")
	}
	if r.Options.Temperature != nil &&
		(math.IsNaN(*r.Options.Temperature) ||
			math.IsInf(*r.Options.Temperature, 0) ||
			*r.Options.Temperature < 0 || *r.Options.Temperature > 2) {
		return invalidCompletionRequest("temperature must be between 0 and 2")
	}
	if r.Options.TopP != nil &&
		(math.IsNaN(*r.Options.TopP) || math.IsInf(*r.Options.TopP, 0) ||
			*r.Options.TopP <= 0 || *r.Options.TopP > 1) {
		return invalidCompletionRequest("top_p must be greater than 0 and at most 1")
	}
	for _, stop := range r.Options.Stop {
		if stop == "" {
			return invalidCompletionRequest("stop sequences cannot be empty")
		}
	}

	if r.Output != nil {
		switch r.Output.Mode {
		case StructuredOutputJSON:
			if len(bytes.TrimSpace(r.Output.Schema)) != 0 {
				return invalidCompletionRequest("JSON mode cannot include a schema")
			}
		case StructuredOutputJSONSchema:
			if !validJSONObject(r.Output.Schema) {
				return invalidCompletionRequest("structured output schema must be a JSON object")
			}
		default:
			return invalidCompletionRequest("unknown structured output mode")
		}
	}

	toolNames := make(map[string]struct{}, len(r.Tools))
	for _, tool := range r.Tools {
		if !validToolName(tool.Name) {
			return invalidCompletionRequest("tool name is invalid")
		}
		if _, exists := toolNames[tool.Name]; exists {
			return invalidCompletionRequest("tool names must be unique")
		}
		toolNames[tool.Name] = struct{}{}
		if !validJSONObject(tool.Parameters) {
			return invalidCompletionRequest("tool parameters must be a JSON object")
		}
	}
	if r.Output != nil && len(r.Tools) > 0 {
		return invalidCompletionRequest("structured output and tools cannot be combined")
	}

	choiceMode := r.ToolChoice.Mode
	if choiceMode == "" {
		choiceMode = ToolChoiceAuto
	}
	switch choiceMode {
	case ToolChoiceAuto, ToolChoiceNone:
		if r.ToolChoice.Name != "" {
			return invalidCompletionRequest("auto or none tool choice cannot name a tool")
		}
	case ToolChoiceRequired:
		if len(r.Tools) == 0 || r.ToolChoice.Name != "" {
			return invalidCompletionRequest("required tool choice needs tools and no name")
		}
	case ToolChoiceNamed:
		if _, exists := toolNames[r.ToolChoice.Name]; !exists ||
			!validToolName(r.ToolChoice.Name) {
			return invalidCompletionRequest("named tool choice must reference a declared tool")
		}
	default:
		return invalidCompletionRequest("unknown tool choice mode")
	}
	if len(r.Tools) == 0 && choiceMode != ToolChoiceAuto &&
		choiceMode != ToolChoiceNone {
		return invalidCompletionRequest("tool choice requires declared tools")
	}

	for _, message := range r.Messages {
		if err := validateMessage(message); err != nil {
			return err
		}
	}

	return nil
}

func validateMessage(message Message) error {
	if message.Role != "" && message.Role != RoleSystem &&
		message.Role != RoleUser && message.Role != RoleAssistant &&
		message.Role != RoleTool {
		return invalidCompletionRequest("message role is invalid")
	}
	if message.Role == RoleTool {
		if message.ToolCallID == "" && !validToolName(message.ToolName) {
			return invalidCompletionRequest("tool result requires a call ID or tool name")
		}
		if message.ToolName != "" && !validToolName(message.ToolName) {
			return invalidCompletionRequest("tool result name is invalid")
		}
		if len(message.ToolCalls) != 0 {
			return invalidCompletionRequest("tool result cannot contain tool calls")
		}
		return nil
	}
	if message.ToolCallID != "" || message.ToolName != "" {
		return invalidCompletionRequest("only tool results can reference a tool call")
	}
	if len(message.ToolCalls) != 0 && message.Role != RoleAssistant {
		return invalidCompletionRequest("only assistant messages can contain tool calls")
	}
	for _, call := range message.ToolCalls {
		if !validToolName(call.Name) || !validJSONObject(call.Arguments) {
			return invalidCompletionRequest("assistant tool call is invalid")
		}
	}

	return nil
}

func validToolName(name string) bool {
	if len(name) == 0 || len(name) > 64 || strings.TrimSpace(name) != name {
		return false
	}
	for _, character := range name {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '-' {
			continue
		}
		return false
	}

	return true
}

func validJSONObject(value json.RawMessage) bool {
	var object map[string]json.RawMessage
	return len(bytes.TrimSpace(value)) > 0 &&
		json.Unmarshal(value, &object) == nil && object != nil
}

func invalidCompletionRequest(message string) error {
	return fmt.Errorf("validate completion request: %s: %w", message, ErrInvalidRequest)
}
