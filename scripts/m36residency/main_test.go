package main

import "testing"

func TestEvaluateResidencyGatesFailClosed(t *testing.T) {
	models := func(name string) []runningModel {
		return []runningModel{{Name: name, Model: name, Size: 100, SizeVRAM: 100}}
	}
	r := report{Cycles: 3}
	for cycle := 1; cycle <= 3; cycle++ {
		process := []hostProcess{{ID: 1, Name: "ollama", Command: "ollama serve", Start: "1", ProviderRoot: true}}
		r.Snapshots = append(r.Snapshots, snapshot{Cycle: cycle, Label: "baseline", ProviderProcesses: process})
		for step, name := range []string{"qwen3.5:9b", "qwen3.5:9b", "qwen2.5-coder:14b", "qwen2.5-coder:14b", "qwen3.5:9b", "qwen3.5:9b"} {
			r.Snapshots = append(r.Snapshots, snapshot{Cycle: cycle, Step: step + 1, Models: models(name), ResidencyPattern: name, ProviderProcesses: process})
			r.Requests = append(r.Requests, requestResult{Cycle: cycle, Step: step + 1, Model: name, ObservedModel: name, Terminal: "completed", Correct: true, WithinLatencyLimit: true})
		}
		r.Snapshots = append(r.Snapshots, snapshot{Cycle: cycle, Step: 7, Label: "final_baseline", ProviderProcesses: process})
	}
	if got := evaluate(r); !got.Passed {
		t.Fatalf("valid evidence rejected: %#v", got)
	}
	broken := r
	broken.Requests = append([]requestResult(nil), r.Requests...)
	broken.Requests[8].ObservedModel = "qwen3.5:9b"
	if got := evaluate(broken); got.Passed || got.CrossProfileFallbacks != 1 {
		t.Fatalf("cross-profile fallback accepted: %#v", got)
	}
}
