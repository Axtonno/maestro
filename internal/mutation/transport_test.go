package mutation

import (
	"errors"
	"fmt"
	"testing"
)

func TestFrozenTransportsCompileToSameCandidateWithoutFallback(t *testing.T) {
	proposal := `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"41","new_text":"42"}`
	inputs := map[Transport]string{
		TransportStructured:     proposal,
		TransportNativeToolCall: fmt.Sprintf(`{"name":"workspace_replace","arguments":%s}`, proposal),
	}
	snapshot := Snapshot{Path: "src/demo.go", Content: "answer = 41\n", Digest: digest("answer = 41\n")}
	var fingerprint string
	for transport, raw := range inputs {
		normalized, err := DecodeTransport(transport, []byte(raw))
		if err != nil {
			t.Fatalf("%s: %v", transport, err)
		}
		candidate, err := Compile(normalized, snapshot)
		if err != nil {
			t.Fatalf("%s compile: %v", transport, err)
		}
		if fingerprint == "" {
			fingerprint = candidate.Fingerprint()
		}
		if candidate.Fingerprint() != fingerprint {
			t.Fatalf("%s changed candidate identity", transport)
		}
	}
}

func TestTransportRejectsMixedMalformedAndCrossTransportOutput(t *testing.T) {
	proposal := `{"version":1,"path":"src/demo.go","operation":"replace","old_text":"41","new_text":"42"}`
	tests := []struct {
		kind Transport
		raw  string
	}{
		{TransportStructured, "Here is the patch: " + proposal},
		{TransportStructured, proposal + proposal},
		{TransportNativeToolCall, proposal},
		{TransportNativeToolCall, fmt.Sprintf(`{"name":"workspace_patch","arguments":%s}`, proposal)},
		{TransportNativeToolCall, fmt.Sprintf(`{"name":"workspace_replace","arguments":%s,"content":"also do this"}`, proposal)},
		{Transport("fallback"), proposal},
	}
	for _, test := range tests {
		if _, err := DecodeTransport(test.kind, []byte(test.raw)); !errors.Is(err, ErrInvalidTransport) {
			t.Fatalf("kind=%s raw=%q: got %v", test.kind, test.raw, err)
		}
	}
}
