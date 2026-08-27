package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestLlamaCPPAdvancedGenerationMapsOptionsToolsAndResults(t *testing.T) {
	temperature := 0.25
	topP := 0.85
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if decoded.MaxTokens != 72 || decoded.Temperature == nil ||
			*decoded.Temperature != 0.25 || decoded.TopP == nil ||
			*decoded.TopP != 0.85 || len(decoded.Stop) != 1 ||
			decoded.Stop[0] != "END" {
			t.Fatalf("unexpected generation options: %#v", decoded)
		}
		if len(decoded.Tools) != 1 || decoded.Tools[0].Function.Name != "weather" ||
			string(decoded.ToolChoice) != `{"function":{"name":"weather"},"type":"function"}` {
			t.Fatalf("unexpected tools or choice: %#v %s", decoded.Tools, decoded.ToolChoice)
		}
		if len(decoded.Messages) != 2 || len(decoded.Messages[0].ToolCalls) != 1 ||
			decoded.Messages[0].ToolCalls[0].ID != "call-old" ||
			decoded.Messages[1].ToolCallID != "call-old" {
			t.Fatalf("unexpected tool history: %#v", decoded.Messages)
		}

		writeJSON(t, writer, chatResponse{
			Model: "local",
			Choices: []chatChoice{{
				Index: 0, FinishReason: stringPointer("tool_calls"),
				Message: chatMessage{Role: "assistant", ToolCalls: []toolCall{{
					ID: "call-new", Type: "function",
					Function: toolCallFunction{
						Name: "weather", Arguments: `{"city":"Rome"}`,
					},
				}}},
			}},
		})
	})

	response, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{
		Model: "local",
		Options: pkgProvider.GenerationOptions{
			MaxTokens: 72, Temperature: &temperature, TopP: &topP,
			Stop: []string{"END"},
		},
		Tools: []pkgProvider.Tool{{
			Name: "weather", Parameters: json.RawMessage(`{"type":"object"}`),
		}},
		ToolChoice: pkgProvider.ToolChoice{
			Mode: pkgProvider.ToolChoiceNamed, Name: "weather",
		},
		Messages: []pkgProvider.Message{
			{Role: pkgProvider.RoleAssistant, ToolCalls: []pkgProvider.ToolCall{{
				ID: "call-old", Name: "weather",
				Arguments: json.RawMessage(`{"city":"Milan"}`),
			}}},
			{Role: pkgProvider.RoleTool, ToolCallID: "call-old", ToolName: "weather", Content: "sunny"},
		},
	})
	if err != nil {
		t.Fatalf("complete with tools: %v", err)
	}
	if len(response.Message.ToolCalls) != 1 ||
		response.Message.ToolCalls[0].ID != "call-new" ||
		response.Message.ToolCalls[0].Name != "weather" ||
		string(response.Message.ToolCalls[0].Arguments) != `{"city":"Rome"}` {
		t.Fatalf("unexpected tool response: %#v", response.Message.ToolCalls)
	}
}

func TestLlamaCPPRejectsPerRequestContextAndThinkingBeforeIO(t *testing.T) {
	provider := newTestProvider(t, "local", "", func(http.ResponseWriter, *http.Request) {
		t.Fatal("unsupported generation controls performed remote I/O")
	})
	for _, options := range []pkgProvider.GenerationOptions{
		{ContextWindow: 4096},
		{Thinking: pkgProvider.ThinkingDefault},
		{Thinking: pkgProvider.ThinkingDisabled},
	} {
		if _, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{Options: options}); !errors.Is(err, pkgProvider.ErrUnsupportedCapability) {
			t.Fatalf("expected unsupported capability for %#v, got %v", options, err)
		}
	}
}

