package contextengine

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestAnalysisCacheInvalidatesOnRenameAndReusesStablePath(t *testing.T) {
	root := t.TempDir()
	first := document(t, "first.go", "package sample\n")
	second := document(t, "renamed.go", "package sample\n")
	source := &mutableSource{id: "context.cache-source", result: pkgContext.ScanResult{Documents: []pkgContext.Document{first}}}
	analyzer := &countingAnalyzer{id: "context.counting-analyzer", version: "1"}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := engine.RegisterAnalyzer(analyzer); err != nil {
		t.Fatalf("register analyzer: %v", err)
	}
	workspace := workspaceWithAnalyzers(t, root, source.id, []pkgContext.AnalyzerID{analyzer.id})
	if _, err := engine.Index(context.Background(), workspace); err != nil {
		t.Fatalf("cold index: %v", err)
	}
	source.set(pkgContext.ScanResult{Documents: []pkgContext.Document{second}}, nil)
	snapshot, err := engine.Index(context.Background(), workspace)
	if err != nil {
		t.Fatalf("warm index: %v", err)
	}
	if analyzer.calls.Load() != 2 {
		t.Fatalf("analyzer calls=%d, rename must invalidate path-dependent analysis", analyzer.calls.Load())
	}
	if len(snapshot.Analyses()) != 1 || snapshot.Analyses()[0].Path() != "renamed.go" {
		t.Fatalf("renamed analysis has wrong provenance: %#v", snapshot.Analyses())
	}
	if _, err := engine.Index(context.Background(), workspace); err != nil {
		t.Fatalf("stable warm index: %v", err)
	}
	if analyzer.calls.Load() != 2 {
		t.Fatalf("stable content was not reused, calls=%d", analyzer.calls.Load())
	}
	stats := engine.CacheStats()
	if stats.Hits == 0 || stats.Misses == 0 {
		t.Fatalf("cache did not record cold/warm access: %#v", stats)
	}
}

func TestSemanticEmbeddingCacheAvoidsRepeatedProviderCalls(t *testing.T) {
	var calls atomic.Int64
	runtime := &embeddingStub{embed: func(context.Context, pkgProvider.ID, pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
		calls.Add(1)
		return pkgProvider.EmbeddingResponse{Embeddings: [][]float32{{1, 0}, {1, 0}}}, nil
	}}
	engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "relevant"})
	query := semanticQuery(t)
	first, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("cold semantic retrieval: %v", err)
	}
	second, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("warm semantic retrieval: %v", err)
	}
	if calls.Load() != 1 || !reflect.DeepEqual(first, second) {
		t.Fatalf("cache changed output or repeated provider call: calls=%d first=%#v second=%#v", calls.Load(), first, second)
	}
}

func TestEmbeddingDimensionChangePurgesTargetAndFailsCurrentRequest(t *testing.T) {
	var calls atomic.Int64
	runtime := &embeddingStub{embed: func(_ context.Context, _ pkgProvider.ID, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
		call := calls.Add(1)
		dimension := 2
		if call >= 2 {
			dimension = 3
		}
		embeddings := make([][]float32, len(request.Inputs))
		for index := range embeddings {
			embeddings[index] = make([]float32, dimension)
			embeddings[index][0] = 1
		}
		return pkgProvider.EmbeddingResponse{Embeddings: embeddings}, nil
	}}
	engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "relevant"})
	if _, err := engine.Retrieve(context.Background(), semanticQuery(t)); err != nil {
		t.Fatalf("initial semantic retrieval: %v", err)
	}
	target := pkgContext.EmbeddingTarget{Provider: "fixture", Model: "embed"}
	changedQuery := retrievalQuery(t, "different query", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalSemantic}, Embedding: &target, TopK: 10,
	})
	if _, err := engine.Retrieve(context.Background(), changedQuery); !errors.Is(err, pkgContext.ErrEmbeddingFailure) {
		t.Fatalf("expected dimension-change failure, got %v", err)
	}
	if _, err := engine.Retrieve(context.Background(), changedQuery); err != nil {
		t.Fatalf("retry after purge: %v", err)
	}
	if engine.CacheStats().Evictions < 1 {
		t.Fatalf("dimension change did not evict target entries: %#v", engine.CacheStats())
	}
}

