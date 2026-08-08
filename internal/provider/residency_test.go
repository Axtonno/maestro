package provider

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type fakeResidencyScheduler struct {
	mu     sync.Mutex
	timers []*fakeResidencyTimer
}

type fakeResidencyTimer struct {
	mu       sync.Mutex
	duration time.Duration
	callback func()
	stopped  bool
	fired    bool
}

func (s *fakeResidencyScheduler) AfterFunc(
	duration time.Duration,
	callback func(),
) residencyTimer {
	timer := &fakeResidencyTimer{duration: duration, callback: callback}
	s.mu.Lock()
	s.timers = append(s.timers, timer)
	s.mu.Unlock()

	return timer
}

func (s *fakeResidencyScheduler) latest(t *testing.T) *fakeResidencyTimer {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.timers) == 0 {
		t.Fatal("expected a scheduled residency timer")
	}

	return s.timers[len(s.timers)-1]
}

func (t *fakeResidencyTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped || t.fired {
		return false
	}
	t.stopped = true

	return true
}

func (t *fakeResidencyTimer) Fire() bool {
	t.mu.Lock()
	if t.stopped || t.fired {
		t.mu.Unlock()

		return false
	}
	t.fired = true
	callback := t.callback
	t.mu.Unlock()

	callback()

	return true
}

type residencyProvider struct {
	identityProvider

	mu sync.Mutex

	state pkgProvider.ModelState

	discoveries int
	loads       int
	unloads     int
	completions int
	embeddings  int

	loadStarted     chan struct{}
	allowLoad       <-chan struct{}
	completionEntry chan struct{}
	allowCompletion <-chan struct{}

	stream pkgProvider.Stream
}

func (p *residencyProvider) DiscoverModels(
	context.Context,
) ([]pkgProvider.ModelInfo, error) {
	p.mu.Lock()
	p.discoveries++
	state := p.state
	p.mu.Unlock()

	return []pkgProvider.ModelInfo{{
		Model: pkgProvider.Model{ID: "qwen"},
		State: state,
	}}, nil
}

func (p *residencyProvider) LoadModel(
	ctx context.Context,
	request pkgProvider.ModelLoadRequest,
) error {
	p.mu.Lock()
	p.loads++
	started := p.loadStarted
	allow := p.allowLoad
	p.mu.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if allow != nil {
		select {
		case <-allow:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	p.mu.Lock()
	p.state = pkgProvider.ModelStateLoaded
	p.mu.Unlock()

	return nil
}

func (p *residencyProvider) UnloadModel(
	_ context.Context,
	_ pkgProvider.ModelUnloadRequest,
) error {
	p.mu.Lock()
	p.unloads++
	p.state = pkgProvider.ModelStateAvailable
	p.mu.Unlock()

	return nil
}

func (p *residencyProvider) Complete(
	ctx context.Context,
	_ pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	p.mu.Lock()
	p.completions++
	entry := p.completionEntry
	allow := p.allowCompletion
	p.mu.Unlock()

	if entry != nil {
		entry <- struct{}{}
	}
	if allow != nil {
		select {
		case <-allow:
		case <-ctx.Done():
			return pkgProvider.CompletionResponse{}, ctx.Err()
		}
	}

	return pkgProvider.CompletionResponse{Model: "qwen"}, nil
}

func (p *residencyProvider) Stream(
	context.Context,
	pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	return p.stream, nil
}

func (p *residencyProvider) Embed(
	context.Context,
	pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	p.mu.Lock()
	p.embeddings++
	p.mu.Unlock()

	return pkgProvider.EmbeddingResponse{Model: "qwen"}, nil
}

func (p *residencyProvider) counts() (int, int, int, int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.discoveries, p.loads, p.unloads, p.completions, p.embeddings
}

func newResidencyRuntime(
	t *testing.T,
	scheduler residencyScheduler,
	provider *residencyProvider,
) *runtime {
	t.Helper()

	providerRuntime := newRuntimeWithResidencyScheduler("test", scheduler)
	if err := providerRuntime.Register(provider); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	return providerRuntime
}

func setResidencyPolicy(
	t *testing.T,
	runtime *runtime,
	policy pkgProvider.ModelResidencyPolicy,
) {
	t.Helper()

	if err := runtime.SetModelResidencyPolicy(
		context.Background(),
		"test",
		policy,
	); err != nil {
		t.Fatalf("set residency policy: %v", err)
	}
}

func TestResidencyPolicyIsOptIn(t *testing.T) {
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)

	if _, err := runtime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}

	discoveries, loads, unloads, completions, _ := provider.counts()
	if discoveries != 0 || loads != 0 || unloads != 0 || completions != 1 {
		t.Fatalf(
			"unexpected calls without policy: discoveries=%d loads=%d unloads=%d completions=%d",
			discoveries,
			loads,
			unloads,
			completions,
		)
	}
}

