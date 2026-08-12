package agent

import (
	"context"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

const ReferenceAgentID pkgAgent.ID = "agent.reference"

// ReferenceAgent provides the smallest deterministic planning policy. The
// shared Agent Runtime remains owner of model calls, tools and permissions.
type ReferenceAgent struct{ descriptor pkgAgent.Descriptor }

func NewReferenceAgent() *ReferenceAgent {
	descriptor, err := pkgAgent.NewDescriptor(
		ReferenceAgentID, "1.0.0", "Provider-neutral workspace reference agent.",
		[]pkgRuntime.Capability{pkgAgent.CapabilityPlanning, pkgAgent.CapabilityRun, pkgAgent.CapabilityWorkspaceAware},
	)
	if err != nil {
		panic(err)
	}
	return &ReferenceAgent{descriptor: descriptor}
}

func (agent *ReferenceAgent) Descriptor() pkgAgent.Descriptor { return agent.descriptor }

func (*ReferenceAgent) Plan(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
	if err := ctx.Err(); err != nil {
		return pkgAgent.Plan{}, err
	}
	step, err := pkgAgent.NewPlanStep("execute", request.Instruction(), nil, pkgAgent.StepPending, "")
	if err != nil {
		return pkgAgent.Plan{}, err
	}
	return pkgAgent.NewPlan(1, []pkgAgent.PlanStep{step})
}
