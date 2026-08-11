package maestro_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/antonio-cafeo/maestro"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestContextEngineIsComposedWithProviderRuntime(t *testing.T) {
	runtime := maestro.New()
	provider := &contextEmbeddingProvider{id: "fixture"}
	if err := runtime.Providers().Register(provider); err != nil {
		t.Fatalf("register embedding provider: %v", err)
	}
	root := t.TempDir()
	writeContextFile(t, root, "a.txt", "relevant evidence")
	workspace := contextWorkspace(t, "provider-workspace", root)
	if _, err := runtime.ContextEngine().Index(t.Context(), workspace); err != nil {
		t.Fatalf("index workspace: %v", err)
	}
	target := pkgContext.EmbeddingTarget{Provider: "fixture", Model: "embed"}
	query, err := pkgContext.NewRetrievalQuery(workspace.ID(), "relevant", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalSemantic}, TopK: 1, Embedding: &target,
	})
	if err != nil {
		t.Fatalf("construct query: %v", err)
	}
	results, err := runtime.ContextEngine().Retrieve(t.Context(), query)
	if err != nil {
		t.Fatalf("retrieve through composed provider: %v", err)
	}
	if len(results) != 1 || results[0].Path != "a.txt" || provider.calls != 1 {
		t.Fatalf("unexpected semantic result or provider calls: %#v calls=%d", results, provider.calls)
	}

	other := maestro.New()
	if runtime.ContextEngine() == other.ContextEngine() {
		t.Fatal("composition roots share a Context Engine")
	}
}

func TestContextEnginePublicPathIndexesAnalyzesRetrievesAndBuilds(t *testing.T) {
	runtime := maestro.New()
	root := t.TempDir()
	writeContextFile(t, root, "service.go", "package service\ntype Runner struct{}\nfunc (Runner) Run() {}\n")
	workspace := contextWorkspace(t, "public-path", root)

	snapshot, err := runtime.ContextEngine().Index(t.Context(), workspace)
	if err != nil {
		t.Fatalf("index public workspace: %v", err)
	}
	if snapshot.Metadata().DocumentCount != 1 || snapshot.Metadata().AnalysisCount != 1 {
		t.Fatalf("unexpected analyzed snapshot: %#v", snapshot.Metadata())
	}
	query, err := pkgContext.NewRetrievalQuery(workspace.ID(), "run method", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalStructured},
		Symbol:  "Run",
		TopK:    1,
	})
	if err != nil {
		t.Fatalf("construct structured query: %v", err)
	}
	results, err := runtime.ContextEngine().Retrieve(t.Context(), query)
	if err != nil || len(results) != 1 || results[0].ReasonCode != "symbol_exact" {
		t.Fatalf("unexpected structured retrieval: results=%#v err=%v", results, err)
	}
	bundle, err := runtime.ContextEngine().Build(t.Context(), pkgContext.BuildRequest{
		Query: query, Budget: pkgContext.Budget{MaxTokens: 30}, Estimator: "context.utf8-estimator",
	})
	if err != nil {
		t.Fatalf("build public context bundle: %v", err)
	}
	if sections := bundle.Sections(); len(sections) != 1 || sections[0].Path != "service.go" || !strings.Contains(sections[0].Text, "Run") {
		t.Fatalf("unexpected public context bundle: %#v", sections)
	}
}

