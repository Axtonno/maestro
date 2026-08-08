package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

const residencyCleanupTimeout = 5 * time.Second

type residencyKey struct {
	providerID pkgProvider.ID
	model      string
}

type residencyEntry struct {
	provider pkgProvider.Provider
	policy   pkgProvider.ModelResidencyPolicy

	active        int
	owned         bool
	transitioning bool
	changed       chan struct{}

	timer      residencyTimer
	generation uint64
}

type residencyTimer interface {
	Stop() bool
}

type residencyScheduler interface {
	AfterFunc(time.Duration, func()) residencyTimer
}

type realResidencyScheduler struct{}

func (realResidencyScheduler) AfterFunc(
	duration time.Duration,
	callback func(),
) residencyTimer {
	return time.AfterFunc(duration, callback)
}

func (r *runtime) SetModelResidencyPolicy(
	ctx context.Context,
	providerID pkgProvider.ID,
	policy pkgProvider.ModelResidencyPolicy,
) error {
	if ctx == nil {
		return fmt.Errorf(
			"set model residency policy: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateResidencyPolicy(policy); err != nil {
		return err
	}

	selected, err := r.Resolve(providerID)
	if err != nil {
		return fmt.Errorf(
			"set model residency policy with provider %q: %w",
			providerID,
			err,
		)
	}
	if policy.Autoload {
		if _, ok := selected.(pkgProvider.ModelDiscoverer); !ok {
			return unsupportedCapability(
				selected.ID(),
				pkgProvider.OperationResidencyPolicy,
				"model discovery",
			)
		}
		if _, ok := selected.(pkgProvider.ModelLoader); !ok {
			return unsupportedCapability(
				selected.ID(),
				pkgProvider.OperationResidencyPolicy,
				"model loading",
			)
		}
		if _, ok := selected.(pkgProvider.ModelUnloader); !ok {
			return unsupportedCapability(
				selected.ID(),
				pkgProvider.OperationResidencyPolicy,
				"model unloading",
			)
		}
	}

	key := residencyKey{providerID: selected.ID(), model: policy.Model}
	for {
		r.residencyMu.Lock()
		if r.residencyShuttingDown {
			r.residencyMu.Unlock()

			return pkgProvider.ErrResidencyShuttingDown
		}

		entry := r.residencies[key]
		if entry == nil {
			entry = &residencyEntry{
				provider: selected,
				changed:  make(chan struct{}),
			}
			r.residencies[key] = entry
		}

		if entry.transitioning {
			changed := entry.changed
			r.residencyMu.Unlock()

			if err := waitForResidencyChange(ctx, changed); err != nil {
				return err
			}
			continue
		}

		entry.policy = policy
		r.stopResidencyTimerLocked(entry)
		signalResidencyLocked(entry)

		if entry.active > 0 || !entry.owned || policy.Persistent {
			r.residencyMu.Unlock()

			return nil
		}
		if policy.Autoload && policy.KeepAlive > 0 {
			r.scheduleResidencyExpiryLocked(key, entry, policy.KeepAlive)
			r.residencyMu.Unlock()

			return nil
		}

		entry.transitioning = true
		signalResidencyLocked(entry)
		r.residencyMu.Unlock()

		err := r.unloadResidency(ctx, entry.provider, key.model)
		r.finishResidencyUnload(key, entry, err)

		if err != nil {
			return fmt.Errorf(
				"apply model residency policy for provider %q model %q: %w",
				key.providerID,
				key.model,
				err,
			)
		}

		return nil
	}
}

func (r *runtime) ResidencyPolicy(
	providerID pkgProvider.ID,
	model string,
) (pkgProvider.ModelResidencyPolicy, bool, error) {
	if !validModelID(model) {
		return pkgProvider.ModelResidencyPolicy{}, false, fmt.Errorf(
			"get model residency policy: invalid model %q: %w",
			model,
			pkgProvider.ErrInvalidResidencyPolicy,
		)
	}

	selected, err := r.Resolve(providerID)
	if err != nil {
		return pkgProvider.ModelResidencyPolicy{}, false, fmt.Errorf(
			"get model residency policy with provider %q: %w",
			providerID,
			err,
		)
	}

	key := residencyKey{providerID: selected.ID(), model: model}
	r.residencyMu.Lock()
	defer r.residencyMu.Unlock()

	entry, exists := r.residencies[key]
	if !exists {
		return pkgProvider.ModelResidencyPolicy{}, false, nil
	}

	return entry.policy, true, nil
}

func (r *runtime) acquireModelResidency(
	ctx context.Context,
	provider pkgProvider.Provider,
	model string,
) (func() error, error) {
	if !validModelID(model) {
		return nil, nil
	}

	key := residencyKey{providerID: provider.ID(), model: model}
	for {
		r.residencyMu.Lock()
		if r.residencyShuttingDown {
			r.residencyMu.Unlock()

			return nil, pkgProvider.ErrResidencyShuttingDown
		}

		entry := r.residencies[key]
		if entry == nil || !entry.policy.Autoload {
			r.residencyMu.Unlock()

			return nil, nil
		}
		if entry.transitioning {
			changed := entry.changed
			r.residencyMu.Unlock()

			if err := waitForResidencyChange(ctx, changed); err != nil {
				return nil, err
			}
			continue
		}
		if entry.active > 0 {
			entry.active++
			signalResidencyLocked(entry)
			r.residencyMu.Unlock()

			return r.newResidencyRelease(key), nil
		}

		r.stopResidencyTimerLocked(entry)
		entry.transitioning = true
		signalResidencyLocked(entry)
		previouslyOwned := entry.owned
		r.residencyMu.Unlock()

		loaded, err := r.ensureModelResident(ctx, entry.provider, key.model)

		r.residencyMu.Lock()
		current := r.residencies[key]
		if current == entry {
			entry.transitioning = false
			if err == nil {
				entry.owned = previouslyOwned || loaded
				entry.active++
			}
			signalResidencyLocked(entry)
		}
		r.residencyMu.Unlock()

		if err != nil {
			return nil, fmt.Errorf(
				"ensure residency with provider %q model %q: %w",
				key.providerID,
				key.model,
				err,
			)
		}

		return r.newResidencyRelease(key), nil
	}
}

func (r *runtime) newResidencyRelease(key residencyKey) func() error {
	var once sync.Once
	var releaseError error

	return func() error {
		once.Do(func() {
			releaseError = r.releaseModelResidency(key)
		})

		return releaseError
	}
}

func (r *runtime) releaseModelResidency(key residencyKey) error {
	r.residencyMu.Lock()
	entry := r.residencies[key]
	if entry == nil || entry.active == 0 {
		r.residencyMu.Unlock()

		return nil
	}

	entry.active--
	signalResidencyLocked(entry)
	if entry.active > 0 || !entry.owned {
		r.residencyMu.Unlock()

		return nil
	}

	policy := entry.policy
	if !r.residencyShuttingDown && policy.Autoload && policy.Persistent {
		r.residencyMu.Unlock()

		return nil
	}
	if !r.residencyShuttingDown && policy.Autoload && policy.KeepAlive > 0 {
		r.scheduleResidencyExpiryLocked(key, entry, policy.KeepAlive)
		r.residencyMu.Unlock()

		return nil
	}

	entry.transitioning = true
	signalResidencyLocked(entry)
	r.residencyMu.Unlock()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		residencyCleanupTimeout,
	)
	defer cancel()

	err := r.unloadResidency(ctx, entry.provider, key.model)
	r.finishResidencyUnload(key, entry, err)

	if err != nil {
		return fmt.Errorf(
			"release residency with provider %q model %q: %w",
			key.providerID,
			key.model,
			err,
		)
	}

	return nil
}

