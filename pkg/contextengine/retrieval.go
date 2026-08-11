package contextengine

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type RetrievalMethod string

const (
	RetrievalLexical    RetrievalMethod = "lexical"
	RetrievalStructured RetrievalMethod = "structured"
	RetrievalSemantic   RetrievalMethod = "semantic"
)

func (method RetrievalMethod) Valid() bool {
	return method == RetrievalLexical || method == RetrievalStructured || method == RetrievalSemantic
}

type EmbeddingTarget struct {
	Provider string
	Model    string
}

func (target EmbeddingTarget) Validate() error {
	if !exactID(target.Provider) || !exactID(target.Model) {
		return fmt.Errorf("embedding provider %q and model %q must be exact: %w", target.Provider, target.Model, ErrInvalidQuery)
	}
	return nil
}

type RetrievalQueryOptions struct {
	Methods   []RetrievalMethod
	Paths     []DocumentPath
	Languages []Language
	Symbol    string
	TopK      int
	Embedding *EmbeddingTarget
}

type RetrievalQuery struct {
	workspace WorkspaceID
	text      string
	methods   []RetrievalMethod
	paths     []DocumentPath
	languages []Language
	symbol    string
	topK      int
	embedding *EmbeddingTarget
}

func NewRetrievalQuery(workspace WorkspaceID, text string, options RetrievalQueryOptions) (RetrievalQuery, error) {
	query := RetrievalQuery{
		workspace: workspace, text: text, methods: slices.Clone(options.Methods),
		paths: slices.Clone(options.Paths), languages: slices.Clone(options.Languages),
		symbol: options.Symbol, topK: options.TopK,
	}
	if options.Embedding != nil {
		target := *options.Embedding
		query.embedding = &target
	}
	if err := query.Validate(); err != nil {
		return RetrievalQuery{}, err
	}
	return query, nil
}

func (query RetrievalQuery) Workspace() WorkspaceID     { return query.workspace }
func (query RetrievalQuery) Text() string               { return query.text }
func (query RetrievalQuery) Methods() []RetrievalMethod { return slices.Clone(query.methods) }
func (query RetrievalQuery) Paths() []DocumentPath      { return slices.Clone(query.paths) }
func (query RetrievalQuery) Languages() []Language      { return slices.Clone(query.languages) }
func (query RetrievalQuery) Symbol() string             { return query.symbol }
func (query RetrievalQuery) TopK() int                  { return query.topK }
func (query RetrievalQuery) Embedding() (EmbeddingTarget, bool) {
	if query.embedding == nil {
		return EmbeddingTarget{}, false
	}
	return *query.embedding, true
}

func (query RetrievalQuery) Validate() error {
	if err := query.workspace.Validate(); err != nil {
		return fmt.Errorf("retrieval workspace: %w: %w", err, ErrInvalidQuery)
	}
	if strings.TrimSpace(query.text) == "" || query.topK <= 0 {
		return fmt.Errorf("retrieval text must not be blank and top-k must be positive: %w", ErrInvalidQuery)
	}
	if len(query.methods) == 0 {
		return fmt.Errorf("retrieval requires at least one method: %w", ErrInvalidQuery)
	}
	seenMethods := make(map[RetrievalMethod]struct{}, len(query.methods))
	semantic := false
	for _, method := range query.methods {
		if !method.Valid() {
			return fmt.Errorf("retrieval method %q is invalid: %w", method, ErrInvalidQuery)
		}
		if _, exists := seenMethods[method]; exists {
			return fmt.Errorf("retrieval method %q is duplicated: %w", method, ErrInvalidQuery)
		}
		seenMethods[method] = struct{}{}
		semantic = semantic || method == RetrievalSemantic
	}
	for _, documentPath := range query.paths {
		if err := documentPath.Validate(); err != nil {
			return fmt.Errorf("retrieval path: %w: %w", err, ErrInvalidQuery)
		}
	}
	for _, language := range query.languages {
		if language == "" {
			return fmt.Errorf("retrieval language cannot be empty: %w", ErrInvalidQuery)
		}
		if err := language.Validate(); err != nil {
			return fmt.Errorf("retrieval language: %w: %w", err, ErrInvalidQuery)
		}
	}
	if query.symbol != "" && strings.TrimSpace(query.symbol) != query.symbol {
		return fmt.Errorf("retrieval symbol %q is not exact: %w", query.symbol, ErrInvalidQuery)
	}
	if semantic {
		if query.embedding == nil {
			return fmt.Errorf("semantic retrieval requires an embedding target: %w", ErrInvalidQuery)
		}
		if err := query.embedding.Validate(); err != nil {
			return err
		}
	} else if query.embedding != nil {
		return fmt.Errorf("embedding target requires semantic retrieval: %w", ErrInvalidQuery)
	}
	return nil
}

type RetrievalResult struct {
	Path       DocumentPath
	Range      SourceRange
	Method     RetrievalMethod
	Score      float64
	ReasonCode string
}

func (result RetrievalResult) Validate(document Document) error {
	if result.Path != document.Path() {
		return fmt.Errorf("retrieval result references another document: %w", ErrInvalidResult)
	}
	if err := result.Range.Validate(document.SizeBytes()); err != nil {
		return fmt.Errorf("retrieval result: %w: %w", err, ErrInvalidResult)
	}
	if !result.Method.Valid() || math.IsNaN(result.Score) || math.IsInf(result.Score, 0) || !safeCode(result.ReasonCode) {
		return fmt.Errorf("retrieval result method, score, or reason is invalid: %w", ErrInvalidResult)
	}
	return nil
}
