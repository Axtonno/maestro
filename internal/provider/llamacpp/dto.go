package llamacpp

import "encoding/json"

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Stream         bool            `json:"stream"`
	N              int             `json:"n"`
	StreamOptions  *streamOptions  `json:"stream_options,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    *float64        `json:"temperature,omitempty"`
	TopP           *float64        `json:"top_p,omitempty"`
	Stop           []string        `json:"stop,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Tools          []chatTool      `json:"tools,omitempty"`
	ToolChoice     json.RawMessage `json:"tool_choice,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
}

type responseFormat struct {
	Type   string          `json:"type"`
	Schema json.RawMessage `json:"schema,omitempty"`
}

type chatTool struct {
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

type toolCall struct {
	Index    int              `json:"index,omitempty"`
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function toolCallFunction `json:"function"`
}

type toolCallFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type chatResponse struct {
	Model   string          `json:"model"`
	Choices []chatChoice    `json:"choices"`
	Usage   *usage          `json:"usage"`
	Error   json.RawMessage `json:"error"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	Delta        chatMessage `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type embeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format"`
}

type embeddingResponse struct {
	Model string          `json:"model"`
	Data  []embeddingData `json:"data"`
	Usage usage           `json:"usage"`
	Error json.RawMessage `json:"error"`
}

type embeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type modelsResponse struct {
	Data  []modelData     `json:"data"`
	Error json.RawMessage `json:"error"`
}

type modelData struct {
	ID     string          `json:"id"`
	Path   string          `json:"path"`
	Status modelStatusData `json:"status"`
	Meta   modelMeta       `json:"meta"`
}

type modelStatusData struct {
	Value  string   `json:"value"`
	Failed bool     `json:"failed"`
	Args   []string `json:"args"`
}

type modelMeta struct {
	Size       int64 `json:"size"`
	ContextLen int   `json:"n_ctx_train"`
}

type modelLifecycleRequest struct {
	Model string `json:"model"`
}

type modelLifecycleResponse struct {
	Success bool            `json:"success"`
	Error   json.RawMessage `json:"error"`
}

type modelEvent struct {
	Model string          `json:"model"`
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type modelDownloadProgress struct {
	Done  int64 `json:"done"`
	Total int64 `json:"total"`
}

type errorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"`
}
