package provider

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type identityProvider struct {
	id pkgProvider.ID
}

func (p *identityProvider) ID() pkgProvider.ID {
	return p.id
}

type capableProvider struct {
	identityProvider
	completionResponse pkgProvider.CompletionResponse
	embeddingResponse  pkgProvider.EmbeddingResponse
	models             []pkgProvider.Model
	modelInfos         []pkgProvider.ModelInfo
	loadedModel        string
	unloadedModel      string
	pulledModel        string
	removedModel       string
	stream             pkgProvider.Stream
	pullStream         pkgProvider.ModelPullStream
	err                error
}

type reentrantProvider struct {
	identityProvider
	runtime pkgProvider.Runtime
}

func (p *reentrantProvider) Complete(
	_ context.Context,
	_ pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	err := p.runtime.Register(&identityProvider{id: "registered-from-provider"})

	return pkgProvider.CompletionResponse{}, err
}

func (p *capableProvider) Complete(
	_ context.Context,
	_ pkgProvider.CompletionRequest,
) (pkgProvider.CompletionResponse, error) {
	return p.completionResponse, p.err
}

func (p *capableProvider) Stream(
	_ context.Context,
	_ pkgProvider.CompletionRequest,
) (pkgProvider.Stream, error) {
	return p.stream, p.err
}

func (p *capableProvider) Embed(
	_ context.Context,
	_ pkgProvider.EmbeddingRequest,
) (pkgProvider.EmbeddingResponse, error) {
	return p.embeddingResponse, p.err
}

func (p *capableProvider) Models(
	_ context.Context,
) ([]pkgProvider.Model, error) {
	return p.models, p.err
}

func (p *capableProvider) DiscoverModels(
	_ context.Context,
) ([]pkgProvider.ModelInfo, error) {
	return p.modelInfos, p.err
}

func (p *capableProvider) LoadModel(
	_ context.Context,
	request pkgProvider.ModelLoadRequest,
) error {
	p.loadedModel = request.Model

	return p.err
}

func (p *capableProvider) UnloadModel(
	_ context.Context,
	request pkgProvider.ModelUnloadRequest,
) error {
	p.unloadedModel = request.Model

	return p.err
}

func (p *capableProvider) PullModel(
	_ context.Context,
	request pkgProvider.ModelPullRequest,
) (pkgProvider.ModelPullStream, error) {
	p.pulledModel = request.Model

	return p.pullStream, p.err
}

func (p *capableProvider) RemoveModel(
	_ context.Context,
	request pkgProvider.ModelRemoveRequest,
) error {
	p.removedModel = request.Model

	return p.err
}

type testStream struct {
	chunks []pkgProvider.StreamChunk
	index  int
	closed bool
}

type testModelPullStream struct {
	progress []pkgProvider.ModelPullProgress
	index    int
	closed   bool
}

func (s *testModelPullStream) Recv() (pkgProvider.ModelPullProgress, error) {
	if s.index == len(s.progress) {
		return pkgProvider.ModelPullProgress{}, io.EOF
	}

	progress := s.progress[s.index]
	s.index++

	return progress, nil
}

func (s *testModelPullStream) Close() error {
	s.closed = true

	return nil
}

func (s *testStream) Recv() (pkgProvider.StreamChunk, error) {
	if s.index == len(s.chunks) {
		return pkgProvider.StreamChunk{}, io.EOF
	}

	chunk := s.chunks[s.index]
	s.index++

	return chunk, nil
}

func (s *testStream) Close() error {
	s.closed = true

	return nil
}

