package mutation

import (
	"errors"
	"strings"
	"testing"
)

func TestCompileQualifiedClassifiesWithoutChangingFrozenCompiler(t *testing.T) {
	content := "one two one\n"
	snapshot := Snapshot{Path: "src/demo.go", Content: content, Digest: digest(content)}
	tests := []struct {
		raw      string
		want     error
		terminal string
	}{
		{`{"version":1,"path":"src/demo.go","operation":"replace","old_text":"absent","new_text":"x"}`, ErrTargetNotFound, "target_not_found"},
		{`{"version":1,"path":"src/demo.go","operation":"replace","old_text":"one","new_text":"x"}`, ErrTargetAmbiguous, "target_ambiguous"},
		{`{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"x"}`, nil, ""},
	}
	for _, test := range tests {
		_, err := CompileQualified([]byte(test.raw), snapshot)
		if !errors.Is(err, test.want) || TerminalForError(err) != test.terminal {
			t.Fatalf("err=%v terminal=%q", err, TerminalForError(err))
		}
	}
	_, err := CompileQualified([]byte(tests[2].raw), Snapshot{Path: snapshot.Path, Content: content, Digest: strings.Repeat("0", 64)})
	if !errors.Is(err, ErrStaleSource) || TerminalForError(err) != "stale_source" {
		t.Fatalf("stale=%v", err)
	}
}
