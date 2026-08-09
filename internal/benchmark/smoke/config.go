package smoke

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	maestro "github.com/antonio-cafeo/maestro"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	"github.com/antonio-cafeo/maestro/pkg/provider/llamacpp"
	"github.com/antonio-cafeo/maestro/pkg/provider/ollama"
)

const (
	ProviderOllama   = "ollama"
	ProviderLlamaCPP = "llama.cpp"
)

type Environment func(string) string

type Config struct {
	ProviderID           pkgProvider.ID
	BaseURL              string
	APIKey               string
	Models               map[string]string
	Enabled              bool
	MissingEnvironment   []string
	AllowCatalogMutation bool
	AdapterTimeout       time.Duration
}

func ConfigFromEnvironment(
	manifest pkgBenchmark.Manifest,
	providerID string,
	timeout time.Duration,
) (Config, error) {
	return configFromEnvironment(manifest, providerID, timeout, os.Getenv)
}

func configFromEnvironment(
	manifest pkgBenchmark.Manifest,
	providerID string,
	timeout time.Duration,
	getenv Environment,
) (Config, error) {
	if err := manifest.Validate(); err != nil {
		return Config{}, err
	}
	providerManifest, exists := manifest.Providers[providerID]
	if !exists {
		return Config{}, fmt.Errorf("smoke provider %q is not declared in the manifest", providerID)
	}
	if providerID != ProviderOllama && providerID != ProviderLlamaCPP {
		return Config{}, fmt.Errorf("smoke provider %q is not implemented", providerID)
	}
	if timeout <= 0 {
		return Config{}, errors.New("smoke adapter timeout must be positive")
	}
	if getenv == nil {
		return Config{}, errors.New("smoke environment reader is nil")
	}

	config := Config{
		ProviderID:     pkgProvider.ID(providerID),
		Models:         make(map[string]string, len(providerManifest.ModelEnvironment)),
		AdapterTimeout: timeout,
	}
	for _, environmentName := range providerManifest.RequiredEnvironment {
		value := strings.TrimSpace(getenv(environmentName))
		if value == "" {
			config.MissingEnvironment = append(config.MissingEnvironment, environmentName)
			continue
		}
		if strings.HasSuffix(environmentName, "_BASE_URL") {
			config.BaseURL = value
		}
	}
	for role, environmentName := range providerManifest.ModelEnvironment {
		if value := strings.TrimSpace(getenv(environmentName)); value != "" {
			config.Models[role] = value
		}
	}
	if providerID == ProviderLlamaCPP {
		config.APIKey = getenv("MAESTRO_LLAMACPP_API_KEY")
	}
	mutationValue := strings.TrimSpace(getenv("MAESTRO_ALLOW_CATALOG_MUTATION"))
	switch mutationValue {
	case "", "false":
	case "true":
		config.AllowCatalogMutation = true
	default:
		return Config{}, fmt.Errorf(
			"MAESTRO_ALLOW_CATALOG_MUTATION must be true or false, got %q",
			mutationValue,
		)
	}
	config.Enabled = len(config.MissingEnvironment) == 0 && config.BaseURL != ""

	return config, nil
}

func NewRuntime(config Config) (maestro.Runtime, error) {
	runtime := maestro.New()
	if !config.Enabled {
		return runtime, nil
	}

	var configured pkgProvider.Provider
	var err error
	switch string(config.ProviderID) {
	case ProviderOllama:
		configured, err = ollama.New(ollama.Config{
			BaseURL: config.BaseURL, Timeout: config.AdapterTimeout,
			DefaultModel: config.Models["chat"],
		})
	case ProviderLlamaCPP:
		configured, err = llamacpp.New(llamacpp.Config{
			BaseURL: config.BaseURL, Timeout: config.AdapterTimeout,
			DefaultModel: config.Models["chat"], APIKey: config.APIKey,
		})
	default:
		return nil, fmt.Errorf("smoke provider %q is not implemented", config.ProviderID)
	}
	if err != nil {
		return nil, err
	}
	if err := runtime.Providers().Register(configured); err != nil {
		return nil, err
	}
	if err := runtime.Providers().SetDefault(config.ProviderID); err != nil {
		return nil, err
	}

	return runtime, nil
}

func Shutdown(ctx context.Context, runtime maestro.Runtime) error {
	if runtime == nil {
		return nil
	}

	return runtime.Providers().Shutdown(ctx)
}

func (c Config) ConfigurationProfile() pkgBenchmark.ConfigurationProfile {
	models := make(map[string]pkgBenchmark.ModelProfile, len(c.Models))
	for role, modelID := range c.Models {
		models[role] = pkgBenchmark.ModelProfile{ID: modelID}
	}

	return pkgBenchmark.ConfigurationProfile{
		Provider: pkgBenchmark.ProviderProfile{
			ID: string(c.ProviderID), Endpoint: c.BaseURL,
		},
		Models: models,
	}
}
