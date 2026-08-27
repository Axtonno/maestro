package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type agentLoop struct {
	providers   generationRuntime
	tools       pkgTool.Runtime
	permissions pkgTool.Runtime
	context     pkgContext.Engine
	events      pkgRuntime.EventBus
}

func (loop *agentLoop) Run(
	ctx context.Context,
	current *session,
	request pkgAgent.RunRequest,
	bundle pkgContext.ContextBundle,
) (string, pkgAgent.TerminalReason, error) {
	providerTools, descriptors, err := loop.resolveTools(request.Tools())
	if err != nil {
		return "", pkgAgent.TerminalToolFailure, err
	}
	permission, err := planningPermission(request, bundle)
	if err != nil {
		return "", pkgAgent.TerminalInternalFailure, err
	}
	seenCalls := make(map[pkgTool.CallID]struct{})
	choreography := &toolChoreography{}
	progressiveEvidence := newProgressiveEvidenceState(request, loop.events)
	lastContent := ""
	turn := 0
	mutationAttempts := 0
	workspaceReadObserved := false

	for {
		step, found, err := nextReadyStep(current.snapshotValue())
		if err != nil {
			return "", pkgAgent.TerminalInternalFailure, err
		}
		if !found {
			if strings.TrimSpace(lastContent) == "" {
				return "", pkgAgent.TerminalBlocked, pkgAgent.ErrProviderFailed
			}
			return lastContent, pkgAgent.TerminalCompleted, nil
		}
		if err := current.transitionStep(step.ID(), pkgAgent.StepRunning, ""); err != nil {
			return "", pkgAgent.TerminalInternalFailure, err
		}
		loop.publishStep(current, step.ID(), pkgAgent.StepRunning)

		messages := initialMessages(request, step, bundle)
		for {
			if err := ctx.Err(); err != nil {
				return "", reasonForError(ctx, pkgAgent.TerminalCanceled), err
			}
			decision, err := loop.permissions.AuthorizeModel(ctx, permission, request.Approver())
			if err != nil {
				return "", permissionTerminal(ctx, err), err
			}
			if decision.Kind() != pkgTool.DecisionAllow {
				_ = current.transitionStep(step.ID(), pkgAgent.StepBlocked, "permission_denied")
				return "", pkgAgent.TerminalPermissionDenied, pkgAgent.ErrPermissionDenied
			}
			if !workspaceReadObserved {
				var bootstrapped bool
				messages, bootstrapped, err = loop.bootstrapReferenceRead(
					ctx, current, request, messages, descriptors, seenCalls, choreography,
				)
				if err != nil {
					return "", toolErrorTerminal(ctx, err), errors.Join(pkgAgent.ErrToolFailed, err)
				}
				workspaceReadObserved = bootstrapped
				if bootstrapped && progressiveEvidence != nil {
					messages = append(messages, provider.Message{
						Role: provider.RoleSystem, Content: progressiveEvidence.observeBootstrap(),
					})
				}
			}
			if err := current.consume(counterDelta{modelTurns: 1}); err != nil {
				return "", pkgAgent.TerminalLimit, err
			}
			publishAgentEvent(loop.events, pkgAgent.EventTurnStarted, sessionEventPayload(current.snapshotValue(), 0, pkgAgent.EventFailureNone))
			turn++
			turnTools := choreography.toolsForTurn(providerTools, descriptors)
			if requiresWorkspaceRead(request, descriptors) && !workspaceReadObserved {
				turnTools = onlyWorkspaceReadTool(turnTools, descriptors)
			}
			completionRequest := provider.CompletionRequest{
				Model: request.Model(), Messages: messages,
				Options:    request.GenerationOptions(),
				Tools:      turnTools,
				ToolChoice: provider.ToolChoice{Mode: provider.ToolChoiceAuto},
			}
			response, err := loop.complete(ctx, request, completionRequest)
			if err != nil {
				_ = current.transitionStep(step.ID(), pkgAgent.StepFailed, "provider_failed")
				return "", reasonForError(ctx, pkgAgent.TerminalProviderFailure), errors.Join(pkgAgent.ErrProviderFailed, err)
			}
			if err := validateProviderResponse(response, request.Limits().MaxToolCallsPerTurn); err != nil {
				if errors.Is(err, pkgAgent.ErrLimitExceeded) {
					return "", pkgAgent.TerminalLimit, err
				}
				_ = current.transitionStep(step.ID(), pkgAgent.StepFailed, "provider_failed")
				return "", pkgAgent.TerminalProviderFailure, errors.Join(pkgAgent.ErrProviderFailed, err)
			}
			responseBytes := len(response.Message.Content)
			for _, call := range response.Message.ToolCalls {
				responseBytes += len(call.ID) + len(call.Name) + len(call.Arguments)
			}
			if err := current.consume(counterDelta{
				inputTokens: response.Usage.InputTokens, outputTokens: response.Usage.OutputTokens,
				sessionBytes: responseBytes,
			}); err != nil {
				return "", pkgAgent.TerminalLimit, err
			}
			publishAgentEvent(loop.events, pkgAgent.EventTurnCompleted, sessionEventPayload(current.snapshotValue(), 0, pkgAgent.EventFailureNone))

			assistant := response.Message
			assistant.Role = provider.RoleAssistant
			messages = append(messages, assistant)
			if len(assistant.ToolCalls) == 0 {
				if response.FinishReason == provider.FinishReasonLength {
					return "", pkgAgent.TerminalLimit, pkgAgent.ErrLimitExceeded
				}
				if strings.TrimSpace(assistant.Content) == "" {
					_ = current.transitionStep(step.ID(), pkgAgent.StepFailed, "provider_failed")
					return "", pkgAgent.TerminalProviderFailure, pkgAgent.ErrProviderFailed
				}
				if len(providerTools) > 0 && describesToolCall(assistant.Content) {
					messages = append(messages, provider.Message{
						Role:    provider.RoleSystem,
						Content: "The previous assistant response emitted tool-call-shaped JSON as text. The step is not complete. Invoke one of the declared tool interfaces through the provider tool channel now; do not print a tool name or arguments as assistant content.",
					})
					continue
				}
				if choreography.requiresMutation() {
					messages = append(messages, provider.Message{
						Role:    provider.RoleSystem,
						Content: "A requested workspace mutation has not completed. Continue with the single declared tool needed by the latest recoverable result; do not return a final answer yet.",
					})
					continue
				}
				if requiresWorkspaceRead(request, descriptors) && !workspaceReadObserved {
					messages = append(messages, provider.Message{
						Role:    provider.RoleSystem,
						Content: "The reference agent has not read the requested workspace file through the declared read tool yet. The step is not complete. Invoke the declared workspace read tool through the provider tool channel now; do not claim to have read tool results before the invocation succeeds.",
					})
					continue
				}
				if accepted, correction := progressiveEvidence.evaluateAssistant(assistant.Content); !accepted {
					messages = append(messages, provider.Message{Role: provider.RoleSystem, Content: correction})
					continue
				}
				if err := current.transitionStep(step.ID(), pkgAgent.StepCompleted, ""); err != nil {
					return "", pkgAgent.TerminalInternalFailure, err
				}
				loop.publishStep(current, step.ID(), pkgAgent.StepCompleted)
				lastContent = assistant.Content
				break
			}

			for callIndex, call := range assistant.ToolCalls {
				callID := pkgTool.CallID(call.ID)
				if callID == "" {
					callID = pkgTool.CallID(fmt.Sprintf("maestro-%s-%d-%d", request.Run(), turn, callIndex))
				}
				if err := callID.Validate(); err != nil {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, err)
				}
				if _, duplicate := seenCalls[callID]; duplicate {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgTool.ErrInvalidInvocation)
				}
				seenCalls[callID] = struct{}{}
				descriptor, ok := descriptors[call.Name]
				if !ok {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgTool.ErrNotFound)
				}
				invocation, err := pkgTool.NewInvocation(descriptor.ID(), callID, pkgTool.RunID(request.Run()), call.Arguments)
				if err != nil {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, err)
				}
				if err := current.consume(counterDelta{toolCalls: 1}); err != nil {
					return "", pkgAgent.TerminalLimit, err
				}
				choreographyResult, execute, err := choreography.beforeCall(descriptor, call.Arguments, callIndex)
				if err != nil {
					return "", pkgAgent.TerminalInternalFailure, err
				}
				if !execute {
					progressiveEvidence.afterTool(descriptor, choreographyResult)
					messageContent, err := encodeToolResult(choreographyResult)
					if err != nil {
						return "", pkgAgent.TerminalInternalFailure, err
					}
					if err := current.consume(counterDelta{sessionBytes: len(messageContent)}); err != nil {
						return "", pkgAgent.TerminalLimit, err
					}
					messages = append(messages, provider.Message{
						Role: provider.RoleTool, Content: messageContent,
						ToolCallID: string(callID), ToolName: call.Name,
					})
					continue
				}
				execution, err := pkgTool.NewExecutionRequest(
					invocation, request.Policy(), request.Approver(),
					pkgTool.ExecutionLimits{
						MaxDuration: request.Limits().MaxDuration, MaxOutputBytes: request.Limits().MaxToolResultBytes,
						MaxItems: 100_000,
					},
				)
				if err != nil {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, err)
				}
				mutating := descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceMutate)
				if mutating {
					if mutationAttempts > 0 {
						loop.publishMutation(current, pkgAgent.MutationStageApply, pkgAgent.MutationFailed, "", false, 0)
						return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrMutationFailed)
					}
					mutationAttempts++
				}
				progressiveEvidence.beforeTool(descriptor)
				result, err := loop.tools.Invoke(ctx, execution)
				if err != nil {
					if mutating {
						if current.snapshotValue().ContextStale() {
							loop.publishMutation(current, pkgAgent.MutationStageApply, mutationStatusForError(ctx), "", false, 0)
						}
						err = errors.Join(pkgAgent.ErrMutationFailed, err)
						return "", toolErrorTerminal(ctx, err), errors.Join(pkgAgent.ErrToolFailed, err)
					}
					recoverable, ok, recoverErr := recoverableReadOnlyInvocationError(descriptor, err)
					if recoverErr != nil {
						return "", pkgAgent.TerminalInternalFailure, recoverErr
					}
					if !ok {
						progressiveEvidence.toolFailed(descriptor)
						return "", toolErrorTerminal(ctx, err), errors.Join(pkgAgent.ErrToolFailed, err)
					}
					result = recoverable
				}
				progressiveEvidence.afterTool(descriptor, result)
				if result.Outcome() == pkgTool.ResultSuccess && descriptor.ID() == workspaceReadToolID {
					workspaceReadObserved = true
				}
				choreography.afterCall(descriptor, result)
				if mutating && result.Outcome() != pkgTool.ResultDenied && !current.snapshotValue().ContextStale() {
					if err := current.markStale(); err != nil {
						return "", pkgAgent.TerminalInternalFailure, err
					}
				}
				messageContent, err := encodeToolResult(result)
				if err != nil {
					return "", pkgAgent.TerminalInternalFailure, err
				}
				if err := current.consume(counterDelta{sessionBytes: len(messageContent)}); err != nil {
					return "", pkgAgent.TerminalLimit, err
				}
				messages = append(messages, provider.Message{
					Role: provider.RoleTool, Content: messageContent,
					ToolCallID: string(callID), ToolName: call.Name,
				})
				if result.Outcome() == pkgTool.ResultDenied && result.Disposition() == pkgTool.DenyTerminal {
					_ = current.transitionStep(step.ID(), pkgAgent.StepBlocked, "permission_denied")
					return "", pkgAgent.TerminalPermissionDenied, pkgAgent.ErrPermissionDenied
				}
				if mutating && result.Outcome() != pkgTool.ResultDenied {
					loop.publishMutation(
						current, pkgAgent.MutationStageApply, mutationStatusForResult(result),
						mutationEffect(result.Effect()), result.Durable(), 0,
					)
				}
				if result.Outcome() == pkgTool.ResultCanceled {
					return "", reasonForError(ctx, pkgAgent.TerminalCanceled), pkgAgent.ErrRunCanceled
				}
				if mutating && current.snapshotValue().ContextStale() {
					workspace, ok := request.WorkspaceTarget()
					if !ok {
						loop.publishMutation(current, pkgAgent.MutationStageReindex, pkgAgent.MutationFailed, mutationEffect(result.Effect()), result.Durable(), 0)
						return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrContextRefreshFailed, pkgContext.ErrInvalidWorkspace)
					}
					loop.publishMutation(current, pkgAgent.MutationStageReindex, pkgAgent.MutationStarted, mutationEffect(result.Effect()), result.Durable(), 0)
					snapshot, refreshErr := loop.context.Index(ctx, workspace)
					if refreshErr != nil {
						loop.publishMutation(current, pkgAgent.MutationStageReindex, mutationStatusForError(ctx), mutationEffect(result.Effect()), result.Durable(), 0)
						return "", reasonForError(ctx, pkgAgent.TerminalToolFailure), errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrContextRefreshFailed, refreshErr)
					}
					refreshed, refreshErr := loop.context.Build(ctx, request.Context())
					if refreshErr != nil {
						loop.publishMutation(current, pkgAgent.MutationStageReindex, mutationStatusForError(ctx), mutationEffect(result.Effect()), result.Durable(), 0)
						return "", reasonForError(ctx, pkgAgent.TerminalToolFailure), errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrContextRefreshFailed, refreshErr)
					}
					if refreshed.Generation() != snapshot.Metadata().Generation {
						loop.publishMutation(current, pkgAgent.MutationStageReindex, pkgAgent.MutationFailed, mutationEffect(result.Effect()), result.Durable(), 0)
						return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrContextRefreshFailed, errors.New("context refresh returned a stale generation"))
					}
					if err := current.markFresh(refreshed.Generation()); err != nil {
						return "", pkgAgent.TerminalInternalFailure, err
					}
					bundle = refreshed
					loop.publishMutation(current, pkgAgent.MutationStageReindex, pkgAgent.MutationSucceeded, mutationEffect(result.Effect()), result.Durable(), refreshed.Generation())
				}
				if mutating && result.Outcome() != pkgTool.ResultSuccess {
					return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgAgent.ErrMutationFailed)
				}
			}
		}
	}
}

