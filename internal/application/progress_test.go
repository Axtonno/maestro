package application

import (
	"bytes"
	"strings"
	"testing"

	"github.com/antonio-cafeo/maestro"
	pkgAgent "github.com/antonio-cafeo/maestro/pkg/agent"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestProgressRendererUsesOnlyRedactedEventPayloads(t *testing.T) {
	runtime := maestro.New()
	var output bytes.Buffer
	renderer := NewProgressRenderer(&output)
	if err := renderer.Subscribe(runtime.EventBus()); err != nil {
		t.Fatal(err)
	}
	renderer.RenderLimits(pkgAgent.Limits{MaxDuration: 10, MaxModelTurns: 2, MaxToolCalls: 3, MaxToolCallsPerTurn: 1, MaxPlanSteps: 2, MaxPlanRevisions: 1, MaxToolResultBytes: 1024, MaxSessionBytes: 2048, MaxInputTokens: 100, MaxOutputTokens: 50})
	events := runtime.EventBus()
	_ = events.Publish(pkgAgent.Event{Topic: pkgAgent.EventPlanCreated, Data: pkgAgent.EventPayload{Run: "run-safe", PlanVersion: 2}})
	_ = events.Publish(pkgAgent.Event{Topic: pkgAgent.EventStepTransitioned, Data: pkgAgent.EventPayload{Run: "run-safe", Step: "execute", StepState: pkgAgent.StepRunning}})
	_ = events.Publish(pkgTool.Event{Topic: pkgTool.EventPermissionDecided, Data: pkgTool.EventPayload{Run: "run-safe", Decision: pkgTool.DecisionAllow, ActionCount: 2}})
	_ = events.Publish(pkgAgent.Event{Topic: pkgAgent.EventSessionCompleted, Data: pkgAgent.EventPayload{Run: "run-safe", Terminal: pkgAgent.TerminalCompleted, ModelTurns: 2, ToolCalls: 1, DurationMillis: 12}})
	got := output.String()
	for _, expected := range []string{"limits\tduration=10ns model_turns=2 tool_calls=3", "plan\trun=run-safe version=2", "step\trun=run-safe id=execute state=running", "permission\trun=run-safe decision=allow actions=2", "terminal\trun=run-safe reason=completed model_turns=2 tool_calls=1 duration_ms=12"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("output %q lacks %q", got, expected)
		}
	}
}

func TestProgressRendererContainsPanickingWriter(t *testing.T) {
	runtime := maestro.New()
	renderer := NewProgressRenderer(panicWriter{})
	if err := renderer.Subscribe(runtime.EventBus()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.EventBus().Publish(pkgAgent.Event{Topic: pkgAgent.EventSessionStarted, Data: pkgAgent.EventPayload{Run: "run-safe"}}); err != nil {
		t.Fatal(err)
	}
}

type panicWriter struct{}

func (panicWriter) Write([]byte) (int, error) { panic("writer failed") }
