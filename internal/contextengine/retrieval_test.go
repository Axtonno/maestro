package contextengine

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestLexicalRetrievalRanksDeterministicallyAndFilters(t *testing.T) {
	engine := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{
		"b.txt": "alpha",
		"a.txt": "alpha beta beta",
		"c.txt": "gamma",
	})
	query := retrievalQuery(t, "alpha beta", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 10,
	})
	results, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("retrieve lexical results: %v", err)
	}
	if len(results) != 2 || results[0].Path != "a.txt" || results[0].Score != 1 || results[1].Path != "b.txt" || results[1].Score != 0.5 {
		t.Fatalf("unexpected lexical ranking: %#v", results)
	}

	filtered := retrievalQuery(t, "alpha", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, Paths: []pkgContext.DocumentPath{"b.txt"}, TopK: 10,
	})
	results, err = engine.Retrieve(context.Background(), filtered)
	if err != nil || len(results) != 1 || results[0].Path != "b.txt" {
		t.Fatalf("unexpected filtered ranking: %#v err=%v", results, err)
	}
}

func TestStructuredRetrievalUsesSymbolRanges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "service.go", "package sample\ntype Service struct{}\nfunc (service *Service) Run() {}\n")
	engine := New()
	if _, err := engine.Index(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, pkgContext.DefaultScanPolicy())); err != nil {
		t.Fatalf("index Go workspace: %v", err)
	}
	query := retrievalQuery(t, "run method", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalStructured}, Symbol: "Run", TopK: 5,
	})
	results, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("retrieve structured result: %v", err)
	}
	if len(results) != 1 || results[0].ReasonCode != "symbol_exact" || results[0].Score != 1 {
		t.Fatalf("unexpected structured results: %#v", results)
	}
	snapshot, _ := engine.Snapshot("workspace")
	document, _ := snapshot.Document("service.go")
	selected := document.Content()[results[0].Range.Start:results[0].Range.End]
	if !strings.Contains(selected, "Run") {
		t.Fatalf("structured range does not contain symbol: %q", selected)
	}
}

func TestSemanticRetrievalUsesExplicitProviderAndValidatesVectors(t *testing.T) {
	runtime := &embeddingStub{embed: func(_ context.Context, provider pkgProvider.ID, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
		if provider != "fixture" || request.Model != "embed" || len(request.Inputs) != 3 {
			t.Fatalf("unexpected embedding request: provider=%q request=%#v", provider, request)
		}
		return pkgProvider.EmbeddingResponse{Model: "embed", Embeddings: [][]float32{{1, 0}, {0, 1}, {1, 0}}}, nil
	}}
	engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "irrelevant", "b.txt": "relevant"})
	query := semanticQuery(t)
	results, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("retrieve semantic results: %v", err)
	}
	if len(results) != 2 || results[0].Path != "b.txt" || results[0].Method != pkgContext.RetrievalSemantic {
		t.Fatalf("unexpected semantic ranking: %#v", results)
	}

	noRuntime := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{"a.txt": "text"})
	if _, err := noRuntime.Retrieve(context.Background(), semanticQuery(t)); !errors.Is(err, pkgContext.ErrUnsupported) {
		t.Fatalf("expected unsupported semantic retrieval, got %v", err)
	}
}

func TestSemanticRetrievalRejectsInvalidEmbeddingShapes(t *testing.T) {
	tests := []struct {
		name       string
		embeddings [][]float32
	}{
		{name: "cardinality", embeddings: [][]float32{{1, 0}}},
		{name: "dimension", embeddings: [][]float32{{1, 0}, {1}}},
		{name: "non finite", embeddings: [][]float32{{1, 0}, {float32(math.NaN()), 1}}},
		{name: "zero norm", embeddings: [][]float32{{1, 0}, {0, 0}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime := &embeddingStub{embed: func(context.Context, pkgProvider.ID, pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
				return pkgProvider.EmbeddingResponse{Embeddings: test.embeddings}, nil
			}}
			engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "text"})
			if _, err := engine.Retrieve(context.Background(), semanticQuery(t)); !errors.Is(err, pkgContext.ErrEmbeddingFailure) {
				t.Fatalf("expected embedding failure, got %v", err)
			}
		})
	}
}

