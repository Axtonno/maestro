package developer

import (
	"fmt"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

const rubricMethod = "deterministic_term_checklist"

func evaluateGeneration(dataset Dataset, task Task, response string) *pkgBenchmark.QualityEvaluation {
	normalized := strings.ToLower(response)
	matched := 0
	for _, criterion := range task.Criteria {
		criterionMatched := len(criterion.AllTerms) > 0
		for _, term := range criterion.AllTerms {
			if !strings.Contains(normalized, strings.ToLower(term)) {
				criterionMatched = false
				break
			}
		}
		for _, term := range criterion.AnyTerms {
			if strings.Contains(normalized, strings.ToLower(term)) {
				criterionMatched = true
				break
			}
		}
		if criterionMatched {
			matched++
		}
	}
	return qualityEvaluation(dataset, task, matched, fmt.Sprintf("criteria_matched_%d_of_3", matched))
}

func evaluateRetrieval(dataset Dataset, task Task, bestRelevantRank int) *pkgBenchmark.QualityEvaluation {
	score := 0
	switch bestRelevantRank {
	case 1:
		score = 3
	case 2:
		score = 2
	case 3:
		score = 1
	}
	rationale := "relevant_rank_outside_top_3"
	if bestRelevantRank > 0 {
		rationale = fmt.Sprintf("relevant_rank_%d", bestRelevantRank)
	}
	return qualityEvaluation(dataset, task, score, rationale)
}

func qualityEvaluation(dataset Dataset, task Task, score int, rationaleCode string) *pkgBenchmark.QualityEvaluation {
	return &pkgBenchmark.QualityEvaluation{
		Evaluator: dataset.ID + "@" + dataset.Version + ":" + task.ID,
		Method:    rubricMethod, Score: score, MaxScore: 3,
		RationaleCode: rationaleCode,
	}
}
