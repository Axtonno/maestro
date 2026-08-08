package ollama

import (
	"context"
	"fmt"
	"net/http"

	internalOllama "github.com/antonio-cafeo/maestro/internal/provider/ollama"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const ID pkgProvider.ID = "ollama"

var (
	_ pkgProvider.Provider            = (*Provider)(nil)
	_ pkgProvider.Completer           = (*Provider)(nil)
	_ pkgProvider.Streamer            = (*Provider)(nil)
	_ pkgProvider.Embedder            = (*Provider)(nil)
	_ pkgProvider.ModelLister         = (*Provider)(nil)
	_ pkgProvider.ModelDiscoverer     = (*Provider)(nil)
	_ pkgProvider.ModelLoader         = (*Provider)(nil)
	_ pkgProvider.ModelUnloader       = (*Provider)(nil)
	_ pkgProvider.ModelPuller         = (*Provider)(nil)
	_ pkgProvider.ModelRemover        = (*Provider)(nil)
	_ pkgProvider.CapabilityInspector = (*Provider)(nil)
)

type adapter interface {
	pkgProvider.Completer
	pkgProvider.Streamer
	pkgProvider.Embedder
	pkgProvider.ModelLister
	pkgProvider.ModelDiscoverer
	pkgProvider.ModelLoader
	pkgProvider.ModelUnloader
	pkgProvider.ModelPuller
	pkgProvider.ModelRemover
	pkgProvider.CapabilityInspector
}

// Provider is a public facade over Maestro's internal Ollama adapter.
type Provider struct {
	adapter adapter
}

func New(config Config) (*Provider, error) {
	if config.Timeout < 0 {
		return nil, fmt.Errorf(
			"create Ollama provider: timeout cannot be negative: %w",
			ErrInvalidConfig,
		)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	client := config.HTTPClient
	if client == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = DefaultTimeout
		}

		client = &http.Client{Timeout: timeout}
	}

	internalAdapter, err := internalOllama.New(
		baseURL,
		config.DefaultModel,
		client,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create Ollama provider: %w: %v",
			ErrInvalidConfig,
			err,
		)
	}

	return &Provider{adapter: internalAdapter}, nil
}

func (p *Provider) ID() pkgProvider.ID {
	return p.adapter.ID()
}

func (p *Provider) Complete(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	return p.adapter.Complete(ctx, request)
}

func (p *Provider) Stream(
	ctx context.Context,
	request pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	return p.adapter.Stream(ctx, request)
}

func (p *Provider) Embed(
	ctx context.Context,
	request pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	return p.adapter.Embed(ctx, request)
}

func (p *Provider) Models(
	ctx context.Context,
) ([]pkgProvider.Model, error) {
	return p.adapter.Models(ctx)
}

func (p *Provider) DiscoverModels(
	ctx context.Context,
) ([]pkgProvider.ModelInfo, error) {
	return p.adapter.DiscoverModels(ctx)
}

func (p *Provider) LoadModel(
	ctx context.Context,
	request pkgProvider.ModelLoadRequest,
) error {
	return p.adapter.LoadModel(ctx, request)
}

func (p *Provider) UnloadModel(
	ctx context.Context,
	request pkgProvider.ModelUnloadRequest,
) error {
	return p.adapter.UnloadModel(ctx, request)
}

func (p *Provider) PullModel(
	ctx context.Context,
	request pkgProvider.ModelPullRequest,
) (pkgProvider.ModelPullStream, error) {
	return p.adapter.PullModel(ctx, request)
}

func (p *Provider) RemoveModel(
	ctx context.Context,
	request pkgProvider.ModelRemoveRequest,
) error {
	return p.adapter.RemoveModel(ctx, request)
}

func (p *Provider) InspectCapabilities(
	ctx context.Context,
	request pkgProvider.CapabilityRequest,
) (pkgProvider.CapabilityReport, error) {
	return p.adapter.InspectCapabilities(ctx, request)
}
