package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

var _ pkgProvider.Runtime = (*runtime)(nil)

type runtime struct {
	mu sync.RWMutex

	providers map[pkgProvider.ID]pkgProvider.Provider
	defaultID pkgProvider.ID

	residencyMu           sync.Mutex
	residencies           map[residencyKey]*residencyEntry
	residencyScheduler    residencyScheduler
	residencyShuttingDown bool

	resilienceMu       sync.RWMutex
	resiliencePolicies map[resilienceKey]pkgProvider.ResiliencePolicy
	circuits           map[resilienceKey]*circuitBreaker
	resilienceClock    resilienceClock
	resilienceJitter   resilienceJitter
}

func NewRuntime(defaultID pkgProvider.ID) pkgProvider.Runtime {
	return newRuntimeWithResidencyScheduler(defaultID, realResidencyScheduler{})
}

func newRuntimeWithResidencyScheduler(
	defaultID pkgProvider.ID,
	scheduler residencyScheduler,
) *runtime {
	return newRuntimeWithDependencies(
		defaultID,
		scheduler,
		realResilienceClock{},
		newLockedResilienceJitter(),
	)
}

func newRuntimeWithDependencies(
	defaultID pkgProvider.ID,
	scheduler residencyScheduler,
	clock resilienceClock,
	jitter resilienceJitter,
) *runtime {
	return &runtime{
		providers:          make(map[pkgProvider.ID]pkgProvider.Provider),
		defaultID:          defaultID,
		residencies:        make(map[residencyKey]*residencyEntry),
		residencyScheduler: scheduler,
		resiliencePolicies: make(map[resilienceKey]pkgProvider.ResiliencePolicy),
		circuits:           make(map[resilienceKey]*circuitBreaker),
		resilienceClock:    clock,
		resilienceJitter:   jitter,
	}
}