func TestRuntimeRegistersAndResolvesProvider(t *testing.T) {
	providerRuntime := NewRuntime("")
	registered := &identityProvider{id: "ollama"}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	resolved, err := providerRuntime.Resolve("ollama")
	if err != nil {
		t.Fatalf("resolve provider: %v", err)
	}

	if resolved != registered {
		t.Fatal("resolved an unexpected provider")
	}

	if err := providerRuntime.Register(registered); !errors.Is(
		err,
		pkgProvider.ErrAlreadyRegistered,
	) {
		t.Fatalf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestRuntimeRejectsInvalidProvider(t *testing.T) {
	providerRuntime := NewRuntime("")
	var typedNil *identityProvider

	tests := []struct {
		name     string
		provider pkgProvider.Provider
	}{
		{name: "nil"},
		{name: "typed nil", provider: typedNil},
		{name: "empty ID", provider: &identityProvider{}},
		{name: "blank ID", provider: &identityProvider{id: "   "}},
		{name: "surrounding whitespace", provider: &identityProvider{id: " ollama"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := providerRuntime.Register(test.provider)
			if !errors.Is(err, pkgProvider.ErrInvalidProvider) {
				t.Fatalf("expected ErrInvalidProvider, got %v", err)
			}
		})
	}
}

func TestRuntimeUsesExplicitDefaultProvider(t *testing.T) {
	providerRuntime := NewRuntime("ollama")
	registered := &identityProvider{id: "ollama"}

	if _, err := providerRuntime.Default(); !errors.Is(
		err,
		pkgProvider.ErrNotFound,
	) {
		t.Fatalf("expected unresolved configured default, got %v", err)
	}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	resolved, err := providerRuntime.Default()
	if err != nil {
		t.Fatalf("resolve default provider: %v", err)
	}

	if resolved != registered {
		t.Fatal("resolved an unexpected default provider")
	}
}

func TestRuntimeRequiresExplicitDefaultProvider(t *testing.T) {
	providerRuntime := NewRuntime("")

	if err := providerRuntime.Register(
		&identityProvider{id: "ollama"},
	); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	if _, err := providerRuntime.Default(); !errors.Is(
		err,
		pkgProvider.ErrDefaultNotConfigured,
	) {
		t.Fatalf("expected ErrDefaultNotConfigured, got %v", err)
	}

	if err := providerRuntime.SetDefault("missing"); !errors.Is(
		err,
		pkgProvider.ErrNotFound,
	) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := providerRuntime.SetDefault("ollama"); err != nil {
		t.Fatalf("set default provider: %v", err)
	}

	if _, err := providerRuntime.Default(); err != nil {
		t.Fatalf("resolve default provider: %v", err)
	}
}

func TestRuntimeRoutesProviderCapabilities(t *testing.T) {
	stream := &testStream{chunks: []pkgProvider.StreamChunk{{Content: "hello"}}}
	pullStream := &testModelPullStream{progress: []pkgProvider.ModelPullProgress{{
		Model: "qwen", Stage: pkgProvider.ModelPullStageCompleted,
	}}}
	registered := &capableProvider{
		identityProvider: identityProvider{id: "ollama"},
		completionResponse: pkgProvider.CompletionResponse{
			Message: pkgProvider.Message{Content: "hello"},
		},
		embeddingResponse: pkgProvider.EmbeddingResponse{
			Embeddings: [][]float32{{1, 2}},
		},
		models: []pkgProvider.Model{{ID: "qwen"}},
		modelInfos: []pkgProvider.ModelInfo{{
			Model: pkgProvider.Model{ID: "qwen"},
			State: pkgProvider.ModelStateLoaded,
		}},
		stream:     stream,
		pullStream: pullStream,
	}
	providerRuntime := NewRuntime("ollama")

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	completion, err := providerRuntime.Complete(
		context.Background(),
		"",
		pkgProvider.CompletionRequest{},
	)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if completion.Message.Content != "hello" {
		t.Fatalf("unexpected completion: %#v", completion)
	}

	acquiredStream, err := providerRuntime.Stream(
		context.Background(),
		"ollama",
		pkgProvider.CompletionRequest{},
	)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	chunk, err := acquiredStream.Recv()
	if err != nil || chunk.Content != "hello" {
		t.Fatalf("unexpected stream chunk %#v, error %v", chunk, err)
	}
	if err := acquiredStream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}
	if !stream.closed {
		t.Fatal("stream was not closed")
	}

	embedding, err := providerRuntime.Embed(
		context.Background(),
		"ollama",
		pkgProvider.EmbeddingRequest{},
	)
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if !reflect.DeepEqual(embedding.Embeddings, [][]float32{{1, 2}}) {
		t.Fatalf("unexpected embedding: %#v", embedding)
	}

	models, err := providerRuntime.Models(context.Background(), "ollama")
	if err != nil {
		t.Fatalf("list models: %v", err)
	}
	if !reflect.DeepEqual(models, []pkgProvider.Model{{ID: "qwen"}}) {
		t.Fatalf("unexpected models: %#v", models)
	}

	modelInfos, err := providerRuntime.DiscoverModels(
		context.Background(),
		"ollama",
	)
	if err != nil {
		t.Fatalf("discover models: %v", err)
	}
	if !reflect.DeepEqual(modelInfos, registered.modelInfos) {
		t.Fatalf("unexpected model info: %#v", modelInfos)
	}

	if err := providerRuntime.LoadModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelLoadRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("load model: %v", err)
	}
	if registered.loadedModel != "qwen" {
		t.Fatalf("unexpected loaded model %q", registered.loadedModel)
	}

	if err := providerRuntime.UnloadModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelUnloadRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("unload model: %v", err)
	}
	if registered.unloadedModel != "qwen" {
		t.Fatalf("unexpected unloaded model %q", registered.unloadedModel)
	}

	acquiredPullStream, err := providerRuntime.PullModel(
		context.Background(),
		"",
		pkgProvider.ModelPullRequest{Model: "qwen"},
	)
	if err != nil {
		t.Fatalf("pull model: %v", err)
	}
	progress, err := acquiredPullStream.Recv()
	if err != nil || progress.Stage != pkgProvider.ModelPullStageCompleted {
		t.Fatalf("unexpected pull progress %#v, error %v", progress, err)
	}
	if err := acquiredPullStream.Close(); err != nil {
		t.Fatalf("close model pull stream: %v", err)
	}
	if !pullStream.closed || registered.pulledModel != "qwen" {
		t.Fatalf("model pull was not routed correctly")
	}

	if err := providerRuntime.RemoveModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelRemoveRequest{Model: "qwen"},
	); err != nil {
		t.Fatalf("remove model: %v", err)
	}
	if registered.removedModel != "qwen" {
		t.Fatalf("unexpected removed model %q", registered.removedModel)
	}
}

