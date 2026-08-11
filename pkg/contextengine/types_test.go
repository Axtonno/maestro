package contextengine_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

func TestPublicContractAssertions(t *testing.T) {
	var _ contextengine.WorkspaceProvider = workspaceProviderStub{}
	var _ contextengine.Source = sourceStub{}
	var _ contextengine.Analyzer = analyzerStub{}
	var _ contextengine.TokenEstimator = estimatorStub{}
	var _ contextengine.Engine = engineStub{}
}

func TestWorkspaceValidationAndDefensiveCopies(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	policy := contextengine.DefaultScanPolicy()
	policy.Include = []string{"*.go"}
	analyzers := []contextengine.AnalyzerID{"context.go-ast"}
	metadata := map[string]string{"framework": "laravel"}

	workspace, err := contextengine.NewWorkspace("workspace", root, contextengine.WorkspaceOptions{
		Source: contextengine.SourceFilesystem, Policy: policy, Analyzers: analyzers, Metadata: metadata,
	})
	if err != nil {
		t.Fatalf("construct workspace: %v", err)
	}
	policy.Include[0] = "*.php"
	analyzers[0] = "context.changed"
	metadata["framework"] = "changed"
	if got := workspace.Policy().Include[0]; got != "*.go" {
		t.Fatalf("workspace policy shares caller storage: %q", got)
	}
	if got := workspace.Metadata()["framework"]; got != "laravel" {
		t.Fatalf("workspace metadata shares caller storage: %q", got)
	}
	if got := workspace.Analyzers()[0]; got != "context.go-ast" {
		t.Fatalf("workspace analyzers share caller storage: %q", got)
	}

	returnedPolicy := workspace.Policy()
	returnedPolicy.Include[0] = "*.md"
	returnedMetadata := workspace.Metadata()
	returnedAnalyzers := workspace.Analyzers()
	returnedAnalyzers[0] = "context.changed"
	returnedMetadata["framework"] = "symfony"
	if workspace.Policy().Include[0] != "*.go" || workspace.Metadata()["framework"] != "laravel" || workspace.Analyzers()[0] != "context.go-ast" {
		t.Fatal("workspace accessors expose mutable storage")
	}
}

func TestWorkspaceRejectsInvalidInputs(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	valid := contextengine.WorkspaceOptions{Source: contextengine.SourceFilesystem, Policy: contextengine.DefaultScanPolicy()}
	tests := []struct {
		name    string
		id      contextengine.WorkspaceID
		root    string
		options contextengine.WorkspaceOptions
		is      error
	}{
		{name: "empty ID", root: root, options: valid, is: contextengine.ErrInvalidWorkspaceID},
		{name: "relative root", id: "workspace", root: ".", options: valid, is: contextengine.ErrInvalidWorkspace},
		{name: "unnormalized root", id: "workspace", root: root + string(filepath.Separator) + "child" + string(filepath.Separator) + "..", options: valid, is: contextengine.ErrInvalidWorkspace},
		{name: "invalid source", id: "workspace", root: root, options: contextengine.WorkspaceOptions{Source: "filesystem", Policy: contextengine.DefaultScanPolicy()}, is: contextengine.ErrInvalidSourceID},
		{name: "zero policy", id: "workspace", root: root, options: contextengine.WorkspaceOptions{Source: contextengine.SourceFilesystem}, is: contextengine.ErrInvalidPolicy},
		{name: "unsafe metadata", id: "workspace", root: root, options: contextengine.WorkspaceOptions{Source: contextengine.SourceFilesystem, Policy: contextengine.DefaultScanPolicy(), Metadata: map[string]string{"bad\nkey": "value"}}, is: contextengine.ErrInvalidWorkspace},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := contextengine.NewWorkspace(test.id, test.root, test.options)
			if !errors.Is(err, test.is) {
				t.Fatalf("expected %v, got %v", test.is, err)
			}
		})
	}
}

func TestDocumentIdentityAndValidation(t *testing.T) {
	document, err := contextengine.NewDocument("src/main.go", "text/x-go", "go", "package main\n")
	if err != nil {
		t.Fatalf("construct document: %v", err)
	}
	if document.Path() != "src/main.go" || document.SizeBytes() != len("package main\n") || len(document.Digest()) != 64 {
		t.Fatalf("unexpected document: path=%q size=%d digest=%q", document.Path(), document.SizeBytes(), document.Digest())
	}

	invalidPaths := []contextengine.DocumentPath{"", ".", "/absolute", "../escape", "a/../b", `a\b`, "bad\x00path"}
	for _, documentPath := range invalidPaths {
		if _, err := contextengine.NewDocument(documentPath, "text/plain", "", "text"); !errors.Is(err, contextengine.ErrInvalidPath) {
			t.Errorf("path %q: expected invalid path, got %v", documentPath, err)
		}
	}
	if _, err := contextengine.NewDocument("file", "invalid", "", "text"); !errors.Is(err, contextengine.ErrInvalidDocument) {
		t.Fatalf("expected invalid media type, got %v", err)
	}
	if _, err := contextengine.NewDocument("file", "text/plain", "Go", "text"); !errors.Is(err, contextengine.ErrInvalidDocument) {
		t.Fatalf("expected invalid language, got %v", err)
	}
}

