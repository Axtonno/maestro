package controlledmutation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	internalTool "github.com/antonio-cafeo/maestro/internal/tool"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgOllama "github.com/antonio-cafeo/maestro/pkg/provider/ollama"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

//go:embed assets/prompt.txt
var prompt []byte

//go:embed assets/schema.json
var schema []byte

const handoffPollInterval = 25 * time.Millisecond

type ProviderFactory func(productconfig.Config, string) (pkgProvider.Provider, error)

type Dependencies struct {
	Getenv          func(string) string
	ProviderFactory ProviderFactory
	RunID           func() (pkgTool.RunID, error)
	AfterPreview    func()
}

type Request struct {
	File        string
	StartLine   int
	EndLine     int
	Instruction string
	Approver    pkgTool.Approver
}

type Result struct {
	Terminal            string
	Model               string
	InputTokens         int
	OutputTokens        int
	RequestedNumCtx     int
	RequestedNumPredict int
	RequestedResidency  time.Duration
	Effect              pkgTool.EffectState
	Durable             bool
}

type mutationProvider interface {
	pkgProvider.Completer
	pkgProvider.CapabilityInspector
	pkgProvider.ModelDiscoverer
	pkgProvider.ModelUnloader
}

type Service struct {
	config   productconfig.Config
	profile  productconfig.ControlledMutationProfileConfig
	provider mutationProvider
	runID    func() (pkgTool.RunID, error)
	after    func()
}

func Build(config productconfig.Config, dependencies Dependencies) (*Service, error) {
	if err := config.ValidateMutationExecutionProfile(); err != nil {
		return nil, err
	}
	if hash(prompt) != productconfig.MutationPromptSHA256 || hash(schema) != productconfig.MutationSchemaSHA256 {
		return nil, ErrProfileRequired
	}
	if dependencies.Getenv == nil {
		dependencies.Getenv = os.Getenv
	}
	if dependencies.ProviderFactory == nil {
		dependencies.ProviderFactory = defaultProvider
	}
	if dependencies.RunID == nil {
		dependencies.RunID = randomRunID
	}
	secret, err := config.Secret(dependencies.Getenv)
	if err != nil {
		return nil, err
	}
	candidate, err := dependencies.ProviderFactory(config, secret)
	if err != nil {
		return nil, ErrProviderUnavailable
	}
	provider, ok := candidate.(mutationProvider)
	if !ok || nilValue(provider) || provider.ID() != pkgProvider.ID(config.Provider.ID) {
		return nil, ErrCapabilityUnsupported
	}
	return &Service{config: config, profile: config.ControlledMutation, provider: provider, runID: dependencies.RunID, after: dependencies.AfterPreview}, nil
}

func defaultProvider(config productconfig.Config, _ string) (pkgProvider.Provider, error) {
	return pkgOllama.New(pkgOllama.Config{BaseURL: config.Provider.BaseURL, Timeout: config.Provider.Timeout.Duration, DefaultModel: config.ControlledMutation.Model})
}