func TestReciprocalRankFusionIsExplicitAndDeterministic(t *testing.T) {
	runtime := &embeddingStub{embed: func(context.Context, pkgProvider.ID, pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
		return pkgProvider.EmbeddingResponse{Embeddings: [][]float32{{1, 0}, {1, 0}, {0, 1}}}, nil
	}}
	engine := indexedTextEngine(t, runtime, map[pkgContext.DocumentPath]string{"a.txt": "alpha", "b.txt": "beta"})
	target := pkgContext.EmbeddingTarget{Provider: "fixture", Model: "embed"}
	query := retrievalQuery(t, "alpha", pkgContext.RetrievalQueryOptions{
		Methods:   []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical, pkgContext.RetrievalSemantic},
		Embedding: &target, Fusion: pkgContext.FusionReciprocalRank, TopK: 10,
	})
	results, err := engine.Retrieve(context.Background(), query)
	if err != nil {
		t.Fatalf("retrieve fused results: %v", err)
	}
	if len(results) != 2 || results[0].Path != "a.txt" || results[0].Method != pkgContext.RetrievalFused || results[0].ReasonCode != "reciprocal_rank_fusion" {
		t.Fatalf("unexpected fused ranking: %#v", results)
	}
}

func TestBuilderRespectsBudgetAndUTF8Boundaries(t *testing.T) {
	engine := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{
		"a.txt": "target ééé long evidence text",
		"b.txt": "target second evidence",
	})
	query := retrievalQuery(t, "target", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 10,
	})
	bundle, err := engine.Build(context.Background(), pkgContext.BuildRequest{
		Query: query, Budget: pkgContext.Budget{MaxTokens: 8, ReservedTokens: 1, SafetyTokens: 1}, Estimator: UTF8EstimatorID,
	})
	if err != nil {
		t.Fatalf("build context bundle: %v", err)
	}
	if bundle.UsedTokens() > bundle.Budget().EvidenceTokens() || len(bundle.Sections()) != 1 {
		t.Fatalf("bundle exceeded budget or selection: used=%d sections=%#v", bundle.UsedTokens(), bundle.Sections())
	}
	section := bundle.Sections()[0]
	if !section.Truncated || !utf8.ValidString(section.Text) || !strings.HasPrefix(section.Text, "target") {
		t.Fatalf("invalid truncated section: %#v", section)
	}
}

func TestEstimatorRegistrationFailureAndCallbacks(t *testing.T) {
	engine := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{"a.txt": "target evidence"})
	var typedNil *estimatorStub
	if err := engine.RegisterEstimator(typedNil); !errors.Is(err, pkgContext.ErrInvalidEstimator) {
		t.Fatalf("expected typed nil rejection, got %v", err)
	}
	if err := engine.RegisterEstimator(&estimatorStub{id: UTF8EstimatorID, version: "2"}); !errors.Is(err, pkgContext.ErrAlreadyRegistered) {
		t.Fatalf("expected duplicate estimator, got %v", err)
	}

	entered := make(chan struct{})
	release := make(chan struct{})
	blocking := &estimatorStub{id: "context.blocking-estimator", version: "1", cost: 1, entered: entered, release: release}
	if err := engine.RegisterEstimator(blocking); err != nil {
		t.Fatalf("register blocking estimator: %v", err)
	}
	query := retrievalQuery(t, "target", pkgContext.RetrievalQueryOptions{Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1})
	done := make(chan error, 1)
	go func() {
		_, err := engine.Build(context.Background(), pkgContext.BuildRequest{
			Query: query, Budget: pkgContext.Budget{MaxTokens: 10}, Estimator: blocking.id,
		})
		done <- err
	}()
	<-entered
	registered := make(chan error, 1)
	go func() {
		registered <- engine.RegisterEstimator(&estimatorStub{id: "context.parallel-estimator", version: "1", cost: 1})
	}()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatalf("register while estimator blocked: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("estimator callback held engine lock")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("build with blocking estimator: %v", err)
	}
}

