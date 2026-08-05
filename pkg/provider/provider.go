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

// Stream is a pull-based completion stream. Recv returns io.EOF when the
// stream has completed. Callers must close streams they acquire.
type Stream interface {
	Recv() (StreamChunk, error)
	Close() error
}