func (service *Service) Execute(ctx context.Context, request Request) (Result, error) {
	if service == nil || ctx == nil || request.Approver == nil || nilValue(request.Approver) ||
		!validInstruction(request.Instruction) || request.File == "" || request.StartLine < 1 || request.EndLine < request.StartLine {
		return Result{}, ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	runContext, cancel := context.WithTimeout(ctx, service.profile.Timeout.Duration)
	defer cancel()
	run, err := service.runID()
	if err != nil || run.Validate() != nil {
		return Result{}, ErrExecutionFailed
	}
	workspace, err := pkgContext.NewWorkspace(pkgContext.WorkspaceID(service.config.Workspace.ID), service.config.Workspace.Root, pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	if err != nil {
		return Result{}, ErrInvalidRequest
	}
	registry := internalTool.NewWorkspaceRegistry()
	if err := registry.Bind(run, workspace); err != nil {
		return Result{}, ErrExecutionFailed
	}
	defer registry.Unbind(run)
	host, err := internalTool.NewHostBoundMutation(registry)
	if err != nil {
		return Result{}, ErrExecutionFailed
	}
	bound, err := host.Capture(runContext, run, []string{request.File}, request.StartLine, request.EndLine)
	if err != nil {
		return Result{}, err
	}
	selected := bound.Target()
	if err := service.preflight(runContext); err != nil {
		return Result{}, err
	}
	if err := handoff(runContext, service.provider, service.config.DirectChat.Model); err != nil {
		return Result{}, executionError(runContext, err)
	}
	payload, err := json.Marshal(struct {
		Request, SelectedText string
		StartLine, EndLine    int
	}{request.Instruction, selected.Text(), selected.StartLine(), selected.EndLine()})
	if err != nil {
		return Result{}, ErrExecutionFailed
	}
	temperature := 0.0
	options := service.profile.GenerationOptions()
	options.Temperature = &temperature
	response, err := service.provider.Complete(runContext, pkgProvider.CompletionRequest{
		Model:    service.profile.Model,
		Messages: []pkgProvider.Message{{Role: pkgProvider.RoleSystem, Content: string(prompt)}, {Role: pkgProvider.RoleUser, Content: string(payload)}},
		Options:  options, KeepAlive: service.profile.Residency.Duration,
		ToolChoice: pkgProvider.ToolChoice{Mode: pkgProvider.ToolChoiceNone},
		Output:     &pkgProvider.StructuredOutput{Mode: pkgProvider.StructuredOutputJSONSchema, Schema: append(json.RawMessage(nil), schema...)},
	})
	if err != nil {
		return Result{}, executionError(runContext, err)
	}
	if response.Model != service.profile.Model || response.Message.Role != pkgProvider.RoleAssistant || len(response.Message.ToolCalls) != 0 || response.FinishReason != pkgProvider.FinishReasonStop || response.Usage.InputTokens < 0 || response.Usage.OutputTokens < 0 || len(response.Message.Content) == 0 || len(response.Message.Content) > service.profile.MaxOutputBytes || !utf8.ValidString(response.Message.Content) || strings.ContainsRune(response.Message.Content, 0) {
		return Result{}, ErrResponseInvalid
	}
	decision, err := mutation.DecodeHostBoundDecision([]byte(response.Message.Content))
	if err != nil {
		return Result{}, ErrResponseInvalid
	}
	if decision.Decision == mutation.BinaryAbstain {
		return Result{Terminal: "insufficient_information", Model: service.profile.Model, InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens}, ErrInsufficientInfo
	}
	prepared, err := host.Prepare(runContext, bound, pkgTool.CallID(string(run)+"-replace"), []byte(response.Message.Content))
	if err != nil {
		return Result{}, err
	}
	permission, err := pkgTool.NewToolPermissionRequest("controlled-mutation.cli", prepared)
	if err != nil {
		return Result{}, ErrExecutionFailed
	}
	if service.after != nil {
		service.after()
	}
	approval, err := request.Approver.Approve(runContext, permission)
	if err != nil {
		return Result{}, executionError(runContext, err)
	}
	if approval.Kind() != pkgTool.ApprovalAllow || approval.Scope() != pkgTool.GrantOneShot {
		return Result{Terminal: "approval_rejected", Model: service.profile.Model}, ErrApprovalRejected
	}
	toolResult, err := host.Execute(runContext, prepared)
	if err != nil {
		return Result{}, ErrExecutionFailed
	}
	if toolResult.Effect() == pkgTool.EffectUnchanged {
		return Result{Terminal: "stale_source", Model: service.profile.Model, Effect: toolResult.Effect()}, ErrStaleSource
	}
	if toolResult.Outcome() != pkgTool.ResultSuccess || toolResult.Effect() != pkgTool.EffectApplied || !toolResult.Durable() {
		return Result{}, ErrExecutionFailed
	}
	return Result{Terminal: "applied", Model: service.profile.Model, InputTokens: response.Usage.InputTokens, OutputTokens: response.Usage.OutputTokens, RequestedNumCtx: service.profile.NumCtx, RequestedNumPredict: service.profile.NumPredict, RequestedResidency: service.profile.Residency.Duration, Effect: toolResult.Effect(), Durable: toolResult.Durable()}, nil
}

func (service *Service) preflight(ctx context.Context) error {
	if err := verifyIdentity(ctx, service.provider, service.profile.Model, service.profile.Digest); err != nil {
		return err
	}
	report, err := service.provider.InspectCapabilities(ctx, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetModel, Model: service.profile.Model})
	if err != nil {
		return executionError(ctx, err)
	}
	for _, capability := range []pkgProvider.Capability{pkgProvider.CapabilityCompletion, pkgProvider.CapabilityStructuredOutput, pkgProvider.CapabilityModelDiscovery, pkgProvider.CapabilityModelUnload} {
		if !capabilityAvailable(report, capability) {
			return ErrCapabilityUnsupported
		}
	}
	if err := pkgProvider.ValidateGenerationCapabilities(report, service.profile.GenerationOptions()); err != nil {
		return ErrCapabilityUnsupported
	}
	return nil
}

func verifyIdentity(ctx context.Context, provider pkgProvider.ModelDiscoverer, model, digest string) error {
	models, err := provider.DiscoverModels(ctx)
	if err != nil {
		return executionError(ctx, err)
	}
	for _, candidate := range models {
		if candidate.Model.ID == model {
			if candidate.Digest == digest {
				return nil
			}
			return ErrModelIdentity
		}
	}
	return ErrModelIdentity
}

func handoff(ctx context.Context, provider mutationProvider, outgoing string) error {
	if err := provider.UnloadModel(ctx, pkgProvider.ModelUnloadRequest{Model: outgoing}); err != nil {
		return err
	}
	for {
		models, err := provider.DiscoverModels(ctx)
		if err != nil {
			return err
		}
		loaded := false
		for _, model := range models {
			loaded = loaded || model.Model.ID == outgoing && model.State == pkgProvider.ModelStateLoaded
		}
		if !loaded {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(handoffPollInterval):
		}
	}
}

func PromptSHA256() string { return hash(prompt) }
func SchemaSHA256() string { return hash(schema) }

func hash(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

func validInstruction(value string) bool {
	return strings.TrimSpace(value) != "" && len(value) <= 1<<20 && utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func executionError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return ErrProviderUnavailable
}

func capabilityAvailable(report pkgProvider.CapabilityReport, capability pkgProvider.Capability) bool {
	for _, item := range report.Capabilities {
		if item.Capability == capability {
			return item.Support == pkgProvider.CapabilitySupported && item.Availability == pkgProvider.CapabilityAvailabilityAvailable
		}
	}
	return false
}

func nilValue(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflect.ValueOf(value).IsNil()
}

func randomRunID() (pkgTool.RunID, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate mutation run ID: %w", err)
	}
	return pkgTool.RunID("mutation-" + hex.EncodeToString(value)), nil
}
