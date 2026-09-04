package mutation

import "testing"

func TestDecodeBinaryDecision(t *testing.T) {
	propose, err := DecodeBinaryDecision([]byte(`{"version":1,"decision":"propose","path":"src/a.go","operation":"replace","old_text":"a","new_text":"b"}`))
	if err != nil || propose.Decision != BinaryPropose || len(propose.Proposal) == 0 {
		t.Fatalf("propose=%#v err=%v", propose, err)
	}
	abstain, err := DecodeBinaryDecision([]byte(`{"version":1,"decision":"abstain"}`))
	if err != nil || abstain.Decision != BinaryAbstain || len(abstain.Proposal) != 0 {
		t.Fatalf("abstain=%#v err=%v", abstain, err)
	}
}

func TestDecodeBinaryDecisionRejectsNonStrictOutput(t *testing.T) {
	for _, raw := range []string{
		`{"version":1,"decision":"propose"}`,
		`{"version":1,"decision":"abstain","path":"src/a.go"}`,
		`{"version":1,"decision":"other"}`,
		`{"version":1,"decision":"abstain","decision":"abstain"}`,
		`{"version":1,"decision":"abstain"} {}`,
	} {
		if _, err := DecodeBinaryDecision([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