func (loop *agentLoop) publishMutation(
	current *session,
	stage pkgAgent.MutationStage,
	status pkgAgent.MutationStatus,
	effect pkgAgent.MutationEffect,
	durable bool,
	generation uint64,
) {
	snapshot := current.snapshotValue()
	payload := pkgAgent.MutationEventPayload{
		Run: snapshot.Run(), Agent: snapshot.Agent(), MutationStage: stage,
		MutationStatus: status, MutationEffect: effect, Durable: durable,
		WorkspaceGeneration: generation,
	}
	publishMutationEvent(loop.events, payload)
}

func mutationStatusForResult(result pkgTool.Result) pkgAgent.MutationStatus {
	switch result.Outcome() {
	case pkgTool.ResultSuccess:
		return pkgAgent.MutationSucceeded
	case pkgTool.ResultCanceled:
		return pkgAgent.MutationCanceled
	default:
		return pkgAgent.MutationFailed
	}
}

func mutationStatusForError(ctx context.Context) pkgAgent.MutationStatus {
	if ctx != nil && (errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded)) {
		return pkgAgent.MutationCanceled
	}
	return pkgAgent.MutationFailed
}

func mutationEffect(effect pkgTool.EffectState) pkgAgent.MutationEffect {
	switch effect {
	case pkgTool.EffectApplied:
		return pkgAgent.MutationEffectApplied
	case pkgTool.EffectUnchanged:
		return pkgAgent.MutationEffectUnchanged
	default:
		return ""
	}
}

