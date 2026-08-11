package contextengine

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

var _ pkgContext.Engine = (*Engine)(nil)

type Engine struct {
	mu         sync.RWMutex
	sources    map[pkgContext.SourceID]pkgContext.Source
	analyzers  map[pkgContext.AnalyzerID]pkgContext.Analyzer
	estimators map[pkgContext.EstimatorID]pkgContext.TokenEstimator
	embedding  embeddingRuntime
	cache      *artifactCache
	snapshots  map[pkgContext.WorkspaceID]pkgContext.Snapshot
}

type Options struct {
	Embedding embeddingRuntime
	Cache     pkgContext.CachePolicy
}

func New() *Engine {
	return NewWithEmbeddingRuntime(nil)
}

func NewWithEmbeddingRuntime(embedding embeddingRuntime) *Engine {
	engine, err := NewWithOptions(Options{Embedding: embedding, Cache: pkgContext.DefaultCachePolicy()})
	if err != nil {
		panic(err)
	}
	return engine
}

func NewWithOptions(options Options) (*Engine, error) {
	if err := options.Cache.Validate(); err != nil {
		return nil, err
	}
	filesystem := NewFilesystemSource()
	goAnalyzer := NewGoAnalyzer()
	estimator := NewUTF8Estimator()
	return &Engine{
		sources:    map[pkgContext.SourceID]pkgContext.Source{filesystem.ID(): filesystem},
		analyzers:  map[pkgContext.AnalyzerID]pkgContext.Analyzer{goAnalyzer.ID(): goAnalyzer},
		estimators: map[pkgContext.EstimatorID]pkgContext.TokenEstimator{estimator.ID(): estimator},
		embedding:  options.Embedding,
		cache:      newArtifactCache(options.Cache),
		snapshots:  make(map[pkgContext.WorkspaceID]pkgContext.Snapshot),
	}, nil
}

func (engine *Engine) CacheStats() pkgContext.CacheStats { return engine.cache.stats() }

