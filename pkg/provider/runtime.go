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
	PullModel(context.Context, ID, ModelPullRequest) (ModelPullStream, error)
	RemoveModel(context.Context, ID, ModelRemoveRequest) error
	Capabilities(context.Context, ID, CapabilityRequest) (CapabilityReport, error)
	SetResiliencePolicy(context.Context, ID, ResiliencePolicy) error
	ResiliencePolicy(ID, Operation, string) (ResiliencePolicy, bool, error)
	CircuitState(ID, Operation, string) (CircuitSnapshot, bool, error)
	SetObserver(ProviderObserver)

	SetModelResidencyPolicy(context.Context, ID, ModelResidencyPolicy) error
	ResidencyPolicy(ID, string) (ModelResidencyPolicy, bool, error)
	Shutdown(context.Context) error
}
