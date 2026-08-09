package benchmark

import (
	"errors"
	"strings"
	"testing"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func TestDecodeManifestLoadsProviderHandoff(t *testing.T) {
	manifest, err := LoadManifest("../../docs/provider-smoke-benchmark-manifest.yaml")
	if err != nil {
		t.Fatalf("load provider handoff: %v", err)
	}

	if manifest.Version != pkgBenchmark.ManifestSchemaVersion ||
		len(manifest.Providers) != 2 || len(manifest.Scenarios) != 14 {
		t.Fatalf("unexpected provider handoff: %#v", manifest)
	}
	if manifest.Scenarios[0].ID != "capability-instance" ||
		manifest.Scenarios[len(manifest.Scenarios)-1].ID != "observability-redaction" {
		t.Fatalf("scenario order was not preserved: %#v", manifest.Scenarios)
	}
}

func TestDecodeManifestIsStrictAndSingleDocument(t *testing.T) {
	base := `
version: 1
owner: tests
source_milestone: tests
report_redaction:
  exclude: []
providers:
  test: {}
scenarios:
  - id: completion
    capability: completion
    model_role: chat
    cleanup: none
result_states: [passed, failed, skipped, unsupported]
`

	tests := []struct {
		name string
		yaml string
	}{
		{name: "unknown field", yaml: base + "unknown: true\n"},
		{name: "second document", yaml: base + "---\nversion: 1\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodeManifest(strings.NewReader(test.yaml)); err == nil {
				t.Fatal("expected strict decoding error")
			}
		})
	}
}

func TestDecodeManifestRejectsNilReader(t *testing.T) {
	_, err := DecodeManifest(nil)
	if !errors.Is(err, pkgBenchmark.ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest, got %v", err)
	}
}
