package mutation

import (
	"context"
	"testing"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func TestDeterministicQualificationCoverageMatchesFrozenMatrix(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	coverage := DeterministicCoverage()
	if len(coverage) != len(profile.MutationMatrix) {
		t.Fatalf("coverage count differs: %d != %d", len(coverage), len(profile.MutationMatrix))
	}
	report, err := NewDeterministicReport(
		context.Background(), profile, pkgBenchmark.HardwareProfile{OS: "linux", Architecture: "amd64"},
		ReportBuild{Version: "test", Commit: "test"},
		time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if report.State != "passed" || len(report.Samples) != 15 {
		t.Fatalf("unexpected deterministic report: %#v", report)
	}
	for _, sample := range report.Samples {
		if sample.Scenario == "cancellation_after_commit" ||
			sample.Scenario == "refresh_failure_after_commit" ||
			sample.Scenario == "second_mutation_attempt" {
			if sample.FinalSHA256 != profile.Fixture.ExpectedSHA256 || sample.ContextFreshness != "stale" {
				t.Fatalf("post-commit evidence is inaccurate: %#v", sample)
			}
		}
	}
}
