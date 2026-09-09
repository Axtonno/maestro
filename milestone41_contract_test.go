package maestro_test

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone41SeparatesOnboardingFromValidation(t *testing.T) {
	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	for _, path := range []string{"README.md", "docs/quick-start.md", "docs/installation.md", "docs/install-and-try.md"} {
		content := strings.ToLower(read(path))
		for _, forbidden := range []string{"fixture", "git -c fixtures", "git status", "milestone gate"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s exposes validation detail %q in the user path", path, forbidden)
			}
		}
	}
	quickStart := read("docs/quick-start.md")
	for _, required := range []string{"maestro setup", "maestro chat", "maestro mutate --preview", "Troubleshooting", "Validation Guide"} {
		if !strings.Contains(quickStart, required) {
			t.Fatalf("quick start is missing %q", required)
		}
	}
	validation := read("docs/validation.md")
	for _, required := range []string{"Baseline Git della fixture", "doctor --mode all", "approval_rejected", "stale_source", "git -C fixtures/laravel-v1 diff"} {
		if !strings.Contains(validation, required) {
			t.Fatalf("validation guide is missing %q", required)
		}
	}

	var matrix struct {
		Status     string
		Verdict    string
		Boundaries struct {
			DoctorRequired  bool `yaml:"doctor_required_for_first_use"`
			GitRequired     bool `yaml:"git_baseline_required_for_first_use"`
			PreviewCanWrite bool `yaml:"preview_can_write"`
		} `yaml:"boundaries"`
		Publication struct {
			V050Unchanged  bool `yaml:"v0_5_0_asset_unchanged"`
			NextAuthorized bool `yaml:"next_release_authorized"`
		} `yaml:"publication"`
	}
	if err := yaml.Unmarshal([]byte(read("docs/milestone-41-installation-onboarding-matrix.yaml")), &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Status != "validation_pending" || matrix.Verdict != "implementation_complete_validation_pending" || matrix.Boundaries.DoctorRequired || matrix.Boundaries.GitRequired || matrix.Boundaries.PreviewCanWrite || !matrix.Publication.V050Unchanged || matrix.Publication.NextAuthorized {
		t.Fatalf("invalid M41 boundary: %#v", matrix)
	}
}