func TestRuntimeRejectsUnsupportedCapabilities(t *testing.T) {
	providerRuntime := NewRuntime("")
	if err := providerRuntime.Register(
		&identityProvider{id: "identity-only"},
	); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	assertUnsupported := func(name string, err error) {
		t.Helper()

		if !errors.Is(err, pkgProvider.ErrUnsupportedCapability) {
			t.Fatalf("%s: expected ErrUnsupportedCapability, got %v", name, err)
		}
	}

	_, err := providerRuntime.Complete(
		context.Background(),
		"identity-only",
		pkgProvider.CompletionRequest{},
	)
	assertUnsupported("complete", err)

	_, err = providerRuntime.Stream(
		context.Background(),
		"identity-only",
		pkgProvider.CompletionRequest{},
	)
	assertUnsupported("stream", err)

	_, err = providerRuntime.Embed(
		context.Background(),
		"identity-only",
		pkgProvider.EmbeddingRequest{},
	)
	assertUnsupported("embed", err)

	_, err = providerRuntime.Models(context.Background(), "identity-only")
	assertUnsupported("models", err)

	_, err = providerRuntime.DiscoverModels(
		context.Background(),
		"identity-only",
	)
	assertUnsupported("model discovery", err)

	err = providerRuntime.LoadModel(
		context.Background(),
		"identity-only",
		pkgProvider.ModelLoadRequest{},
	)
	assertUnsupported("model loading", err)

	err = providerRuntime.UnloadModel(
		context.Background(),
		"identity-only",
		pkgProvider.ModelUnloadRequest{},
	)
	assertUnsupported("model unloading", err)

	_, err = providerRuntime.PullModel(
		context.Background(),
		"identity-only",
		pkgProvider.ModelPullRequest{},
	)
	assertUnsupported("model pulling", err)

	err = providerRuntime.RemoveModel(
		context.Background(),
		"identity-only",
		pkgProvider.ModelRemoveRequest{},
	)
	assertUnsupported("model removal", err)
}

