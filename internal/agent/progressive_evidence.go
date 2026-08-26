package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const workspaceSearchToolID pkgTool.ID = "workspace.search"

var progressiveEvidenceStages = []pkgAgent.EvidenceStage{
	pkgAgent.EvidenceStageControllerAction,
	pkgAgent.EvidenceStageReferencedSymbols,
	pkgAgent.EvidenceStageEventsJobsServices,
}

type progressiveEvidenceState struct {
	events        pkgRuntime.EventBus
	run           pkgAgent.RunID
	agent         pkgAgent.ID
	stageIndex    int
	reads         int
	emptySearches int
}

type evidenceDeclaration struct {
	Stage  pkgAgent.EvidenceStage  `json:"evidence_stage"`
	Status pkgAgent.EvidenceStatus `json:"status"`
	Reason string                  `json:"reason"`
}

func newProgressiveEvidenceState(request pkgAgent.RunRequest, events pkgRuntime.EventBus) *progressiveEvidenceState {
	if request.Agent() != ProgressiveReferenceAgentID {
		return nil
	}
	return &progressiveEvidenceState{events: events, run: request.Run(), agent: request.Agent()}
}

func (state *progressiveEvidenceState) observeBootstrap() string {
	if state == nil {
		return ""
	}
	publishEvidenceEvent(
		state.events, state.run, state.agent, pkgAgent.EvidenceStageRoute,
		pkgAgent.EvidenceCovered, pkgAgent.EvidenceDecisionBootstrap,
		string(workspaceReadToolID), pkgAgent.EvidenceResultSuccess,
		pkgAgent.EvidenceStopStageComplete,
	)
	return state.stagePrompt("The route evidence state was covered by the deterministic bootstrap read.")
}

func (state *progressiveEvidenceState) beforeTool(descriptor pkgTool.Descriptor) {
	if state == nil || state.complete() {
		return
	}
	decision, ok := evidenceToolDecision(descriptor.ID())
	if !ok {
		return
	}
	publishEvidenceEvent(
		state.events, state.run, state.agent, state.current(), pkgAgent.EvidencePending,
		decision, string(descriptor.ID()), pkgAgent.EvidenceResultPending, "",
	)
}

func (state *progressiveEvidenceState) afterTool(descriptor pkgTool.Descriptor, result pkgTool.Result) {
	if state == nil || state.complete() {
		return
	}
	decision, ok := evidenceToolDecision(descriptor.ID())
	if !ok {
		return
	}
	evidenceResult := pkgAgent.EvidenceResultFailed
	if result.Outcome() == pkgTool.ResultSuccess {
		evidenceResult = pkgAgent.EvidenceResultSuccess
		switch descriptor.ID() {
		case workspaceReadToolID:
			state.reads++
		case workspaceSearchToolID:
			if result.ItemCount() == 0 {
				state.emptySearches++
				evidenceResult = pkgAgent.EvidenceResultEmpty
			}
		}
	}
	publishEvidenceEvent(
		state.events, state.run, state.agent, state.current(), pkgAgent.EvidencePending,
		decision, string(descriptor.ID()), evidenceResult, "",
	)
}

func (state *progressiveEvidenceState) toolFailed(descriptor pkgTool.Descriptor) {
	if state == nil || state.complete() {
		return
	}
	decision, ok := evidenceToolDecision(descriptor.ID())
	if !ok {
		return
	}
	publishEvidenceEvent(
		state.events, state.run, state.agent, state.current(), pkgAgent.EvidencePending,
		decision, string(descriptor.ID()), pkgAgent.EvidenceResultFailed, "",
	)
}

