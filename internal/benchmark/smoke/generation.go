package smoke

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

var smokeJSONSchema = json.RawMessage(`{
  "type":"object",
  "properties":{"status":{"type":"string","const":"ok"}},
  "required":["status"],
  "additionalProperties":false
}`)

var smokeTool = pkgProvider.Tool{
	Name:        "echo_message",
	Description: "Return the supplied message",
	Parameters: json.RawMessage(`{
  "type":"object",
  "properties":{"message":{"type":"string"}},
  "required":["message"],
  "additionalProperties":false
}`),
}

func (s *Suite) structuredJSON(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return s.structuredScenario(definition, &pkgProvider.StructuredOutput{
		Mode: pkgProvider.StructuredOutputJSON,
	})
}

func (s *Suite) structuredJSONSchema(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return s.structuredScenario(definition, &pkgProvider.StructuredOutput{
		Mode:   pkgProvider.StructuredOutputJSONSchema,
		Schema: append(json.RawMessage(nil), smokeJSONSchema...),
	})
}

func (s *Suite) structuredScenario(
	definition pkgBenchmark.ScenarioDefinition,
	output *pkgProvider.StructuredOutput,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityStructuredOutput,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			response, err := s.runtime.Complete(
				ctx,
				s.config.ProviderID,
				pkgProvider.CompletionRequest{
					Model: model,
					Messages: []pkgProvider.Message{{
						Role:    pkgProvider.RoleUser,
						Content: "Return a JSON object with the key status and the value ok.",
					}},
					Options: pkgProvider.GenerationOptions{MaxTokens: 64},
					Output:  output,
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			var object map[string]json.RawMessage
			if err := json.Unmarshal([]byte(response.Message.Content), &object); err != nil ||
				object == nil {
				return failed("structured_output_not_object"), nil
			}
			if output.Mode == pkgProvider.StructuredOutputJSONSchema {
				var status string
				if err := json.Unmarshal(object["status"], &status); err != nil ||
					status != "ok" {
					return failed("structured_output_schema_mismatch"), nil
				}
			}

			return passed(countMeasurement("json_field_count", len(object))), nil
		},
	}
}

func (s *Suite) toolCallResult(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityToolCalling,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			userMessage := pkgProvider.Message{
				Role:    pkgProvider.RoleUser,
				Content: "Call echo_message exactly once with message set to Maestro smoke.",
			}
			response, err := s.runtime.Complete(
				ctx,
				s.config.ProviderID,
				pkgProvider.CompletionRequest{
					Model: model, Messages: []pkgProvider.Message{userMessage},
					Options: pkgProvider.GenerationOptions{MaxTokens: 128},
					Tools:   []pkgProvider.Tool{smokeTool},
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if len(response.Message.ToolCalls) == 0 {
				return failed("tool_call_missing"), nil
			}
			assistant := response.Message
			if assistant.Role == "" {
				assistant.Role = pkgProvider.RoleAssistant
			}
			messages := []pkgProvider.Message{userMessage, assistant}
			for _, call := range response.Message.ToolCalls {
				if call.Name != smokeTool.Name || !jsonObject(call.Arguments) {
					return failed("tool_call_invalid"), nil
				}
				messages = append(messages, pkgProvider.Message{
					Role: pkgProvider.RoleTool, ToolCallID: call.ID,
					ToolName: call.Name, Content: "smoke-ok",
				})
			}
			final, err := s.runtime.Complete(
				ctx,
				s.config.ProviderID,
				pkgProvider.CompletionRequest{
					Model: model, Messages: messages,
					Options:    pkgProvider.GenerationOptions{MaxTokens: 64},
					Tools:      []pkgProvider.Tool{smokeTool},
					ToolChoice: pkgProvider.ToolChoice{Mode: pkgProvider.ToolChoiceNone},
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if strings.TrimSpace(final.Message.Content) == "" {
				return failed("tool_result_response_empty"), nil
			}

			return passed(countMeasurement(
				"tool_call_count",
				len(response.Message.ToolCalls),
			)), nil
		},
	}
}

type streamedToolCall struct {
	name      string
	arguments strings.Builder
}

func (s *Suite) toolCallStream(
	definition pkgBenchmark.ScenarioDefinition,
) pkgBenchmark.Scenario {
	var active pkgProvider.Stream

	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(
			ctx context.Context,
			_ pkgBenchmark.Iteration,
		) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(
				ctx,
				"chat",
				pkgProvider.CapabilityToolCalling,
				pkgProvider.CapabilityStreaming,
			)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			active, err = s.runtime.Stream(
				ctx,
				s.config.ProviderID,
				pkgProvider.CompletionRequest{
					Model: model,
					Messages: []pkgProvider.Message{{
						Role:    pkgProvider.RoleUser,
						Content: "Call echo_message with message set to Maestro smoke.",
					}},
					Options: pkgProvider.GenerationOptions{MaxTokens: 128},
					Tools:   []pkgProvider.Tool{smokeTool},
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			calls := make(map[int]*streamedToolCall)
			terminalSeen := false
			for {
				chunk, receiveError := active.Recv()
				if errors.Is(receiveError, io.EOF) {
					break
				}
				if receiveError != nil {
					return pkgBenchmark.IterationResult{}, receiveError
				}
				if chunk.FinishReason == pkgProvider.FinishReasonToolCalls {
					terminalSeen = true
				}
				for _, delta := range chunk.ToolCalls {
					if delta.Index < 0 {
						return failed("tool_stream_index_invalid"), nil
					}
					call := calls[delta.Index]
					if call == nil {
						call = &streamedToolCall{}
						calls[delta.Index] = call
					}
					if delta.Name != "" {
						call.name = delta.Name
					}
					call.arguments.WriteString(delta.Arguments)
				}
			}
			if !terminalSeen || len(calls) == 0 {
				return failed("tool_stream_terminal_missing"), nil
			}
			for _, call := range calls {
				if call.name != smokeTool.Name ||
					!jsonObject(json.RawMessage(call.arguments.String())) {
					return failed("tool_stream_call_invalid"), nil
				}
			}

			return passed(countMeasurement("tool_call_count", len(calls))), nil
		},
		CleanupFunc: func(context.Context, pkgBenchmark.Iteration) error {
			if active == nil {
				return nil
			}
			err := active.Close()
			active = nil

			return err
		},
	}
}

func jsonObject(value json.RawMessage) bool {
	var object map[string]json.RawMessage

	return json.Unmarshal(value, &object) == nil && object != nil
}
