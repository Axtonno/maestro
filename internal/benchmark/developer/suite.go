package developer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/antonio-cafeo/maestro/internal/benchmark/smoke"
	internalContext "github.com/antonio-cafeo/maestro/internal/contextengine"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type Suite struct {
	dataset Dataset
	config  smoke.Config
	runtime pkgProvider.Runtime
}

func NewScenarios(
	manifest pkgBenchmark.Manifest,
	dataset Dataset,
	config smoke.Config,
	runtime pkgProvider.Runtime,
) ([]pkgBenchmark.Scenario, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	if err := dataset.validate(); err != nil {
		return nil, err
	}
	if nilProviderRuntime(runtime) {
		return nil, errors.New("developer benchmark provider runtime is nil")
	}
	suite := &Suite{dataset: dataset, config: config, runtime: runtime}
	scenarios := make([]pkgBenchmark.Scenario, 0, len(manifest.Scenarios))
	for _, definition := range manifest.Scenarios {
		task, exists := dataset.Task(definition.ID)
		if !exists {
			return nil, fmt.Errorf("developer benchmark scenario %q has no dataset task", definition.ID)
		}
		if task.ModelRole != definition.ModelRole {
			return nil, fmt.Errorf("developer benchmark scenario %q model role differs from dataset", definition.ID)
		}
		switch task.Kind {
		case "generation":
			scenarios = append(scenarios, suite.generationScenario(definition, task))
		case "retrieval":
			scenarios = append(scenarios, suite.retrievalScenario(definition, task))
		default:
			return nil, fmt.Errorf("developer benchmark task %q has unknown kind %q", task.ID, task.Kind)
		}
	}
	return scenarios, nil
}

func nilProviderRuntime(runtime pkgProvider.Runtime) bool {
	if runtime == nil {
		return true
	}
	value := reflect.ValueOf(runtime)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (s *Suite) generationScenario(definition pkgBenchmark.ScenarioDefinition, task Task) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, task.ModelRole, pkgProvider.CapabilityCompletion)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			prompt, err := s.dataset.Prompt(task)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			started := time.Now()
			response, err := s.runtime.Complete(ctx, s.config.ProviderID, pkgProvider.CompletionRequest{
				Model:    model,
				Messages: []pkgProvider.Message{{Role: pkgProvider.RoleUser, Content: prompt}},
				Options:  pkgProvider.GenerationOptions{MaxTokens: 1024},
			})
			elapsed := time.Since(started)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			content := strings.TrimSpace(response.Message.Content)
			if content == "" {
				return failed("developer_response_empty"), nil
			}
			evaluation := evaluateGeneration(s.dataset, task, content)
			return pkgBenchmark.IterationResult{
				State: pkgBenchmark.ResultPassed,
				Measurements: []pkgBenchmark.Measurement{
					milliseconds("total_latency_ms", elapsed),
					count("input_tokens", response.Usage.InputTokens),
					count("output_tokens", response.Usage.OutputTokens),
					count("response_bytes", len([]byte(content))),
					scoreMeasurement(evaluation.Score),
				},
				Evaluation: evaluation,
			}, nil
		},
	}
}

func (s *Suite) retrievalScenario(definition pkgBenchmark.ScenarioDefinition, task Task) pkgBenchmark.Scenario {
	return pkgBenchmark.ScenarioFuncs{
		DefinitionValue: definition,
		RunFunc: func(ctx context.Context, _ pkgBenchmark.Iteration) (pkgBenchmark.IterationResult, error) {
			model, preflight, err := s.requireModel(ctx, task.ModelRole, pkgProvider.CapabilityEmbedding)
			if err != nil || preflight != nil {
				return resultOrZero(preflight), err
			}
			documents := make([]pkgContext.Document, 0, len(task.Files))
			for _, name := range task.Files {
				content, err := s.dataset.File(name)
				if err != nil {
					return pkgBenchmark.IterationResult{}, err
				}
				document, err := pkgContext.NewDocument(
					pkgContext.DocumentPath(name),
					"text/x-php",
					"php",
					"Fixture file "+name+"\n"+string(content),
				)
				if err != nil {
					return pkgBenchmark.IterationResult{}, err
				}
				documents = append(documents, document)
			}
			embedding := &observedEmbeddingRuntime{runtime: s.runtime}
			contextEngine := internalContext.NewWithEmbeddingRuntime(embedding)
			source := developerContextSource{documents: documents}
			if err := contextEngine.RegisterSource(source); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			workspace, err := pkgContext.NewWorkspace(
				"developer-benchmark",
				"/developer-benchmark",
				pkgContext.WorkspaceOptions{Source: source.ID(), Policy: pkgContext.DefaultScanPolicy()},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if _, err := contextEngine.Index(ctx, workspace); err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			target := pkgContext.EmbeddingTarget{Provider: string(s.config.ProviderID), Model: model}
			query, err := pkgContext.NewRetrievalQuery(
				workspace.ID(),
				task.Instruction,
				pkgContext.RetrievalQueryOptions{
					Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalSemantic},
					TopK:    len(task.Files), Embedding: &target,
				},
			)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			started := time.Now()
			results, err := contextEngine.Retrieve(ctx, query)
			elapsed := time.Since(started)
			if err != nil {
				return pkgBenchmark.IterationResult{}, err
			}
			if len(results) != len(task.Files) || embedding.Dimension() == 0 {
				return failed("developer_embedding_shape_invalid"), nil
			}
			ranking := make([]rankedDocument, 0, len(results))
			for _, result := range results {
				ranking = append(ranking, rankedDocument{name: string(result.Path), similarity: result.Score})
			}
			bestRank := bestRelevantRank(ranking, task.RelevantFiles)
			evaluation := evaluateRetrieval(s.dataset, task, bestRank)
			return pkgBenchmark.IterationResult{
				State: pkgBenchmark.ResultPassed,
				Measurements: []pkgBenchmark.Measurement{
					milliseconds("embedding_latency_ms", elapsed),
					count("document_count", len(task.Files)),
					count("embedding_dimension", embedding.Dimension()),
					count("best_relevant_rank", bestRank),
					scoreMeasurement(evaluation.Score),
				},
				Evaluation: evaluation,
			}, nil
		},
	}
}