func TestEstimatorCacheProducesEquivalentWarmBundle(t *testing.T) {
	engine := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{"a.txt": "target evidence"})
	estimator := &countingEstimator{id: "context.counting-estimator", version: "1"}
	if err := engine.RegisterEstimator(estimator); err != nil {
		t.Fatalf("register estimator: %v", err)
	}
	request := pkgContext.BuildRequest{
		Query:  retrievalQuery(t, "target", pkgContext.RetrievalQueryOptions{Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1}),
		Budget: pkgContext.Budget{MaxTokens: 100}, Estimator: estimator.id,
	}
	first, err := engine.Build(context.Background(), request)
	if err != nil {
		t.Fatalf("cold build: %v", err)
	}
	calls := estimator.calls.Load()
	second, err := engine.Build(context.Background(), request)
	if err != nil {
		t.Fatalf("warm build: %v", err)
	}
	if estimator.calls.Load() != calls || !reflect.DeepEqual(first.Sections(), second.Sections()) || first.UsedTokens() != second.UsedTokens() {
		t.Fatalf("warm bundle differs or estimator repeated: calls=%d/%d", calls, estimator.calls.Load())
	}
}

func TestCacheEvictionIsBoundedAndDeterministic(t *testing.T) {
	cache := newArtifactCache(pkgContext.CachePolicy{MaxEntries: 2, MaxBytes: 200})
	cache.put("a", 1, 50)
	cache.put("b", 2, 50)
	if _, found := cache.get("a"); !found {
		t.Fatal("expected cache hit")
	}
	cache.put("c", 3, 50)
	if _, found := cache.get("b"); found {
		t.Fatal("least recently used entry was retained")
	}
	if _, found := cache.get("a"); !found {
		t.Fatal("recent entry was evicted")
	}
	stats := cache.stats()
	if stats.Entries != 2 || stats.Bytes > 200 || stats.Evictions != 1 {
		t.Fatalf("unexpected bounded cache stats: %#v", stats)
	}
	cache.put("oversized", 4, 201)
	if cache.stats().Entries != 2 {
		t.Fatal("oversized entry entered the cache")
	}
}

func TestConcurrentColdSemanticRequestsAreIndependent(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	var calls atomic.Int64
	runtime := &embeddingStub{embed: func(ctx context.Context, _ pkgProvider.ID, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
		calls.Add(1)
		entered <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			return pkgProvider.EmbeddingResponse{}, ctx.Err()
		}
		embeddings := make([][]float32, len(request.Inputs))
		for index := range embeddings {
			embeddings[index] = []float32{1, 0}
		}
		return pkgProvider.EmbeddingResponse{Embeddings: embeddings}, nil
	}}
	engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "relevant"})
	query := semanticQuery(t)
	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := engine.Retrieve(context.Background(), query)
			results <- err
		}()
	}
	<-entered
	<-entered
	close(release)
	for range 2 {
		if err := <-results; err != nil {
			t.Fatalf("concurrent semantic retrieval: %v", err)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("expected independent cold requests, calls=%d", calls.Load())
	}
}

func TestCachePolicyValidation(t *testing.T) {
	if _, err := NewWithOptions(Options{}); !errors.Is(err, pkgContext.ErrInvalidCachePolicy) {
		t.Fatalf("expected invalid cache policy, got %v", err)
	}
	if _, err := NewWithOptions(Options{Cache: pkgContext.CachePolicy{MaxEntries: 1, MaxBytes: 1}}); err != nil {
		t.Fatalf("construct engine with minimal cache: %v", err)
	}
}

type countingAnalyzer struct {
	id      pkgContext.AnalyzerID
	version string
	calls   atomic.Int64
}

func (analyzer *countingAnalyzer) ID() pkgContext.AnalyzerID { return analyzer.id }
func (analyzer *countingAnalyzer) Version() string           { return analyzer.version }
func (*countingAnalyzer) Supports(pkgContext.Document) bool  { return true }
func (analyzer *countingAnalyzer) Analyze(_ context.Context, document pkgContext.Document) (pkgContext.Analysis, error) {
	analyzer.calls.Add(1)
	return pkgContext.NewAnalysis(document, analyzer.id, analyzer.version, nil, nil, nil, nil)
}

type countingEstimator struct {
	id      pkgContext.EstimatorID
	version string
	calls   atomic.Int64
}

func (estimator *countingEstimator) ID() pkgContext.EstimatorID { return estimator.id }
func (estimator *countingEstimator) Version() string            { return estimator.version }
func (estimator *countingEstimator) Estimate(context.Context, string) (int, error) {
	estimator.calls.Add(1)
	return 5, nil
}
