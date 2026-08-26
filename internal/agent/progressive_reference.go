package agent

import (
	"context"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

const ProgressiveReferenceAgentID pkgAgent.ID = "agent.progressive-reference"

// ProgressiveReferenceAgent is a development-only reference-agent candidate.
// Production composition does not register it unless the
// maestro_development build tag is enabled.
type ProgressiveReferenceAgent struct{ descriptor pkgAgent.Descriptor }

func NewProgressiveReferenceAgent() *ProgressiveReferenceAgent {
	descriptor, err := pkgAgent.NewDescriptor(
		ProgressiveReferenceAgentID, "0.1.0-development",
		"Development-only progressive evidence reference agent.",
		[]pkgRuntime.Capability{pkgAgent.CapabilityPlanning, pkgAgent.CapabilityRun, pkgAgent.CapabilityWorkspaceAware},
	)
	if err != nil {
		panic(err)
	}
	return &ProgressiveReferenceAgent{descriptor: descriptor}
}

func (agent *ProgressiveReferenceAgent) Descriptor() pkgAgent.Descriptor { return agent.descriptor }

func (*ProgressiveReferenceAgent) Plan(ctx context.Context, request pkgAgent.PlanningRequest) (pkgAgent.Plan, error) {
	if err := ctx.Err(); err != nil {
		return pkgAgent.Plan{}, err
	}
	step, err := pkgAgent.NewPlanStep("execute", request.Instruction(), nil, pkgAgent.StepPending, "")
	if err != nil {
		return pkgAgent.Plan{}, err
	}
	return pkgAgent.NewPlan(1, []pkgAgent.PlanStep{step})
}

func isReferenceAgent(id pkgAgent.ID) bool {
	return id == ReferenceAgentID || id == ProgressiveReferenceAgentID
}
