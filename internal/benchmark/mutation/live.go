package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/antonio-cafeo/maestro/internal/application"
	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type LiveOptions struct {
	Profile      Profile
	Dependencies application.Dependencies
	Hardware     pkgBenchmark.HardwareProfile
	Build        ReportBuild
	Approver     func(int) pkgTool.Approver
}

func RunLiveGate(ctx context.Context, gate string, options LiveOptions) (Report, error) {
	runner := RunnerOptions{
		Profile: options.Profile, Gate: gate, Hardware: options.Hardware, Build: options.Build,
	}
	switch gate {
	case GateA:
		return RunGate(ctx, runner, func(attemptContext context.Context, attempt int) (AttemptResult, error) {
			return runDirectProtocol(attemptContext, options, attempt)
		})
	case GateB:
		return RunGate(ctx, runner, func(attemptContext context.Context, attempt int) (AttemptResult, error) {
			return runReferenceAgent(attemptContext, options, attempt, false)
		})
	case GateC:
		if options.Approver == nil {
			return Report{}, errors.New("Gate C requires an approver factory")
		}
		return RunGate(ctx, runner, func(attemptContext context.Context, attempt int) (AttemptResult, error) {
			return runReferenceAgent(attemptContext, options, attempt, true)
		})
	default:
		return Report{}, fmt.Errorf("live mutation gate %q is invalid", gate)
	}
}

func runDirectProtocol(ctx context.Context, options LiveOptions, _ int) (AttemptResult, error) {
	fixture, err := MaterializeFixture(ctx, options.Profile)
	if err != nil {
		return AttemptResult{}, err
	}
	defer fixture.Cleanup()
	result := baseAttempt(options.Profile, fixture.Initial, "direct_provider_protocol")
	config, err := qualificationConfig(options.Profile, fixture.Root, false)
	if err != nil {
		return failedAttempt(result, "configuration_invalid", "harness", "internal_failure"), nil
	}
	configured, err := application.Build(config, options.Dependencies)
	if err != nil {
		return failedAttempt(result, "composition_failed", "harness", "internal_failure"), nil
	}
	defer closeQualificationApplication(configured, options.Profile)
	if err := configured.Start(ctx); err != nil {
		return failedAttempt(result, "runtime_start_failed", "environment", "provider_failure"), nil
	}
	readTool, patchTool, err := qualificationProviderTools(configured.Runtime().Tools().Descriptors())
	if err != nil {
		return failedAttempt(result, "tool_contract_missing", "product", "tool_failure"), nil
	}
	content, err := osReadQualificationFile(fixture.Root, options.Profile.Fixture.Target)
	if err != nil {
		return failedAttempt(result, "fixture_read_failed", "harness", "internal_failure"), nil
	}
	temperature := options.Profile.Protocol.Temperature
	messages := []pkgProvider.Message{
		{Role: pkgProvider.RoleSystem, Content: "Use the single declared tool through the native tool channel. Do not print a tool call as text."},
		{Role: pkgProvider.RoleUser, Content: options.Profile.Protocol.GateAInstruction},
	}
	readResponse, err := configured.Runtime().Providers().Complete(ctx, pkgProvider.ID(config.Provider.ID), pkgProvider.CompletionRequest{
		Model: config.Models.Chat, Messages: messages,
		Options: pkgProvider.GenerationOptions{MaxTokens: options.Profile.Protocol.DirectMaxTokensPerTurn, Temperature: &temperature},
		Tools:   []pkgProvider.Tool{readTool}, ToolChoice: pkgProvider.ToolChoice{Mode: pkgProvider.ToolChoiceAuto},
	})
	if err != nil {
		return failedAttempt(result, "read_completion_failed", "environment", "provider_failure"), nil
	}
	result.ModelTurns++
	result.ToolCalls += len(readResponse.Message.ToolCalls)
	readCall, ok := exactToolCall(readResponse.Message.ToolCalls, readTool.Name)
	if !ok || !validReadArguments(readCall.Arguments, options.Profile.Fixture.Target) {
		return failedAttempt(result, "read_tool_call_invalid", "model", "provider_failure"), nil
	}
	if readCall.ID == "" {
		readCall.ID = "gate-a-read"
	}
	payload, _ := json.Marshal(map[string]string{
		"path": options.Profile.Fixture.Target, "digest": options.Profile.Fixture.InitialSHA256,
		"content": string(content),
	})
	assistant := readResponse.Message
	assistant.Role = pkgProvider.RoleAssistant
	messages = append(messages, assistant, pkgProvider.Message{
		Role: pkgProvider.RoleTool, ToolCallID: readCall.ID, ToolName: readTool.Name, Content: string(payload),
	})
	patchResponse, err := configured.Runtime().Providers().Complete(ctx, pkgProvider.ID(config.Provider.ID), pkgProvider.CompletionRequest{
		Model: config.Models.Chat, Messages: messages,
		Options: pkgProvider.GenerationOptions{MaxTokens: options.Profile.Protocol.DirectMaxTokensPerTurn, Temperature: &temperature},
		Tools:   []pkgProvider.Tool{patchTool}, ToolChoice: pkgProvider.ToolChoice{Mode: pkgProvider.ToolChoiceAuto},
	})
	if err != nil {
		return failedAttempt(result, "patch_completion_failed", "environment", "provider_failure"), nil
	}
	result.ModelTurns++
	result.ToolCalls += len(patchResponse.Message.ToolCalls)
	patchCall, ok := exactToolCall(patchResponse.Message.ToolCalls, patchTool.Name)
	if !ok || !validPatchArguments(patchCall.Arguments, options.Profile.Fixture) {
		return failedAttempt(result, "patch_tool_call_invalid", "model", "provider_failure"), nil
	}
	final, err := SnapshotWorkspace(ctx, fixture.Root)
	if err != nil || final.Digest != fixture.Initial.Digest {
		return failedAttempt(result, "gate_a_workspace_changed", "product", "tool_failure"), nil
	}
	result.State = "passed"
	result.Terminal = "completed"
	result.FinalSHA256 = options.Profile.Fixture.InitialSHA256
	result.WorkspaceDigest = final.Digest
	result.Lifecycle = []string{"terminal_completed"}
	return result, nil
}