func TestEstimatorErrorsAndPanicsAreClassified(t *testing.T) {
	engine := indexedTextEngine(t, nil, map[pkgContext.DocumentPath]string{"a.txt": "target evidence"})
	query := retrievalQuery(t, "target", pkgContext.RetrievalQueryOptions{Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1})
	for _, estimator := range []*estimatorStub{
		{id: "context.error-estimator", version: "1", err: errors.New("failed")},
		{id: "context.panic-estimator", version: "1", panicEstimate: true},
		{id: "context.zero-estimator", version: "1"},
	} {
		if err := engine.RegisterEstimator(estimator); err != nil {
			t.Fatalf("register estimator: %v", err)
		}
		_, err := engine.Build(context.Background(), pkgContext.BuildRequest{
			Query: query, Budget: pkgContext.Budget{MaxTokens: 10}, Estimator: estimator.id,
		})
		if !errors.Is(err, pkgContext.ErrEstimatorFailure) {
			t.Errorf("estimator %q: expected failure, got %v", estimator.id, err)
		}
	}
}

type embeddingStub struct {
	embed func(context.Context, pkgProvider.ID, pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error)
}

func (runtime *embeddingStub) Embed(ctx context.Context, id pkgProvider.ID, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
	return runtime.embed(ctx, id, request)
}

type estimatorStub struct {
	id            pkgContext.EstimatorID
	version       string
	cost          int
	err           error
	panicEstimate bool
	entered       chan<- struct{}
	release       <-chan struct{}
	once          sync.Once
}

func (estimator *estimatorStub) ID() pkgContext.EstimatorID { return estimator.id }
func (estimator *estimatorStub) Version() string            { return estimator.version }
func (estimator *estimatorStub) Estimate(ctx context.Context, _ string) (int, error) {
	if estimator.panicEstimate {
		panic("estimate panic")
	}
	if estimator.entered != nil {
		estimator.once.Do(func() { close(estimator.entered) })
		select {
		case <-estimator.release:
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	return estimator.cost, estimator.err
}

func indexedTextEngine(t *testing.T, embedding embeddingRuntime, contents map[pkgContext.DocumentPath]string) *Engine {
	t.Helper()
	documents := make([]pkgContext.Document, 0, len(contents))
	for path, content := range contents {
		document, err := pkgContext.NewDocument(path, "text/plain", "", content)
		if err != nil {
			t.Fatalf("construct document: %v", err)
		}
		documents = append(documents, document)
	}
	source := &mutableSource{id: "context.retrieval-source", result: pkgContext.ScanResult{Documents: documents}}
	engine := NewWithEmbeddingRuntime(embedding)
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	workspace, err := pkgContext.NewWorkspace("workspace", filepath.Clean(t.TempDir()), pkgContext.WorkspaceOptions{
		Source: source.id, Policy: pkgContext.DefaultScanPolicy(),
	})
	if err != nil {
		t.Fatalf("construct workspace: %v", err)
	}
	if _, err := engine.Index(context.Background(), workspace); err != nil {
		t.Fatalf("index workspace: %v", err)
	}
	return engine
}

func retrievalQuery(t *testing.T, text string, options pkgContext.RetrievalQueryOptions) pkgContext.RetrievalQuery {
	t.Helper()
	query, err := pkgContext.NewRetrievalQuery("workspace", text, options)
	if err != nil {
		t.Fatalf("construct retrieval query: %v", err)
	}
	return query
}

func semanticQuery(t *testing.T) pkgContext.RetrievalQuery {
	t.Helper()
	target := pkgContext.EmbeddingTarget{Provider: "fixture", Model: "embed"}
	return retrievalQuery(t, "relevant", pkgContext.RetrievalQueryOptions{
		Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalSemantic}, Embedding: &target, TopK: 10,
	})
}
