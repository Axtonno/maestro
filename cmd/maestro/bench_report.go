package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	internalBenchmark "github.com/antonio-cafeo/maestro/internal/benchmark"
	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

func runBenchRender(arguments []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("maestro bench render", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "input benchmark JSON report")
	outputPath := flags.String("output", "-", "Markdown report path, or - for stdout")
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || strings.TrimSpace(*inputPath) == "" {
		fmt.Fprintln(stderr, "maestro bench render requires --input and no positional arguments")
		return 2
	}
	if *outputPath != "-" && samePath(*inputPath, *outputPath) {
		fmt.Fprintln(stderr, "maestro bench render input and output must be different files")
		return 2
	}
	input, err := os.Open(*inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "open benchmark JSON report: %v\n", err)
		return 1
	}
	report, decodeError := internalBenchmark.DecodeReportJSON(input)
	closeError := input.Close()
	if decodeError != nil {
		fmt.Fprintf(stderr, "decode benchmark JSON report: %v\n", decodeError)
		return 1
	}
	if closeError != nil {
		fmt.Fprintf(stderr, "close benchmark JSON report: %v\n", closeError)
		return 1
	}
	if *outputPath == "-" {
		err = internalBenchmark.EncodeReportMarkdown(stdout, report)
	} else {
		err = internalBenchmark.WriteReportMarkdown(*outputPath, report)
	}
	if err != nil {
		fmt.Fprintf(stderr, "write benchmark Markdown report: %v\n", err)
		return 1
	}
	return 0
}

func writeBenchmarkReports(
	jsonPath string,
	markdownPath string,
	stdout io.Writer,
	report pkgBenchmark.Report,
) error {
	if err := validateBenchmarkReportPaths(jsonPath, markdownPath); err != nil {
		return err
	}
	var err error
	if jsonPath == "-" {
		err = internalBenchmark.EncodeReportJSON(stdout, report)
	} else {
		err = internalBenchmark.WriteReportJSON(jsonPath, report)
	}
	if err != nil {
		return err
	}
	if markdownPath != "" {
		if err := internalBenchmark.WriteReportMarkdown(markdownPath, report); err != nil {
			return err
		}
	}
	return nil
}

func validateBenchmarkReportPaths(jsonPath, markdownPath string) error {
	if markdownPath == "-" {
		return fmt.Errorf("Markdown output - conflicts with the JSON stdout contract")
	}
	if markdownPath != "" && jsonPath != "-" && samePath(jsonPath, markdownPath) {
		return fmt.Errorf("JSON and Markdown report paths must differ")
	}
	return nil
}

func samePath(left, right string) bool {
	leftPath, leftError := filepath.Abs(left)
	rightPath, rightError := filepath.Abs(right)
	if leftError != nil || rightError != nil {
		return filepath.Clean(left) == filepath.Clean(right)
	}
	return filepath.Clean(leftPath) == filepath.Clean(rightPath)
}
