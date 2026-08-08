package ollama

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

func TestOllamaAdvancedGenerationMapsOptionsToolsAndResults(t *testing.T) {
	temperature := 0.1
	topP := 0.8
	provider := newTestProvider(t, "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if decoded.Options == nil || decoded.Options.NumPredict != 96 ||
			decoded.Options.Temperature == nil || *decoded.Options.Temperature != 0.1 ||
			decoded.Options.TopP == nil || *decoded.Options.TopP != 0.8 ||
			len(decoded.Options.Stop) != 1 || decoded.Options.Stop[0] != "END" {
			t.Fatalf("unexpected options: %#v", decoded.Options)
		}
		if len(decoded.Tools) != 1 || decoded.Tools[0].Function.Name != "weather" ||
			string(decoded.Tools[0].Function.Parameters) != `{"type":"object"}` {
			t.Fatalf("unexpected tools: %#v", decoded.Tools)
		}
		if len(decoded.Messages) != 2 || len(decoded.Messages[0].ToolCalls) != 1 ||
			decoded.Messages[0].ToolCalls[0].ID != "call-old" ||
			decoded.Messages[1].ToolName != "weather" ||
			decoded.Messages[1].ToolCallID != "call-old" {
			t.Fatalf("unexpected tool history: %#v", decoded.Messages)
		}

		writeJSON(t, writer, chatResponse{
			Model: "qwen", Done: true, DoneReason: "tool_calls",
			Message: chatMessage{Role: "assistant", ToolCalls: []toolCall{{
				ID: "call-new",
				Function: toolCallFunction{
					Name: "weather", Arguments: json.RawMessage(`{"city":"Rome"}`),
				},
			}}},
		})
	})

	response, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{
		Model: "qwen",
		Options: pkgProvider.GenerationOptions{
			MaxTokens: 96, Temperature: &temperature, TopP: &topP,
			Stop: []string{"END"},
		},
		Tools: []pkgProvider.Tool{{
			Name: "weather", Description: "Current weather",
			Parameters: json.RawMessage(`{"type":"object"}`),
		}},
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
		t.Fatalf("unexpected translated tool call: %#v", response.Message.ToolCalls)
	}
}

func TestOllamaStructuredOutputMapsSchemaAndValidatesResponse(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","required":["name"]}`)
	provider := newTestProvider(t, "qwen", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		decoded := chatRequest{}
		_ = json.NewDecoder(request.Body).Decode(&decoded)
		if string(decoded.Format) != string(schema) {
			t.Fatalf("unexpected format: %s", decoded.Format)
		}
		writeJSON(t, writer, chatResponse{
			Done:    true,
			Message: chatMessage{Role: "assistant", Content: `{"name":"Maestro"}`},
		})
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

func TestOllamaStructuredOutputMapsJSONMode(t *testing.T) {
	request := newChatRequest("qwen", pkgProvider.CompletionRequest{
		Output: &pkgProvider.StructuredOutput{Mode: pkgProvider.StructuredOutputJSON},
	}, false)

	if string(request.Format) != `"json"` {
		t.Fatalf("unexpected JSON format: %s", request.Format)
	}
}

func TestOllamaRejectsUnsupportedToolChoiceBeforeIO(t *testing.T) {
	provider := newTestProvider(t, "qwen", func(http.ResponseWriter, *http.Request) {
		t.Fatal("unsupported tool choice performed remote I/O")
	})
	_, err := provider.Complete(context.Background(), pkgProvider.CompletionRequest{
		Tools: []pkgProvider.Tool{{
			Name: "weather", Parameters: json.RawMessage(`{}`),
		}},
		ToolChoice: pkgProvider.ToolChoice{
			Mode: pkgProvider.ToolChoiceNamed, Name: "weather",
		},
	})
	if !errors.Is(err, pkgProvider.ErrUnsupportedCapability) {
		t.Fatalf("expected unsupported capability, got %v", err)
	}
}

func TestOllamaStreamMapsToolCallsAndValidatesStructuredOutput(t *testing.T) {
	t.Run("tool call", func(t *testing.T) {
		body := &trackingBody{reader: strings.NewReader(
			`{"model":"qwen","message":{"role":"assistant","tool_calls":[{"id":"call-1","function":{"index":0,"name":"weather","arguments":{"city":"Rome"}}}]},"done":true,"done_reason":"tool_calls"}` + "\n",
		)}
		provider := newProviderWithBody(t, body)
		stream, err := provider.Stream(context.Background(), pkgProvider.CompletionRequest{
			Tools: []pkgProvider.Tool{{Name: "weather", Parameters: json.RawMessage(`{}`)}},
		})
		if err != nil {
			t.Fatalf("open tool stream: %v", err)
		}
		chunk, err := stream.Recv()
		if err != nil || len(chunk.ToolCalls) != 1 ||
			chunk.ToolCalls[0].ID != "call-1" ||
			chunk.ToolCalls[0].Arguments != `{"city":"Rome"}` ||
			chunk.FinishReason != "tool_calls" {
			t.Fatalf("unexpected tool chunk: %#v, %v", chunk, err)
		}
		if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
			t.Fatalf("expected EOF, got %v", err)
		}
	})

	t.Run("invalid structured terminal", func(t *testing.T) {
		body := &trackingBody{reader: strings.NewReader(
			`{"message":{"role":"assistant","content":"not-json"},"done":true,"done_reason":"stop"}` + "\n",
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
	})
}
