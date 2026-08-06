package llamacpp

import "encoding/json"

type chatRequest struct {
	Model         string         `json:"model"`
	Messages      []chatMessage  `json:"messages"`
	Stream        bool           `json:"stream"`
	N             int            `json:"n"`
	StreamOptions *streamOptions `json:"stream_options,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	Value  string `json:"value"`
	Failed bool   `json:"failed"`
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
