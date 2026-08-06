package provider

import "context"

// Provider identifies a provider implementation. Operational behavior is
// expressed by the optional capability interfaces below.
type Provider interface {
	ID() ID
}

type Completer interface {
	Provider

	Complete(context.Context, CompletionRequest) (CompletionResponse, error)
}

type Streamer interface {
	Provider

	Stream(context.Context, CompletionRequest) (Stream, error)
}

type Embedder interface {
	Provider

	Embed(context.Context, EmbeddingRequest) (EmbeddingResponse, error)
}

type ModelLister interface {
	Provider

	Models(context.Context) ([]Model, error)
}

type ModelDiscoverer interface {
	Provider

	DiscoverModels(context.Context) ([]ModelInfo, error)
}

type ModelLoader interface {
	Provider

	LoadModel(context.Context, ModelLoadRequest) error
}

type ModelUnloader interface {
	Provider

	UnloadModel(context.Context, ModelUnloadRequest) error
}

type ModelPuller interface {
	Provider

	PullModel(context.Context, ModelPullRequest) (ModelPullStream, error)
}

type ModelRemover interface {
	Provider

	RemoveModel(context.Context, ModelRemoveRequest) error
}

// Stream is a pull-based completion stream. Recv returns io.EOF when the
// stream has completed. Callers must close streams they acquire.
type Stream interface {
	Recv() (StreamChunk, error)
	Close() error
}

// ModelPullStream is a pull-based model acquisition stream. Recv returns the
// terminal completed progress before returning io.EOF on the next call.
// Callers must close streams they acquire.
type ModelPullStream interface {
	Recv() (ModelPullProgress, error)
	Close() error
}