func (engine *Engine) RegisterSource(source pkgContext.Source) error {
	if nilInterface(source) {
		return fmt.Errorf("register context source: %w", pkgContext.ErrInvalidSource)
	}
	id := source.ID()
	if err := id.Validate(); err != nil {
		return fmt.Errorf("register context source: %w: %w", err, pkgContext.ErrInvalidSource)
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if _, exists := engine.sources[id]; exists {
		return fmt.Errorf("register context source %q: %w", id, pkgContext.ErrAlreadyRegistered)
	}
	engine.sources[id] = source
	return nil
}

func (engine *Engine) RegisterAnalyzer(analyzer pkgContext.Analyzer) error {
	if nilInterface(analyzer) {
		return fmt.Errorf("register context analyzer: %w", pkgContext.ErrInvalidAnalyzer)
	}
	id := analyzer.ID()
	if err := id.Validate(); err != nil {
		return fmt.Errorf("register context analyzer: %w: %w", err, pkgContext.ErrInvalidAnalyzer)
	}
	if !exactVersion(analyzer.Version()) {
		return fmt.Errorf("register context analyzer %q with invalid version: %w", id, pkgContext.ErrInvalidAnalyzer)
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if _, exists := engine.analyzers[id]; exists {
		return fmt.Errorf("register context analyzer %q: %w", id, pkgContext.ErrAlreadyRegistered)
	}
	engine.analyzers[id] = analyzer
	return nil
}

func (engine *Engine) RegisterEstimator(estimator pkgContext.TokenEstimator) error {
	if nilInterface(estimator) {
		return fmt.Errorf("register token estimator: %w", pkgContext.ErrInvalidEstimator)
	}
	id := estimator.ID()
	if err := id.Validate(); err != nil {
		return fmt.Errorf("register token estimator: %w: %w", err, pkgContext.ErrInvalidEstimator)
	}
	if !exactVersion(estimator.Version()) {
		return fmt.Errorf("register token estimator %q with invalid version: %w", id, pkgContext.ErrInvalidEstimator)
	}
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if _, exists := engine.estimators[id]; exists {
		return fmt.Errorf("register token estimator %q: %w", id, pkgContext.ErrAlreadyRegistered)
	}
	engine.estimators[id] = estimator
	return nil
}

func (engine *Engine) Index(ctx context.Context, workspace pkgContext.Workspace) (pkgContext.Snapshot, error) {
	if ctx == nil {
		return pkgContext.Snapshot{}, fmt.Errorf("index workspace with nil context: %w", pkgContext.ErrInvalidWorkspace)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.Snapshot{}, err
	}
	if err := workspace.Validate(); err != nil {
		return pkgContext.Snapshot{}, err
	}
	engine.mu.RLock()
	source, exists := engine.sources[workspace.Source()]
	analyzers, analyzerErr := engine.resolveAnalyzersLocked(workspace.Analyzers())
	engine.mu.RUnlock()
	if !exists {
		return pkgContext.Snapshot{}, fmt.Errorf("context source %q: %w", workspace.Source(), pkgContext.ErrNotFound)
	}
	if analyzerErr != nil {
		return pkgContext.Snapshot{}, analyzerErr
	}

	result, err := source.Scan(ctx, workspace)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return pkgContext.Snapshot{}, ctxErr
		}
		return pkgContext.Snapshot{}, fmt.Errorf("scan workspace %q with source %q: %w: %w", workspace.ID(), workspace.Source(), err, pkgContext.ErrSourceFailure)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.Snapshot{}, err
	}
	analyses, err := analyzeDocuments(ctx, result.Documents, analyzers, len(workspace.Analyzers()) > 0, engine.cache)
	if err != nil {
		return pkgContext.Snapshot{}, err
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()
	generation := uint64(1)
	if current, found := engine.snapshots[workspace.ID()]; found {
		generation = current.Metadata().Generation + 1
	}
	snapshot, err := pkgContext.NewSnapshot(workspace, generation, result.Documents, analyses, result.Diagnostics)
	if err != nil {
		return pkgContext.Snapshot{}, fmt.Errorf("publish workspace %q snapshot: %w", workspace.ID(), err)
	}
	engine.snapshots[workspace.ID()] = snapshot
	cacheAnalyses(engine.cache, snapshot, analyzers)
	return snapshot, nil
}

func (engine *Engine) resolveAnalyzersLocked(configured []pkgContext.AnalyzerID) ([]pkgContext.Analyzer, error) {
	if len(configured) > 0 {
		resolved := make([]pkgContext.Analyzer, 0, len(configured))
		for _, id := range configured {
			analyzer, exists := engine.analyzers[id]
			if !exists {
				return nil, fmt.Errorf("context analyzer %q: %w", id, pkgContext.ErrNotFound)
			}
			resolved = append(resolved, analyzer)
		}
		return resolved, nil
	}
	ids := make([]pkgContext.AnalyzerID, 0, len(engine.analyzers))
	for id := range engine.analyzers {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	resolved := make([]pkgContext.Analyzer, 0, len(ids))
	for _, id := range ids {
		resolved = append(resolved, engine.analyzers[id])
	}
	return resolved, nil
}

func analyzeDocuments(ctx context.Context, documents []pkgContext.Document, analyzers []pkgContext.Analyzer, explicit bool, cache *artifactCache) ([]pkgContext.Analysis, error) {
	analyses := make([]pkgContext.Analysis, 0)
	for _, document := range documents {
		if document.MediaType() == "application/octet-stream" {
			continue
		}
		matching := make([]pkgContext.Analyzer, 0, len(analyzers))
		for _, analyzer := range analyzers {
			supported, err := supportsSafely(analyzer, document)
			if err != nil {
				return nil, err
			}
			if supported {
				matching = append(matching, analyzer)
			}
		}
		if !explicit && len(matching) > 1 {
			return nil, fmt.Errorf("document %q matches multiple analyzers: %w", document.Path(), pkgContext.ErrAmbiguous)
		}
		for _, analyzer := range matching {
			if cached, found := cache.get(analysisCacheKey(document, analyzer)); found {
				analysis, err := rebindAnalysis(cached.(pkgContext.Analysis), document)
				if err != nil {
					return nil, err
				}
				analyses = append(analyses, analysis)
				continue
			}
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			analysis, err := analyzeSafely(ctx, analyzer, document)
			if err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return nil, ctxErr
				}
				return nil, fmt.Errorf("analyze document %q with %q: %w: %w", document.Path(), analyzer.ID(), err, pkgContext.ErrAnalyzerFailure)
			}
			if analysis.Analyzer() != analyzer.ID() || analysis.Version() != analyzer.Version() {
				return nil, fmt.Errorf("analyzer %q returned mismatched identity or version: %w", analyzer.ID(), pkgContext.ErrInvalidAnalysis)
			}
			analyses = append(analyses, analysis)
		}
	}
	return analyses, nil
}