func runReferenceAgent(ctx context.Context, options LiveOptions, attempt int, mutating bool) (AttemptResult, error) {
	fixture, err := MaterializeFixture(ctx, options.Profile)
	if err != nil {
		return AttemptResult{}, err
	}
	defer fixture.Cleanup()
	evidenceCode := "reference_agent_read_only"
	if mutating {
		evidenceCode = "reference_agent_controlled_mutation"
	}
	result := baseAttempt(options.Profile, fixture.Initial, evidenceCode)
	config, err := qualificationConfig(options.Profile, fixture.Root, mutating)
	if err != nil {
		return failedAttempt(result, "configuration_invalid", "harness", "internal_failure"), nil
	}
	configured, err := application.Build(config, options.Dependencies)
	if err != nil {
		return failedAttempt(result, "composition_failed", "harness", "internal_failure"), nil
	}
	recorder := &liveRecorder{}
	if err := recorder.Subscribe(configured.Runtime().EventBus()); err != nil {
		_ = closeQualificationApplication(configured, options.Profile)
		return failedAttempt(result, "event_subscription_failed", "harness", "internal_failure"), nil
	}
	instruction := options.Profile.Protocol.GateBInstruction
	var approver pkgTool.Approver
	if mutating {
		instruction = options.Profile.Protocol.GateCInstruction
		approver = options.Approver(attempt)
	}
	runResult, runError := configured.ExecuteWithOptions(ctx, instruction, application.ExecuteOptions{Approver: approver})
	closeError := closeQualificationApplication(configured, options.Profile)
	final, snapshotError := SnapshotWorkspace(context.Background(), fixture.Root)
	if snapshotError != nil {
		return failedAttempt(result, "workspace_snapshot_failed", "harness", "internal_failure"), nil
	}
	result.FinalSHA256, _ = final.File(options.Profile.Fixture.Target)
	result.WorkspaceDigest = final.Digest
	recorded := recorder.Result()
	result.Lifecycle = recorded.Lifecycle
	result.Approval = recorded.Approval
	result.MutationAttempts = recorded.MutationAttempts
	if runResult.Validate() == nil {
		session := runResult.Session()
		result.Terminal = string(session.Terminal())
		result.ModelTurns = session.Counters().ModelTurns
		result.ToolCalls = session.Counters().ToolCalls
		if session.ContextStale() {
			result.ContextFreshness = "stale"
		} else if mutating {
			result.ContextFreshness = "fresh"
		}
	}
	if result.Terminal == "" {
		result.Terminal = recorded.Terminal
	}
	if closeError != nil {
		return failedAttempt(result, "runtime_cleanup_failed", "product", "internal_failure"), nil
	}
	if runError != nil {
		failureClass := "product"
		if result.Terminal == "provider_failure" || result.Terminal == "deadline_exceeded" {
			failureClass = "model"
		}
		return failedAttempt(result, "reference_agent_failed", failureClass, defaultTerminal(result.Terminal)), nil
	}
	if result.Terminal != "completed" {
		return failedAttempt(result, "terminal_not_completed", "product", defaultTerminal(result.Terminal)), nil
	}
	if !mutating {
		if !containsEvidence(runResult.Content(), options.Profile.Protocol.GateBEvidence) {
			return failedAttempt(result, "read_only_evidence_missing", "model", result.Terminal), nil
		}
		if final.Digest != fixture.Initial.Digest || result.FinalSHA256 != options.Profile.Fixture.InitialSHA256 {
			return failedAttempt(result, "read_only_workspace_changed", "product", "tool_failure"), nil
		}
		result.Approval = "not_requested"
		result.ContextFreshness = "not_applicable"
	} else {
		if result.Approval != "allowed_once" || result.MutationAttempts != 1 ||
			result.ContextFreshness != "fresh" || result.FinalSHA256 != options.Profile.Fixture.ExpectedSHA256 ||
			!onlyTargetChanged(fixture.Initial, final, options.Profile.Fixture.Target) {
			return failedAttempt(result, "controlled_mutation_invariant", "product", "tool_failure"), nil
		}
	}
	result.State = "passed"
	result.ReasonCode = ""
	result.FailureClass = ""
	return result, nil
}

