package benchmark

import (
	"testing"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func TestAggregateSamplesPublishesP95OnlyWithEnoughSamples(t *testing.T) {
	samples := make([]pkgBenchmark.Sample, 20)
	for index := range samples {
		samples[index] = pkgBenchmark.Sample{
			Iteration: pkgBenchmark.Iteration{Index: index + 1},
			StartedAt: time.Now(),
			State:     pkgBenchmark.ResultPassed,
			Measurements: []pkgBenchmark.Measurement{{
				Name: "latency", Unit: "ms", Value: float64(index + 1),
			}},
		}
	}

	aggregates := aggregateSamples(samples)
	if len(aggregates) != 1 || aggregates[0].P95 == nil ||
		*aggregates[0].P95 != 19 {
		t.Fatalf("unexpected p95: %#v", aggregates)
	}
}