func TestContextEventsAreOrderedRedactedReentrantAndBestEffort(t *testing.T) {
	runtime := maestro.New()
	root := t.TempDir()
	writeContextFile(t, root, "secret.txt", "private query marker")
	workspace := contextWorkspace(t, "events", root)

	var mu sync.Mutex
	topics := make([]string, 0)
	payloads := make([]pkgContext.EventPayload, 0)
	recorder := func(event pkgRuntime.Event) {
		payload, ok := event.Payload().(pkgContext.EventPayload)
		if !ok {
			t.Errorf("unexpected context payload %T", event.Payload())
			return
		}
		_ = runtime.ContextEngine().CacheStats()
		mu.Lock()
		topics = append(topics, event.Name())
		payloads = append(payloads, payload)
		mu.Unlock()
	}
	for _, topic := range []string{
		pkgContext.EventIndexStarted,
		pkgContext.EventIndexCompleted,
		pkgContext.EventCacheObserved,
		pkgContext.EventBuildStarted,
		pkgContext.EventBuildCompleted,
		pkgContext.EventBuildFailed,
	} {
		if err := runtime.EventBus().Subscribe(topic, recorder); err != nil {
			t.Fatalf("subscribe %q: %v", topic, err)
		}
	}
	if err := runtime.EventBus().Subscribe(pkgContext.EventIndexCompleted, func(pkgRuntime.Event) {
		panic("observer panic")
	}); err != nil {
		t.Fatalf("subscribe panic observer: %v", err)
	}

	if _, err := runtime.ContextEngine().Index(t.Context(), workspace); err != nil {
		t.Fatalf("index with panic observer: %v", err)
	}
	query, err := pkgContext.NewRetrievalQuery(workspace.ID(), "private query marker", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1,
	})
	if err != nil {
		t.Fatalf("construct build query: %v", err)
	}
	if _, err := runtime.ContextEngine().Build(t.Context(), pkgContext.BuildRequest{
		Query: query, Budget: pkgContext.Budget{MaxTokens: 20}, Estimator: "context.utf8-estimator",
	}); err != nil {
		t.Fatalf("build context with observers: %v", err)
	}
	missingQuery, err := pkgContext.NewRetrievalQuery("missing", "private query marker", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1,
	})
	if err != nil {
		t.Fatalf("construct missing query: %v", err)
	}
	if _, err := runtime.ContextEngine().Build(t.Context(), pkgContext.BuildRequest{
		Query: missingQuery, Budget: pkgContext.Budget{MaxTokens: 20}, Estimator: "context.utf8-estimator",
	}); !errors.Is(err, pkgContext.ErrNotFound) {
		t.Fatalf("expected missing snapshot, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	wantPrefix := []string{
		pkgContext.EventIndexStarted, pkgContext.EventIndexCompleted,
		pkgContext.EventCacheObserved, pkgContext.EventBuildStarted,
		pkgContext.EventBuildCompleted, pkgContext.EventCacheObserved, pkgContext.EventBuildStarted,
		pkgContext.EventBuildFailed,
	}
	// The panic observer stops only the current Event Bus delivery. The committed
	// operation and the following cache-summary publication remain successful.
	if !reflect.DeepEqual(topics, wantPrefix) {
		t.Fatalf("unexpected context event order: got=%v want=%v", topics, wantPrefix)
	}
	if payloads[len(payloads)-1].Failure != pkgContext.EventFailureNotFound {
		t.Fatalf("unexpected failure classification: %#v", payloads[len(payloads)-1])
	}
	encoded, err := json.Marshal(payloads)
	if err != nil {
		t.Fatalf("encode event payloads: %v", err)
	}
	for _, private := range []string{root, "secret.txt", "private query marker", "fixture", "embed"} {
		if strings.Contains(string(encoded), private) {
			t.Fatalf("event payload leaked %q: %s", private, encoded)
		}
	}
}

func TestContextIndexCancellationPublishesFailureOnly(t *testing.T) {
	runtime := maestro.New()
	workspace := contextWorkspace(t, "canceled-events", t.TempDir())

	var mu sync.Mutex
	topics := make([]string, 0, 2)
	var failure pkgContext.EventFailure
	for _, topic := range []string{
		pkgContext.EventIndexStarted,
		pkgContext.EventIndexCompleted,
		pkgContext.EventIndexFailed,
	} {
		if err := runtime.EventBus().Subscribe(topic, func(event pkgRuntime.Event) {
			payload := event.Payload().(pkgContext.EventPayload)
			mu.Lock()
			topics = append(topics, event.Name())
			failure = payload.Failure
			mu.Unlock()
		}); err != nil {
			t.Fatalf("subscribe %q: %v", topic, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runtime.ContextEngine().Index(ctx, workspace); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled index, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if want := []string{pkgContext.EventIndexStarted, pkgContext.EventIndexFailed}; !reflect.DeepEqual(topics, want) {
		t.Fatalf("unexpected cancellation events: got=%v want=%v", topics, want)
	}
	if failure != pkgContext.EventFailureCanceled {
		t.Fatalf("unexpected cancellation classification: %q", failure)
	}
}

func TestContextObserverIsSynchronousAndRunsAfterSnapshotCommit(t *testing.T) {
	runtime := maestro.New()
	root := t.TempDir()
	writeContextFile(t, root, "a.txt", "evidence")
	workspace := contextWorkspace(t, "slow-observer", root)
	entered := make(chan struct{})
	release := make(chan struct{})
	if err := runtime.EventBus().Subscribe(pkgContext.EventIndexCompleted, func(pkgRuntime.Event) {
		if _, ok := runtime.ContextEngine().Snapshot(workspace.ID()); !ok {
			t.Error("snapshot is not committed during completion event")
		}
		close(entered)
		<-release
	}); err != nil {
		t.Fatalf("subscribe blocking observer: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := runtime.ContextEngine().Index(t.Context(), workspace)
		done <- err
	}()
	<-entered
	select {
	case err := <-done:
		t.Fatalf("index returned before synchronous observer release: %v", err)
	default:
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("index changed by blocking observer: %v", err)
	}
}

type contextEmbeddingProvider struct {
	id    pkgProvider.ID
	calls int
}

func (provider *contextEmbeddingProvider) ID() pkgProvider.ID { return provider.id }
func (provider *contextEmbeddingProvider) Embed(_ context.Context, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
	provider.calls++
	embeddings := make([][]float32, len(request.Inputs))
	for index := range embeddings {
		embeddings[index] = []float32{1, 0}
	}
	return pkgProvider.EmbeddingResponse{Model: request.Model, Embeddings: embeddings}, nil
}

func contextWorkspace(t *testing.T, id pkgContext.WorkspaceID, root string) pkgContext.Workspace {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace(id, filepath.Clean(root), pkgContext.WorkspaceOptions{
		Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy(),
	})
	if err != nil {
		t.Fatalf("construct context workspace: %v", err)
	}
	return workspace
}

func writeContextFile(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write context fixture: %v", err)
	}
}
