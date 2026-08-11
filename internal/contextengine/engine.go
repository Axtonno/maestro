package contextengine

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

var _ pkgContext.Engine = (*Engine)(nil)

type Engine struct {
	mu        sync.RWMutex
	sources   map[pkgContext.SourceID]pkgContext.Source
	snapshots map[pkgContext.WorkspaceID]pkgContext.Snapshot
}

func New() *Engine {
	filesystem := NewFilesystemSource()
	return &Engine{
		sources:   map[pkgContext.SourceID]pkgContext.Source{filesystem.ID(): filesystem},
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

func (engine *Engine) RegisterAnalyzer(pkgContext.Analyzer) error {
	return pkgContext.ErrUnsupported
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
	engine.mu.RUnlock()
	if !exists {
		return pkgContext.Snapshot{}, fmt.Errorf("context source %q: %w", workspace.Source(), pkgContext.ErrNotFound)
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

	engine.mu.Lock()
	defer engine.mu.Unlock()
	generation := uint64(1)
	if current, found := engine.snapshots[workspace.ID()]; found {
		generation = current.Metadata().Generation + 1
	}
	snapshot, err := pkgContext.NewSnapshot(workspace, generation, result.Documents, nil, result.Diagnostics)
	if err != nil {
		return pkgContext.Snapshot{}, fmt.Errorf("publish workspace %q snapshot: %w", workspace.ID(), err)
	}
	engine.snapshots[workspace.ID()] = snapshot
	return snapshot, nil
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
