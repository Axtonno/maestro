package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestPublicContractAssertions(t *testing.T) {
	var _ tool.Tool = toolStub{}
	var _ tool.Policy = policyStub{}
	var _ tool.Approver = (*approverStub)(nil)
	var _ tool.Runtime = runtimeStub{}
}

func TestExtensionValidationRejectsTypedNil(t *testing.T) {
	var candidate *toolStub
	if !errors.Is(tool.ValidateTool(candidate), tool.ErrInvalidTool) {
		t.Fatalf("expected typed nil tool rejection, got %v", tool.ValidateTool(candidate))
	}
	var policy *policyStub
	if !errors.Is(tool.ValidatePolicy(policy), tool.ErrInvalidPolicy) {
		t.Fatalf("expected typed nil policy rejection, got %v", tool.ValidatePolicy(policy))
	}
}

func TestDescriptorValidationAndDefensiveCopies(t *testing.T) {
	parameters := json.RawMessage(`{"b":2,"a":1}`)
	effects := []tool.Effect{tool.EffectWorkspaceMutate, tool.EffectWorkspaceInspect}
	descriptor, err := tool.NewDescriptor(
		"workspace.file", "workspace_file", "1.0.0", "Read or update a workspace file.",
		parameters, effects,
	)
	if err != nil {
		t.Fatalf("construct descriptor: %v", err)
	}
	parameters[2] = 'x'
	effects[0] = tool.EffectNetworkAccess
	if got := string(descriptor.Parameters()); got != `{"a":1,"b":2}` {
		t.Fatalf("parameters are not canonical and defensive: %s", got)
	}
	if got := descriptor.Effects(); !reflect.DeepEqual(got, []tool.Effect{tool.EffectWorkspaceInspect, tool.EffectWorkspaceMutate}) {
		t.Fatalf("effects are not ordered and defensive: %#v", got)
	}
	returned := descriptor.Parameters()
	returned[2] = 'z'
	if string(descriptor.Parameters()) != `{"a":1,"b":2}` {
		t.Fatal("descriptor exposes parameter storage")
	}

	invalid := []struct {
		id      tool.ID
		name    tool.Name
		version tool.Version
		schema  json.RawMessage
		effects []tool.Effect
		is      error
	}{
		{id: "workspace", name: "valid", version: "1", schema: json.RawMessage(`{}`), effects: []tool.Effect{tool.EffectLocalCompute}, is: tool.ErrInvalidToolID},
		{id: "workspace.file", name: "bad name", version: "1", schema: json.RawMessage(`{}`), effects: []tool.Effect{tool.EffectLocalCompute}, is: tool.ErrInvalidToolName},
		{id: "workspace.file", name: "valid", version: "bad version", schema: json.RawMessage(`{}`), effects: []tool.Effect{tool.EffectLocalCompute}, is: tool.ErrInvalidVersion},
		{id: "workspace.file", name: "valid", version: "1", schema: json.RawMessage(`{} {}`), effects: []tool.Effect{tool.EffectLocalCompute}, is: tool.ErrInvalidDescriptor},
		{id: "workspace.file", name: "valid", version: "1", schema: json.RawMessage(`{}`), is: tool.ErrInvalidDescriptor},
	}
	for _, test := range invalid {
		_, err := tool.NewDescriptor(test.id, test.name, test.version, "description", test.schema, test.effects)
		if !errors.Is(err, test.is) {
			t.Errorf("expected %v, got %v", test.is, err)
		}
	}
}

