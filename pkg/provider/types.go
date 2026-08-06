package provider

import "time"

type ID string

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role    Role
	Content string
}

type CompletionRequest struct {
	Model    string
	Messages []Message
}

type CompletionResponse struct {
	Model        string
	Message      Message
	FinishReason string
	Usage        Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}

type StreamChunk struct {
	Model        string
	Content      string
	FinishReason string
	Usage        Usage
}

type EmbeddingRequest struct {
	Model  string
	Inputs []string
}

type EmbeddingResponse struct {
	Model      string
	Embeddings [][]float32
	Usage      Usage
}

type Model struct {
	ID          string
	Name        string
	Description string
}

type ModelState string

const (
	ModelStateUnknown     ModelState = "unknown"
	ModelStateAvailable   ModelState = "available"
	ModelStateDownloading ModelState = "downloading"
	ModelStateLoading     ModelState = "loading"
	ModelStateLoaded      ModelState = "loaded"
	ModelStateSleeping    ModelState = "sleeping"
	ModelStateFailed      ModelState = "failed"
)

// ModelInfo is a provider-neutral snapshot of a model and its observed state.
// Fields unavailable from a provider retain their zero value.
type ModelInfo struct {
	Model Model
	State ModelState

	Digest        string
	SizeBytes     int64
	VRAMBytes     int64
	ContextLength int

	Format        string
	Family        string
	ParameterSize string
	Quantization  string
	ModifiedAt    time.Time
	ExpiresAt     time.Time
}

type ModelLoadRequest struct {
	Model string
}

type ModelUnloadRequest struct {
	Model string
}
