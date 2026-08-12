package agent

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	internalTool "github.com/antonio-cafeo/maestro/internal/tool"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	"github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestProviderPlannerUsesExplicitTargetStrictJSONAndEvidence(t *testing.T) {
	completion := &completionStub{response: provider.CompletionResponse{
		Message: provider.Message{Role: provider.RoleAssistant, Content: `{"steps":[{"id":"inspect","objective":"Inspect files","dependencies":[]}]}`},
		Usage:   provider.Usage{InputTokens: 7, OutputTokens: 3},
	}}
	planner, err := NewProviderPlanner(completion, "ollama", "model")
	if err != nil {
		t.Fatal(err)
	}
	bundle := testBundle(t, "workspace")
	request, _ := pkgAgent.NewPlanningRequest("run-1", "Plan the task.", bundle, 5)
	plan, usage, err := planner.PlanMeasured(context.Background(), request)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(plan.Steps()) != 1 || usage.InputTokens != 7 || completion.provider != "ollama" || completion.request.Model != "model" {
		t.Fatalf("unexpected planner result/target: %#v %#v provider=%q", plan, usage, completion.provider)
	}
	if !strings.Contains(completion.request.Messages[1].Content, "package main") || !strings.Contains(completion.request.Messages[1].Content, "main.go") {
		t.Fatal("authorized context evidence was not included in the planning request")
	}
	if completion.request.Output == nil || completion.request.Output.Mode != provider.StructuredOutputJSONSchema {
		t.Fatal("planner did not request structured output")
	}

	completion.response.Message.Content += ` {}`
	if _, _, err := planner.PlanMeasured(context.Background(), request); !errors.Is(err, pkgAgent.ErrInvalidPlan) {
		t.Fatalf("expected trailing JSON rejection, got %v", err)
	}
}

func TestRuntimeAuthorizesProviderPlannerBeforeCompletionAndAccountsUsage(t *testing.T) {
	bundle := testBundle(t, "workspace")
	completion := &completionStub{response: provider.CompletionResponse{
		Message: provider.Message{Role: provider.RoleAssistant, Content: `{"steps":[{"id":"inspect","objective":"Inspect files","dependencies":[]}]}`},
		Usage:   provider.Usage{InputTokens: 7, OutputTokens: 3},
	}}
	planner, _ := NewProviderPlanner(completion, "ollama", "model")
	descriptor, _ := pkgAgent.NewDescriptor("agent.provider", "1", "Provider planning agent.", []pkgRuntime.Capability{pkgAgent.CapabilityRun, pkgAgent.CapabilityPlanning})
	candidate := &providerAgent{descriptor: descriptor, planner: planner}
	permissions := internalTool.NewRuntime()
	_ = permissions.RegisterPolicy(&decisionPolicy{id: "policy.test", decision: permissionDecision(t, tool.DecisionAllow, "model_allowed", "", tool.GrantOneShot)})
	options := DefaultOptions()
	options.Permissions = permissions
	runtime, _ := NewRuntimeWithOptions(&contextStub{bundle: bundle}, options)
	_ = runtime.Register(candidate)
	result, err := runtime.Run(context.Background(), runRequest(t, "run-1", "agent.provider", "workspace", 5))
	if err != nil {
		t.Fatalf("run provider planner: %v", err)
	}
	counters := result.Session().Counters()
	if counters.ModelTurns != 1 || counters.InputTokens != 7 || counters.OutputTokens != 3 || completion.calls.Load() != 1 {
		t.Fatalf("unexpected accounting/calls: %#v calls=%d", counters, completion.calls.Load())
	}

	deniedCompletion := &completionStub{response: completion.response}
	deniedPlanner, _ := NewProviderPlanner(deniedCompletion, "ollama", "model")
	deniedAgent := &providerAgent{descriptor: descriptorWithID(t, "agent.denied"), planner: deniedPlanner}
	deniedPermissions := internalTool.NewRuntime()
	_ = deniedPermissions.RegisterPolicy(&decisionPolicy{id: "policy.test", decision: permissionDecision(t, tool.DecisionDeny, "model_denied", tool.DenyTerminal, "")})
	deniedOptions := DefaultOptions()
	deniedOptions.Permissions = deniedPermissions
	deniedRuntime, _ := NewRuntimeWithOptions(&contextStub{bundle: bundle}, deniedOptions)
	_ = deniedRuntime.Register(deniedAgent)
	_, err = deniedRuntime.Run(context.Background(), runRequest(t, "run-denied", "agent.denied", "workspace", 5))
	if !errors.Is(err, pkgAgent.ErrPermissionDenied) || deniedCompletion.calls.Load() != 0 {
		t.Fatalf("provider called despite deny: err=%v calls=%d", err, deniedCompletion.calls.Load())
	}
	snapshot, _ := deniedRuntime.Session("run-denied")
	if snapshot.Terminal() != pkgAgent.TerminalPermissionDenied {
		t.Fatalf("unexpected deny terminal: %q", snapshot.Terminal())
	}
}

type completionStub struct {
	response provider.CompletionResponse
	err      error
	calls    atomic.Int32
	provider provider.ID
	request  provider.CompletionRequest
}

func (stub *completionStub) Complete(_ context.Context, providerID provider.ID, request provider.CompletionRequest) (provider.CompletionResponse, error) {
	stub.calls.Add(1)
	stub.provider = providerID
	stub.request = request
	return stub.response, stub.err
}

type providerAgent struct {
	descriptor pkgAgent.Descriptor
	planner    *ProviderPlanner
}

func (candidate *providerAgent) Descriptor() pkgAgent.Descriptor { return candidate.descriptor }
func (candidate *providerAgent) Plan(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
	return candidate.planner.Plan(ctx, request)
}
func (candidate *providerAgent) Target() (provider.ID, string) { return candidate.planner.Target() }
func (candidate *providerAgent) PlanMeasured(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, provider.Usage, error) {
	return candidate.planner.PlanMeasured(ctx, request)
}

type decisionPolicy struct {
	id       tool.PolicyID
	decision tool.Decision
}

func (policy *decisionPolicy) ID() tool.PolicyID { return policy.id }
func (policy *decisionPolicy) Decide(context.Context, tool.PermissionRequest) (tool.Decision, error) {
	return policy.decision, nil
}

func permissionDecision(t *testing.T, kind tool.DecisionKind, reason string, disposition tool.DenyDisposition, scope tool.GrantScope) tool.Decision {
	t.Helper()
	decision, err := tool.NewDecision(kind, reason, disposition, scope)
	if err != nil {
		t.Fatal(err)
	}
	return decision
}

func descriptorWithID(t *testing.T, id pkgAgent.ID) pkgAgent.Descriptor {
	t.Helper()
	descriptor, err := pkgAgent.NewDescriptor(id, "1", "Provider planning agent.", []pkgRuntime.Capability{pkgAgent.CapabilityRun, pkgAgent.CapabilityPlanning})
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}
