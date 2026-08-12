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
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type agentLoop struct {
	providers   generationRuntime
	tools       pkgTool.Runtime
	permissions pkgTool.Runtime
	context     pkgContext.Engine
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
	lastContent := ""
	turn := 0

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
			if err := current.consume(counterDelta{modelTurns: 1}); err != nil {
				return "", pkgAgent.TerminalLimit, err
			}
			turn++
			completionRequest := provider.CompletionRequest{
				Model: request.Model(), Messages: messages, Tools: providerTools,
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
				if err := current.transitionStep(step.ID(), pkgAgent.StepCompleted, ""); err != nil {
					return "", pkgAgent.TerminalInternalFailure, err
				}
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
				result, err := loop.tools.Invoke(ctx, execution)
				if err != nil {
					if mutating && !current.snapshotValue().ContextStale() {
						_ = current.markStale()
					}
					return "", toolErrorTerminal(ctx, err), errors.Join(pkgAgent.ErrToolFailed, err)
				}
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
				if result.Outcome() == pkgTool.ResultCanceled {
					return "", reasonForError(ctx, pkgAgent.TerminalCanceled), pkgAgent.ErrRunCanceled
				}
				if mutating && current.snapshotValue().ContextStale() {
					workspace, ok := request.WorkspaceTarget()
					if !ok {
						return "", pkgAgent.TerminalToolFailure, errors.Join(pkgAgent.ErrToolFailed, pkgContext.ErrInvalidWorkspace)
					}
					snapshot, refreshErr := loop.context.Index(ctx, workspace)
					if refreshErr != nil {
						return "", reasonForError(ctx, pkgAgent.TerminalToolFailure), errors.Join(pkgAgent.ErrToolFailed, refreshErr)
					}
					refreshed, refreshErr := loop.context.Build(ctx, request.Context())
					if refreshErr != nil {
						return "", reasonForError(ctx, pkgAgent.TerminalToolFailure), errors.Join(pkgAgent.ErrToolFailed, refreshErr)
					}
					if refreshed.Generation() != snapshot.Metadata().Generation {
						return "", pkgAgent.TerminalInternalFailure, errors.New("context refresh returned a stale generation")
					}
					if err := current.markFresh(refreshed.Generation()); err != nil {
						return "", pkgAgent.TerminalInternalFailure, err
					}
					bundle = refreshed
				}
			}
		}
	}
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
		{Role: provider.RoleSystem, Content: "Execute the current plan step using only declared tools. Return a final answer when the step is complete."},
		{Role: provider.RoleUser, Content: user.String()},
	}
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
		Content     string                  `json:"content,omitempty"`
		MediaType   string                  `json:"media_type,omitempty"`
		Reason      string                  `json:"reason"`
		ItemCount   int                     `json:"item_count"`
		Truncated   bool                    `json:"truncated"`
		Disposition pkgTool.DenyDisposition `json:"disposition,omitempty"`
	}{result.Outcome(), result.Content(), result.MediaType(), result.Reason(), result.ItemCount(), result.Truncated(), result.Disposition()}
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
