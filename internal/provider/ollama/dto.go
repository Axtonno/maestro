package ollama

import (
	"encoding/json"
	"time"
)

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []chatMessage   `json:"messages"`
	Stream   bool            `json:"stream"`
	Think    *bool           `json:"think,omitempty"`
	Format   json.RawMessage `json:"format,omitempty"`
	Options  *chatOptions    `json:"options,omitempty"`
	Tools    []chatTool      `json:"tools,omitempty"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Thinking   string     `json:"thinking,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolName   string     `json:"tool_name,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
}

type chatOptions struct {
	NumPredict  int      `json:"num_predict,omitempty"`
	NumCtx      int      `json:"num_ctx,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
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
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function toolCallFunction `json:"function"`
}

type toolCallFunction struct {
	Index     int             `json:"index"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type chatResponse struct {
	Model           string      `json:"model"`
	Message         chatMessage `json:"message"`
	Done            bool        `json:"done"`
	DoneReason      string      `json:"done_reason"`
	PromptEvalCount int         `json:"prompt_eval_count"`
	EvalCount       int         `json:"eval_count"`
	Error           string      `json:"error"`
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	PromptEvalCount int         `json:"prompt_eval_count"`
	Error           string      `json:"error"`
}

type tagsResponse struct {
	Models []modelResponse `json:"models"`
	Error  string          `json:"error"`
}

type modelResponse struct {
	Name          string       `json:"name"`
	Model         string       `json:"model"`
	ModifiedAt    time.Time    `json:"modified_at"`
	Size          int64        `json:"size"`
	Digest        string       `json:"digest"`
	Details       modelDetails `json:"details"`
	ExpiresAt     time.Time    `json:"expires_at"`
	SizeVRAM      int64        `json:"size_vram"`
	ContextLength int          `json:"context_length"`
}

type modelDetails struct {
	Format        string `json:"format"`
	Family        string `json:"family"`
	ParameterSize string `json:"parameter_size"`
	Quantization  string `json:"quantization_level"`
}

type modelLifecycleRequest struct {
	Model     string `json:"model"`
	Stream    bool   `json:"stream"`
	KeepAlive int    `json:"keep_alive"`
}

type modelLifecycleResponse struct {
	Model string `json:"model"`
	Done  bool   `json:"done"`
	Error string `json:"error"`
}

type modelPullRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

type modelPullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Error     string `json:"error"`
}

type modelRemoveRequest struct {
	Model string `json:"model"`
}

type modelShowRequest struct {
	Model   string `json:"model"`
	Verbose bool   `json:"verbose"`
}

type modelShowResponse struct {
	Capabilities []string `json:"capabilities"`
	Error        string   `json:"error"`
}

type errorResponse struct {
	Error string `json:"error"`
}
