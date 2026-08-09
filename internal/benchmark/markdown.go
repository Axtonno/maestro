package benchmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

const maxBenchmarkReportBytes = 64 << 20

func DecodeReportJSON(reader io.Reader) (pkgBenchmark.Report, error) {
	if reader == nil {
		return pkgBenchmark.Report{}, errors.New("decode benchmark report: reader is nil")
	}
	encoded, err := io.ReadAll(io.LimitReader(reader, maxBenchmarkReportBytes+1))
	if err != nil {
		return pkgBenchmark.Report{}, fmt.Errorf("read benchmark report JSON: %w", err)
	}
	if len(encoded) > maxBenchmarkReportBytes {
		return pkgBenchmark.Report{}, fmt.Errorf("benchmark report exceeds %d bytes", maxBenchmarkReportBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	var report pkgBenchmark.Report
	if err := decoder.Decode(&report); err != nil {
		return pkgBenchmark.Report{}, fmt.Errorf("decode benchmark report JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return pkgBenchmark.Report{}, errors.New("benchmark report contains multiple JSON values")
		}
		return pkgBenchmark.Report{}, fmt.Errorf("decode benchmark report trailing content: %w", err)
	}
	if err := report.Validate(); err != nil {
		return pkgBenchmark.Report{}, err
	}
	return report, nil
}

func WriteReportMarkdown(path string, report pkgBenchmark.Report) (writeError error) {
	if strings.TrimSpace(path) == "" {
		return errors.New("write benchmark Markdown report: path is empty")
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".maestro-benchmark-*.md")
	if err != nil {
		return fmt.Errorf("create temporary Markdown report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if writeError != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect temporary Markdown report: %w", err)
	}
	if err := EncodeReportMarkdown(temporary, report); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary Markdown report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary Markdown report: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish Markdown report: %w", err)
	}
	return nil
}

func EncodeReportMarkdown(writer io.Writer, report pkgBenchmark.Report) error {
	if writer == nil {
		return errors.New("encode benchmark Markdown report: writer is nil")
	}
	if err := report.Validate(); err != nil {
		return err
	}
	report = redactReport(report)
	var output strings.Builder
	output.WriteString("# Maestro Benchmark Report\n\n")
	output.WriteString("Generated from benchmark JSON schema `" + markdownInline(report.SchemaVersion) + "`.\n\n")

	output.WriteString("## Run\n\n")
	writeMarkdownRows(&output, [][2]string{
		{"Run ID", report.RunID},
		{"Command", valueOrDash(report.Metadata.Command)},
		{"Created", formatTime(report.CreatedAt)},
		{"Completed", formatTime(report.CompletedAt)},
		{"Duration", formatNumber(report.DurationMS) + " ms"},
		{"Manifest", fmt.Sprintf("v%d — %s", report.Metadata.ManifestVersion, report.Metadata.ManifestOwner)},
		{"Maestro version", valueOrDash(report.Metadata.MaestroVersion)},
		{"Maestro commit", valueOrDash(report.Metadata.MaestroCommit)},
	})

	configuration := report.Configuration
	output.WriteString("## Configuration\n\n")
	writeMarkdownRows(&output, [][2]string{
		{"Operating system", valueOrDash(configuration.Hardware.OS)},
		{"Architecture", valueOrDash(configuration.Hardware.Architecture)},
		{"CPU", valueOrDash(configuration.Hardware.CPU)},
		{"Logical CPUs", optionalInteger(int64(configuration.Hardware.LogicalCPUs))},
		{"Memory", optionalSize(configuration.Hardware.MemoryMB)},
		{"GPU", valueOrDash(configuration.Hardware.GPU)},
		{"Backend", valueOrDash(configuration.Hardware.Backend)},
		{"VRAM", optionalSize(configuration.Hardware.VRAMMB)},
		{"Provider", valueOrDash(configuration.Provider.ID)},
		{"Provider version", valueOrDash(configuration.Provider.ServerVersion)},
		{"Endpoint", valueOrDash(configuration.Provider.Endpoint)},
		{"Primary model", valueOrDash(configuration.Model.ID)},
		{"Dataset", joinedIdentity(configuration.Dataset.ID, configuration.Dataset.Version)},
		{"Warmup / runs", fmt.Sprintf("%d / %d", configuration.Execution.Warmup, configuration.Execution.Runs)},
	})

	writeModelProfiles(&output, configuration.Models)
	writePluginProfiles(&output, configuration.Plugins)

	output.WriteString("## Scenario summary\n\n")
	output.WriteString("| Scenario | State | Measured samples | Quality |\n")
	output.WriteString("|---|---:|---:|---:|\n")
	for _, scenario := range report.Scenarios {
		fmt.Fprintf(
			&output,
			"| %s | %s | %d | %s |\n",
			markdownCell(scenario.Scenario.ID),
			markdownCell(string(scenario.State)),
			measuredSampleCount(scenario.Samples),
			markdownCell(qualitySummary(scenario.Samples)),
		)
	}
	output.WriteString("\n")

	for index, scenario := range report.Scenarios {
		fmt.Fprintf(&output, "## Scenario %d\n\n", index+1)
		writeMarkdownRows(&output, [][2]string{
			{"ID", scenario.Scenario.ID},
			{"Capability", scenario.Scenario.Capability},
			{"Model role", scenario.Scenario.ModelRole},
			{"State", string(scenario.State)},
		})
		writeAggregates(&output, scenario.Aggregates)
		writeSamples(&output, scenario.Samples)
	}
	if _, err := io.WriteString(writer, output.String()); err != nil {
		return fmt.Errorf("encode benchmark Markdown report: %w", err)
	}
	return nil
}

func writeMarkdownRows(output *strings.Builder, rows [][2]string) {
	output.WriteString("| Field | Value |\n|---|---|\n")
	for _, row := range rows {
		fmt.Fprintf(output, "| %s | %s |\n", markdownCell(row[0]), markdownCell(row[1]))
	}
	output.WriteString("\n")
}

func writeModelProfiles(output *strings.Builder, models map[string]pkgBenchmark.ModelProfile) {
	if len(models) == 0 {
		return
	}
	roles := make([]string, 0, len(models))
	for role := range models {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	output.WriteString("### Models by role\n\n| Role | Model | Format | Quantization | Context |\n|---|---|---|---|---:|\n")
	for _, role := range roles {
		model := models[role]
		fmt.Fprintf(output, "| %s | %s | %s | %s | %s |\n",
			markdownCell(role), markdownCell(valueOrDash(model.ID)),
			markdownCell(valueOrDash(model.Format)), markdownCell(valueOrDash(model.Quantization)),
			markdownCell(optionalInteger(int64(model.ContextLength))))
	}
	output.WriteString("\n")
}

func writePluginProfiles(output *strings.Builder, plugins []pkgBenchmark.PluginProfile) {
	if len(plugins) == 0 {
		return
	}
	output.WriteString("### Plugins\n\n| Plugin | Version |\n|---|---|\n")
	for _, plugin := range plugins {
		fmt.Fprintf(output, "| %s | %s |\n", markdownCell(plugin.ID), markdownCell(valueOrDash(plugin.Version)))
	}
	output.WriteString("\n")
}

func writeAggregates(output *strings.Builder, aggregates []pkgBenchmark.Aggregate) {
	if len(aggregates) == 0 {
		return
	}
	output.WriteString("### Aggregates\n\n| Metric | Scope | Count | Min | Median | P95 | Max | Unit |\n|---|---|---:|---:|---:|---:|---:|---|\n")
	for _, aggregate := range aggregates {
		p95 := "—"
		if aggregate.P95 != nil {
			p95 = formatNumber(*aggregate.P95)
		}
		fmt.Fprintf(output, "| %s | %s | %d | %s | %s | %s | %s | %s |\n",
			markdownCell(aggregate.Name), markdownCell(valueOrDash(aggregate.Scope)), aggregate.Count,
			formatNumber(aggregate.Min), formatNumber(aggregate.Median), p95,
			formatNumber(aggregate.Max), markdownCell(aggregate.Unit))
	}
	output.WriteString("\n")
}

func writeSamples(output *strings.Builder, samples []pkgBenchmark.Sample) {
	output.WriteString("### Samples\n\n| Iteration | Type | State | Duration | Result | Quality |\n|---:|---|---|---:|---|---|\n")
	for _, sample := range samples {
		iterationType := "measured"
		if sample.Iteration.Warmup {
			iterationType = "warmup"
		}
		result := sampleResult(sample)
		quality := "—"
		if sample.Evaluation != nil {
			quality = fmt.Sprintf("%d/%d (%s)", sample.Evaluation.Score, sample.Evaluation.MaxScore, sample.Evaluation.RationaleCode)
		}
		fmt.Fprintf(output, "| %d | %s | %s | %s ms | %s | %s |\n",
			sample.Iteration.Index, iterationType, markdownCell(string(sample.State)),
			formatNumber(sample.DurationMS), markdownCell(result), markdownCell(quality))
	}
	output.WriteString("\n")
}

func sampleResult(sample pkgBenchmark.Sample) string {
	if sample.CleanupError != nil {
		return "cleanup:" + sample.CleanupError.Kind + ":" + sample.CleanupError.Code
	}
	if sample.Error != nil {
		return sample.Error.Kind + ":" + sample.Error.Code
	}
	if sample.ReasonCode != "" {
		return sample.ReasonCode
	}
	return "ok"
}

func measuredSampleCount(samples []pkgBenchmark.Sample) int {
	count := 0
	for _, sample := range samples {
		if !sample.Iteration.Warmup {
			count++
		}
	}
	return count
}

func qualitySummary(samples []pkgBenchmark.Sample) string {
	minimum, maximum, count := 4, -1, 0
	for _, sample := range samples {
		if sample.Iteration.Warmup || sample.Evaluation == nil {
			continue
		}
		score := sample.Evaluation.Score
		if score < minimum {
			minimum = score
		}
		if score > maximum {
			maximum = score
		}
		count++
	}
	if count == 0 {
		return "—"
	}
	if minimum == maximum {
		return fmt.Sprintf("%d/3", minimum)
	}
	return fmt.Sprintf("%d–%d/3", minimum, maximum)
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func optionalInteger(value int64) string {
	if value == 0 {
		return "—"
	}
	return strconv.FormatInt(value, 10)
}

func optionalSize(value int64) string {
	if value == 0 {
		return "—"
	}
	return strconv.FormatInt(value, 10) + " MiB"
}

func joinedIdentity(id, version string) string {
	if id == "" {
		return "—"
	}
	if version == "" {
		return id
	}
	return id + "@" + version
}

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func markdownCell(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.ReplaceAll(value, "\\", "\\\\")
	return strings.ReplaceAll(value, "|", "\\|")
}

func markdownInline(value string) string {
	return strings.ReplaceAll(value, "`", "'")
}
