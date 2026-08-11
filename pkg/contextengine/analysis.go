package contextengine

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type SymbolKind string

const (
	SymbolUnknown   SymbolKind = "unknown"
	SymbolPackage   SymbolKind = "package"
	SymbolNamespace SymbolKind = "namespace"
	SymbolType      SymbolKind = "type"
	SymbolFunction  SymbolKind = "function"
	SymbolMethod    SymbolKind = "method"
	SymbolField     SymbolKind = "field"
	SymbolVariable  SymbolKind = "variable"
	SymbolConstant  SymbolKind = "constant"
	SymbolImport    SymbolKind = "import"
)

func (kind SymbolKind) Valid() bool {
	switch kind {
	case SymbolUnknown, SymbolPackage, SymbolNamespace, SymbolType, SymbolFunction,
		SymbolMethod, SymbolField, SymbolVariable, SymbolConstant, SymbolImport:
		return true
	default:
		return false
	}
}

type Symbol struct {
	ID        string
	Name      string
	Kind      SymbolKind
	Range     SourceRange
	Container string
}

func (symbol Symbol) Validate(size int) error {
	if !safeCode(symbol.ID) || symbol.Name == "" || strings.TrimSpace(symbol.Name) != symbol.Name || !symbol.Kind.Valid() {
		return fmt.Errorf("symbol %q has invalid identity, name, or kind: %w", symbol.ID, ErrInvalidAnalysis)
	}
	if err := symbol.Range.Validate(size); err != nil {
		return fmt.Errorf("symbol %q: %w: %w", symbol.ID, err, ErrInvalidAnalysis)
	}
	if symbol.Container != "" && !safeCode(symbol.Container) {
		return fmt.Errorf("symbol %q container %q is invalid: %w", symbol.ID, symbol.Container, ErrInvalidAnalysis)
	}
	return nil
}

type RelationKind string

const (
	RelationContains   RelationKind = "contains"
	RelationImports    RelationKind = "imports"
	RelationCalls      RelationKind = "calls"
	RelationExtends    RelationKind = "extends"
	RelationImplements RelationKind = "implements"
)

func (kind RelationKind) Valid() bool {
	switch kind {
	case RelationContains, RelationImports, RelationCalls, RelationExtends, RelationImplements:
		return true
	default:
		return false
	}
}

type Relation struct {
	From string
	To   string
	Kind RelationKind
}

func (relation Relation) Validate(symbols map[string]struct{}) error {
	if _, exists := symbols[relation.From]; !exists {
		return fmt.Errorf("relation source %q is unknown: %w", relation.From, ErrInvalidAnalysis)
	}
	if !exactID(relation.To) || !relation.Kind.Valid() {
		return fmt.Errorf("relation target %q or kind %q is invalid: %w", relation.To, relation.Kind, ErrInvalidAnalysis)
	}
	return nil
}

type Chunk struct {
	ID    string
	Kind  string
	Range SourceRange
}

func (chunk Chunk) Validate(size int) error {
	if !safeCode(chunk.ID) || !safeCode(chunk.Kind) {
		return fmt.Errorf("chunk ID %q or kind %q is invalid: %w", chunk.ID, chunk.Kind, ErrInvalidAnalysis)
	}
	if err := chunk.Range.Validate(size); err != nil {
		return fmt.Errorf("chunk %q: %w: %w", chunk.ID, err, ErrInvalidAnalysis)
	}
	return nil
}

type Analysis struct {
	path        DocumentPath
	digest      Digest
	analyzer    AnalyzerID
	version     string
	symbols     []Symbol
	relations   []Relation
	chunks      []Chunk
	diagnostics []Diagnostic
}

func NewAnalysis(
	document Document,
	analyzer AnalyzerID,
	version string,
	symbols []Symbol,
	relations []Relation,
	chunks []Chunk,
	diagnostics []Diagnostic,
) (Analysis, error) {
	analysis := Analysis{
		path: document.Path(), digest: document.Digest(), analyzer: analyzer, version: version,
		symbols: slices.Clone(symbols), relations: slices.Clone(relations), chunks: slices.Clone(chunks),
		diagnostics: slices.Clone(diagnostics),
	}
	if err := analysis.validate(document); err != nil {
		return Analysis{}, err
	}
	return analysis, nil
}

func (analysis Analysis) Path() DocumentPath        { return analysis.path }
func (analysis Analysis) Digest() Digest            { return analysis.digest }
func (analysis Analysis) Analyzer() AnalyzerID      { return analysis.analyzer }
func (analysis Analysis) Version() string           { return analysis.version }
func (analysis Analysis) Symbols() []Symbol         { return slices.Clone(analysis.symbols) }
func (analysis Analysis) Relations() []Relation     { return slices.Clone(analysis.relations) }
func (analysis Analysis) Chunks() []Chunk           { return slices.Clone(analysis.chunks) }
func (analysis Analysis) Diagnostics() []Diagnostic { return slices.Clone(analysis.diagnostics) }

func (analysis Analysis) validate(document Document) error {
	if analysis.path != document.Path() || analysis.digest != document.Digest() {
		return fmt.Errorf("analysis does not identify its document: %w", ErrInvalidAnalysis)
	}
	if err := analysis.analyzer.Validate(); err != nil {
		return fmt.Errorf("analysis analyzer: %w: %w", err, ErrInvalidAnalysis)
	}
	if !exactID(analysis.version) {
		return fmt.Errorf("analysis version %q is invalid: %w", analysis.version, ErrInvalidAnalysis)
	}
	symbolSet := make(map[string]struct{}, len(analysis.symbols))
	for index, symbol := range analysis.symbols {
		if err := symbol.Validate(document.SizeBytes()); err != nil {
			return fmt.Errorf("analysis symbol %d: %w", index, err)
		}
		if _, exists := symbolSet[symbol.ID]; exists {
			return fmt.Errorf("analysis symbol %q is duplicated: %w", symbol.ID, ErrInvalidAnalysis)
		}
		symbolSet[symbol.ID] = struct{}{}
	}
	for index, relation := range analysis.relations {
		if err := relation.Validate(symbolSet); err != nil {
			return fmt.Errorf("analysis relation %d: %w", index, err)
		}
	}
	chunkSet := make(map[string]struct{}, len(analysis.chunks))
	for index, chunk := range analysis.chunks {
		if err := chunk.Validate(document.SizeBytes()); err != nil {
			return fmt.Errorf("analysis chunk %d: %w", index, err)
		}
		if _, exists := chunkSet[chunk.ID]; exists {
			return fmt.Errorf("analysis chunk %q is duplicated: %w", chunk.ID, ErrInvalidAnalysis)
		}
		chunkSet[chunk.ID] = struct{}{}
	}
	for index, diagnostic := range analysis.diagnostics {
		if diagnostic.Path != document.Path() {
			return fmt.Errorf("analysis diagnostic %d belongs to another document: %w", index, ErrInvalidAnalysis)
		}
		if err := diagnostic.Validate(document.SizeBytes()); err != nil {
			return fmt.Errorf("analysis diagnostic %d: %w", index, err)
		}
	}
	return nil
}

type Analyzer interface {
	ID() AnalyzerID
	Version() string
	Supports(Document) bool
	Analyze(context.Context, Document) (Analysis, error)
}
