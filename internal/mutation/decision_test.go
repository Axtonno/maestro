package mutation

import "testing"

func TestDecodeDecisionAcceptsStrictProposalAndAbstentions(t *testing.T) {
	valid := []string{
		`{"version":1,"decision":"propose","proposal":{"version":1,"path":"src/a.go","operation":"replace","old_text":"a","new_text":"b"}}`,
		`{"version":1,"decision":"abstain_missing_information"}`,
		`{"version":1,"decision":"abstain_target_not_found"}`,
		`{"version":1,"decision":"abstain_target_ambiguous"}`,
	}
	for _, raw := range valid {
		if _, err := DecodeDecision([]byte(raw)); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
	}
}

func TestDecodeDecisionRejectsNonStrictEnvelopes(t *testing.T) {
	invalid := []string{
		`{"version":1,"decision":"propose"}`,
		`{"version":1,"decision":"abstain_target_not_found","proposal":{"version":1}}`,
		`{"version":1,"decision":"unknown"}`,
		`{"version":1,"decision":"abstain_missing_information","extra":true}`,
		`{"version":1,"version":1,"decision":"abstain_missing_information"}`,
		`{"version":1,"decision":"abstain_missing_information"} {}`,
	}
	for _, raw := range invalid {
		if _, err := DecodeDecision([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