func (r *runtime) Register(provider pkgProvider.Provider) error {
	if nilProvider(provider) {
		return fmt.Errorf(
			"register provider: provider is nil: %w",
			pkgProvider.ErrInvalidProvider,
		)
	}

	providerID := provider.ID()
	if !validProviderID(providerID) {
		return fmt.Errorf(
			"register provider: invalid ID %q: %w",
			providerID,
			pkgProvider.ErrInvalidProvider,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[providerID]; exists {
		return fmt.Errorf(
			"register provider %q: %w",
			providerID,
			pkgProvider.ErrAlreadyRegistered,
		)
	}

	r.providers[providerID] = provider

	return nil
}

func (r *runtime) Resolve(
	providerID pkgProvider.ID,
) (pkgProvider.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.resolveLocked(providerID)
}

func (r *runtime) SetDefault(providerID pkgProvider.ID) error {
	if !validProviderID(providerID) {
		return fmt.Errorf(
			"set default provider: invalid ID %q: %w",
			providerID,
			pkgProvider.ErrInvalidProvider,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[providerID]; !exists {
		return fmt.Errorf(
			"set default provider %q: %w",
			providerID,
			pkgProvider.ErrNotFound,
		)
	}

	r.defaultID = providerID

	return nil
}

func (r *runtime) Default() (pkgProvider.Provider, error) {
	return r.Resolve("")
}

func (r *runtime) Complete(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with provider %q: %w",
			providerID,
			err,
		)
	}

	completer, ok := provider.(pkgProvider.Completer)
	if !ok {
		return pkgProvider.CompletionResponse{}, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationCompletion,
			"completion",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationCompletion,
		request.Model,
	)
	if err != nil {
		return pkgProvider.CompletionResponse{}, err
	}

	release, err := r.acquireModelResidency(ctx, provider, request.Model)
	if err != nil {
		execution.finish(err)
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	response, operationErr := executeWithResilience(
		ctx,
		execution,
		func() (pkgProvider.CompletionResponse, error) {
			return completer.Complete(ctx, request)
		},
	)
	err = errors.Join(operationErr, releaseResidency(release))
	if err != nil {
		return pkgProvider.CompletionResponse{}, fmt.Errorf(
			"complete with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return response, nil
}

func (r *runtime) Stream(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return nil, fmt.Errorf(
			"stream with provider %q: %w",
			providerID,
			err,
		)
	}

	streamer, ok := provider.(pkgProvider.Streamer)
	if !ok {
		return nil, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationStreaming,
			"streaming",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationStreaming,
		request.Model,
	)
	if err != nil {
		return nil, err
	}

	release, err := r.acquireModelResidency(ctx, provider, request.Model)
	if err != nil {
		execution.finish(err)
		return nil, fmt.Errorf(
			"stream with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	stream, operationErr := openStreamWithResilience(
		ctx,
		execution,
		func() (pkgProvider.Stream, error) {
			return streamer.Stream(ctx, request)
		},
	)
	if operationErr != nil {
		err = errors.Join(operationErr, releaseResidency(release))

		return nil, fmt.Errorf(
			"stream with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	if nilStream(stream) {
		execution.finish(pkgProvider.ErrInvalidStream)
		err = errors.Join(
			fmt.Errorf(
				"provider returned a nil stream: %w",
				pkgProvider.ErrInvalidStream,
			),
			releaseResidency(release),
		)

		return nil, fmt.Errorf(
			"stream with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	if release == nil {
		return stream, nil
	}

	return &residencyStream{stream: stream, release: release}, nil
}

func (r *runtime) Embed(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with provider %q: %w",
			providerID,
			err,
		)
	}

	embedder, ok := provider.(pkgProvider.Embedder)
	if !ok {
		return pkgProvider.EmbeddingResponse{}, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationEmbedding,
			"embedding",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationEmbedding,
		request.Model,
	)
	if err != nil {
		return pkgProvider.EmbeddingResponse{}, err
	}

	release, err := r.acquireModelResidency(ctx, provider, request.Model)
	if err != nil {
		execution.finish(err)
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	response, operationErr := executeWithResilience(
		ctx,
		execution,
		func() (pkgProvider.EmbeddingResponse, error) {
			return embedder.Embed(ctx, request)
		},
	)
	err = errors.Join(operationErr, releaseResidency(release))
	if err != nil {
		return pkgProvider.EmbeddingResponse{}, fmt.Errorf(
			"embed with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return response, nil
}

func (r *runtime) Models(
	ctx context.Context,
	providerID pkgProvider.ID,
) ([]pkgProvider.Model, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return nil, fmt.Errorf(
			"list models with provider %q: %w",
			providerID,
			err,
		)
	}

	modelLister, ok := provider.(pkgProvider.ModelLister)
	if !ok {
		return nil, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelListing,
			"model listing",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelListing,
		"",
	)
	if err != nil {
		return nil, err
	}

	models, err := executeWithResilience(
		ctx,
		execution,
		func() ([]pkgProvider.Model, error) {
			return modelLister.Models(ctx)
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list models with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return models, nil
}

func (r *runtime) DiscoverModels(
	ctx context.Context,
	providerID pkgProvider.ID,
) ([]pkgProvider.ModelInfo, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return nil, fmt.Errorf(
			"discover models with provider %q: %w",
			providerID,
			err,
		)
	}

	discoverer, ok := provider.(pkgProvider.ModelDiscoverer)
	if !ok {
		return nil, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelDiscovery,
			"model discovery",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelDiscovery,
		"",
	)
	if err != nil {
		return nil, err
	}

	models, err := executeWithResilience(
		ctx,
		execution,
		func() ([]pkgProvider.ModelInfo, error) {
			return discoverer.DiscoverModels(ctx)
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"discover models with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return models, nil
}

func (r *runtime) LoadModel(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.ModelLoadRequest,
) error {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return fmt.Errorf(
			"load model with provider %q: %w",
			providerID,
			err,
		)
	}

	loader, ok := provider.(pkgProvider.ModelLoader)
	if !ok {
		return unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelLoad,
			"model loading",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelLoad,
		request.Model,
	)
	if err != nil {
		return err
	}

	if err := executeErrorWithResilience(ctx, execution, func() error {
		return loader.LoadModel(ctx, request)
	}); err != nil {
		return fmt.Errorf(
			"load model with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return nil
}

func (r *runtime) UnloadModel(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.ModelUnloadRequest,
) error {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return fmt.Errorf(
			"unload model with provider %q: %w",
			providerID,
			err,
		)
	}

	unloader, ok := provider.(pkgProvider.ModelUnloader)
	if !ok {
		return unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelUnload,
			"model unloading",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelUnload,
		request.Model,
	)
	if err != nil {
		return err
	}

	if err := executeErrorWithResilience(ctx, execution, func() error {
		return unloader.UnloadModel(ctx, request)
	}); err != nil {
		return fmt.Errorf(
			"unload model with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return nil
}

func (r *runtime) PullModel(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.ModelPullRequest,
) (pkgProvider.ModelPullStream, error) {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return nil, fmt.Errorf(
			"pull model with provider %q: %w",
			providerID,
			err,
		)
	}

	puller, ok := provider.(pkgProvider.ModelPuller)
	if !ok {
		return nil, unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelPull,
			"model pulling",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelPull,
		request.Model,
	)
	if err != nil {
		return nil, err
	}

	stream, err := puller.PullModel(ctx, request)
	if err != nil {
		execution.finish(err)
		return nil, fmt.Errorf(
			"pull model with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	if nilModelPullStream(stream) {
		execution.finish(pkgProvider.ErrInvalidStream)
		return nil, fmt.Errorf(
			"pull model with provider %q: provider returned a nil stream: %w",
			provider.ID(),
			pkgProvider.ErrInvalidStream,
		)
	}

	if execution == nil {
		return stream, nil
	}

	return &resilienceModelPullStream{
		stream: stream, execution: execution,
	}, nil
}

func (r *runtime) RemoveModel(
	ctx context.Context,
	providerID pkgProvider.ID,
	request pkgProvider.ModelRemoveRequest,
) error {
	provider, err := r.Resolve(providerID)
	if err != nil {
		return fmt.Errorf(
			"remove model with provider %q: %w",
			providerID,
			err,
		)
	}

	remover, ok := provider.(pkgProvider.ModelRemover)
	if !ok {
		return unsupportedCapability(
			provider.ID(),
			pkgProvider.OperationModelRemove,
			"model removal",
		)
	}
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelRemove,
		request.Model,
	)
	if err != nil {
		return err
	}

	if err := executeErrorWithResilience(ctx, execution, func() error {
		return remover.RemoveModel(ctx, request)
	}); err != nil {
		return fmt.Errorf(
			"remove model with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return nil
}

func (r *runtime) resolveLocked(
	providerID pkgProvider.ID,
) (pkgProvider.Provider, error) {
	selectedID := providerID
	if selectedID == "" {
		selectedID = r.defaultID
		if selectedID == "" {
			return nil, pkgProvider.ErrDefaultNotConfigured
		}
	}

	if !validProviderID(selectedID) {
		return nil, fmt.Errorf(
			"resolve provider: invalid ID %q: %w",
			selectedID,
			pkgProvider.ErrInvalidProvider,
		)
	}

	provider, exists := r.providers[selectedID]
	if !exists {
		return nil, fmt.Errorf(
			"resolve provider %q: %w",
			selectedID,
			pkgProvider.ErrNotFound,
		)
	}

	return provider, nil
}

func unsupportedCapability(
	providerID pkgProvider.ID,
	operation pkgProvider.Operation,
	capability string,
) error {
	return pkgProvider.NewProviderError(
		pkgProvider.ProviderErrorDetails{
			Kind:      pkgProvider.ErrorKindCapabilityNotFound,
			Operation: operation,
			Provider:  providerID,
			Message:   fmt.Sprintf("%s is not supported", capability),
		},
		pkgProvider.ErrUnsupportedCapability,
	)
}

func validProviderID(providerID pkgProvider.ID) bool {
	value := string(providerID)

	return value != "" && strings.TrimSpace(value) == value
}

func nilProvider(provider pkgProvider.Provider) bool {
	return nilInterface(provider)
}

func nilStream(stream pkgProvider.Stream) bool {
	return nilInterface(stream)
}

func nilModelPullStream(stream pkgProvider.ModelPullStream) bool {
	return nilInterface(stream)
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)

	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
