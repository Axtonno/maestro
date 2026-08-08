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
	Role       Role
	Content    string
	ToolCallID string
	ToolName   string
	ToolCalls  []ToolCall
}

type CompletionRequest struct {
	Model      string
	Messages   []Message
	Options    GenerationOptions
	Output     *StructuredOutput
	Tools      []Tool
	ToolChoice ToolChoice
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
	ToolCalls    []ToolCallDelta
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

// ModelResidencyPolicy configures runtime-managed autoload and release for one
// exact model ID. An absent policy leaves provider behavior unchanged.
type ModelResidencyPolicy struct {
	Model      string
	Autoload   bool
	KeepAlive  time.Duration
	Persistent bool
}

type ModelPullRequest struct {
	Model string
}

type ModelRemoveRequest struct {
	Model string
}

type ModelPullStage string

const (
	ModelPullStageUnknown     ModelPullStage = "unknown"
	ModelPullStageResolving   ModelPullStage = "resolving"
	ModelPullStageDownloading ModelPullStage = "downloading"
	ModelPullStageVerifying   ModelPullStage = "verifying"
	ModelPullStageFinalizing  ModelPullStage = "finalizing"
	ModelPullStageCompleted   ModelPullStage = "completed"
)

// ModelPullProgress is a provider-neutral snapshot of an acquisition. Detail
// is informational and must not be used to make control-flow decisions.
type ModelPullProgress struct {
	Model          string
	Stage          ModelPullStage
	Detail         string
	Digest         string
	TotalBytes     int64
	CompletedBytes int64
}
