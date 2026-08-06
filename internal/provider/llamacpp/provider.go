package llamacpp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const providerID pkgProvider.ID = "llama.cpp"

var (
	_ pkgProvider.Provider        = (*Provider)(nil)
	_ pkgProvider.Completer       = (*Provider)(nil)
	_ pkgProvider.Streamer        = (*Provider)(nil)
	_ pkgProvider.Embedder        = (*Provider)(nil)
	_ pkgProvider.ModelLister     = (*Provider)(nil)
	_ pkgProvider.ModelDiscoverer = (*Provider)(nil)
	_ pkgProvider.ModelLoader     = (*Provider)(nil)
	_ pkgProvider.ModelUnloader   = (*Provider)(nil)
)

type Provider struct {
	baseURL      string
	defaultModel string
	apiKey       string
	client       *http.Client
}

func New(
	baseURL string,
	defaultModel string,
	apiKey string,
	client *http.Client,
) (*Provider, error) {
	normalizedBaseURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, fmt.Errorf("HTTP client is nil")
	}

	if strings.TrimSpace(defaultModel) != defaultModel {
		return nil, fmt.Errorf("default model contains surrounding whitespace")
	}

	if strings.TrimSpace(apiKey) != apiKey {
		return nil, fmt.Errorf("API key contains surrounding whitespace")
	}

	return &Provider{
		baseURL:      normalizedBaseURL,
		defaultModel: defaultModel,
		apiKey:       apiKey,
		client:       client,
	}, nil
}

func (p *Provider) ID() pkgProvider.ID {
	return providerID
}

func (p *Provider) model(requestModel string) (string, error) {
	model := requestModel
	if model == "" {
		model = p.defaultModel
	}

	if strings.TrimSpace(model) == "" || strings.TrimSpace(model) != model {
		return "", fmt.Errorf(
			"resolve llama.cpp model: model is missing or invalid: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}

	return model, nil
}

func (p *Provider) endpoint(path string) string {
	return p.baseURL + path
}

func normalizeBaseURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("base URL must use http or https")
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("base URL host is missing")
	}

	if parsed.User != nil {
		return "", fmt.Errorf("base URL must not contain credentials")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("base URL must not contain a path")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base URL must not contain a query or fragment")
	}

	parsed.Path = ""
	parsed.RawPath = ""

	return strings.TrimSuffix(parsed.String(), "/"), nil
}
