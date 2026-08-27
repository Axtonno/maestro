package agent

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	contextengine "github.com/antonio-cafeo/maestro/pkg/contextengine"
	"github.com/antonio-cafeo/maestro/pkg/provider"
	"github.com/antonio-cafeo/maestro/pkg/tool"
)

type RunRequestOptions struct {
	Context    contextengine.BuildRequest
	Tools      []tool.ID
	Approver   tool.Approver
	Streaming  bool
	Generation provider.GenerationOptions
	Workspace  *contextengine.Workspace
}

type RunRequest struct {
	run             RunID
	agent           ID
	provider        provider.ID
	model           string
	workspace       contextengine.WorkspaceID
	policy          tool.PolicyID
	instruction     string
	limits          Limits
	context         contextengine.BuildRequest
	tools           []tool.ID
	approver        tool.Approver
	streaming       bool
	generation      provider.GenerationOptions
	workspaceTarget *contextengine.Workspace
}

func NewRunRequest(
	run RunID,
	agent ID,
	providerID provider.ID,
	model string,
	workspace contextengine.WorkspaceID,
	policy tool.PolicyID,
	instruction string,
	limits Limits,
	options RunRequestOptions,
) (RunRequest, error) {
	if err := run.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent run: %w: %w", err, ErrInvalidRequest)
	}
	if err := agent.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent identity: %w: %w", err, ErrInvalidRequest)
	}
	if !exactValue(string(providerID), 256) || strings.ContainsAny(string(providerID), "\r\n") ||
		!exactValue(model, 512) || strings.ContainsAny(model, "\r\n") {
		return RunRequest{}, fmt.Errorf("provider and model must be explicit and exact: %w", ErrInvalidRequest)
	}
	if err := workspace.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent workspace: %w: %w", err, ErrInvalidRequest)
	}
	if err := policy.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent policy: %w: %w", err, ErrInvalidRequest)
	}
	if strings.TrimSpace(instruction) == "" || len(instruction) > 1<<20 ||
		!utf8.ValidString(instruction) || strings.ContainsRune(instruction, 0) {
		return RunRequest{}, fmt.Errorf("agent instruction is blank, oversized, or unsafe: %w", ErrInvalidRequest)
	}
	if err := limits.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent limits: %w: %w", err, ErrInvalidRequest)
	}
	if err := options.Generation.Validate(); err != nil {
		return RunRequest{}, fmt.Errorf("agent generation options: %w: %w", err, ErrInvalidRequest)
	}
	if err := options.Context.Validate(); err != nil || options.Context.Query.Workspace() != workspace {
		return RunRequest{}, fmt.Errorf("agent context request is invalid or references another workspace: %w", ErrInvalidRequest)
	}
	tools := slices.Clone(options.Tools)
	if len(tools) == 0 {
		return RunRequest{}, fmt.Errorf("agent run requires an explicit tool set: %w", ErrInvalidRequest)
	}
	seen := make(map[tool.ID]struct{}, len(tools))
	for _, toolID := range tools {
		if err := toolID.Validate(); err != nil {
			return RunRequest{}, fmt.Errorf("agent tool: %w: %w", err, ErrInvalidRequest)
		}
		if _, exists := seen[toolID]; exists {
			return RunRequest{}, fmt.Errorf("agent tool %q is duplicated: %w", toolID, ErrInvalidRequest)
		}
		seen[toolID] = struct{}{}
	}
	slices.Sort(tools)
	if options.Approver != nil && nilInterface(options.Approver) {
		return RunRequest{}, fmt.Errorf("agent approver is typed nil: %w", ErrInvalidRequest)
	}
	var workspaceTarget *contextengine.Workspace
	if options.Workspace != nil {
		if err := options.Workspace.Validate(); err != nil || options.Workspace.ID() != workspace {
			return RunRequest{}, fmt.Errorf("agent workspace target is invalid or has another ID: %w", ErrInvalidRequest)
		}
		copyWorkspace := *options.Workspace
		workspaceTarget = &copyWorkspace
	}
	return RunRequest{
		run: run, agent: agent, provider: providerID, model: model,
		workspace: workspace, policy: policy, instruction: instruction, limits: limits,
		context: options.Context, tools: tools, approver: options.Approver,
		streaming: options.Streaming, generation: cloneGenerationOptions(options.Generation),
		workspaceTarget: workspaceTarget,
	}, nil
}

func (request RunRequest) Run() RunID                           { return request.run }
func (request RunRequest) Agent() ID                            { return request.agent }
func (request RunRequest) Provider() provider.ID                { return request.provider }
func (request RunRequest) Model() string                        { return request.model }
func (request RunRequest) Workspace() contextengine.WorkspaceID { return request.workspace }
func (request RunRequest) Policy() tool.PolicyID                { return request.policy }
func (request RunRequest) Instruction() string                  { return request.instruction }
func (request RunRequest) Limits() Limits                       { return request.limits }
func (request RunRequest) Context() contextengine.BuildRequest  { return request.context }
func (request RunRequest) Tools() []tool.ID                     { return slices.Clone(request.tools) }
func (request RunRequest) Approver() tool.Approver              { return request.approver }
func (request RunRequest) Streaming() bool                      { return request.streaming }
func (request RunRequest) GenerationOptions() provider.GenerationOptions {
	return cloneGenerationOptions(request.generation)
}
func (request RunRequest) WorkspaceTarget() (contextengine.Workspace, bool) {
	if request.workspaceTarget == nil {
		return contextengine.Workspace{}, false
	}
	return *request.workspaceTarget, true
}

func (request RunRequest) Validate() error {
	_, err := NewRunRequest(
		request.run, request.agent, request.provider, request.model,
		request.workspace, request.policy, request.instruction, request.limits,
		RunRequestOptions{Context: request.context, Tools: request.tools, Approver: request.approver, Streaming: request.streaming, Generation: request.generation, Workspace: request.workspaceTarget},
	)
	return err
}

func cloneGenerationOptions(options provider.GenerationOptions) provider.GenerationOptions {
	cloned := options
	cloned.Stop = slices.Clone(options.Stop)
	if options.Temperature != nil {
		value := *options.Temperature
		cloned.Temperature = &value
	}
	if options.TopP != nil {
		value := *options.TopP
		cloned.TopP = &value
	}
	return cloned
}

func nilInterface(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
