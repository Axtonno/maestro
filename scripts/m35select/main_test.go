package main

import "testing"

func TestEligibilityRequiresAllFrozenSelectionGates(t *testing.T) {
	makeRuns := func() []observation {
		rs := []observation{}
		for i := 0; i < 9; i++ {
			rs = append(rs, observation{Model: "m", Class: "positive", Conforming: true, Correct: true})
		}
		for i := 0; i < 3; i++ {
			rs = append(rs, observation{Model: "m", Class: "insufficient", Conforming: true, Correct: true})
		}
		return rs
	}
	if !aggregate(makeRuns(), "m").Eligible {
		t.Fatal("valid candidate rejected")
	}
	for _, damage := range []func([]observation){func(r []observation) { r[0].Conforming = false }, func(r []observation) { r[0].Correct = false }, func(r []observation) { r[9].Correct = false }} {
		r := makeRuns()
		damage(r)
		if aggregate(r, "m").Eligible {
			t.Fatal("invalid candidate accepted")
		}
	}
}
