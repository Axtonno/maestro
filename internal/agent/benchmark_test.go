package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	internalTool "github.com/antonio-cafeo/maestro/internal/tool"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func BenchmarkAgentLoopDeterministic(b *testing.B) {
	workspace, _ := pkgContext.NewWorkspace("workspace", b.TempDir(), pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	document, _ := pkgContext.NewDocument("main.go", "text/plain", "", "package main\n")
	snapshot, _ := pkgContext.NewSnapshot(workspace, 1, []pkgContext.Document{document}, nil, nil)
	section := pkgContext.ContextSection{Path: "main.go", Range: pkgContext.SourceRange{Start: 0, End: document.SizeBytes()}, Role: "evidence", Method: pkgContext.RetrievalLexical, ReasonCode: "term_match", Text: document.Content(), Tokens: 3}
	bundle, _ := pkgContext.NewContextBundle(snapshot, "context.test", "1", pkgContext.Budget{MaxTokens: 100, ReservedTokens: 10, SafetyTokens: 10}, []pkgContext.ContextSection{section})
	query, _ := pkgContext.NewRetrievalQuery("workspace", "task", pkgContext.RetrievalQueryOptions{Methods: []pkgContext.RetrievalMethod{pkgContext.RetrievalLexical}, TopK: 1})
	descriptor, _ := pkgTool.NewDescriptor("fixture.noop", "fixture_noop", "1", "Benchmark no-op tool.", json.RawMessage(`{"type":"object"}`), []pkgTool.Effect{pkgTool.EffectLocalCompute})

	b.ResetTimer()
	for range b.N {
		tools := internalTool.NewRuntime()
		_ = tools.Register(benchmarkTool{descriptor: descriptor})
		allow, _ := pkgTool.NewDecision(pkgTool.DecisionAllow, "benchmark_allow", "", pkgTool.GrantRun)
		_ = tools.RegisterPolicy(&decisionPolicy{id: "policy.benchmark", decision: allow})
		options := DefaultOptions()
		options.Providers = benchmarkProvider{}
		options.Tools = tools
		runtime, _ := NewRuntimeWithOptions(&contextStub{bundle: bundle}, options)
		_ = runtime.Register(NewReferenceAgent())
		request, _ := pkgAgent.NewRunRequest(
			"run-benchmark", ReferenceAgentID, "benchmark", "model", "workspace", "policy.benchmark", "Complete the step.",
			pkgAgent.Limits{MaxDuration: time.Second, MaxModelTurns: 2, MaxToolCalls: 1, MaxToolCallsPerTurn: 1, MaxPlanSteps: 1, MaxPlanRevisions: 1, MaxToolResultBytes: 1024, MaxSessionBytes: 1 << 16, MaxInputTokens: 100, MaxOutputTokens: 100},
			pkgAgent.RunRequestOptions{Context: pkgContext.BuildRequest{Query: query, Budget: pkgContext.Budget{MaxTokens: 100, ReservedTokens: 10, SafetyTokens: 10}, Estimator: "context.test"}, Tools: []pkgTool.ID{"fixture.noop"}},
		)
		if _, err := runtime.Run(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

type benchmarkProvider struct{}

func (benchmarkProvider) Complete(context.Context, provider.ID, provider.CompletionRequest) (provider.CompletionResponse, error) {
	return provider.CompletionResponse{Message: provider.Message{Role: provider.RoleAssistant, Content: "done"}, FinishReason: provider.FinishReasonStop}, nil
}
func (benchmarkProvider) Stream(context.Context, provider.ID, provider.CompletionRequest) (provider.Stream, error) {
	return nil, context.Canceled
}

type benchmarkTool struct{ descriptor pkgTool.Descriptor }

func (tool benchmarkTool) Descriptor() pkgTool.Descriptor { return tool.descriptor }
func (tool benchmarkTool) Prepare(_ context.Context, invocation pkgTool.Invocation) (pkgTool.PreparedInvocation, error) {
	action, _ := pkgTool.NewAction(pkgTool.EffectLocalCompute, "benchmark", "")
	return pkgTool.NewPreparedInvocation(invocation, tool.descriptor.Version(), invocation.Arguments(), []pkgTool.Action{action})
}
func (benchmarkTool) Execute(context.Context, pkgTool.PreparedInvocation) (pkgTool.Result, error) {
	return pkgTool.NewResult(pkgTool.ResultSuccess, "ok", "text/plain", "completed", 1, false, "")
}