func (r *runtime) scheduleResidencyExpiryLocked(
	key residencyKey,
	entry *residencyEntry,
	duration time.Duration,
) {
	entry.generation++
	generation := entry.generation
	entry.timer = r.residencyScheduler.AfterFunc(duration, func() {
		r.expireModelResidency(key, generation)
	})
}

func (r *runtime) expireModelResidency(
	key residencyKey,
	generation uint64,
) {
	r.residencyMu.Lock()
	entry := r.residencies[key]
	if entry == nil || entry.generation != generation ||
		entry.active > 0 || !entry.owned || entry.transitioning {
		r.residencyMu.Unlock()

		return
	}

	entry.timer = nil
	entry.transitioning = true
	signalResidencyLocked(entry)
	r.residencyMu.Unlock()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		residencyCleanupTimeout,
	)
	defer cancel()

	err := r.unloadResidency(ctx, entry.provider, key.model)
	r.finishResidencyUnload(key, entry, err)
}

func (r *runtime) finishResidencyUnload(
	key residencyKey,
	entry *residencyEntry,
	err error,
) {
	r.residencyMu.Lock()
	defer r.residencyMu.Unlock()

	if current := r.residencies[key]; current != entry {
		return
	}

	entry.transitioning = false
	if err == nil {
		entry.owned = false
	}
	signalResidencyLocked(entry)
}