func qualificationConfig(profile Profile, root string, mutating bool) (productconfig.Config, error) {
	name, err := profile.Resolve(profile.Configuration.ProductProfile)
	if err != nil {
		return productconfig.Config{}, err
	}
	config, err := productconfig.Load(name)
	if err != nil {
		return productconfig.Config{}, err
	}
	config.Workspace.Root = root
	if !mutating {
		config.Agent.Tools = []string{"workspace.list", "workspace.read", "workspace.search"}
		config.Policy.WorkspaceMutation = "deny"
	}
	if err := config.ValidateExecutionProfile(); err != nil {
		return productconfig.Config{}, err
	}
	return config, nil
}

func closeQualificationApplication(configured *application.Application, profile Profile) error {
	ctx, cancel := context.WithTimeout(context.Background(), profile.Protocol.CleanupTimeout.Duration)
	defer cancel()
	return configured.Close(ctx)
}

func qualificationProviderTools(descriptors []pkgTool.Descriptor) (pkgProvider.Tool, pkgProvider.Tool, error) {
	var readTool, patchTool pkgProvider.Tool
	for _, descriptor := range descriptors {
		candidate := pkgProvider.Tool{
			Name: string(descriptor.Name()), Description: descriptor.Description(), Parameters: descriptor.Parameters(),
		}
		switch descriptor.ID() {
		case "workspace.read":
			readTool = candidate
		case "workspace.patch":
			patchTool = candidate
		}
	}
	if readTool.Name == "" || patchTool.Name == "" {
		return pkgProvider.Tool{}, pkgProvider.Tool{}, errors.New("qualification workspace tools are missing")
	}
	return readTool, patchTool, nil
}

func exactToolCall(calls []pkgProvider.ToolCall, name string) (pkgProvider.ToolCall, bool) {
	if len(calls) != 1 || calls[0].Name != name {
		return pkgProvider.ToolCall{}, false
	}
	return calls[0], true
}

func validReadArguments(raw json.RawMessage, target string) bool {
	var arguments struct {
		Path     string `json:"path"`
		MaxBytes int64  `json:"max_bytes,omitempty"`
	}
	return decodeStrictJSON(raw, &arguments) == nil && arguments.Path == target && arguments.MaxBytes >= 0
}

func validPatchArguments(raw json.RawMessage, fixture Fixture) bool {
	var arguments struct {
		Path           string `json:"path"`
		Old            string `json:"old"`
		New            string `json:"new"`
		ExpectedDigest string `json:"expected_digest"`
	}
	return decodeStrictJSON(raw, &arguments) == nil && arguments.Path == fixture.Target &&
		arguments.Old == fixture.Old && arguments.New == fixture.Replacement &&
		arguments.ExpectedDigest == fixture.InitialSHA256
}

func decodeStrictJSON(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("JSON contains trailing content")
	}
	return nil
}

func baseAttempt(profile Profile, snapshot WorkspaceSnapshot, evidence string) AttemptResult {
	return AttemptResult{
		State: "failed", Scenario: "", EvidenceCode: evidence, Terminal: "internal_failure",
		Approval: "not_requested", InitialSHA256: profile.Fixture.InitialSHA256,
		ExpectedSHA256: profile.Fixture.ExpectedSHA256, FinalSHA256: profile.Fixture.InitialSHA256,
		WorkspaceDigest: snapshot.Digest, ContextFreshness: "not_applicable", TemporaryCleanup: "clean",
	}
}

