package benchmark

import (
	"math"
	"sort"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

type measurementSeries struct {
	name   string
	unit   string
	scope  string
	method string
	values []float64
}

func aggregateSamples(samples []pkgBenchmark.Sample) []pkgBenchmark.Aggregate {
	seriesByKey := make(map[string]*measurementSeries)
	for _, sample := range samples {
		if sample.Iteration.Warmup || sample.State != pkgBenchmark.ResultPassed ||
			sample.CleanupError != nil {
			continue
		}
		for _, measurement := range sample.Measurements {
			key := measurement.Name + "\x00" + measurement.Unit + "\x00" +
				measurement.Scope + "\x00" + measurement.Method
			series := seriesByKey[key]
			if series == nil {
				series = &measurementSeries{
					name: measurement.Name, unit: measurement.Unit,
					scope: measurement.Scope, method: measurement.Method,
				}
				seriesByKey[key] = series
			}
			series.values = append(series.values, measurement.Value)
		}
	}

	keys := make([]string, 0, len(seriesByKey))
	for key := range seriesByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	aggregates := make([]pkgBenchmark.Aggregate, 0, len(keys))
	for _, key := range keys {
		series := seriesByKey[key]
		sort.Float64s(series.values)
		count := len(series.values)
		aggregate := pkgBenchmark.Aggregate{
			Name: series.name, Unit: series.unit, Scope: series.scope,
			Method: series.method, Count: count,
			Min:    series.values[0],
			Median: median(series.values),
			Max:    series.values[count-1],
		}
		if count >= 20 {
			p95 := series.values[int(math.Ceil(float64(count)*0.95))-1]
			aggregate.P95 = &p95
		}
		aggregates = append(aggregates, aggregate)
	}

	return aggregates
}

func median(sorted []float64) float64 {
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}

	return (sorted[middle-1] + sorted[middle]) / 2
}