func TestRuntimeRejectsNilStream(t *testing.T) {
	providerRuntime := NewRuntime("")
	registered := &capableProvider{
		identityProvider: identityProvider{id: "ollama"},
	}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	_, err := providerRuntime.Stream(
		context.Background(),
		"ollama",
		pkgProvider.CompletionRequest{},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidStream) {
		t.Fatalf("expected ErrInvalidStream, got %v", err)
	}
}

func TestRuntimeRejectsNilModelPullStream(t *testing.T) {
	providerRuntime := NewRuntime("")
	registered := &capableProvider{
		identityProvider: identityProvider{id: "ollama"},
	}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	_, err := providerRuntime.PullModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelPullRequest{},
	)
	if !errors.Is(err, pkgProvider.ErrInvalidStream) {
		t.Fatalf("expected ErrInvalidStream, got %v", err)
	}
}

func TestRuntimePreservesProviderError(t *testing.T) {
	cause := errors.New("transport failed")
	providerRuntime := NewRuntime("")
	registered := &capableProvider{
		identityProvider: identityProvider{id: "ollama"},
		err:              cause,
	}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	_, err := providerRuntime.Complete(
		context.Background(),
		"ollama",
		pkgProvider.CompletionRequest{},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("expected provider cause, got %v", err)
	}

	_, err = providerRuntime.DiscoverModels(context.Background(), "ollama")
	if !errors.Is(err, cause) {
		t.Fatalf("expected discovery provider cause, got %v", err)
	}

	err = providerRuntime.LoadModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelLoadRequest{},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("expected load provider cause, got %v", err)
	}

	err = providerRuntime.UnloadModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelUnloadRequest{},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("expected unload provider cause, got %v", err)
	}

	_, err = providerRuntime.PullModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelPullRequest{},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("expected pull provider cause, got %v", err)
	}

	err = providerRuntime.RemoveModel(
		context.Background(),
		"ollama",
		pkgProvider.ModelRemoveRequest{},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("expected removal provider cause, got %v", err)
	}
}

func TestRuntimeDoesNotHoldLockWhileCallingProvider(t *testing.T) {
	providerRuntime := NewRuntime("")
	registered := &reentrantProvider{
		identityProvider: identityProvider{id: "reentrant"},
		runtime:          providerRuntime,
	}

	if err := providerRuntime.Register(registered); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	completed := make(chan error, 1)
	go func() {
		_, err := providerRuntime.Complete(
			context.Background(),
			"reentrant",
			pkgProvider.CompletionRequest{},
		)
		completed <- err
	}()

	select {
	case err := <-completed:
		if err != nil {
			t.Fatalf("complete through reentrant provider: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("provider call blocked while reentering the runtime")
	}

	if _, err := providerRuntime.Resolve("registered-from-provider"); err != nil {
		t.Fatalf("resolve provider registered during callback: %v", err)
	}
}

func TestRuntimeSupportsConcurrentRegistrationAndResolution(t *testing.T) {
	providerRuntime := NewRuntime("")
	const providerCount = 32

	var registrations sync.WaitGroup
	for index := range providerCount {
		registrations.Add(1)

		go func() {
			defer registrations.Done()

			providerID := pkgProvider.ID("provider-" + strconv.Itoa(index))
			if err := providerRuntime.Register(
				&identityProvider{id: providerID},
			); err != nil {
				t.Errorf("register %q: %v", providerID, err)
			}
		}()
	}
	registrations.Wait()

	var resolutions sync.WaitGroup
	for index := range providerCount {
		resolutions.Add(1)

		go func() {
			defer resolutions.Done()

			providerID := pkgProvider.ID("provider-" + strconv.Itoa(index))
			if _, err := providerRuntime.Resolve(providerID); err != nil {
				t.Errorf("resolve %q: %v", providerID, err)
			}
		}()
	}
	resolutions.Wait()
}