func describesToolCall(content string) bool {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		if newline := strings.IndexByte(trimmed, '\n'); newline >= 0 {
			trimmed = strings.TrimSpace(trimmed[newline+1:])
		}
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	if decodeContainsToolCall(trimmed) {
		return true
	}
	for index := 0; index < len(trimmed); index++ {
		if trimmed[index] != '{' && trimmed[index] != '[' {
			continue
		}
		if decodeContainsToolCall(trimmed[index:]) {
			return true
		}
	}
	return false
}

func decodeContainsToolCall(candidate string) bool {
	decoder := json.NewDecoder(strings.NewReader(candidate))
	var value any
	if decoder.Decode(&value) != nil {
		return false
	}
	return containsToolCall(value)
}

func containsToolCall(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if containsToolCall(item) {
				return true
			}
		}
	case map[string]any:
		name, hasName := typed["name"].(string)
		_, hasArguments := typed["arguments"]
		_, hasParameters := typed["parameters"]
		_, hasInput := typed["input"]
		if hasName && strings.TrimSpace(name) != "" && (hasArguments || hasParameters || hasInput) {
			return true
		}
		for key, item := range typed {
			if key == "function" || key == "tool_call" || key == "tool_calls" {
				if containsToolCall(item) {
					return true
				}
			}
		}
	}
	return false
}

