package provider

import "context"

const ConfigDefaultProvider = "providers.default"

// Runtime registers providers and routes operations to their capabilities.
// Passing an empty provider ID selects the explicitly configured default.
type Runtime interface {
	Register(Provider) error
	Resolve(ID) (Provider, error)

	SetDefault(ID) error
	Default() (Provider, error)

	Complete(
		context.Context,
		ID,
		CompletionRequest,
	) (CompletionResponse, error)

	Stream(
		context.Context,
		ID,
		CompletionRequest,
	) (Stream, error)

	Embed(
		context.Context,
		ID,
		EmbeddingRequest,
	) (EmbeddingResponse, error)

	Models(context.Context, ID) ([]Model, error)
	DiscoverModels(context.Context, ID) ([]ModelInfo, error)
	LoadModel(context.Context, ID, ModelLoadRequest) error
	UnloadModel(context.Context, ID, ModelUnloadRequest) error
}
