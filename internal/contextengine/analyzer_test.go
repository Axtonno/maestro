package contextengine

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

func TestGoAnalyzerProducesSymbolsRelationsAndChunks(t *testing.T) {
	content := `package sample

import "context"

const Version = "1"

type Service struct {
	Name string
}

func (service *Service) Run(ctx context.Context) error {
	return nil
}

func Helper() {}
`
	document := document(t, "service.go", content)
	analysis, err := NewGoAnalyzer().Analyze(context.Background(), document)
	if err != nil {
		t.Fatalf("analyze Go document: %v", err)
	}
	if analysis.Analyzer() != GoAnalyzerID || analysis.Version() != GoAnalyzerVersion || analysis.Path() != document.Path() || analysis.Digest() != document.Digest() {
		t.Fatalf("unexpected analysis identity: %#v", analysis)
	}
	symbols := analysis.Symbols()
	names := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		names = append(names, symbol.Name)
		if symbol.Range.Start < 0 || symbol.Range.End > document.SizeBytes() || symbol.Range.Start >= symbol.Range.End {
			t.Fatalf("invalid symbol range: %#v", symbol)
		}
	}
	for _, expected := range []string{"sample", "context", "Version", "Service", "Name", "Run", "Helper"} {
		if !slices.Contains(names, expected) {
			t.Errorf("missing symbol %q in %v", expected, names)
		}
	}
	if len(analysis.Relations()) == 0 || len(analysis.Chunks()) != 5 || len(analysis.Diagnostics()) != 0 {
		t.Fatalf("unexpected structural output: relations=%d chunks=%d diagnostics=%d", len(analysis.Relations()), len(analysis.Chunks()), len(analysis.Diagnostics()))
	}
}

func TestGoAnalyzerKeepsPartialAnalysisForMalformedSource(t *testing.T) {
	document := document(t, "broken.go", "package main\nfunc broken(")
	analysis, err := NewGoAnalyzer().Analyze(context.Background(), document)
	if err != nil {
		t.Fatalf("malformed source should produce diagnostics: %v", err)
	}
	if len(analysis.Diagnostics()) != 1 || analysis.Diagnostics()[0].Code != "go_parse_error" {
		t.Fatalf("unexpected diagnostics: %#v", analysis.Diagnostics())
	}
	if len(analysis.Symbols()) == 0 {
		t.Fatal("partial AST did not preserve available symbols")
	}
}

func TestIndexRunsGoAnalyzerAutomatically(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\nfunc main() {}\n")
	snapshot, err := New().Index(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, pkgContext.DefaultScanPolicy()))
	if err != nil {
		t.Fatalf("index Go workspace: %v", err)
	}
	if snapshot.Metadata().AnalysisCount != 1 || snapshot.Analyses()[0].Analyzer() != GoAnalyzerID {
		t.Fatalf("Go analysis was not published: %#v", snapshot.Analyses())
	}
}

func TestAnalyzerAmbiguityRequiresExplicitSelection(t *testing.T) {
	root := t.TempDir()
	document := document(t, "main.go", "package main\n")
	source := &mutableSource{id: "context.analysis-source", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document}}}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	custom := &analyzerStub{id: "context.custom", version: "1", supports: true}
	custom.analysis = validAnalysis(custom, document)
	if err := engine.RegisterAnalyzer(custom); err != nil {
		t.Fatalf("register analyzer: %v", err)
	}
	if _, err := engine.Index(context.Background(), workspace(t, root, source.id, pkgContext.DefaultScanPolicy())); !errors.Is(err, pkgContext.ErrAmbiguous) {
		t.Fatalf("expected analyzer ambiguity, got %v", err)
	}

	explicit := workspaceWithAnalyzers(t, root, source.id, []pkgContext.AnalyzerID{GoAnalyzerID, custom.id})
	snapshot, err := engine.Index(context.Background(), explicit)
	if err != nil {
		t.Fatalf("explicit analyzer composition: %v", err)
	}
	if snapshot.Metadata().AnalysisCount != 2 {
		t.Fatalf("expected two analyses, got %d", snapshot.Metadata().AnalysisCount)
	}
}

func TestAnalyzerFailurePanicAndInvalidOutputAreAtomic(t *testing.T) {
	tests := []struct {
		name     string
		analyzer *analyzerStub
		is       error
	}{
		{name: "error", analyzer: &analyzerStub{id: "context.error", version: "1", supports: true, err: errors.New("failed")}, is: pkgContext.ErrAnalyzerFailure},
		{name: "panic", analyzer: &analyzerStub{id: "context.panic", version: "1", supports: true, panicAnalyze: true}, is: pkgContext.ErrAnalyzerFailure},
		{name: "invalid output", analyzer: &analyzerStub{id: "context.invalid-output", version: "1", supports: true}, is: pkgContext.ErrInvalidAnalysis},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			document := document(t, "main.go", "package main\n")
			source := &mutableSource{id: "context.fixture", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document}}}
			engine := New()
			if err := engine.RegisterSource(source); err != nil {
				t.Fatalf("register source: %v", err)
			}
			if err := engine.RegisterAnalyzer(test.analyzer); err != nil {
				t.Fatalf("register analyzer: %v", err)
			}
			workspace := workspaceWithAnalyzers(t, root, source.id, []pkgContext.AnalyzerID{test.analyzer.id})
			if _, err := engine.Index(context.Background(), workspace); !errors.Is(err, test.is) {
				t.Fatalf("expected %v, got %v", test.is, err)
			}
			if _, found := engine.Snapshot("workspace"); found {
				t.Fatal("failed analysis published a snapshot")
			}
		})
	}
}