func TestPreparedInvocationBindsIdentityArgumentsActionsAndRun(t *testing.T) {
	raw := json.RawMessage(`{"path":"src/main.go","mode":"read"}`)
	invocation, err := tool.NewInvocation("workspace.file", "call-1", "run-1", raw)
	if err != nil {
		t.Fatalf("construct invocation: %v", err)
	}
	raw[2] = 'x'
	action, _ := tool.NewAction(tool.EffectWorkspaceInspect, "src/main.go", "workspace")
	prepared, err := tool.NewPreparedInvocation(
		invocation, "1.0.0", json.RawMessage(`{"mode":"read","path":"src/main.go"}`), []tool.Action{action},
	)
	if err != nil {
		t.Fatalf("construct prepared invocation: %v", err)
	}
	if err := prepared.Fingerprint().Validate(); err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	changed, _ := tool.NewInvocation("workspace.file", "call-2", "run-1", invocation.Arguments())
	changedPrepared, _ := tool.NewPreparedInvocation(changed, "1.0.0", prepared.Arguments(), prepared.Actions())
	if prepared.Fingerprint() == changedPrepared.Fingerprint() {
		t.Fatal("call identity is absent from fingerprint")
	}
	actions := prepared.Actions()
	actions[0], _ = tool.NewAction(tool.EffectWorkspaceMutate, "src/main.go", "workspace")
	actionChanged, _ := tool.NewPreparedInvocation(invocation, "1.0.0", prepared.Arguments(), actions)
	if prepared.Fingerprint() == actionChanged.Fingerprint() {
		t.Fatal("actions are absent from fingerprint")
	}
	argumentChanged, _ := tool.NewPreparedInvocation(invocation, "1.0.0", json.RawMessage(`{"mode":"write","path":"src/main.go"}`), prepared.Actions())
	if prepared.Fingerprint() == argumentChanged.Fingerprint() {
		t.Fatal("normalized arguments are absent from fingerprint")
	}
	if prepared.Actions()[0].Effect() != tool.EffectWorkspaceInspect {
		t.Fatal("prepared invocation exposes action storage")
	}
	if _, err := tool.NewInvocation("workspace.file", "call", "run", json.RawMessage(`{} trailing`)); !errors.Is(err, tool.ErrInvalidInvocation) {
		t.Fatalf("expected trailing JSON rejection, got %v", err)
	}
	if _, err := tool.NewInvocation("workspace.file", "call\nbad", "run", json.RawMessage(`{}`)); !errors.Is(err, tool.ErrInvalidCallID) {
		t.Fatalf("expected unsafe call ID rejection, got %v", err)
	}
}

func TestPermissionRequestsKeepToolAndModelAuthoritySeparate(t *testing.T) {
	invocation, _ := tool.NewInvocation("workspace.file", "call-1", "run-1", json.RawMessage(`{"path":"src/main.go"}`))
	inspect, _ := tool.NewAction(tool.EffectWorkspaceInspect, "src/main.go", "workspace")
	network, _ := tool.NewAction(tool.EffectNetworkAccess, "example.invalid:443", "")
	prepared, _ := tool.NewPreparedInvocation(invocation, "1", invocation.Arguments(), []tool.Action{inspect, network})
	request, err := tool.NewToolPermissionRequest("policy.default-deny", prepared)
	if err != nil {
		t.Fatalf("construct tool permission request: %v", err)
	}
	if request.Subject() != tool.PermissionSubjectTool || len(request.Actions()) != 2 {
		t.Fatalf("unexpected tool permission request: %#v", request)
	}
	otherPolicy, _ := tool.NewToolPermissionRequest("policy.other", prepared)
	if request.Fingerprint() == otherPolicy.Fingerprint() {
		t.Fatal("policy identity is absent from permission fingerprint")
	}
	actions := request.Actions()
	actions[0] = network
	if request.Actions()[0].Effect() != tool.EffectWorkspaceInspect {
		t.Fatal("permission request exposes action storage")
	}

	target, _ := tool.NewModelTarget("ollama", "model")
	manifest, _ := tool.NewDisclosureManifest(
		contextengine.WorkspaceID("workspace"), 3, 2, 50, 200,
		tool.Fingerprint(strings.Repeat("a", 64)),
	)
	modelRequest, err := tool.NewModelPermissionRequest("policy.default-deny", "run-1", target, &manifest)
	if err != nil {
		t.Fatalf("construct model permission request: %v", err)
	}
	if modelRequest.Subject() != tool.PermissionSubjectModel || len(modelRequest.Actions()) != 2 {
		t.Fatalf("unexpected model permission request: %#v", modelRequest)
	}
	if _, ok := modelRequest.Prepared(); ok {
		t.Fatal("model permission request masquerades as a tool invocation")
	}
	if got, ok := modelRequest.Disclosure(); !ok || got.Fingerprint() != manifest.Fingerprint() {
		t.Fatalf("missing disclosure manifest: %#v %v", got, ok)
	}
}

func TestToolInvocationRejectsModelOnlyEffects(t *testing.T) {
	if _, err := tool.NewDescriptor(
		"model.fake", "model_fake", "1", "Invalid model-shaped tool.",
		json.RawMessage(`{}`), []tool.Effect{tool.EffectModelInvoke},
	); !errors.Is(err, tool.ErrInvalidDescriptor) {
		t.Fatalf("expected model effect descriptor rejection, got %v", err)
	}
	invocation, _ := tool.NewInvocation("model.fake", "call-1", "run-1", json.RawMessage(`{}`))
	action, _ := tool.NewAction(tool.EffectModelInvoke, "ollama/model", "")
	if _, err := tool.NewPreparedInvocation(invocation, "1", json.RawMessage(`{}`), []tool.Action{action}); !errors.Is(err, tool.ErrInvalidPreparedInvocation) {
		t.Fatalf("expected model action prepared invocation rejection, got %v", err)
	}
}