func (loop *agentLoop) publishStep(current *session, step pkgAgent.StepID, state pkgAgent.StepStatus) {
	payload := sessionEventPayload(current.snapshotValue(), 0, pkgAgent.EventFailureNone)
	payload.Step = step
	payload.StepState = state
	publishAgentEvent(loop.events, pkgAgent.EventStepTransitioned, payload)
}

func descriptorHasEffect(descriptor pkgTool.Descriptor, effect pkgTool.Effect) bool {
	for _, candidate := range descriptor.Effects() {
		if candidate == effect {
			return true
		}
	}
	return false
}

func (loop *agentLoop) resolveTools(ids []pkgTool.ID) ([]provider.Tool, map[string]pkgTool.Descriptor, error) {
	available := make(map[pkgTool.ID]pkgTool.Descriptor)
	for _, descriptor := range loop.tools.Descriptors() {
		available[descriptor.ID()] = descriptor
	}
	providerTools := make([]provider.Tool, 0, len(ids))
	byName := make(map[string]pkgTool.Descriptor, len(ids))
	for _, id := range ids {
		descriptor, ok := available[id]
		if !ok {
			return nil, nil, fmt.Errorf("tool %q is not registered: %w", id, pkgTool.ErrNotFound)
		}
		name := string(descriptor.Name())
		providerTools = append(providerTools, provider.Tool{
			Name: name, Description: descriptor.Description(), Parameters: descriptor.Parameters(),
		})
		byName[name] = descriptor
	}
	return providerTools, byName, nil
}