func TestAnalysisAndSnapshotAreValidatedAndOrdered(t *testing.T) {
	workspace := newWorkspace(t)
	left := newDocument(t, "b.go", "package b\n")
	right := newDocument(t, "a.go", "package a\n")
	analysis, err := contextengine.NewAnalysis(
		right,
		"context.go-ast",
		"1",
		[]contextengine.Symbol{{ID: "package_0", Name: "a", Kind: contextengine.SymbolPackage, Range: contextengine.SourceRange{Start: 0, End: 9}}},
		nil,
		[]contextengine.Chunk{{ID: "chunk_0", Kind: "declaration", Range: contextengine.SourceRange{Start: 0, End: right.SizeBytes()}}},
		nil,
	)
	if err != nil {
		t.Fatalf("construct analysis: %v", err)
	}
	snapshot, err := contextengine.NewSnapshot(workspace, 1, []contextengine.Document{left, right}, []contextengine.Analysis{analysis}, nil)
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	documents := snapshot.Documents()
	if documents[0].Path() != "a.go" || documents[1].Path() != "b.go" {
		t.Fatalf("documents are not ordered: %#v", documents)
	}
	documents[0] = left
	if snapshot.Documents()[0].Path() != "a.go" {
		t.Fatal("snapshot documents share caller storage")
	}
	metadata := snapshot.Metadata()
	if metadata.Generation != 1 || metadata.DocumentCount != 2 || metadata.AnalysisCount != 1 || metadata.TotalBytes != int64(left.SizeBytes()+right.SizeBytes()) {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if found, ok := snapshot.Document("a.go"); !ok || found.Digest() != right.Digest() {
		t.Fatalf("snapshot lookup failed: %#v %v", found, ok)
	}
}

func TestAnalysisRejectsInvalidReferences(t *testing.T) {
	document := newDocument(t, "main.go", "package main\n")
	_, err := contextengine.NewAnalysis(
		document,
		"context.go-ast",
		"1",
		[]contextengine.Symbol{{ID: "symbol_0", Name: "main", Kind: contextengine.SymbolPackage, Range: contextengine.SourceRange{Start: 0, End: 7}}},
		[]contextengine.Relation{{From: "missing", To: "external", Kind: contextengine.RelationImports}},
		nil,
		nil,
	)
	if !errors.Is(err, contextengine.ErrInvalidAnalysis) {
		t.Fatalf("expected invalid analysis, got %v", err)
	}
}

func TestRetrievalQueryValidationAndCopies(t *testing.T) {
	target := contextengine.EmbeddingTarget{Provider: "ollama", Model: "embed"}
	methods := []contextengine.RetrievalMethod{contextengine.RetrievalLexical, contextengine.RetrievalSemantic}
	paths := []contextengine.DocumentPath{"main.go"}
	query, err := contextengine.NewRetrievalQuery("workspace", "main package", contextengine.RetrievalQueryOptions{
		Methods: methods, Paths: paths, Languages: []contextengine.Language{"go"}, TopK: 5,
		Embedding: &target, Fusion: contextengine.FusionReciprocalRank,
	})
	if err != nil {
		t.Fatalf("construct query: %v", err)
	}
	methods[0] = contextengine.RetrievalStructured
	paths[0] = "other.go"
	target.Model = "changed"
	if !reflect.DeepEqual(query.Methods(), []contextengine.RetrievalMethod{contextengine.RetrievalLexical, contextengine.RetrievalSemantic}) || query.Paths()[0] != "main.go" {
		t.Fatal("query shares caller storage")
	}
	if embedding, ok := query.Embedding(); !ok || embedding.Model != "embed" {
		t.Fatalf("unexpected embedding target: %#v %v", embedding, ok)
	}

	_, err = contextengine.NewRetrievalQuery("workspace", "query", contextengine.RetrievalQueryOptions{
		Methods: []contextengine.RetrievalMethod{contextengine.RetrievalSemantic}, TopK: 1,
	})
	if !errors.Is(err, contextengine.ErrInvalidQuery) {
		t.Fatalf("expected semantic target error, got %v", err)
	}
}

func TestContextBundleEnforcesBudgetAndCopiesSections(t *testing.T) {
	workspace := newWorkspace(t)
	document := newDocument(t, "main.go", "package main\n")
	snapshot, err := contextengine.NewSnapshot(workspace, 3, []contextengine.Document{document}, nil, nil)
	if err != nil {
		t.Fatalf("construct snapshot: %v", err)
	}
	sections := []contextengine.ContextSection{{
		Path: "main.go", Range: contextengine.SourceRange{Start: 0, End: document.SizeBytes()},
		Role: "evidence", Method: contextengine.RetrievalLexical, ReasonCode: "term_match",
		Text: document.Content(), Tokens: 3,
	}}
	budget := contextengine.Budget{MaxTokens: 10, ReservedTokens: 2, SafetyTokens: 1}
	bundle, err := contextengine.NewContextBundle(snapshot, "context.bytes-estimator", "1", budget, sections)
	if err != nil {
		t.Fatalf("construct bundle: %v", err)
	}
	sections[0].Text = "changed"
	if bundle.Workspace() != "workspace" || bundle.Generation() != 3 || bundle.UsedTokens() != 3 || bundle.Sections()[0].Text != document.Content() {
		t.Fatalf("unexpected or mutable bundle: %#v", bundle)
	}

	sections = bundle.Sections()
	sections[0].Tokens = 8
	_, err = contextengine.NewContextBundle(snapshot, "context.bytes-estimator", "1", budget, sections)
	if !errors.Is(err, contextengine.ErrBudgetExceeded) || !errors.Is(err, contextengine.ErrInvalidBundle) {
		t.Fatalf("expected composed budget error, got %v", err)
	}
}

func newWorkspace(t *testing.T) contextengine.Workspace {
	t.Helper()
	workspace, err := contextengine.NewWorkspace("workspace", filepath.Clean(t.TempDir()), contextengine.WorkspaceOptions{
		Source: contextengine.SourceFilesystem, Policy: contextengine.DefaultScanPolicy(),
	})
	if err != nil {
		t.Fatalf("construct workspace: %v", err)
	}
	return workspace
}

func newDocument(t *testing.T, path contextengine.DocumentPath, content string) contextengine.Document {
	t.Helper()
	document, err := contextengine.NewDocument(path, "text/x-go", "go", content)
	if err != nil {
		t.Fatalf("construct document: %v", err)
	}
	return document
}

type workspaceProviderStub struct{}

func (workspaceProviderStub) Workspace(context.Context) (contextengine.Workspace, error) {
	return contextengine.Workspace{}, nil
}

type sourceStub struct{}

func (sourceStub) ID() contextengine.SourceID { return contextengine.SourceFilesystem }
func (sourceStub) Scan(context.Context, contextengine.Workspace) (contextengine.ScanResult, error) {
	return contextengine.ScanResult{}, nil
}

type analyzerStub struct{}

func (analyzerStub) ID() contextengine.AnalyzerID         { return "context.stub" }
func (analyzerStub) Version() string                      { return "1" }
func (analyzerStub) Supports(contextengine.Document) bool { return true }
func (analyzerStub) Analyze(context.Context, contextengine.Document) (contextengine.Analysis, error) {
	return contextengine.Analysis{}, nil
}

type estimatorStub struct{}

func (estimatorStub) ID() contextengine.EstimatorID                 { return "context.stub" }
func (estimatorStub) Version() string                               { return "1" }
func (estimatorStub) Estimate(context.Context, string) (int, error) { return 1, nil }

type engineStub struct{}

func (engineStub) CacheStats() contextengine.CacheStats { return contextengine.CacheStats{} }

func (engineStub) RegisterSource(contextengine.Source) error            { return nil }
func (engineStub) RegisterAnalyzer(contextengine.Analyzer) error        { return nil }
func (engineStub) RegisterEstimator(contextengine.TokenEstimator) error { return nil }
func (engineStub) Index(context.Context, contextengine.Workspace) (contextengine.Snapshot, error) {
	return contextengine.Snapshot{}, nil
}
func (engineStub) Snapshot(contextengine.WorkspaceID) (contextengine.Snapshot, bool) {
	return contextengine.Snapshot{}, false
}
func (engineStub) Retrieve(context.Context, contextengine.RetrievalQuery) ([]contextengine.RetrievalResult, error) {
	return nil, nil
}
func (engineStub) Build(context.Context, contextengine.BuildRequest) (contextengine.ContextBundle, error) {
	return contextengine.ContextBundle{}, nil
}
