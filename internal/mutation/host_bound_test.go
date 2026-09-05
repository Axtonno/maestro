package mutation

import (
	"errors"
	"testing"
)

func TestHostBoundSelectionSplicesExactCoordinates(t *testing.T) {
	for _, tc := range []struct {
		before                       string
		start, end                   int
		selected, replacement, after string
	}{
		{"same\nsame\n", 2, 2, "same", "other", "same\nother\n"},
		{"à\r\nβ\r\n尾\r\n", 1, 2, "à\r\nβ", "è\r\nγ", "è\r\nγ\r\n尾\r\n"},
		{"one\ntwo", 2, 2, "two", "", "one\n"},
		{"\nlast\n", 1, 1, "", "first", "first\nlast\n"},
	} {
		s, err := Select(Snapshot{"app/Test.php", tc.before, digest(tc.before)}, tc.start, tc.end)
		if err != nil || s.Text() != tc.selected {
			t.Fatalf("selection: %q %v", s.Text(), err)
		}
		after, err := s.Replace(tc.replacement)
		if err != nil || after != tc.after {
			t.Fatalf("splice: %q %v", after, err)
		}
		if s.Fingerprint(tc.replacement, "diff") == s.Fingerprint(tc.replacement, "changed diff") {
			t.Fatal("diff not bound")
		}
	}
	for _, span := range [][2]int{{0, 1}, {2, 1}, {1, 3}, {3, 3}} {
		_, err := Select(Snapshot{"app/Test.php", "a\nb\n", digest("a\nb\n")}, span[0], span[1])
		if !errors.Is(err, ErrSelectionOutOfBounds) {
			t.Fatalf("range %v: %v", span, err)
		}
	}
	a, _ := Select(Snapshot{"app/Test.php", "x\nx", digest("x\nx")}, 1, 1)
	b, _ := Select(Snapshot{"app/Test.php", "x\nx", digest("x\nx")}, 2, 2)
	if a.Fingerprint("y", "diff") == b.Fingerprint("y", "diff") {
		t.Fatal("coordinates not bound")
	}
}

func TestHostBoundDecoderStrict(t *testing.T) {
	for _, raw := range []string{`{"decision":"abstain"}`, `{"decision":"propose","new_text":""}`} {
		if _, err := DecodeHostBoundDecision([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{`null`, `[]`, `{"decision":"propose"}`, `{"decision":"propose","new_text":null}`, `{"decision":"abstain","new_text":"x"}`, `{"decision":"abstain","decision":"abstain"}`, `{"decision":"propose","new_text":"x","path":"app/Other.php"}`, `{"decision":"propose","new_text":"x","old_text":"y"}`, `{"decision":"propose","new_text":"x","start_line":2}`, `{"decision":"abstain"} {}`, `{"decision":"propose","new_text":12}`} {
		if _, err := DecodeHostBoundDecision([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