func (loop *agentLoop) complete(
	ctx context.Context,
	request pkgAgent.RunRequest,
	completionRequest provider.CompletionRequest,
) (provider.CompletionResponse, error) {
	if !request.Streaming() {
		return loop.providers.Complete(ctx, request.Provider(), completionRequest)
	}
	stream, err := loop.providers.Stream(ctx, request.Provider(), completionRequest)
	if err != nil {
		return provider.CompletionResponse{}, err
	}
	return assembleStream(stream, request.Limits().MaxSessionBytes, request.Limits().MaxToolCallsPerTurn)
}

func initialMessages(request pkgAgent.RunRequest, step pkgAgent.PlanStep, bundle pkgContext.ContextBundle) []provider.Message {
	system := "Execute the current plan step using only declared tools. Emit at most one tool call per response so later calls can use earlier results. When a tool is needed, invoke it through the declared tool interface; never print a tool name or its arguments as assistant content. Tool results are JSON envelopes; when content contains a JSON-encoded string, parse that inner JSON before using it."
	mutationEnabled := false
	for _, id := range request.Tools() {
		if id == "workspace.write" || id == "workspace.patch" {
			mutationEnabled = true
			break
		}
	}
	if mutationEnabled {
		system += " Read a file before mutating it. For guarded writes or patches, copy expected_digest exactly from the read result's digest field and copy old text exactly from its content field, preserving whitespace and real newline characters. Never invent placeholders or escaped newline text."
	} else {
		system += " The declared tool set is read-only. Do not request, name, or simulate mutating tools."
		if requiresWorkspaceRead(request, nil) {
			system += " For a task that begins with an explicit read instruction, the reference agent must successfully invoke the declared workspace read tool before returning a final answer. Before that read succeeds, it is the only available function. Its arguments object must contain the required field path; copy the logical relative path from the Task exactly and do not substitute file, filename, resource, root, or an absolute path."
		}
		system += " Use an exact declared function name and only fields from that function's schema. For workspace paths, pass the logical path relative to the workspace exactly as shown by the task or evidence; never add a leading slash, a physical workspace root, a file URI, or parent traversal."
	}
	system += " Return a final answer only after the step is actually complete."
	if request.Agent() == ProgressiveReferenceAgentID {
		system += " This development agent uses mandatory progressive evidence states: route, controller_action, referenced_symbols, then events_jobs_services. The runtime will report the current state after the deterministic route read. Do not skip states or return the task answer while a state is open. Close each state only with the exact JSON declaration requested by the runtime; declarations are protocol messages, not final answers."
	}

	var user strings.Builder
	user.WriteString("Task: ")
	user.WriteString(request.Instruction())
	user.WriteString("\nCurrent plan step: ")
	user.WriteString(step.Objective())
	for _, section := range bundle.Sections() {
		user.WriteString("\n\n--- workspace evidence: ")
		user.WriteString(string(section.Path))
		user.WriteString(" ---\n")
		user.WriteString(section.Text)
	}
	return []provider.Message{
		{Role: provider.RoleSystem, Content: system},
		{Role: provider.RoleUser, Content: user.String()},
	}
}

