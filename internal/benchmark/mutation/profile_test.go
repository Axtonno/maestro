package mutation

import (
	"os"
	"strings"
	"testing"
)

func TestPublishedProfileIsStrictAndFrozen(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Digest() == "" || profile.Target.Model != "ibm/granite4.1:8b" {
		t.Fatalf("unexpected published profile: %#v", profile)
	}
	if gate, ok := profile.Gate(GateC); !ok || gate.RequiredPasses != 3 || !gate.FailFast {
		t.Fatalf("unexpected Gate C: %#v", gate)
	}
	resolved, err := profile.Resolve(profile.Configuration.ProductProfile)
	if err != nil || !strings.HasSuffix(resolved, "/configs/maestro.mutating.example.yaml") {
		t.Fatalf("unexpected product profile resolution: %q, %v", resolved, err)
	}
}

func TestProfileRejectsUnknownFieldsAndFrozenGateChanges(t *testing.T) {
	encoded := readPublishedProfile(t)
	if _, err := DecodeProfile(strings.NewReader(encoded + "\nunknown: true\n")); err == nil {
		t.Fatal("unknown profile field was accepted")
	}
	changed := strings.Replace(encoded, "required_passes: 3", "required_passes: 2", 1)
	if _, err := DecodeProfile(strings.NewReader(changed)); err == nil {
		t.Fatal("changed frozen gate was accepted")
	}
}

func readPublishedProfile(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
