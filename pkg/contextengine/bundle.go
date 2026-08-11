package contextengine

import (
	"context"
	"fmt"
	"slices"
)

type Budget struct {
	MaxTokens      int
	ReservedTokens int
	SafetyTokens   int
}

func (budget Budget) Validate() error {
	if budget.MaxTokens <= 0 || budget.ReservedTokens < 0 || budget.SafetyTokens < 0 ||
		budget.ReservedTokens+budget.SafetyTokens >= budget.MaxTokens {
		return fmt.Errorf("token budget has no positive evidence allowance: %w", ErrInvalidBudget)
	}
	return nil
}

func (budget Budget) EvidenceTokens() int {
	return budget.MaxTokens - budget.ReservedTokens - budget.SafetyTokens
}

type TokenEstimator interface {
	ID() EstimatorID
	Version() string
	Estimate(context.Context, string) (int, error)
}

type ContextSection struct {
	Path       DocumentPath
	Range      SourceRange
	Role       string
	Method     RetrievalMethod
	ReasonCode string
	Text       string
	Tokens     int
	Truncated  bool
}

func (section ContextSection) Validate(document Document) error {
	if section.Path != document.Path() || section.Text == "" || section.Tokens <= 0 ||
		!safeCode(section.Role) || !section.Method.Valid() || !safeCode(section.ReasonCode) {
		return fmt.Errorf("context section identity, role, method, reason, text, or cost is invalid: %w", ErrInvalidSection)
	}
	if err := section.Range.Validate(document.SizeBytes()); err != nil {
		return fmt.Errorf("context section range: %w: %w", err, ErrInvalidSection)
	}
	if document.Content()[section.Range.Start:section.Range.End] != section.Text {
		return fmt.Errorf("context section text does not match its source range: %w", ErrInvalidSection)
	}
	return nil
}

type ContextBundle struct {
	workspace        WorkspaceID
	generation       uint64
	estimator        EstimatorID
	estimatorVersion string
	budget           Budget
	usedTokens       int
	sections         []ContextSection
}

func NewContextBundle(
	snapshot Snapshot,
	estimator EstimatorID,
	estimatorVersion string,
	budget Budget,
	sections []ContextSection,
) (ContextBundle, error) {
	if err := estimator.Validate(); err != nil {
		return ContextBundle{}, fmt.Errorf("bundle estimator: %w: %w", err, ErrInvalidBundle)
	}
	if !exactID(estimatorVersion) {
		return ContextBundle{}, fmt.Errorf("bundle estimator version %q is invalid: %w", estimatorVersion, ErrInvalidBundle)
	}
	if err := budget.Validate(); err != nil {
		return ContextBundle{}, fmt.Errorf("bundle budget: %w: %w", err, ErrInvalidBundle)
	}
	cloned := slices.Clone(sections)
	used := 0
	for index, section := range cloned {
		document, found := snapshot.Document(section.Path)
		if !found {
			return ContextBundle{}, fmt.Errorf("bundle section %d document %q is absent: %w", index, section.Path, ErrInvalidBundle)
		}
		if err := section.Validate(document); err != nil {
			return ContextBundle{}, fmt.Errorf("bundle section %d: %w: %w", index, err, ErrInvalidBundle)
		}
		used += section.Tokens
	}
	if used > budget.EvidenceTokens() {
		return ContextBundle{}, fmt.Errorf("bundle uses %d evidence tokens with allowance %d: %w: %w", used, budget.EvidenceTokens(), ErrBudgetExceeded, ErrInvalidBundle)
	}
	metadata := snapshot.Metadata()
	return ContextBundle{
		workspace: metadata.Workspace, generation: metadata.Generation,
		estimator: estimator, estimatorVersion: estimatorVersion,
		budget: budget, usedTokens: used, sections: cloned,
	}, nil
}

func (bundle ContextBundle) Workspace() WorkspaceID     { return bundle.workspace }
func (bundle ContextBundle) Generation() uint64         { return bundle.generation }
func (bundle ContextBundle) Estimator() EstimatorID     { return bundle.estimator }
func (bundle ContextBundle) EstimatorVersion() string   { return bundle.estimatorVersion }
func (bundle ContextBundle) Budget() Budget             { return bundle.budget }
func (bundle ContextBundle) UsedTokens() int            { return bundle.usedTokens }
func (bundle ContextBundle) Sections() []ContextSection { return slices.Clone(bundle.sections) }

type BuildRequest struct {
	Query     RetrievalQuery
	Budget    Budget
	Estimator EstimatorID
}

func (request BuildRequest) Validate() error {
	if err := request.Query.Validate(); err != nil {
		return err
	}
	if err := request.Budget.Validate(); err != nil {
		return err
	}
	if err := request.Estimator.Validate(); err != nil {
		return err
	}
	return nil
}
