package provider

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
