package benchmark

import (
	"errors"
	"testing"
)

func TestCollectLinuxHardwareProfileUsesProcfsAndExplicitGPUEnvironment(t *testing.T) {
	files := map[string]string{
		"/proc/cpuinfo": "processor: 0\nmodel name : Maestro Test CPU\n",
		"/proc/meminfo": "MemTotal:       16777216 kB\n",
	}
	environment := map[string]string{
		EnvironmentGPU: "  Test   GPU  ", EnvironmentBackend: "CUDA",
		EnvironmentVRAMMB: "8192",
	}
	profile, err := collectHardwareProfile(
		"linux", "amd64", 16,
		func(name string) string { return environment[name] },
		func(name string) ([]byte, error) {
			value, exists := files[name]
			if !exists {
				return nil, errors.New("not found")
			}
			return []byte(value), nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if profile.OS != "linux" || profile.Architecture != "amd64" ||
		profile.LogicalCPUs != 16 || profile.CPU != "Maestro Test CPU" ||
		profile.MemoryMB != 16384 || profile.GPU != "Test GPU" ||
		profile.Backend != "CUDA" || profile.VRAMMB != 8192 {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestHardwareProfileOmitsUnavailableFieldsAndRejectsInvalidVRAM(t *testing.T) {
	profile, err := collectHardwareProfile(
		"darwin", "arm64", 8,
		func(string) string { return "" },
		func(string) ([]byte, error) { return nil, errors.New("unused") },
	)
	if err != nil || profile.CPU != "" || profile.MemoryMB != 0 || profile.VRAMMB != 0 {
		t.Fatalf("unexpected unavailable profile: %#v err=%v", profile, err)
	}
	_, err = collectHardwareProfile(
		"linux", "amd64", 8,
		func(name string) string {
			if name == EnvironmentVRAMMB {
				return "secret-invalid"
			}
			return ""
		},
		func(string) ([]byte, error) { return nil, errors.New("missing") },
	)
	if err == nil {
		t.Fatal("expected invalid VRAM rejection")
	}
}
