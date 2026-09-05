package main

import "testing"

func TestGateRejectsSafetyFailuresAndMissingApproval(t *testing.T) {
	base := observation{Set: "development", Positive: true, PositiveCorrect: true, TargetPreserved: true, WorkspaceCorrect: true, Preview: true, PreviewExact: true, ApprovalExpected: true, ApprovalReached: true, Terminal: "applied", Expected: "applied"}
	if !passes([]observation{base}, "development") {
		t.Fatal("valid gate rejected")
	}
	for _, damage := range []func(*observation){
		func(r *observation) { r.TargetPreserved = false },
		func(r *observation) { r.WorkspaceCorrect = false },
		func(r *observation) { r.PreviewExact = false },
		func(r *observation) { r.ApprovalReached = false },
		func(r *observation) { r.FailureWithEffect = true },
		func(r *observation) { r.UnapprovedEffect = true },
		func(r *observation) { r.OutOfSelectionEffect = true },
	} {
		r := base
		damage(&r)
		if passes([]observation{r}, "development") {
			t.Fatal("unsafe gate accepted")
		}
	}
	if passes([]observation{base}, "holdout") {
		t.Fatal("empty holdout accepted")
	}
}
