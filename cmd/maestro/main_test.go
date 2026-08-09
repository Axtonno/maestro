package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBenchValidateLoadsVersionedManifest(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(
		[]string{
			"bench", "validate", "--manifest",
			"../../docs/provider-smoke-benchmark-manifest.yaml",
		},
		&stdout,
		&stderr,
	)

	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stderr=%q", exitCode, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "providers=2 scenarios=14") {
		t.Fatalf("unexpected validation output: %q", got)
	}
}

func TestBenchHelpAndUnknownCommandHaveStableExitCodes(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := run([]string{"bench", "--help"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("help exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "maestro bench <command>") {
		t.Fatalf("unexpected help: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := run([]string{"bench", "unknown"}, &stdout, &stderr); exitCode != 2 {
		t.Fatalf("unknown command exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "unknown bench command") {
		t.Fatalf("unexpected error: %q", stderr.String())
	}
}

func TestRootOutputIsWrittenOnlyToProvidedWriter(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := run(nil, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("root exit code: %d", exitCode)
	}
	if strings.Count(stdout.String(), "Maestro AI Runtime") != 1 || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
