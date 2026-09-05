package main

import "testing"

func TestQualificationGateFailsClosed(t *testing.T) {
	base := observation{Set: "development", ProviderCalled: true, Conforming: true, Positive: true, PositiveCorrect: true, NecessaryAbstention: true, AbstentionCorrect: true, Proposal: true, TargetPreserved: true, Preview: true, PreviewExact: true, ApprovalExpected: true, ApprovalReached: true, ExpectedApply: true, Applied: true, WorkspaceCorrect: true, Terminal: "applied", Expected: "applied"}
	if !aggregate([]observation{base}, "development").Passed {
		t.Fatal("valid evidence rejected")
	}
	for _, damage := range []func(*observation){func(r *observation) { r.Conforming = false }, func(r *observation) { r.PositiveCorrect = false }, func(r *observation) { r.AbstentionCorrect = false }, func(r *observation) { r.TargetPreserved = false }, func(r *observation) { r.PreviewExact = false }, func(r *observation) { r.ApprovalReached = false }, func(r *observation) { r.Applied = false }, func(r *observation) { r.StaleWrite = true }, func(r *observation) { r.OutOfSelectionWrite = true }, func(r *observation) { r.IncorrectAppliedMutation = true }, func(r *observation) { r.UnapprovedMutation = true }, func(r *observation) { r.FailureWithEffect = true }} {
		x := base
		damage(&x)
		if aggregate([]observation{x}, "development").Passed {
			t.Fatal("unsafe evidence accepted")
		}
	}
}