func TestResidencyPolicyLoadsAndReleasesOwnedModel(t *testing.T) {
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	policy := pkgProvider.ModelResidencyPolicy{Model: "qwen", Autoload: true}
	setResidencyPolicy(t, runtime, policy)

	stored, exists, err := runtime.ResidencyPolicy("", "qwen")
	if err != nil || !exists || stored != policy {
		t.Fatalf("unexpected stored policy %#v, exists=%t, error=%v", stored, exists, err)
	}

	if _, err := runtime.Embed(
		context.Background(),
		"",
		pkgProvider.EmbeddingRequest{Model: "qwen", Inputs: []string{"hello"}},
	); err != nil {
		t.Fatalf("embed: %v", err)
	}

	discoveries, loads, unloads, _, embeddings := provider.counts()
	if discoveries != 1 || loads != 1 || unloads != 1 || embeddings != 1 {
		t.Fatalf(
			"unexpected calls: discoveries=%d loads=%d unloads=%d embeddings=%d",
			discoveries,
			loads,
			unloads,
			embeddings,
		)
	}
}

func TestResidencyPolicyDoesNotReleaseExternallyResidentModel(t *testing.T) {
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateLoaded,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true,
	})

	if _, err := runtime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}

	_, loads, unloads, _, _ := provider.counts()
	if loads != 0 || unloads != 0 {
		t.Fatalf("external residency was mutated: loads=%d unloads=%d", loads, unloads)
	}
}

func TestResidencyPolicyExpiresWithDeterministicKeepAlive(t *testing.T) {
	scheduler := &fakeResidencyScheduler{}
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
	}
	runtime := newResidencyRuntime(t, scheduler, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true, KeepAlive: 90 * time.Second,
	})

	if _, err := runtime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}

	timer := scheduler.latest(t)
	if timer.duration != 90*time.Second {
		t.Fatalf("unexpected keep-alive duration %s", timer.duration)
	}
	_, _, unloads, _, _ := provider.counts()
	if unloads != 0 {
		t.Fatalf("model unloaded before expiry: %d", unloads)
	}
	if !timer.Fire() {
		t.Fatal("expected timer to fire")
	}
	_, _, unloads, _, _ = provider.counts()
	if unloads != 1 {
		t.Fatalf("expected one unload after expiry, got %d", unloads)
	}
}

func TestResidencyPolicyKeepsStreamLeaseUntilTerminalEvent(t *testing.T) {
	providerStream := &testStream{chunks: []pkgProvider.StreamChunk{{Content: "hello"}}}
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
		stream:           providerStream,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true,
	})

	stream, err := runtime.Stream(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatalf("receive chunk: %v", err)
	}
	_, _, unloads, _, _ := provider.counts()
	if unloads != 0 {
		t.Fatalf("model unloaded while stream was active: %d", unloads)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}
	_, _, unloads, _, _ = provider.counts()
	if unloads != 1 {
		t.Fatalf("expected one unload after stream completion, got %d", unloads)
	}
}

