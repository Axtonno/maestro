package contextengine

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type retrievalCandidate struct {
	path     pkgContext.DocumentPath
	range_   pkgContext.SourceRange
	text     string
	language pkgContext.Language
}

func retrieveSnapshot(ctx context.Context, snapshot pkgContext.Snapshot, query pkgContext.RetrievalQuery, embedding embeddingRuntime) ([]pkgContext.RetrievalResult, error) {
	candidates := retrievalCandidates(snapshot, query)
	byMethod := make([][]pkgContext.RetrievalResult, 0, len(query.Methods()))
	for _, method := range query.Methods() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var results []pkgContext.RetrievalResult
		var err error
		switch method {
		case pkgContext.RetrievalLexical:
			results = lexicalResults(query, candidates)
		case pkgContext.RetrievalStructured:
			results = structuredResults(snapshot, query)
		case pkgContext.RetrievalSemantic:
			results, err = semanticResults(ctx, query, candidates, embedding)
		}
		if err != nil {
			return nil, err
		}
		sortResults(results)
		byMethod = append(byMethod, results)
	}
	var results []pkgContext.RetrievalResult
	if len(byMethod) == 1 {
		results = byMethod[0]
	} else {
		results = fuseResults(byMethod)
		sortResults(results)
	}
	if len(results) > query.TopK() {
		results = results[:query.TopK()]
	}
	for index, result := range results {
		document, found := snapshot.Document(result.Path)
		if !found {
			return nil, fmt.Errorf("retrieval result %d references missing document: %w", index, pkgContext.ErrInvalidResult)
		}
		if err := result.Validate(document); err != nil {
			return nil, err
		}
	}
	return slices.Clone(results), nil
}

func retrievalCandidates(snapshot pkgContext.Snapshot, query pkgContext.RetrievalQuery) []retrievalCandidate {
	pathFilter := make(map[pkgContext.DocumentPath]struct{}, len(query.Paths()))
	for _, path := range query.Paths() {
		pathFilter[path] = struct{}{}
	}
	languageFilter := make(map[pkgContext.Language]struct{}, len(query.Languages()))
	for _, language := range query.Languages() {
		languageFilter[language] = struct{}{}
	}
	chunks := make(map[pkgContext.DocumentPath][]pkgContext.SourceRange)
	for _, analysis := range snapshot.Analyses() {
		for _, chunk := range analysis.Chunks() {
			ranges := chunks[analysis.Path()]
			if !slices.Contains(ranges, chunk.Range) {
				chunks[analysis.Path()] = append(ranges, chunk.Range)
			}
		}
	}
	candidates := make([]retrievalCandidate, 0)
	for _, document := range snapshot.Documents() {
		if document.MediaType() == "application/octet-stream" {
			continue
		}
		if len(pathFilter) > 0 {
			if _, allowed := pathFilter[document.Path()]; !allowed {
				continue
			}
		}
		if len(languageFilter) > 0 {
			if _, allowed := languageFilter[document.Language()]; !allowed {
				continue
			}
		}
		ranges := chunks[document.Path()]
		if len(ranges) == 0 && document.SizeBytes() > 0 {
			ranges = []pkgContext.SourceRange{{Start: 0, End: document.SizeBytes()}}
		}
		slices.SortFunc(ranges, compareRange)
		for _, sourceRange := range ranges {
			candidates = append(candidates, retrievalCandidate{
				path: document.Path(), range_: sourceRange,
				text: document.Content()[sourceRange.Start:sourceRange.End], language: document.Language(),
			})
		}
	}
	return candidates
}

func lexicalResults(query pkgContext.RetrievalQuery, candidates []retrievalCandidate) []pkgContext.RetrievalResult {
	terms := uniqueTerms(query.Text())
	results := make([]pkgContext.RetrievalResult, 0)
	for _, candidate := range candidates {
		contentTerms := make(map[string]struct{})
		for _, term := range tokenize(candidate.text) {
			contentTerms[term] = struct{}{}
		}
		matched := 0
		for _, term := range terms {
			if _, exists := contentTerms[term]; exists {
				matched++
			}
		}
		if matched == 0 {
			continue
		}
		results = append(results, pkgContext.RetrievalResult{
			Path: candidate.path, Range: candidate.range_, Method: pkgContext.RetrievalLexical,
			Score: float64(matched) / float64(len(terms)), ReasonCode: "term_coverage",
		})
	}
	return results
}