type developerContextSource struct {
	documents []pkgContext.Document
}

func (developerContextSource) ID() pkgContext.SourceID { return "benchmark.developer" }
func (source developerContextSource) Scan(context.Context, pkgContext.Workspace) (pkgContext.ScanResult, error) {
	return pkgContext.ScanResult{Documents: slices.Clone(source.documents)}, nil
}

type observedEmbeddingRuntime struct {
	runtime   pkgProvider.Runtime
	dimension atomic.Int64
}

func (runtime *observedEmbeddingRuntime) Embed(ctx context.Context, id pkgProvider.ID, request pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error) {
	response, err := runtime.runtime.Embed(ctx, id, request)
	if err == nil && len(response.Embeddings) > 0 {
		runtime.dimension.Store(int64(len(response.Embeddings[0])))
	}
	return response, err
}

func (runtime *observedEmbeddingRuntime) Dimension() int {
	return int(runtime.dimension.Load())
}

func (s *Suite) requireModel(ctx context.Context, role string, capability pkgProvider.Capability) (string, *pkgBenchmark.IterationResult, error) {
	if !s.config.Enabled {
		result := skipped("provider_not_configured")
		return "", &result, nil
	}
	model := strings.TrimSpace(s.config.Models[role])
	if model == "" {
		result := skipped("model_not_configured")
		return "", &result, nil
	}
	report, err := s.runtime.Capabilities(ctx, s.config.ProviderID, pkgProvider.CapabilityRequest{
		Target: pkgProvider.CapabilityTargetModel, Model: model,
	})
	if err != nil {
		return "", nil, err
	}
	for _, descriptor := range report.Capabilities {
		if descriptor.Capability != capability {
			continue
		}
		if descriptor.Support == pkgProvider.CapabilityUnsupported {
			result := unsupported("capability_not_supported")
			return "", &result, nil
		}
		if descriptor.Availability == pkgProvider.CapabilityAvailabilityUnavailable {
			result := skipped("capability_unavailable")
			return "", &result, nil
		}
		return model, nil, nil
	}
	return "", nil, fmt.Errorf("capability report omitted %q", capability)
}

type rankedDocument struct {
	name       string
	similarity float64
}

func rankDocuments(query []float32, documents [][]float32, names []string) ([]rankedDocument, error) {
	if len(documents) != len(names) {
		return nil, errors.New("embedding document count differs from file count")
	}
	ranking := make([]rankedDocument, len(documents))
	for index, document := range documents {
		similarity, err := cosineSimilarity(query, document)
		if err != nil {
			return nil, err
		}
		ranking[index] = rankedDocument{name: names[index], similarity: similarity}
	}
	sort.SliceStable(ranking, func(left, right int) bool {
		if ranking[left].similarity == ranking[right].similarity {
			return ranking[left].name < ranking[right].name
		}
		return ranking[left].similarity > ranking[right].similarity
	})
	return ranking, nil
}

func cosineSimilarity(left, right []float32) (float64, error) {
	if len(left) == 0 || len(left) != len(right) {
		return 0, errors.New("embedding dimensions differ")
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		leftValue := float64(left[index])
		rightValue := float64(right[index])
		if math.IsNaN(leftValue) || math.IsInf(leftValue, 0) || math.IsNaN(rightValue) || math.IsInf(rightValue, 0) {
			return 0, errors.New("embedding contains non-finite values")
		}
		dot += leftValue * rightValue
		leftNorm += leftValue * leftValue
		rightNorm += rightValue * rightValue
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0, errors.New("embedding norm is zero")
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)), nil
}

func bestRelevantRank(ranking []rankedDocument, relevant []string) int {
	set := make(map[string]struct{}, len(relevant))
	for _, name := range relevant {
		set[name] = struct{}{}
	}
	for index, document := range ranking {
		if _, exists := set[document.name]; exists {
			return index + 1
		}
	}
	return 0
}

func milliseconds(name string, value time.Duration) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Value: float64(value) / float64(time.Millisecond), Unit: "ms", Method: "monotonic_clock"}
}

func count(name string, value int) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: name, Value: float64(value), Unit: "count"}
}

func scoreMeasurement(score int) pkgBenchmark.Measurement {
	return pkgBenchmark.Measurement{Name: "quality_score", Value: float64(score), Unit: "score_0_3", Method: rubricMethod}
}

func failed(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultFailed, Error: &pkgBenchmark.ErrorRecord{Kind: "developer_gate", Code: code}}
}

func skipped(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultSkipped, ReasonCode: code}
}

func unsupported(code string) pkgBenchmark.IterationResult {
	return pkgBenchmark.IterationResult{State: pkgBenchmark.ResultUnsupported, ReasonCode: code}
}

func resultOrZero(result *pkgBenchmark.IterationResult) pkgBenchmark.IterationResult {
	if result == nil {
		return pkgBenchmark.IterationResult{}
	}
	return *result
}