func requiresWorkspaceRead(request pkgAgent.RunRequest, descriptors map[string]pkgTool.Descriptor) bool {
	instruction := strings.ToLower(strings.TrimSpace(request.Instruction()))
	if !isReferenceAgent(request.Agent()) || !strings.HasPrefix(instruction, "read ") {
		return false
	}
	if descriptors == nil {
		return true
	}
	for _, descriptor := range descriptors {
		if descriptor.ID() == workspaceReadToolID {
			return true
		}
	}
	return false
}

func onlyWorkspaceReadTool(tools []provider.Tool, descriptors map[string]pkgTool.Descriptor) []provider.Tool {
	for _, tool := range tools {
		if descriptor, ok := descriptors[tool.Name]; ok && descriptor.ID() == workspaceReadToolID {
			return []provider.Tool{tool}
		}
	}
	return nil
}

func (loop *agentLoop) bootstrapReferenceRead(
	ctx context.Context,
	current *session,
	request pkgAgent.RunRequest,
	messages []provider.Message,
	descriptors map[string]pkgTool.Descriptor,
	seenCalls map[pkgTool.CallID]struct{},
	choreography *toolChoreography,
) ([]provider.Message, bool, error) {
	logical, required := explicitReferenceReadPath(request)
	if !required {
		return messages, false, nil
	}
	var descriptor pkgTool.Descriptor
	var providerName string
	for name, candidate := range descriptors {
		if candidate.ID() == workspaceReadToolID {
			descriptor, providerName = candidate, name
			break
		}
	}
	if providerName == "" {
		return messages, false, nil
	}
	arguments, err := json.Marshal(struct {
		Path string `json:"path"`
	}{Path: logical})
	if err != nil {
		return messages, false, err
	}
	callID := pkgTool.CallID(fmt.Sprintf("maestro-%s-pre-read", request.Run()))
	invocation, err := pkgTool.NewInvocation(descriptor.ID(), callID, pkgTool.RunID(request.Run()), arguments)
	if err != nil {
		return messages, false, err
	}
	if _, duplicate := seenCalls[callID]; duplicate {
		return messages, false, pkgTool.ErrInvalidInvocation
	}
	seenCalls[callID] = struct{}{}
	if err := current.consume(counterDelta{
		toolCalls:    1,
		sessionBytes: len(callID) + len(providerName) + len(arguments),
	}); err != nil {
		return messages, false, err
	}
	execution, err := pkgTool.NewExecutionRequest(
		invocation, request.Policy(), request.Approver(),
		pkgTool.ExecutionLimits{
			MaxDuration:    request.Limits().MaxDuration,
			MaxOutputBytes: request.Limits().MaxToolResultBytes,
			MaxItems:       100_000,
		},
	)
	if err != nil {
		return messages, false, err
	}
	result, err := loop.tools.Invoke(ctx, execution)
	if err != nil {
		return messages, false, err
	}
	if result.Outcome() == pkgTool.ResultDenied {
		return messages, false, pkgTool.ErrPermissionDenied
	}
	if result.Outcome() != pkgTool.ResultSuccess {
		return messages, false, pkgTool.ErrExecutionFailed
	}
	messageContent, err := encodeToolResult(result)
	if err != nil {
		return messages, false, err
	}
	if err := current.consume(counterDelta{sessionBytes: len(messageContent)}); err != nil {
		return messages, false, err
	}
	choreography.afterCall(descriptor, result)
	return append(messages,
		provider.Message{
			Role: provider.RoleAssistant,
			ToolCalls: []provider.ToolCall{{
				ID: string(callID), Name: providerName, Arguments: arguments,
			}},
		},
		provider.Message{
			Role: provider.RoleTool, Content: messageContent,
			ToolCallID: string(callID), ToolName: providerName,
		},
	), true, nil
}

func explicitReferenceReadPath(request pkgAgent.RunRequest) (string, bool) {
	if !isReferenceAgent(request.Agent()) {
		return "", false
	}
	fields := strings.Fields(request.Instruction())
	if len(fields) < 2 || !strings.EqualFold(fields[0], "read") ||
		strings.ContainsAny(fields[1], "\"'`") {
		return "", false
	}
	return fields[1], true
}