func (r *runtime) stopResidencyTimerLocked(entry *residencyEntry) {
	entry.generation++
	if entry.timer != nil {
		entry.timer.Stop()
		entry.timer = nil
	}
}

func (r *runtime) Shutdown(ctx context.Context) (shutdownError error) {
	if ctx == nil {
		return fmt.Errorf(
			"shutdown provider residency: context is nil: %w",
			pkgProvider.ErrInvalidRequest,
		)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	r.residencyMu.Lock()
	if r.residencyShuttingDown {
		r.residencyMu.Unlock()

		return pkgProvider.ErrResidencyShuttingDown
	}
	r.residencyShuttingDown = true
	keys := make([]residencyKey, 0, len(r.residencies))
	for key, entry := range r.residencies {
		r.stopResidencyTimerLocked(entry)
		signalResidencyLocked(entry)
		keys = append(keys, key)
	}
	r.residencyMu.Unlock()

	defer func() {
		r.residencyMu.Lock()
		r.residencyShuttingDown = false
		for _, entry := range r.residencies {
			signalResidencyLocked(entry)
		}
		r.residencyMu.Unlock()
	}()

	var shutdownErrors []error
	for _, key := range keys {
		if err := r.shutdownResidency(ctx, key); err != nil {
			shutdownErrors = append(shutdownErrors, err)
			if ctx.Err() != nil {
				break
			}
		}
	}

	return errors.Join(shutdownErrors...)
}

func (r *runtime) shutdownResidency(
	ctx context.Context,
	key residencyKey,
) error {
	for {
		r.residencyMu.Lock()
		entry := r.residencies[key]
		if entry == nil || !entry.owned {
			r.residencyMu.Unlock()

			return nil
		}
		if entry.active > 0 || entry.transitioning {
			changed := entry.changed
			r.residencyMu.Unlock()

			if err := waitForResidencyChange(ctx, changed); err != nil {
				return err
			}
			continue
		}

		entry.transitioning = true
		signalResidencyLocked(entry)
		r.residencyMu.Unlock()

		err := r.unloadResidency(ctx, entry.provider, key.model)
		r.finishResidencyUnload(key, entry, err)

		if err != nil {
			return fmt.Errorf(
				"shutdown residency with provider %q model %q: %w",
				key.providerID,
				key.model,
				err,
			)
		}

		return nil
	}
}

func (r *runtime) ensureModelResident(
	ctx context.Context,
	provider pkgProvider.Provider,
	model string,
) (bool, error) {
	discoverer := provider.(pkgProvider.ModelDiscoverer)
	discoveryExecution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelDiscovery,
		"",
	)
	if err != nil {
		return false, err
	}
	models, err := executeWithResilience(
		ctx,
		discoveryExecution,
		func() ([]pkgProvider.ModelInfo, error) {
			return discoverer.DiscoverModels(ctx)
		},
	)
	if err != nil {
		return false, err
	}

	for _, info := range models {
		if info.Model.ID != model {
			continue
		}
		if modelStateIsResident(info.State) {
			return false, nil
		}
		break
	}

	loader := provider.(pkgProvider.ModelLoader)
	loadExecution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelLoad,
		model,
	)
	if err != nil {
		return false, err
	}
	if err := executeErrorWithResilience(ctx, loadExecution, func() error {
		return loader.LoadModel(
			ctx,
			pkgProvider.ModelLoadRequest{Model: model},
		)
	}); err != nil {
		return false, err
	}

	return true, nil
}

