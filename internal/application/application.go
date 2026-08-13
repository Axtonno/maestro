// Package application composes Maestro's user-facing v0.1.0 execution path.
package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/antonio-cafeo/maestro"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgLaravel "github.com/antonio-cafeo/maestro/pkg/plugin/laravel"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgLlamaCPP "github.com/antonio-cafeo/maestro/pkg/provider/llamacpp"
	pkgOllama "github.com/antonio-cafeo/maestro/pkg/provider/ollama"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type ProviderFactory func(productconfig.Config, string) (pkgProvider.Provider, error)

type Dependencies struct {
	Getenv          func(string) string
	ProviderFactory ProviderFactory
	RunID           func() (pkgAgent.RunID, error)
}

func DefaultDependencies() Dependencies {
	return Dependencies{Getenv: os.Getenv, ProviderFactory: defaultProvider, RunID: randomRunID}
}

type Application struct {
	config    productconfig.Config
	runtime   maestro.Runtime
	plugin    pkgLaravel.Plugin
	workspace pkgContext.Workspace
	started   bool
	runID     func() (pkgAgent.RunID, error)
}

func Build(config productconfig.Config, dependencies Dependencies) (*Application, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	dependencies = normalizeDependencies(dependencies)
	secret, err := config.Secret(dependencies.Getenv)
	if err != nil {
		return nil, err
	}
	configuredProvider, err := dependencies.ProviderFactory(config, secret)
	if err != nil {
		return nil, fmt.Errorf("compose configured provider: %w", err)
	}
	runtime := maestro.New()
	if err := runtime.Providers().Register(configuredProvider); err != nil {
		return nil, fmt.Errorf("register configured provider: %w", err)
	}
	if err := runtime.Providers().SetDefault(pkgProvider.ID(config.Provider.ID)); err != nil {
		return nil, fmt.Errorf("select configured provider: %w", err)
	}
	policy, err := NewProductPolicy(config)
	if err != nil {
		return nil, err
	}
	if err := runtime.Tools().RegisterPolicy(policy); err != nil {
		return nil, fmt.Errorf("register configured policy: %w", err)
	}
	if err := runtime.Plugins().RegisterLoader(
		pkgLaravel.ID,
		pkgLaravel.NewLoader(pkgLaravel.Config{Root: config.Workspace.Root}),
	); err != nil {
		return nil, fmt.Errorf("register Laravel loader: %w", err)
	}
	loaded, err := runtime.Plugins().Load(context.Background(), pkgLaravel.ID)
	if err != nil {
		return nil, fmt.Errorf("load Laravel plugin: %w", err)
	}
	plugin, ok := loaded.(pkgLaravel.Plugin)
	if !ok {
		return nil, fmt.Errorf("loaded Laravel plugin has unexpected type %T", loaded)
	}
	return &Application{config: config, runtime: runtime, plugin: plugin, runID: dependencies.RunID}, nil
}

func (application *Application) Runtime() maestro.Runtime { return application.runtime }

func (application *Application) Start(ctx context.Context) error {
	if application == nil || application.runtime == nil {
		return errors.New("application is not composed")
	}
	if application.started {
		return nil
	}
	if err := application.runtime.Start(ctx); err != nil {
		return fmt.Errorf("start Maestro: %w", err)
	}
	application.started = true
	workspace, err := application.plugin.Workspace(ctx)
	if err != nil {
		return fmt.Errorf("resolve Laravel workspace: %w", err)
	}
	if workspace.ID() != pkgContext.WorkspaceID(application.config.Workspace.ID) {
		return fmt.Errorf("Laravel workspace ID %q differs from configured ID %q", workspace.ID(), application.config.Workspace.ID)
	}
	application.workspace = workspace
	return nil
}

func (application *Application) Close(ctx context.Context) error {
	if application == nil || application.runtime == nil {
		return nil
	}
	var stopError error
	if application.started {
		stopError = application.runtime.Stop(ctx)
		application.started = false
	}
	return errors.Join(stopError, application.runtime.Providers().Shutdown(ctx))
}

func (application *Application) Models(ctx context.Context) ([]pkgProvider.Model, error) {
	models, err := application.runtime.Providers().Models(ctx, pkgProvider.ID(application.config.Provider.ID))
	if err != nil {
		return nil, fmt.Errorf("list configured provider models: %w", err)
	}
	slices.SortFunc(models, func(left, right pkgProvider.Model) int {
		return strings.Compare(left.ID, right.ID)
	})
	return models, nil
}