func TestAnalyzerRegistrationValidation(t *testing.T) {
	engine := New()
	var typedNil *analyzerStub
	if err := engine.RegisterAnalyzer(typedNil); !errors.Is(err, pkgContext.ErrInvalidAnalyzer) {
		t.Fatalf("expected typed nil rejection, got %v", err)
	}
	invalid := &analyzerStub{id: "invalid", version: "1"}
	if err := engine.RegisterAnalyzer(invalid); !errors.Is(err, pkgContext.ErrInvalidAnalyzerID) {
		t.Fatalf("expected invalid analyzer ID, got %v", err)
	}
	duplicate := &analyzerStub{id: GoAnalyzerID, version: "2"}
	if err := engine.RegisterAnalyzer(duplicate); !errors.Is(err, pkgContext.ErrAlreadyRegistered) {
		t.Fatalf("expected duplicate analyzer, got %v", err)
	}
}

func TestAnalyzerCallbacksRunWithoutEngineLock(t *testing.T) {
	root := t.TempDir()
	document := document(t, "main.go", "package main\n")
	source := &mutableSource{id: "context.callback-source", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document}}}
	entered := make(chan struct{})
	release := make(chan struct{})
	blocking := &blockingAnalyzer{id: "context.blocking-analyzer", entered: entered, release: release}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := engine.RegisterAnalyzer(blocking); err != nil {
		t.Fatalf("register blocking analyzer: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := engine.Index(context.Background(), workspaceWithAnalyzers(t, root, source.id, []pkgContext.AnalyzerID{blocking.id}))
		done <- err
	}()
	<-entered
	registered := make(chan error, 1)
	go func() {
		registered <- engine.RegisterAnalyzer(&analyzerStub{id: "context.parallel-analyzer", version: "1"})
	}()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatalf("register analyzer while callback blocked: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("analyzer callback held the engine lock")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("index with blocking analyzer: %v", err)
	}
}

func TestCancellationDuringAnalyzerDoesNotPublish(t *testing.T) {
	root := t.TempDir()
	document := document(t, "main.go", "package main\n")
	source := &mutableSource{id: "context.cancel-analysis-source", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document}}}
	entered := make(chan struct{})
	analyzer := &cancelingAnalyzer{id: "context.canceling-analyzer", entered: entered}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if err := engine.RegisterAnalyzer(analyzer); err != nil {
		t.Fatalf("register analyzer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := engine.Index(ctx, workspaceWithAnalyzers(t, root, source.id, []pkgContext.AnalyzerID{analyzer.id}))
		done <- err
	}()
	<-entered
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, found := engine.Snapshot("workspace"); found {
		t.Fatal("canceled analysis published a snapshot")
	}
}

type analyzerStub struct {
	id           pkgContext.AnalyzerID
	version      string
	supports     bool
	panicSupport bool
	panicAnalyze bool
	analysis     pkgContext.Analysis
	err          error
}

type blockingAnalyzer struct {
	id      pkgContext.AnalyzerID
	entered chan<- struct{}
	release <-chan struct{}
}

func (analyzer *blockingAnalyzer) ID() pkgContext.AnalyzerID { return analyzer.id }
func (*blockingAnalyzer) Version() string                    { return "1" }
func (analyzer *blockingAnalyzer) Supports(pkgContext.Document) bool {
	close(analyzer.entered)
	<-analyzer.release
	return false
}
func (*blockingAnalyzer) Analyze(context.Context, pkgContext.Document) (pkgContext.Analysis, error) {
	return pkgContext.Analysis{}, errors.New("unexpected analyze")
}

type cancelingAnalyzer struct {
	id      pkgContext.AnalyzerID
	entered chan<- struct{}
}

func (analyzer *cancelingAnalyzer) ID() pkgContext.AnalyzerID { return analyzer.id }
func (*cancelingAnalyzer) Version() string                    { return "1" }
func (*cancelingAnalyzer) Supports(pkgContext.Document) bool  { return true }
func (analyzer *cancelingAnalyzer) Analyze(ctx context.Context, _ pkgContext.Document) (pkgContext.Analysis, error) {
	close(analyzer.entered)
	<-ctx.Done()
	return pkgContext.Analysis{}, ctx.Err()
}

func (analyzer *analyzerStub) ID() pkgContext.AnalyzerID { return analyzer.id }
func (analyzer *analyzerStub) Version() string           { return analyzer.version }
func (analyzer *analyzerStub) Supports(pkgContext.Document) bool {
	if analyzer.panicSupport {
		panic("support panic")
	}
	return analyzer.supports
}
func (analyzer *analyzerStub) Analyze(context.Context, pkgContext.Document) (pkgContext.Analysis, error) {
	if analyzer.panicAnalyze {
		panic("analyze panic")
	}
	return analyzer.analysis, analyzer.err
}

func validAnalysis(analyzer *analyzerStub, document pkgContext.Document) pkgContext.Analysis {
	analysis, err := pkgContext.NewAnalysis(document, analyzer.id, analyzer.version, nil, nil, nil, nil)
	if err != nil {
		panic(err)
	}
	return analysis
}

func workspaceWithAnalyzers(t *testing.T, root string, source pkgContext.SourceID, analyzers []pkgContext.AnalyzerID) pkgContext.Workspace {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace("workspace", filepath.Clean(root), pkgContext.WorkspaceOptions{
		Source: source, Policy: pkgContext.DefaultScanPolicy(), Analyzers: analyzers,
	})
	if err != nil {
		t.Fatalf("construct workspace: %v", err)
	}
	return workspace
}
