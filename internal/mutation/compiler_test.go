package mutation

import (
	"errors"
	"strings"
	"testing"
)

func TestCompileBindsExactProposalToAuthoritativeSnapshot(t *testing.T) {
	before := "package demo\n\nconst answer = 41\n"
	raw := []byte(`{"version":1,"path":"src/demo.go","operation":"replace","old_text":"answer = 41","new_text":"answer = 42"}`)
	candidate, err := Compile(raw, Snapshot{Path: "src/demo.go", Content: before, Digest: digest(before)})
	if err != nil {
		t.Fatal(err)
	}
	if candidate.After() != "package demo\n\nconst answer = 42\n" || candidate.BeforeDigest() != digest(before) ||
		candidate.AfterDigest() != digest(candidate.After()) || len(candidate.Fingerprint()) != 64 {
		t.Fatalf("unexpected candidate: %#v", candidate)
	}
	again, err := Compile(raw, Snapshot{Path: "src/demo.go", Content: before, Digest: digest(before)})
	if err != nil || again.Fingerprint() != candidate.Fingerprint() {
		t.Fatal("compilation is not deterministic")
	}
}

func TestCompileRejectsUntrustedProposalMatrix(t *testing.T) {
	content := "one two one\n"
	validSnapshot := Snapshot{Path: "src/demo.go", Content: content, Digest: digest(content)}
	tests := []struct {
		name     string
		raw      string
		snapshot Snapshot
		want     error
	}{
		{"missing", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two"}`, validSnapshot, ErrInvalidProposal},
		{"unknown", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"three","extra":true}`, validSnapshot, ErrInvalidProposal},
		{"duplicate", `{"version":1,"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"three"}`, validSnapshot, ErrInvalidProposal},
		{"trailing", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"three"} text`, validSnapshot, ErrInvalidProposal},
		{"wrong version", `{"version":2,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"three"}`, validSnapshot, ErrInvalidProposal},
		{"wrong operation", `{"version":1,"path":"src/demo.go","operation":"write","old_text":"two","new_text":"three"}`, validSnapshot, ErrInvalidProposal},
		{"path mismatch", `{"version":1,"path":"src/other.go","operation":"replace","old_text":"two","new_text":"three"}`, validSnapshot, ErrPathMismatch},
		{"traversal", `{"version":1,"path":"../demo.go","operation":"replace","old_text":"two","new_text":"three"}`, validSnapshot, ErrInvalidProposal},
		{"secret", `{"version":1,"path":"secrets/key.txt","operation":"replace","old_text":"two","new_text":"three"}`, Snapshot{Path: "secrets/key.txt", Content: content, Digest: digest(content)}, ErrSensitiveTarget},
		{"stale digest", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"three"}`, Snapshot{Path: "src/demo.go", Content: content, Digest: strings.Repeat("0", 64)}, ErrPrecondition},
		{"absent", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"absent","new_text":"three"}`, validSnapshot, ErrPrecondition},
		{"ambiguous", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"one","new_text":"three"}`, validSnapshot, ErrPrecondition},
		{"noop", `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"two","new_text":"two"}`, validSnapshot, ErrInvalidProposal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Compile([]byte(test.raw), test.snapshot); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
