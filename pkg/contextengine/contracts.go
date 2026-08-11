package contextengine

import "context"

type ScanResult struct {
	Documents   []Document
	Diagnostics []Diagnostic
}

type Source interface {
	ID() SourceID
	Scan(context.Context, Workspace) (ScanResult, error)
}

type Engine interface {
	RegisterSource(Source) error
	RegisterAnalyzer(Analyzer) error
	RegisterEstimator(TokenEstimator) error
	Index(context.Context, Workspace) (Snapshot, error)
	Snapshot(WorkspaceID) (Snapshot, bool)
	Retrieve(context.Context, RetrievalQuery) ([]RetrievalResult, error)
	Build(context.Context, BuildRequest) (ContextBundle, error)
}