func TestResidencyPolicyCoalescesConcurrentLoads(t *testing.T) {
	const operations = 8

	allowLoad := make(chan struct{})
	allowCompletion := make(chan struct{})
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
		loadStarted:      make(chan struct{}, 1),
		allowLoad:        allowLoad,
		completionEntry:  make(chan struct{}, operations),
		allowCompletion:  allowCompletion,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true,
	})

	errorsChannel := make(chan error, operations)
	for range operations {
		go func() {
			_, err := runtime.Complete(
				context.Background(),
				"test",
				pkgProvider.CompletionRequest{Model: "qwen"},
			)
			errorsChannel <- err
		}()
	}

	select {
	case <-provider.loadStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first load")
	}
	close(allowLoad)

	for range operations {
		select {
		case <-provider.completionEntry:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for concurrent completions")
		}
	}
	_, loads, unloads, _, _ := provider.counts()
	if loads != 1 || unloads != 0 {
		t.Fatalf("unexpected concurrent transitions: loads=%d unloads=%d", loads, unloads)
	}

	close(allowCompletion)
	for range operations {
		if err := <-errorsChannel; err != nil {
			t.Fatalf("complete: %v", err)
		}
	}
	_, loads, unloads, _, _ = provider.counts()
	if loads != 1 || unloads != 1 {
		t.Fatalf("expected one load and unload, got loads=%d unloads=%d", loads, unloads)
	}
}

func TestResidencyPolicyPersistentModelUnloadsAtShutdown(t *testing.T) {
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true, Persistent: true,
	})

	if _, err := runtime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}
	_, _, unloads, _, _ := provider.counts()
	if unloads != 0 {
		t.Fatalf("persistent model unloaded after request: %d", unloads)
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	_, _, unloads, _, _ = provider.counts()
	if unloads != 1 {
		t.Fatalf("expected one shutdown unload, got %d", unloads)
	}
}

func TestDisablingResidencyPolicyReleasesOwnedIdleModel(t *testing.T) {
	provider := &residencyProvider{
		identityProvider: identityProvider{id: "test"},
		state:            pkgProvider.ModelStateAvailable,
	}
	runtime := newResidencyRuntime(t, &fakeResidencyScheduler{}, provider)
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen", Autoload: true, Persistent: true,
	})

	if _, err := runtime.Complete(
		context.Background(),
		"test",
		pkgProvider.CompletionRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("complete: %v", err)
	}
	setResidencyPolicy(t, runtime, pkgProvider.ModelResidencyPolicy{
		Model: "qwen",
	})

	_, _, unloads, _, _ := provider.counts()
	if unloads != 1 {
		t.Fatalf("expected one unload after disabling policy, got %d", unloads)
	}
}

func TestResidencyPolicyValidationAndCapabilities(t *testing.T) {
	providerRuntime := newRuntimeWithResidencyScheduler("identity", &fakeResidencyScheduler{})
	if err := providerRuntime.Register(&identityProvider{id: "identity"}); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	invalidPolicies := []pkgProvider.ModelResidencyPolicy{
		{Autoload: true},
		{Model: " qwen", Autoload: true},
		{Model: "qwen", Autoload: true, KeepAlive: -time.Second},
		{Model: "qwen", Autoload: true, KeepAlive: time.Second, Persistent: true},
		{Model: "qwen", KeepAlive: time.Second},
		{Model: "qwen", Persistent: true},
	}
	for _, policy := range invalidPolicies {
		err := providerRuntime.SetModelResidencyPolicy(
			context.Background(),
			"identity",
			policy,
		)
		if !errors.Is(err, pkgProvider.ErrInvalidResidencyPolicy) {
			t.Fatalf("policy %#v: expected ErrInvalidResidencyPolicy, got %v", policy, err)
		}
	}

	err := providerRuntime.SetModelResidencyPolicy(
		context.Background(),
		"identity",
		pkgProvider.ModelResidencyPolicy{Model: "qwen", Autoload: true},
	)
	if !errors.Is(err, pkgProvider.ErrUnsupportedCapability) {
		t.Fatalf("expected ErrUnsupportedCapability, got %v", err)
	}

	if err := providerRuntime.SetModelResidencyPolicy(
		context.Background(),
		"identity",
		pkgProvider.ModelResidencyPolicy{Model: "qwen"},
	); err != nil {
		t.Fatalf("store disabled policy: %v", err)
	}

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	err = providerRuntime.SetModelResidencyPolicy(
		canceledContext,
		"identity",
		pkgProvider.ModelResidencyPolicy{Model: "qwen"},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
