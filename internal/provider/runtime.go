package provider

import (
	"context"
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
}

func NewRuntime(defaultID pkgProvider.ID) pkgProvider.Runtime {
	return &runtime{
		providers: make(map[pkgProvider.ID]pkgProvider.Provider),
		defaultID: defaultID,
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
			"completion",
		)
	}

	response, err := completer.Complete(ctx, request)
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
		return nil, unsupportedCapability(provider.ID(), "streaming")
	}

	stream, err := streamer.Stream(ctx, request)
	if err != nil {
		return nil, fmt.Errorf(
			"stream with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	if nilStream(stream) {
		return nil, fmt.Errorf(
			"stream with provider %q: provider returned a nil stream: %w",
			provider.ID(),
			pkgProvider.ErrInvalidStream,
		)
	}

	return stream, nil
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
			"embedding",
		)
	}

	response, err := embedder.Embed(ctx, request)
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
		return nil, unsupportedCapability(provider.ID(), "model listing")
	}

	models, err := modelLister.Models(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"list models with provider %q: %w",
			provider.ID(),
			err,
		)
	}

	return models, nil
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
	capability string,
) error {
	return fmt.Errorf(
		"provider %q does not support %s: %w",
		providerID,
		capability,
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
