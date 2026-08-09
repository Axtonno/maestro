package benchmark

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func WriteReportJSON(path string, report pkgBenchmark.Report) (writeError error) {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("write benchmark report: path is empty")
	}

	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".maestro-benchmark-*.json")
	if err != nil {
		return fmt.Errorf("create temporary benchmark report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if writeError != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect temporary benchmark report: %w", err)
	}
	if err := EncodeReportJSON(temporary, report); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary benchmark report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary benchmark report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish benchmark report: %w", err)
	}

	return nil
}

// EncodeReportJSON is the supported serialization boundary for benchmark
// reports. It validates and redacts a copy before writing any bytes.
func EncodeReportJSON(writer io.Writer, report pkgBenchmark.Report) error {
	if writer == nil {
		return fmt.Errorf("encode benchmark report: writer is nil")
	}
	if err := report.Validate(); err != nil {
		return err
	}

	redacted := redactReport(report)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(redacted); err != nil {
		return fmt.Errorf("encode benchmark report JSON: %w", err)
	}

	return nil
}

func redactReport(report pkgBenchmark.Report) pkgBenchmark.Report {
	report.Configuration.Provider.Endpoint = redactEndpoint(
		report.Configuration.Provider.Endpoint,
	)
	report.Configuration.Model.ID = redactUserPath(report.Configuration.Model.ID)
	if report.Configuration.Models != nil {
		models := make(map[string]pkgBenchmark.ModelProfile, len(report.Configuration.Models))
		for role, model := range report.Configuration.Models {
			model.ID = redactUserPath(model.ID)
			models[role] = model
		}
		report.Configuration.Models = models
	}
	report.Configuration.Dataset.ID = redactUserPath(report.Configuration.Dataset.ID)
	report.Configuration.Plugins = append(
		[]pkgBenchmark.PluginProfile(nil),
		report.Configuration.Plugins...,
	)
	report.Scenarios = append([]pkgBenchmark.ScenarioReport(nil), report.Scenarios...)
	for scenarioIndex := range report.Scenarios {
		scenario := &report.Scenarios[scenarioIndex]
		scenario.Samples = append([]pkgBenchmark.Sample(nil), scenario.Samples...)
		scenario.Aggregates = append(
			[]pkgBenchmark.Aggregate(nil),
			scenario.Aggregates...,
		)
		for sampleIndex := range scenario.Samples {
			sample := &scenario.Samples[sampleIndex]
			sample.Measurements = append(
				[]pkgBenchmark.Measurement(nil),
				sample.Measurements...,
			)
			sample.Error = redactError(sample.Error)
			sample.CleanupError = redactError(sample.CleanupError)
		}
	}

	return report
}

func redactError(record *pkgBenchmark.ErrorRecord) *pkgBenchmark.ErrorRecord {
	if record == nil {
		return nil
	}

	redacted := *record
	redacted.Model = redactUserPath(redacted.Model)

	return &redacted
}

func redactEndpoint(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[redacted-endpoint]"
	}

	return (&url.URL{Scheme: parsed.Scheme, Host: parsed.Host}).String()
}

func redactUserPath(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	homeDirectory, _ := os.UserHomeDir()
	if (homeDirectory != "" && strings.Contains(value, homeDirectory)) ||
		filepath.IsAbs(value) || looksLikeWindowsPath(value) {
		return "[redacted-path]"
	}

	return value
}

func looksLikeWindowsPath(value string) bool {
	return len(value) >= 3 &&
		((value[0] >= 'a' && value[0] <= 'z') ||
			(value[0] >= 'A' && value[0] <= 'Z')) &&
		value[1] == ':' && (value[2] == '\\' || value[2] == '/')
}
