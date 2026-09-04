package main

import "testing"

func TestDecisionRejectsWhenNeitherTransportPasses(t *testing.T) {
	selected, verdict, authorized := decide(map[string]metrics{
		"native_tool_call":              {Pass: false},
		"constrained_structured_output": {Pass: false},
	})
	if selected != nil || verdict != "controlled_mutation_model_transport_rejected" || authorized {
		t.Fatalf("unexpected decision: selected=%v verdict=%s authorized=%t", selected, verdict, authorized)
	}
}

func TestDecisionUsesFrozenTieBreak(t *testing.T) {
	selected, verdict, authorized := decide(map[string]metrics{
		"native_tool_call":              {Pass: true, Completions: 10, ValidProposals: 7, P95MS: 100},
		"constrained_structured_output": {Pass: true, Completions: 10, ValidProposals: 7, P95MS: 100},
	})
	if selected == nil || *selected != "constrained_structured_output" || verdict != "controlled_mutation_transport_qualified" || !authorized {
		t.Fatalf("unexpected tie decision: selected=%v verdict=%s authorized=%t", selected, verdict, authorized)
	}
}