func failedAttempt(result AttemptResult, reason, class, terminal string) AttemptResult {
	result.State = "failed"
	result.ReasonCode = reason
	result.FailureClass = class
	result.Terminal = defaultTerminal(terminal)
	if len(result.Lifecycle) == 0 {
		if result.Terminal == "canceled" || result.Terminal == "deadline_exceeded" {
			result.Lifecycle = []string{"terminal_canceled"}
		} else {
			result.Lifecycle = []string{"terminal_failed"}
		}
	}
	return result
}

func defaultTerminal(value string) string {
	if value == "" {
		return "internal_failure"
	}
	return value
}

func containsEvidence(content string, evidence []string) bool {
	normalized := strings.ToLower(content)
	for _, value := range evidence {
		if !strings.Contains(normalized, strings.ToLower(value)) {
			return false
		}
	}
	return true
}

func onlyTargetChanged(initial, final WorkspaceSnapshot, target string) bool {
	if len(initial.Files) != len(final.Files) {
		return false
	}
	for index := range initial.Files {
		if initial.Files[index].Path != final.Files[index].Path {
			return false
		}
		if initial.Files[index].Path != target && initial.Files[index].SHA256 != final.Files[index].SHA256 {
			return false
		}
	}
	return true
}

func osReadQualificationFile(root, logical string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
}

type liveRecorder struct {
	mu               sync.Mutex
	lifecycle        []string
	approval         string
	mutationAttempts int
	terminal         string
}

type recordedLiveResult struct {
	Lifecycle        []string
	Approval         string
	MutationAttempts int
	Terminal         string
}

func (recorder *liveRecorder) Subscribe(events pkgRuntime.EventBus) error {
	if recorder == nil || events == nil {
		return errors.New("live mutation recorder dependencies are nil")
	}
	for _, topic := range []string{pkgTool.EventInvocationPrepared, pkgTool.EventPermissionDecided} {
		if err := events.Subscribe(topic, recorder.toolEvent); err != nil {
			return err
		}
	}
	for _, topic := range []string{pkgAgent.EventMutationTransitioned, pkgAgent.EventSessionCompleted, pkgAgent.EventSessionFailed} {
		if err := events.Subscribe(topic, recorder.agentEvent); err != nil {
			return err
		}
	}
	return nil
}

func (recorder *liveRecorder) toolEvent(event pkgRuntime.Event) {
	payload, ok := event.Payload().(pkgTool.EventPayload)
	if !ok || payload.Tool != "workspace.patch" {
		return
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	switch event.Name() {
	case pkgTool.EventInvocationPrepared:
		recorder.lifecycle = append(recorder.lifecycle, "proposal_prepared")
	case pkgTool.EventPermissionDecided:
		if payload.Decision == pkgTool.DecisionAllow {
			recorder.approval = "allowed_once"
			recorder.lifecycle = append(recorder.lifecycle, "approval_allowed")
		} else {
			recorder.approval = "denied"
			recorder.lifecycle = append(recorder.lifecycle, "approval_denied")
		}
	}
}

func (recorder *liveRecorder) agentEvent(event pkgRuntime.Event) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if event.Name() == pkgAgent.EventMutationTransitioned {
		payload, ok := event.Payload().(pkgAgent.MutationEventPayload)
		if !ok {
			return
		}
		stage := string(payload.MutationStage)
		status := string(payload.MutationStatus)
		if stage == "apply" && (status == "succeeded" || status == "failed" || status == "canceled") {
			recorder.mutationAttempts = 1
		}
		if status == "canceled" {
			status = "failed"
		}
		recorder.lifecycle = append(recorder.lifecycle, stage+"_"+status)
		return
	}
	payload, ok := event.Payload().(pkgAgent.EventPayload)
	if !ok {
		return
	}
	recorder.terminal = string(payload.Terminal)
	terminal := "terminal_failed"
	if payload.Terminal == pkgAgent.TerminalCompleted {
		terminal = "terminal_completed"
	} else if payload.Terminal == pkgAgent.TerminalCanceled || payload.Terminal == pkgAgent.TerminalDeadline {
		terminal = "terminal_canceled"
	}
	recorder.lifecycle = append(recorder.lifecycle, terminal)
}

func (recorder *liveRecorder) Result() recordedLiveResult {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	approval := recorder.approval
	if approval == "" {
		approval = "not_requested"
	}
	return recordedLiveResult{
		Lifecycle: slices.Clone(recorder.lifecycle), Approval: approval,
		MutationAttempts: recorder.mutationAttempts, Terminal: recorder.terminal,
	}
}