func TestDecisionAndApprovalValidation(t *testing.T) {
	allow, err := tool.NewDecision(tool.DecisionAllow, "explicit_rule", "", tool.GrantOneShot)
	if err != nil || allow.Validate() != nil {
		t.Fatalf("valid allow: %v", err)
	}
	deny, err := tool.NewDecision(tool.DecisionDeny, "workspace_denied", tool.DenyRecoverable, "")
	if err != nil || deny.Disposition() != tool.DenyRecoverable {
		t.Fatalf("valid deny: %#v %v", deny, err)
	}
	if _, err := tool.NewDecision(tool.DecisionAllow, "bad", "", ""); !errors.Is(err, tool.ErrInvalidDecision) {
		t.Fatalf("expected allow without grant rejection, got %v", err)
	}
	if _, err := tool.NewDecision(tool.DecisionPrompt, "approval_required", tool.DenyTerminal, ""); !errors.Is(err, tool.ErrInvalidDecision) {
		t.Fatalf("expected prompt state rejection, got %v", err)
	}
	approval, err := tool.NewApproval(tool.ApprovalDeny, "user_denied", tool.DenyTerminal, "")
	if err != nil || approval.Validate() != nil {
		t.Fatalf("valid approval: %v", err)
	}
}

func TestExecutionRequestRejectsTypedNilAndHasNoDecisionInput(t *testing.T) {
	invocation, _ := tool.NewInvocation("workspace.file", "call-1", "run-1", json.RawMessage(`{}`))
	limits := tool.ExecutionLimits{MaxDuration: time.Second, MaxOutputBytes: 1024, MaxItems: 10}
	var typedNil *approverStub
	_, err := tool.NewExecutionRequest(invocation, "policy.default-deny", typedNil, limits)
	if !errors.Is(err, tool.ErrInvalidApprover) || !errors.Is(err, tool.ErrInvalidExecutionRequest) {
		t.Fatalf("expected typed nil rejection, got %v", err)
	}
	request, err := tool.NewExecutionRequest(invocation, "policy.default-deny", nil, limits)
	if err != nil || request.Policy() != "policy.default-deny" {
		t.Fatalf("construct execution request: %#v %v", request, err)
	}
	typeOfRequest := reflect.TypeOf(request)
	for index := 0; index < typeOfRequest.NumField(); index++ {
		if typeOfRequest.Field(index).Type == reflect.TypeOf(tool.Decision{}) {
			t.Fatal("a public Decision can be supplied as execution authority")
		}
	}
}

func TestResultAndErrorSemantics(t *testing.T) {
	result, err := tool.NewResult(tool.ResultDenied, "", "", "policy_denied", 0, false, tool.DenyRecoverable)
	if err != nil || result.Validate() != nil {
		t.Fatalf("valid denied result: %v", err)
	}
	if _, err := tool.NewResult(tool.ResultSuccess, "ok", "", "completed", 1, false, ""); !errors.Is(err, tool.ErrInvalidResult) {
		t.Fatalf("expected missing media type rejection, got %v", err)
	}
	cause := errors.New("cause")
	executionErr := tool.NewExecutionError(tool.ErrorPermission, "run-1", "workspace.file", "call-1", "policy_denied", cause)
	if !errors.Is(executionErr, tool.ErrPermissionDenied) || !errors.Is(executionErr, cause) {
		t.Fatalf("execution error does not preserve kind/cause: %v", executionErr)
	}
}

type toolStub struct{}

func (toolStub) Descriptor() tool.Descriptor { return tool.Descriptor{} }
func (toolStub) Prepare(context.Context, tool.Invocation) (tool.PreparedInvocation, error) {
	return tool.PreparedInvocation{}, nil
}
func (toolStub) Execute(context.Context, tool.PreparedInvocation) (tool.Result, error) {
	return tool.Result{}, nil
}

type policyStub struct{}

func (policyStub) ID() tool.PolicyID { return "policy.stub" }
func (policyStub) Decide(context.Context, tool.PermissionRequest) (tool.Decision, error) {
	return tool.Decision{}, nil
}

type approverStub struct{}

func (*approverStub) Approve(context.Context, tool.PermissionRequest) (tool.Approval, error) {
	return tool.Approval{}, nil
}

type runtimeStub struct{}

func (runtimeStub) Register(tool.Tool) error         { return nil }
func (runtimeStub) Descriptors() []tool.Descriptor   { return nil }
func (runtimeStub) RegisterPolicy(tool.Policy) error { return nil }
func (runtimeStub) Policies() []tool.PolicyID        { return nil }
func (runtimeStub) Invoke(context.Context, tool.ExecutionRequest) (tool.Result, error) {
	return tool.Result{}, nil
}
func (runtimeStub) AuthorizeModel(context.Context, tool.PermissionRequest, tool.Approver) (tool.Decision, error) {
	return tool.Decision{}, nil
}
