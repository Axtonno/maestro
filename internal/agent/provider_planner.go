package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	"github.com/antonio-cafeo/maestro/pkg/provider"
)

type completionRuntime interface {
	Complete(context.Context, provider.ID, provider.CompletionRequest) (provider.CompletionResponse, error)
}

type ProviderPlanner struct {
	runtime  completionRuntime
	provider provider.ID
	model    string
}

func NewProviderPlanner(runtime completionRuntime, providerID provider.ID, model string) (*ProviderPlanner, error) {
	if runtime == nil || typedNil(runtime) || strings.TrimSpace(string(providerID)) == "" ||
		strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("provider planner target is invalid: %w", pkgAgent.ErrInvalidRequest)
	}
	return &ProviderPlanner{runtime: runtime, provider: providerID, model: model}, nil
}

func (planner *ProviderPlanner) Plan(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
	plan, _, err := planner.PlanMeasured(ctx, request)
	return plan, err
}

func (planner *ProviderPlanner) Target() (provider.ID, string) {
	return planner.provider, planner.model
}

func (planner *ProviderPlanner) PlanMeasured(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, provider.Usage, error) {
	if ctx == nil {
		return pkgAgent.Plan{}, provider.Usage{}, pkgAgent.ErrInvalidRequest
	}
	schema := json.RawMessage(`{"type":"object","additionalProperties":false,"required":["steps"],"properties":{"steps":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["id","objective","dependencies"],"properties":{"id":{"type":"string"},"objective":{"type":"string"},"dependencies":{"type":"array","items":{"type":"string"}}}}}}}`)
	var user strings.Builder
	user.WriteString(request.Instruction())
	for _, section := range request.Bundle().Sections() {
		user.WriteString("\n\n--- workspace evidence: ")
		user.WriteString(string(section.Path))
		user.WriteString(" ---\n")
		user.WriteString(section.Text)
	}
	response, err := planner.runtime.Complete(ctx, planner.provider, provider.CompletionRequest{
		Model: planner.model,
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "Produce a bounded execution plan. Return only the requested JSON object."},
			{Role: provider.RoleUser, Content: user.String()},
		},
		Output: &provider.StructuredOutput{Mode: provider.StructuredOutputJSONSchema, Schema: schema},
	})
	if err != nil {
		return pkgAgent.Plan{}, provider.Usage{}, fmt.Errorf("provider planning: %w: %w", err, pkgAgent.ErrPlanningFailed)
	}
	var document struct {
		Steps []struct {
			ID           string   `json:"id"`
			Objective    string   `json:"objective"`
			Dependencies []string `json:"dependencies"`
		} `json:"steps"`
	}
	decoder := json.NewDecoder(bytes.NewBufferString(response.Message.Content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return pkgAgent.Plan{}, provider.Usage{}, fmt.Errorf("decode provider plan: %w: %w", err, pkgAgent.ErrInvalidPlan)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return pkgAgent.Plan{}, provider.Usage{}, fmt.Errorf("provider plan contains trailing JSON: %w", pkgAgent.ErrInvalidPlan)
	}
	if len(document.Steps) == 0 || len(document.Steps) > request.MaxSteps() {
		return pkgAgent.Plan{}, provider.Usage{}, fmt.Errorf("provider plan cardinality is invalid: %w", pkgAgent.ErrLimitExceeded)
	}
	steps := make([]pkgAgent.PlanStep, 0, len(document.Steps))
	for _, item := range document.Steps {
		dependencies := make([]pkgAgent.StepID, len(item.Dependencies))
		for index, dependency := range item.Dependencies {
			dependencies[index] = pkgAgent.StepID(dependency)
		}
		step, err := pkgAgent.NewPlanStep(pkgAgent.StepID(item.ID), item.Objective, dependencies, pkgAgent.StepPending, "")
		if err != nil {
			return pkgAgent.Plan{}, provider.Usage{}, fmt.Errorf("provider plan step %q: %w: %w", item.ID, err, pkgAgent.ErrInvalidPlan)
		}
		steps = append(steps, step)
	}
	plan, err := pkgAgent.NewPlan(1, steps)
	if err != nil {
		return pkgAgent.Plan{}, provider.Usage{}, err
	}
	return plan, response.Usage, nil
}
