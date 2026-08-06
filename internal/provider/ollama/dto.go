package ollama

import "time"

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

type errorResponse struct {
	Error string `json:"error"`
}
