package contextengine

import (
	"fmt"
	"slices"
)

type SnapshotMetadata struct {
	Workspace       WorkspaceID
	Generation      uint64
	DocumentCount   int
	AnalysisCount   int
	DiagnosticCount int
	TotalBytes      int64
}

// Snapshot is an immutable, deterministically ordered workspace view.
type Snapshot struct {
	metadata    SnapshotMetadata
	documents   []Document
	analyses    []Analysis
	diagnostics []Diagnostic
}

func NewSnapshot(
	workspace Workspace,
	generation uint64,
	documents []Document,
	analyses []Analysis,
	diagnostics []Diagnostic,
) (Snapshot, error) {
	if err := workspace.Validate(); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot workspace: %w: %w", err, ErrInvalidSnapshot)
	}
	if generation == 0 {
		return Snapshot{}, fmt.Errorf("snapshot generation must be positive: %w", ErrInvalidSnapshot)
	}
	orderedDocuments := slices.Clone(documents)
	slices.SortFunc(orderedDocuments, func(left, right Document) int {
		return left.Path().Compare(right.Path())
	})
	documentByPath := make(map[DocumentPath]Document, len(orderedDocuments))
	var totalBytes int64
	for index, document := range orderedDocuments {
		if err := document.Validate(); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot document %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
		if _, exists := documentByPath[document.Path()]; exists {
			return Snapshot{}, fmt.Errorf("snapshot document path %q is duplicated: %w", document.Path(), ErrInvalidSnapshot)
		}
		documentByPath[document.Path()] = document
		totalBytes += int64(document.SizeBytes())
	}

	orderedAnalyses := slices.Clone(analyses)
	slices.SortFunc(orderedAnalyses, compareAnalysis)
	seenAnalysis := make(map[string]struct{}, len(orderedAnalyses))
	for index, analysis := range orderedAnalyses {
		document, exists := documentByPath[analysis.Path()]
		if !exists {
			return Snapshot{}, fmt.Errorf("snapshot analysis %d references missing document %q: %w", index, analysis.Path(), ErrInvalidSnapshot)
		}
		if err := analysis.validate(document); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot analysis %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
		key := string(analysis.Path()) + "\x00" + string(analysis.Analyzer())
		if _, exists := seenAnalysis[key]; exists {
			return Snapshot{}, fmt.Errorf("snapshot analysis for %q by %q is duplicated: %w", analysis.Path(), analysis.Analyzer(), ErrInvalidSnapshot)
		}
		seenAnalysis[key] = struct{}{}
	}

	orderedDiagnostics := slices.Clone(diagnostics)
	slices.SortFunc(orderedDiagnostics, compareDiagnostic)
	for index, diagnostic := range orderedDiagnostics {
		document, exists := documentByPath[diagnostic.Path]
		if !exists {
			return Snapshot{}, fmt.Errorf("snapshot diagnostic %d references missing document %q: %w", index, diagnostic.Path, ErrInvalidSnapshot)
		}
		if err := diagnostic.Validate(document.SizeBytes()); err != nil {
			return Snapshot{}, fmt.Errorf("snapshot diagnostic %d: %w: %w", index, err, ErrInvalidSnapshot)
		}
	}

	return Snapshot{
		metadata: SnapshotMetadata{
			Workspace: workspace.ID(), Generation: generation,
			DocumentCount: len(orderedDocuments), AnalysisCount: len(orderedAnalyses),
			DiagnosticCount: len(orderedDiagnostics), TotalBytes: totalBytes,
		},
		documents: orderedDocuments, analyses: orderedAnalyses, diagnostics: orderedDiagnostics,
	}, nil
}

func (snapshot Snapshot) Metadata() SnapshotMetadata { return snapshot.metadata }
func (snapshot Snapshot) Documents() []Document      { return slices.Clone(snapshot.documents) }
func (snapshot Snapshot) Analyses() []Analysis       { return slices.Clone(snapshot.analyses) }
func (snapshot Snapshot) Diagnostics() []Diagnostic  { return slices.Clone(snapshot.diagnostics) }

func (snapshot Snapshot) Document(documentPath DocumentPath) (Document, bool) {
	index, found := slices.BinarySearchFunc(snapshot.documents, documentPath, func(document Document, target DocumentPath) int {
		return document.Path().Compare(target)
	})
	if !found {
		return Document{}, false
	}
	return snapshot.documents[index], true
}

func compareAnalysis(left, right Analysis) int {
	if result := left.Path().Compare(right.Path()); result != 0 {
		return result
	}
	return left.Analyzer().Compare(right.Analyzer())
}

func compareDiagnostic(left, right Diagnostic) int {
	if result := left.Path.Compare(right.Path); result != 0 {
		return result
	}
	if left.Range.Start < right.Range.Start {
		return -1
	}
	if left.Range.Start > right.Range.Start {
		return 1
	}
	if left.Range.End < right.Range.End {
		return -1
	}
	if left.Range.End > right.Range.End {
		return 1
	}
	if left.Code < right.Code {
		return -1
	}
	if left.Code > right.Code {
		return 1
	}
	return 0
}