func rebindAnalysis(cached pkgContext.Analysis, document pkgContext.Document) (pkgContext.Analysis, error) {
	diagnostics := cached.Diagnostics()
	for index := range diagnostics {
		diagnostics[index].Path = document.Path()
	}
	return pkgContext.NewAnalysis(
		document, cached.Analyzer(), cached.Version(), cached.Symbols(),
		cached.Relations(), cached.Chunks(), diagnostics,
	)
}

func cacheAnalyses(cache *artifactCache, snapshot pkgContext.Snapshot, analyzers []pkgContext.Analyzer) {
	byID := make(map[pkgContext.AnalyzerID]pkgContext.Analyzer, len(analyzers))
	for _, analyzer := range analyzers {
		byID[analyzer.ID()] = analyzer
	}
	for _, analysis := range snapshot.Analyses() {
		document, found := snapshot.Document(analysis.Path())
		analyzer, analyzerFound := byID[analysis.Analyzer()]
		if found && analyzerFound {
			cache.put(analysisCacheKey(document, analyzer), analysis, analysisSize(analysis))
		}
	}
}

func supportsSafely(analyzer pkgContext.Analyzer, document pkgContext.Document) (supported bool, err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("analyzer %q panicked during support check: %w", analyzer.ID(), pkgContext.ErrAnalyzerFailure)
		}
	}()
	return analyzer.Supports(document), nil
}

func analyzeSafely(ctx context.Context, analyzer pkgContext.Analyzer, document pkgContext.Document) (analysis pkgContext.Analysis, err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("analyzer %q panicked: %w", analyzer.ID(), pkgContext.ErrAnalyzerFailure)
		}
	}()
	return analyzer.Analyze(ctx, document)
}

func exactVersion(version string) bool {
	return version != "" && len(version) <= 64 && strings.TrimSpace(version) == version && !strings.ContainsRune(version, 0)
}

func (engine *Engine) Snapshot(id pkgContext.WorkspaceID) (pkgContext.Snapshot, bool) {
	engine.mu.RLock()
	snapshot, found := engine.snapshots[id]
	engine.mu.RUnlock()
	return snapshot, found
}

func (engine *Engine) Retrieve(ctx context.Context, query pkgContext.RetrievalQuery) ([]pkgContext.RetrievalResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("retrieve with nil context: %w", pkgContext.ErrInvalidQuery)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := query.Validate(); err != nil {
		return nil, err
	}
	engine.mu.RLock()
	snapshot, found := engine.snapshots[query.Workspace()]
	embedding := engine.embedding
	engine.mu.RUnlock()
	if !found {
		return nil, fmt.Errorf("workspace %q snapshot: %w", query.Workspace(), pkgContext.ErrNotFound)
	}
	return retrieveSnapshot(ctx, snapshot, query, embedding, engine.cache)
}

func (engine *Engine) Build(ctx context.Context, request pkgContext.BuildRequest) (pkgContext.ContextBundle, error) {
	if ctx == nil {
		return pkgContext.ContextBundle{}, fmt.Errorf("build with nil context: %w", pkgContext.ErrInvalidBundle)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.ContextBundle{}, err
	}
	if err := request.Validate(); err != nil {
		return pkgContext.ContextBundle{}, err
	}
	engine.mu.RLock()
	snapshot, found := engine.snapshots[request.Query.Workspace()]
	estimator, estimatorFound := engine.estimators[request.Estimator]
	embedding := engine.embedding
	engine.mu.RUnlock()
	if !found {
		return pkgContext.ContextBundle{}, fmt.Errorf("workspace %q snapshot: %w", request.Query.Workspace(), pkgContext.ErrNotFound)
	}
	if !estimatorFound {
		return pkgContext.ContextBundle{}, fmt.Errorf("token estimator %q: %w", request.Estimator, pkgContext.ErrNotFound)
	}
	results, err := retrieveSnapshot(ctx, snapshot, request.Query, embedding, engine.cache)
	if err != nil {
		return pkgContext.ContextBundle{}, err
	}
	return buildBundle(ctx, snapshot, estimator, request.Budget, results, engine.cache)
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflection := reflect.ValueOf(value)
	switch reflection.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflection.IsNil()
	default:
		return false
	}
}