func (state *progressiveEvidenceState) evaluateAssistant(content string) (bool, string) {
	if state == nil {
		return true, ""
	}
	if state.complete() {
		if _, declared := parseEvidenceDeclaration(content); declared {
			return false, "All evidence states are already closed. Return the final task answer now; do not emit another evidence declaration."
		}
		publishEvidenceEvent(
			state.events, state.run, state.agent, pkgAgent.EvidenceStageEventsJobsServices,
			pkgAgent.EvidenceCovered, pkgAgent.EvidenceDecisionFinalize, "",
			pkgAgent.EvidenceResultAccepted, pkgAgent.EvidenceStopComplete,
		)
		return true, ""
	}
	declaration, declared := parseEvidenceDeclaration(content)
	if !declared {
		publishEvidenceEvent(
			state.events, state.run, state.agent, state.current(), pkgAgent.EvidencePending,
			pkgAgent.EvidenceDecisionFinalize, "", pkgAgent.EvidenceResultRejected,
			pkgAgent.EvidenceStopIncomplete,
		)
		return false, state.stagePrompt("Finalization was rejected because the current evidence state is incomplete.")
	}
	if declaration.Stage != state.current() || strings.TrimSpace(declaration.Reason) == "" {
		return false, state.rejectDeclaration("The declaration did not match the current stage or omitted its reason.")
	}
	switch declaration.Status {
	case pkgAgent.EvidenceCovered:
		if state.reads == 0 {
			return false, state.rejectDeclaration("Covered requires at least one successful relevant workspace read during this stage.")
		}
	case pkgAgent.EvidenceUnavailable:
		if state.emptySearches == 0 {
			return false, state.rejectDeclaration("Unavailable requires at least one successful empty workspace search during this stage.")
		}
	default:
		return false, state.rejectDeclaration("Status must be covered or unavailable.")
	}
	stop := pkgAgent.EvidenceStopStageComplete
	if declaration.Status == pkgAgent.EvidenceUnavailable {
		stop = pkgAgent.EvidenceStopDeclaredUnavailable
	}
	publishEvidenceEvent(
		state.events, state.run, state.agent, state.current(), declaration.Status,
		pkgAgent.EvidenceDecisionDeclare, "", pkgAgent.EvidenceResultAccepted, stop,
	)
	state.stageIndex++
	state.reads = 0
	state.emptySearches = 0
	if state.complete() {
		return false, "All required evidence states are now closed. Return the final answer, distinguish verified facts from uncertainty, and do not invent missing components."
	}
	return false, state.stagePrompt("The previous evidence state was accepted.")
}

func (state *progressiveEvidenceState) rejectDeclaration(reason string) string {
	publishEvidenceEvent(
		state.events, state.run, state.agent, state.current(), pkgAgent.EvidencePending,
		pkgAgent.EvidenceDecisionDeclare, "", pkgAgent.EvidenceResultRejected,
		pkgAgent.EvidenceStopIncomplete,
	)
	return state.stagePrompt(reason)
}

func (state *progressiveEvidenceState) stagePrompt(prefix string) string {
	stage := state.current()
	objective := map[pkgAgent.EvidenceStage]string{
		pkgAgent.EvidenceStageControllerAction:   "follow the route to the concrete controller and action, then inspect their source",
		pkgAgent.EvidenceStageReferencedSymbols:  "inspect the application symbols referenced by the controller/action that are needed by the task",
		pkgAgent.EvidenceStageEventsJobsServices: "inspect the relevant events or notifications, queued jobs, and services, including the synchronous versus queued boundary",
	}[stage]
	return fmt.Sprintf(
		"%s Current evidence stage: %s. You must %s. Use declared read-only workspace tools. Searches are discovery only; read relevant source before declaring covered. Do not return the task answer yet. When the stage is complete, return only this JSON object: {\"evidence_stage\":%q,\"status\":\"covered\",\"reason\":\"concise evidence reason\"}. If relevant evidence cannot be found after a successful empty search, use status \"unavailable\" and explain why in reason.",
		prefix, stage, objective, stage,
	)
}

func (state *progressiveEvidenceState) current() pkgAgent.EvidenceStage {
	if state.complete() {
		return pkgAgent.EvidenceStageEventsJobsServices
	}
	return progressiveEvidenceStages[state.stageIndex]
}

func (state *progressiveEvidenceState) complete() bool {
	return state == nil || state.stageIndex >= len(progressiveEvidenceStages)
}

func parseEvidenceDeclaration(content string) (evidenceDeclaration, bool) {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	var declaration evidenceDeclaration
	if decoder.Decode(&declaration) != nil {
		return evidenceDeclaration{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return evidenceDeclaration{}, false
	}
	return declaration, true
}

func evidenceToolDecision(id pkgTool.ID) (pkgAgent.EvidenceDecision, bool) {
	switch id {
	case workspaceReadToolID:
		return pkgAgent.EvidenceDecisionRead, true
	case workspaceSearchToolID:
		return pkgAgent.EvidenceDecisionSearch, true
	default:
		return "", false
	}
}