func (r *runtime) unloadResidency(
	ctx context.Context,
	provider pkgProvider.Provider,
	model string,
) error {
	unloader := provider.(pkgProvider.ModelUnloader)
	execution, err := r.beginResilience(
		ctx,
		provider.ID(),
		pkgProvider.OperationModelUnload,
		model,
	)
	if err != nil {
		return err
	}

	return executeErrorWithResilience(ctx, execution, func() error {
		return unloader.UnloadModel(
			ctx,
			pkgProvider.ModelUnloadRequest{Model: model},
		)
	})
}

func modelStateIsResident(state pkgProvider.ModelState) bool {
	switch state {
	case pkgProvider.ModelStateLoading,
		pkgProvider.ModelStateLoaded,
		pkgProvider.ModelStateSleeping:
		return true
	default:
		return false
	}
}

func validateResidencyPolicy(policy pkgProvider.ModelResidencyPolicy) error {
	if !validModelID(policy.Model) {
		return fmt.Errorf(
			"set model residency policy: invalid model %q: %w",
			policy.Model,
			pkgProvider.ErrInvalidResidencyPolicy,
		)
	}
	if policy.KeepAlive < 0 {
		return fmt.Errorf(
			"set model residency policy for model %q: keep-alive cannot be negative: %w",
			policy.Model,
			pkgProvider.ErrInvalidResidencyPolicy,
		)
	}
	if policy.Persistent && policy.KeepAlive > 0 {
		return fmt.Errorf(
			"set model residency policy for model %q: persistent and keep-alive are mutually exclusive: %w",
			policy.Model,
			pkgProvider.ErrInvalidResidencyPolicy,
		)
	}
	if !policy.Autoload && (policy.Persistent || policy.KeepAlive > 0) {
		return fmt.Errorf(
			"set model residency policy for model %q: disabled autoload cannot retain residency: %w",
			policy.Model,
			pkgProvider.ErrInvalidResidencyPolicy,
		)
	}

	return nil
}

func validModelID(model string) bool {
	return model != "" && strings.TrimSpace(model) == model
}

func signalResidencyLocked(entry *residencyEntry) {
	close(entry.changed)
	entry.changed = make(chan struct{})
}

func waitForResidencyChange(ctx context.Context, changed <-chan struct{}) error {
	select {
	case <-changed:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseResidency(release func() error) error {
	if release == nil {
		return nil
	}

	return release()
}

type residencyStream struct {
	stream  pkgProvider.Stream
	release func() error

	once sync.Once
	err  error
}

func (s *residencyStream) Recv() (pkgProvider.StreamChunk, error) {
	chunk, err := s.stream.Recv()
	if err == nil {
		return chunk, nil
	}

	releaseError := s.releaseOnce()
	if releaseError == nil {
		return chunk, err
	}
	if errors.Is(err, io.EOF) {
		return chunk, releaseError
	}

	return chunk, errors.Join(err, releaseError)
}

func (s *residencyStream) Close() error {
	return errors.Join(s.stream.Close(), s.releaseOnce())
}

func (s *residencyStream) releaseOnce() error {
	s.once.Do(func() {
		s.err = s.release()
	})

	return s.err
}