func AgentDescriptors() []pkgAgent.Descriptor {
	runtime := maestro.New()
	return runtime.Agents().Descriptors()
}

func (application *Application) Execute(ctx context.Context, instruction string) (pkgAgent.RunResult, error) {
	if strings.TrimSpace(instruction) == "" {
		return pkgAgent.RunResult{}, fmt.Errorf("instruction is blank: %w", pkgAgent.ErrInvalidRequest)
	}
	if err := application.Start(ctx); err != nil {
		return pkgAgent.RunResult{}, err
	}
	if _, err := application.runtime.ContextEngine().Index(ctx, application.workspace); err != nil {
		return pkgAgent.RunResult{}, fmt.Errorf("index configured workspace: %w", err)
	}
	query, err := pkgContext.NewRetrievalQuery(
		application.workspace.ID(), instruction,
		pkgContext.RetrievalQueryOptions{
			Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical},
			TopK:    application.config.Context.TopK,
		},
	)
	if err != nil {
		return pkgAgent.RunResult{}, fmt.Errorf("construct configured retrieval: %w", err)
	}
	runID, err := application.runID()
	if err != nil {
		return pkgAgent.RunResult{}, fmt.Errorf("allocate run ID: %w", err)
	}
	request, err := pkgAgent.NewRunRequest(
		runID,
		pkgAgent.ID(application.config.Agent.ID),
		pkgProvider.ID(application.config.Provider.ID),
		application.config.Models.Chat,
		application.workspace.ID(),
		pkgTool.PolicyID(application.config.Policy.ID),
		instruction,
		application.config.AgentLimits(),
		pkgAgent.RunRequestOptions{
			Context: pkgContext.BuildRequest{
				Query:     query,
				Budget:    application.config.ContextBudget(),
				Estimator: "context.utf8-estimator",
			},
			Tools:     application.config.ToolIDs(),
			Streaming: application.config.Agent.Streaming,
			Workspace: &application.workspace,
		},
	)
	if err != nil {
		return pkgAgent.RunResult{}, fmt.Errorf("construct configured agent run: %w", err)
	}
	return application.runtime.Agents().Run(ctx, request)
}

func normalizeDependencies(dependencies Dependencies) Dependencies {
	defaults := DefaultDependencies()
	if dependencies.Getenv == nil {
		dependencies.Getenv = defaults.Getenv
	}
	if dependencies.ProviderFactory == nil {
		dependencies.ProviderFactory = defaults.ProviderFactory
	}
	if dependencies.RunID == nil {
		dependencies.RunID = defaults.RunID
	}
	return dependencies
}

func defaultProvider(config productconfig.Config, secret string) (pkgProvider.Provider, error) {
	switch config.Provider.ID {
	case "ollama":
		return pkgOllama.New(pkgOllama.Config{
			BaseURL:      config.Provider.BaseURL,
			Timeout:      config.Provider.Timeout.Duration,
			DefaultModel: config.Models.Chat,
		})
	case "llama.cpp":
		return pkgLlamaCPP.New(pkgLlamaCPP.Config{
			BaseURL:      config.Provider.BaseURL,
			Timeout:      config.Provider.Timeout.Duration,
			DefaultModel: config.Models.Chat,
			APIKey:       secret,
		})
	default:
		return nil, fmt.Errorf("provider %q is not implemented", config.Provider.ID)
	}
}

func randomRunID() (pkgAgent.RunID, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return pkgAgent.RunID("run-" + hex.EncodeToString(value[:])), nil
}

// HTTPClientProviderFactory is useful for deterministic CLI and application
// tests while preserving the production adapter constructors.
func HTTPClientProviderFactory(client *http.Client) ProviderFactory {
	return func(config productconfig.Config, secret string) (pkgProvider.Provider, error) {
		switch config.Provider.ID {
		case "ollama":
			return pkgOllama.New(pkgOllama.Config{BaseURL: config.Provider.BaseURL, DefaultModel: config.Models.Chat, HTTPClient: client})
		case "llama.cpp":
			return pkgLlamaCPP.New(pkgLlamaCPP.Config{BaseURL: config.Provider.BaseURL, DefaultModel: config.Models.Chat, APIKey: secret, HTTPClient: client})
		default:
			return nil, fmt.Errorf("provider %q is not implemented", config.Provider.ID)
		}
	}
}
