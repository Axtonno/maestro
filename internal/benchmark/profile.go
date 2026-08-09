package benchmark

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	pkgBenchmark "github.com/antonio-cafeo/maestro/pkg/benchmark"
)

const (
	EnvironmentGPU     = "MAESTRO_BENCHMARK_GPU"
	EnvironmentBackend = "MAESTRO_BENCHMARK_BACKEND"
	EnvironmentVRAMMB  = "MAESTRO_BENCHMARK_VRAM_MB"
)

func CollectHardwareProfile() (pkgBenchmark.HardwareProfile, error) {
	return collectHardwareProfile(
		runtime.GOOS,
		runtime.GOARCH,
		runtime.NumCPU(),
		os.Getenv,
		os.ReadFile,
	)
}

func collectHardwareProfile(
	goos string,
	architecture string,
	logicalCPUs int,
	getenv func(string) string,
	readFile func(string) ([]byte, error),
) (pkgBenchmark.HardwareProfile, error) {
	if logicalCPUs < 1 || getenv == nil || readFile == nil {
		return pkgBenchmark.HardwareProfile{}, errors.New("hardware profile dependencies are invalid")
	}
	profile := pkgBenchmark.HardwareProfile{
		OS: goos, Architecture: architecture, LogicalCPUs: logicalCPUs,
		GPU:     normalizeProfileText(getenv(EnvironmentGPU)),
		Backend: normalizeProfileText(getenv(EnvironmentBackend)),
	}
	if rawVRAM := strings.TrimSpace(getenv(EnvironmentVRAMMB)); rawVRAM != "" {
		vram, err := strconv.ParseInt(rawVRAM, 10, 64)
		if err != nil || vram <= 0 {
			return pkgBenchmark.HardwareProfile{}, fmt.Errorf(
				"%s must be a positive integer, got %q",
				EnvironmentVRAMMB,
				rawVRAM,
			)
		}
		profile.VRAMMB = vram
	}
	if goos != "linux" {
		return profile, nil
	}
	if encoded, err := readFile("/proc/cpuinfo"); err == nil {
		profile.CPU = parseCPUModel(string(encoded))
	}
	if encoded, err := readFile("/proc/meminfo"); err == nil {
		profile.MemoryMB = parseMemoryMB(string(encoded))
	}
	return profile, nil
}

func parseCPUModel(value string) string {
	for _, line := range strings.Split(value, "\n") {
		name, content, exists := strings.Cut(line, ":")
		if !exists {
			continue
		}
		switch strings.TrimSpace(name) {
		case "model name", "Hardware", "Processor":
			if normalized := normalizeProfileText(content); normalized != "" {
				return normalized
			}
		}
	}
	return ""
}

func parseMemoryMB(value string) int64 {
	for _, line := range strings.Split(value, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "MemTotal:" {
			continue
		}
		kilobytes, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || kilobytes < 0 {
			return 0
		}
		return kilobytes / 1024
	}
	return 0
}

func normalizeProfileText(value string) string {
	normalized := strings.Join(strings.Fields(value), " ")
	runes := []rune(normalized)
	if len(runes) > 128 {
		normalized = string(runes[:128])
	}
	return normalized
}

func BuildMetadata() (version string, commit string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			commit = setting.Value
			break
		}
	}
	return version, commit
}