func structuredResults(snapshot pkgContext.Snapshot, query pkgContext.RetrievalQuery) []pkgContext.RetrievalResult {
	terms := uniqueTerms(query.Text())
	results := make([]pkgContext.RetrievalResult, 0)
	allowedPaths := make(map[pkgContext.DocumentPath]struct{}, len(query.Paths()))
	for _, path := range query.Paths() {
		allowedPaths[path] = struct{}{}
	}
	allowedLanguages := make(map[pkgContext.Language]struct{}, len(query.Languages()))
	for _, language := range query.Languages() {
		allowedLanguages[language] = struct{}{}
	}
	for _, analysis := range snapshot.Analyses() {
		document, found := snapshot.Document(analysis.Path())
		if !found || document.MediaType() == "application/octet-stream" {
			continue
		}
		if len(allowedPaths) > 0 {
			if _, allowed := allowedPaths[document.Path()]; !allowed {
				continue
			}
		}
		if len(allowedLanguages) > 0 {
			if _, allowed := allowedLanguages[document.Language()]; !allowed {
				continue
			}
		}
		for _, symbol := range analysis.Symbols() {
			score := 0.0
			reason := "symbol_term_match"
			if query.Symbol() != "" && strings.EqualFold(symbol.Name, query.Symbol()) {
				score, reason = 1, "symbol_exact"
			} else if query.Symbol() == "" {
				name := strings.ToLower(symbol.Name)
				for _, term := range terms {
					if name == term {
						score = 0.9
						break
					}
				}
			}
			if score > 0 {
				results = append(results, pkgContext.RetrievalResult{
					Path: document.Path(), Range: symbol.Range, Method: pkgContext.RetrievalStructured,
					Score: score, ReasonCode: reason,
				})
			}
		}
	}
	return results
}

func semanticResults(ctx context.Context, query pkgContext.RetrievalQuery, candidates []retrievalCandidate, runtime embeddingRuntime) ([]pkgContext.RetrievalResult, error) {
	if runtime == nil {
		return nil, fmt.Errorf("semantic retrieval has no embedding runtime: %w", pkgContext.ErrUnsupported)
	}
	target, _ := query.Embedding()
	inputs := make([]string, 1, len(candidates)+1)
	inputs[0] = query.Text()
	for _, candidate := range candidates {
		inputs = append(inputs, candidate.text)
	}
	response, err := runtime.Embed(ctx, pkgProvider.ID(target.Provider), pkgProvider.EmbeddingRequest{Model: target.Model, Inputs: inputs})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("embed retrieval inputs: %w: %w", err, pkgContext.ErrEmbeddingFailure)
	}
	if len(response.Embeddings) != len(inputs) || len(response.Embeddings) == 0 {
		return nil, fmt.Errorf("embedding cardinality %d differs from %d: %w", len(response.Embeddings), len(inputs), pkgContext.ErrEmbeddingFailure)
	}
	queryVector := response.Embeddings[0]
	results := make([]pkgContext.RetrievalResult, 0, len(candidates))
	for index, candidate := range candidates {
		score, err := cosineSimilarity(queryVector, response.Embeddings[index+1])
		if err != nil {
			return nil, fmt.Errorf("embedding %d: %w: %w", index+1, err, pkgContext.ErrEmbeddingFailure)
		}
		results = append(results, pkgContext.RetrievalResult{
			Path: candidate.path, Range: candidate.range_, Method: pkgContext.RetrievalSemantic,
			Score: score, ReasonCode: "cosine_similarity",
		})
	}
	return results, nil
}

func cosineSimilarity(left, right []float32) (float64, error) {
	if len(left) == 0 || len(left) != len(right) {
		return 0, fmt.Errorf("embedding dimensions differ")
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		leftValue, rightValue := float64(left[index]), float64(right[index])
		if math.IsNaN(leftValue) || math.IsInf(leftValue, 0) || math.IsNaN(rightValue) || math.IsInf(rightValue, 0) {
			return 0, fmt.Errorf("embedding contains non-finite values")
		}
		dot += leftValue * rightValue
		leftNorm += leftValue * leftValue
		rightNorm += rightValue * rightValue
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0, fmt.Errorf("embedding norm is zero")
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)), nil
}

func fuseResults(groups [][]pkgContext.RetrievalResult) []pkgContext.RetrievalResult {
	type key struct {
		path       pkgContext.DocumentPath
		start, end int
	}
	merged := make(map[key]pkgContext.RetrievalResult)
	for _, group := range groups {
		for index, result := range group {
			key := key{path: result.Path, start: result.Range.Start, end: result.Range.End}
			current := merged[key]
			current.Path, current.Range = result.Path, result.Range
			current.Method, current.ReasonCode = pkgContext.RetrievalFused, "reciprocal_rank_fusion"
			current.Score += 1 / float64(60+index+1)
			merged[key] = current
		}
	}
	results := make([]pkgContext.RetrievalResult, 0, len(merged))
	for _, result := range merged {
		results = append(results, result)
	}
	return results
}

func sortResults(results []pkgContext.RetrievalResult) {
	slices.SortFunc(results, func(left, right pkgContext.RetrievalResult) int {
		if left.Score > right.Score {
			return -1
		}
		if left.Score < right.Score {
			return 1
		}
		if result := left.Path.Compare(right.Path); result != 0 {
			return result
		}
		if left.Range.Start < right.Range.Start {
			return -1
		}
		if left.Range.Start > right.Range.Start {
			return 1
		}
		return compareRange(left.Range, right.Range)
	})
}

func compareRange(left, right pkgContext.SourceRange) int {
	if left.Start < right.Start {
		return -1
	}
	if left.Start > right.Start {
		return 1
	}
	if left.End < right.End {
		return -1
	}
	if left.End > right.End {
		return 1
	}
	return 0
}

func uniqueTerms(text string) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0)
	for _, term := range tokenize(text) {
		if _, exists := seen[term]; exists {
			continue
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '_'
	})
}