func TestLlamaCPPStructuredOutputMapsSchemaAndValidatesResponse(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","required":["name"]}`)
	provider := newTestProvider(t, "local", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		_ = json.NewDecoder(request.Body).Decode(&decoded)
		if decoded.ResponseFormat == nil || decoded.ResponseFormat.Type != "json_object" ||
			string(decoded.ResponseFormat.Schema) != string(schema) {
			t.Fatalf("unexpected response format: %#v", decoded.ResponseFormat)
		}
		writeJSON(t, writer, chatResponse{Choices: []chatChoice{{
			Index: 0, FinishReason: stringPointer("stop"),
			Message: chatMessage{Role: "assistant", Content: `{"name":"Maestro"}`},
		}}})
	})

	response, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{
		Output: &pkgProvider.StructuredOutput{
			Mode: pkgProvider.StructuredOutputJSONSchema, Schema: schema,
		},
	})
	if err != nil || response.Message.Content != `{"name":"Maestro"}` {
		t.Fatalf("structured completion: %#v, %v", response, err)
	}
}

func TestLlamaCPPStructuredOutputMapsJSONMode(t *testing.T) {
	request := newChatRequest("local", pkgProvider.CompletionRequest{
		Output: &pkgProvider.StructuredOutput{Mode: pkgProvider.StructuredOutputJSON},
	}, false)

	if request.ResponseFormat == nil ||
		request.ResponseFormat.Type != "json_object" ||
		len(request.ResponseFormat.Schema) != 0 {
		t.Fatalf("unexpected JSON response format: %#v", request.ResponseFormat)
	}
}

func TestLlamaCPPRejectsInvalidToolHistoryBeforeIO(t *testing.T) {
	provider := newTestProvider(t, "local", "", func(http.ResponseWriter, *http.Request) {
		t.Fatal("invalid tool history performed remote I/O")
	})
	_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{
		Messages: []pkgProvider.Message{{
			Role: pkgProvider.RoleTool, ToolName: "weather", Content: "sunny",
		}},
	})
	if !errors.Is(err, pkgProvider.ErrInvalidRequest) {
		t.Fatalf("expected invalid request, got %v", err)
	}
}

func TestLlamaCPPStreamMapsIncrementalToolCalls(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"data: {\"model\":\"local\",\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call-1\",\"type\":\"function\",\"function\":{\"name\":\"weather\",\"arguments\":\"{\\\"city\\\":\\\"\"}}]},\"finish_reason\":null}]}\n\n" +
			"data: {\"model\":\"local\",\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"Rome\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}\n\n" +
			"data: {\"model\":\"local\",\"choices\":[],\"usage\":{\"prompt_tokens\":9,\"completion_tokens\":4}}\n\n" +
			"data: [DONE]\n\n",
	)}
	provider := newProviderWithBody(t, body)
	stream, err := provider.Stream(context.Background(), pkgProvider.CompletionRequest{
		Tools: []pkgProvider.Tool{{Name: "weather", Parameters: json.RawMessage(`{}`)}},
	})
	if err != nil {
		t.Fatalf("open tool stream: %v", err)
	}
	first, err := stream.Recv()
	if err != nil || len(first.ToolCalls) != 1 ||
		first.ToolCalls[0].ID != "call-1" || first.ToolCalls[0].Name != "weather" {
		t.Fatalf("unexpected first tool delta: %#v, %v", first, err)
	}
	second, err := stream.Recv()
	if err != nil || len(second.ToolCalls) != 1 ||
		second.ToolCalls[0].Arguments != `Rome"}` ||
		second.FinishReason != "tool_calls" {
		t.Fatalf("unexpected final tool delta: %#v, %v", second, err)
	}
	usage, err := stream.Recv()
	if err != nil || usage.Usage.InputTokens != 9 || usage.Usage.OutputTokens != 4 {
		t.Fatalf("unexpected usage: %#v, %v", usage, err)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestLlamaCPPStreamRejectsInvalidStructuredTerminal(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"not-json\"},\"finish_reason\":\"stop\"}]}\n\n",
	)}
	provider := newProviderWithBody(t, body)
	stream, err := provider.Stream(context.Background(), pkgProvider.CompletionRequest{
		Output: &pkgProvider.StructuredOutput{Mode: pkgProvider.StructuredOutputJSON},
	})
	if err != nil {
		t.Fatalf("open structured stream: %v", err)
	}
	if _, err := stream.Recv(); !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected invalid response, got %v", err)
	}
}
