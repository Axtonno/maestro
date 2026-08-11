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
	mu        sync.RWMutex
	sources   map[pkgContext.SourceID]pkgContext.Source
	analyzers map[pkgContext.AnalyzerID]pkgContext.Analyzer
	snapshots map[pkgContext.WorkspaceID]pkgContext.Snapshot
}

func New() *Engine {
	filesystem := NewFilesystemSource()
	goAnalyzer := NewGoAnalyzer()
	return &Engine{
		sources:   map[pkgContext.SourceID]pkgContext.Source{filesystem.ID(): filesystem},
		analyzers: map[pkgContext.AnalyzerID]pkgContext.Analyzer{goAnalyzer.ID(): goAnalyzer},
		snapshots: make(map[pkgContext.WorkspaceID]pkgContext.Snapshot),
	}
}

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

func (engine *Engine) RegisterEstimator(pkgContext.TokenEstimator) error {
	return pkgContext.ErrUnsupported
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
	analyses, err := analyzeDocuments(ctx, result.Documents, analyzers, len(workspace.Analyzers()) > 0)
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

func analyzeDocuments(ctx context.Context, documents []pkgContext.Document, analyzers []pkgContext.Analyzer, explicit bool) ([]pkgContext.Analysis, error) {
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

func (engine *Engine) Retrieve(context.Context, pkgContext.RetrievalQuery) ([]pkgContext.RetrievalResult, error) {
	return nil, pkgContext.ErrUnsupported
}

func (engine *Engine) Build(context.Context, pkgContext.BuildRequest) (pkgContext.ContextBundle, error) {
	return pkgContext.ContextBundle{}, pkgContext.ErrUnsupported
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