func recoverableReadOnlyInvocationError(descriptor pkgTool.Descriptor, err error) (pkgTool.Result, bool, error) {
	var executionErr *pkgTool.ExecutionError
	if descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceMutate) ||
		!descriptorHasEffect(descriptor, pkgTool.EffectWorkspaceInspect) ||
		!errors.As(err, &executionErr) || executionErr.Kind != pkgTool.ErrorInvalid ||
		executionErr.Reason != "prepare_failed" {
		return pkgTool.Result{}, false, nil
	}
	result, resultErr := recoverableChoreographyResult("invalid_arguments", "use_exact_declared_schema")
	return result, resultErr == nil, resultErr
}

func nextReadyStep(snapshot pkgAgent.SessionSnapshot) (pkgAgent.PlanStep, bool, error) {
	plan, ok := snapshot.Plan()
	if !ok {
		return pkgAgent.PlanStep{}, false, pkgAgent.ErrInvalidPlan
	}
	states := make(map[pkgAgent.StepID]pkgAgent.StepStatus)
	for _, step := range plan.Steps() {
		states[step.ID()] = step.Status()
	}
	for _, step := range plan.Steps() {
		if step.Status() != pkgAgent.StepPending {
			continue
		}
		ready := true
		for _, dependency := range step.Dependencies() {
			if states[dependency] != pkgAgent.StepCompleted && states[dependency] != pkgAgent.StepSkipped {
				ready = false
				break
			}
		}
		if ready {
			return step, true, nil
		}
	}
	for _, status := range states {
		if status != pkgAgent.StepCompleted && status != pkgAgent.StepSkipped {
			return pkgAgent.PlanStep{}, false, pkgAgent.ErrInvalidPlan
		}
	}
	return pkgAgent.PlanStep{}, false, nil
}

func validateProviderResponse(response provider.CompletionResponse, maxCalls int) error {
	switch response.FinishReason {
	case "", provider.FinishReasonStop, provider.FinishReasonLength, provider.FinishReasonToolCalls:
	default:
		return fmt.Errorf("unknown finish reason %q", response.FinishReason)
	}
	if response.Usage.InputTokens < 0 || response.Usage.OutputTokens < 0 || len(response.Message.ToolCalls) > maxCalls {
		return pkgAgent.ErrLimitExceeded
	}
	if response.FinishReason == provider.FinishReasonToolCalls && len(response.Message.ToolCalls) == 0 {
		return errors.New("tool_calls finish reason without calls")
	}
	message := response.Message
	message.Role = provider.RoleAssistant
	return (provider.CompletionRequest{Model: "validation", Messages: []provider.Message{message}}).Validate()
}

func encodeToolResult(result pkgTool.Result) (string, error) {
	payload := struct {
		Outcome     pkgTool.ResultOutcome   `json:"outcome"`
		Content     json.RawMessage         `json:"content,omitempty"`
		MediaType   string                  `json:"media_type,omitempty"`
		Reason      string                  `json:"reason"`
		ItemCount   int                     `json:"item_count"`
		Truncated   bool                    `json:"truncated"`
		Disposition pkgTool.DenyDisposition `json:"disposition,omitempty"`
	}{
		Outcome: result.Outcome(), MediaType: result.MediaType(),
		Reason: result.Reason(), ItemCount: result.ItemCount(), Truncated: result.Truncated(),
		Disposition: result.Disposition(),
	}
	if result.Content() != "" && result.MediaType() == "application/json" && json.Valid([]byte(result.Content())) {
		payload.Content = json.RawMessage(result.Content())
	} else if result.Content() != "" {
		payload.Content, _ = json.Marshal(result.Content())
	}
	encoded, err := json.Marshal(payload)
	return string(encoded), err
}

func permissionTerminal(ctx context.Context, err error) pkgAgent.TerminalReason {
	if ctx.Err() != nil {
		return reasonForError(ctx, pkgAgent.TerminalPermissionDenied)
	}
	return pkgAgent.TerminalPermissionDenied
}

func toolErrorTerminal(ctx context.Context, err error) pkgAgent.TerminalReason {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return pkgAgent.TerminalDeadline
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return pkgAgent.TerminalCanceled
	}
	if errors.Is(err, pkgTool.ErrLimitExceeded) {
		return pkgAgent.TerminalLimit
	}
	if errors.Is(err, pkgTool.ErrPermissionDenied) {
		return pkgAgent.TerminalPermissionDenied
	}
	return pkgAgent.TerminalToolFailure
}
