package contextengine

import (
	"context"
	"fmt"
	"sort"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

func buildBundle(ctx context.Context, snapshot pkgContext.Snapshot, estimator pkgContext.TokenEstimator, budget pkgContext.Budget, results []pkgContext.RetrievalResult, cache *artifactCache) (pkgContext.ContextBundle, error) {
	remaining := budget.EvidenceTokens()
	sections := make([]pkgContext.ContextSection, 0, len(results))
	selected := make(map[pkgContext.DocumentPath][]pkgContext.SourceRange)
	for _, result := range results {
		if remaining <= 0 {
			break
		}
		if overlapsAny(result.Range, selected[result.Path]) {
			continue
		}
		document, found := snapshot.Document(result.Path)
		if !found {
			return pkgContext.ContextBundle{}, fmt.Errorf("build result document %q: %w", result.Path, pkgContext.ErrNotFound)
		}
		text := document.Content()[result.Range.Start:result.Range.End]
		cost, err := estimateSafely(ctx, estimator, text, cache)
		if err != nil {
			return pkgContext.ContextBundle{}, err
		}
		selectedRange := result.Range
		truncated := false
		if cost > remaining {
			text, cost, err = truncateToBudget(ctx, estimator, text, remaining, cache)
			if err != nil {
				return pkgContext.ContextBundle{}, err
			}
			if text == "" {
				continue
			}
			selectedRange.End = selectedRange.Start + len(text)
			truncated = true
		}
		section := pkgContext.ContextSection{
			Path: result.Path, Range: selectedRange, Role: "evidence", Method: result.Method,
			ReasonCode: result.ReasonCode, Text: text, Tokens: cost, Truncated: truncated,
		}
		if err := section.Validate(document); err != nil {
			return pkgContext.ContextBundle{}, err
		}
		sections = append(sections, section)
		selected[result.Path] = append(selected[result.Path], selectedRange)
		remaining -= cost
	}
	return pkgContext.NewContextBundle(snapshot, estimator.ID(), estimator.Version(), budget, sections)
}

func estimateSafely(ctx context.Context, estimator pkgContext.TokenEstimator, text string, cache *artifactCache) (cost int, err error) {
	key := estimatorCacheKey(estimator, text)
	if cached, found := cache.get(key); found {
		return cached.(int), nil
	}
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("token estimator %q panicked: %w", estimator.ID(), pkgContext.ErrEstimatorFailure)
		}
	}()
	cost, err = estimator.Estimate(ctx, text)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, ctxErr
		}
		return 0, fmt.Errorf("token estimator %q: %w: %w", estimator.ID(), err, pkgContext.ErrEstimatorFailure)
	}
	if cost < 0 || (text != "" && cost == 0) {
		return 0, fmt.Errorf("token estimator %q returned invalid cost %d: %w", estimator.ID(), cost, pkgContext.ErrEstimatorFailure)
	}
	cache.put(key, cost, int64(len(key)+8))
	return cost, nil
}

func truncateToBudget(ctx context.Context, estimator pkgContext.TokenEstimator, text string, allowance int, cache *artifactCache) (string, int, error) {
	if allowance <= 0 {
		return "", 0, nil
	}
	boundaries := []int{0}
	for index := range text {
		if index > 0 {
			boundaries = append(boundaries, index)
		}
	}
	boundaries = append(boundaries, len(text))
	boundaries = compactBoundaries(boundaries)
	bestEnd, bestCost := 0, 0
	left, right := 1, len(boundaries)-1
	for left <= right {
		middle := left + (right-left)/2
		end := boundaries[middle]
		if !utf8.ValidString(text[:end]) {
			return "", 0, fmt.Errorf("truncation produced invalid UTF-8: %w", pkgContext.ErrEstimatorFailure)
		}
		cost, err := estimateSafely(ctx, estimator, text[:end], cache)
		if err != nil {
			return "", 0, err
		}
		if cost <= allowance {
			bestEnd, bestCost = end, cost
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	return text[:bestEnd], bestCost, nil
}

func compactBoundaries(values []int) []int {
	sort.Ints(values)
	output := values[:0]
	for _, value := range values {
		if len(output) == 0 || output[len(output)-1] != value {
			output = append(output, value)
		}
	}
	return output
}

func overlapsAny(candidate pkgContext.SourceRange, selected []pkgContext.SourceRange) bool {
	for _, existing := range selected {
		if candidate.Start < existing.End && existing.Start < candidate.End {
			return true
		}
	}
	return false
}
