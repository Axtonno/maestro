package llamacpp

import (
	"context"
	"fmt"
	"net/http"

	internalLlamaCPP "github.com/antonio-cafeo/maestro/internal/provider/llamacpp"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const ID pkgProvider.ID = "llama.cpp"

var (
	_ pkgProvider.Provider    = (*Provider)(nil)
	_ pkgProvider.Completer   = (*Provider)(nil)
	_ pkgProvider.Streamer    = (*Provider)(nil)
	_ pkgProvider.Embedder    = (*Provider)(nil)
	_ pkgProvider.ModelLister = (*Provider)(nil)
)

type adapter interface {
	pkgProvider.Completer
	pkgProvider.Streamer
	pkgProvider.Embedder
	pkgProvider.ModelLister
}

// Provider is a public facade over Maestro's internal llama.cpp adapter.
type Provider struct {
	adapter adapter
}

func New(config Config) (*Provider, error) {
	if config.Timeout < 0 {
		return nil, fmt.Errorf(
			"create llama.cpp provider: timeout cannot be negative: %w",
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

	internalAdapter, err := internalLlamaCPP.New(
		baseURL,
		config.DefaultModel,
		config.APIKey,
		client,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create llama.cpp provider: %w: %v",
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
