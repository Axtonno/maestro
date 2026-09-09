package maestro_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPublicDocumentationDescribesV050OperationalScope(t *testing.T) {
	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	assertContains := func(path string, required ...string) {
		t.Helper()
		content := read(path)
		for _, value := range required {
			if !strings.Contains(content, value) {
				t.Fatalf("%s is missing %q", path, value)
			}
		}
	}

	assertContains("README.md",
		"workstation AI locale",
		"maestro setup",
		"maestro mutate --preview",
		"docs/quick-start.md",
		"docs/validation.md",
		"docs/current-capabilities.md",
	)
	assertContains("docs/current-capabilities.md",
		"## Cosa funziona oggi",
		"## Sperimentale, non parte del prodotto v0.5.0",
		"## Non supportato oggi",
		"## Hardware consigliato",
		"Difetto documentale dell'archive v0.5.0",
		"qwen3.5:9b",
		"qwen2.5-coder:14b",
	)
	assertContains("docs/install-and-try.md",
		"Quick Start",
		"maestro setup",
		"maestro mutate --preview",
		"Validation Guide",
	)
	assertContains("docs/validation.md",
		"Baseline Git della fixture",
		"doctor --mode all",
		"approval_rejected",
		"git -C fixtures/laravel-v1 diff",
	)
	assertContains("docs/controlled-mutation-support.md",
		"Un file PHP regolare e non symlink sotto `app/`",
		"allow-once",
		"`stale_source`",
		"Fuori perimetro oggi",
	)
	assertContains("docs/compatibility.md",
		"Maestro v0.5.0 Compatibility Matrix",
		"Controlled Mutation",
		"Non supportato oggi",
	)
	assertContains("docs/identity.md",
		"workstation AI locale",
		"nucleo operativo attuale",
	)
	assertContains("docs/vision.md", "## Oggi", "## Direzione")

	currentDocs := []string{
		"README.md",
		"docs/cli.md",
		"docs/compatibility.md",
		"docs/configuration.md",
		"docs/known-issues.md",
		"docs/quick-start.md",
		"docs/security-model.md",
		"docs/troubleshooting.md",
		"docs/validation.md",
	}
	for _, path := range currentDocs {
		content := read(path)
		for _, stale := range []string{
			"Maestro v0.3.1",
			"Maestro v0.3.0",
			"La linea candidata v0.5.0",
			"mutazioni non sono supportate",
		} {
			if strings.Contains(content, stale) {
				t.Fatalf("%s contains stale public positioning %q", path, stale)
			}
		}
	}

	for _, script := range []string{
		"scripts/package-candidate.sh",
		"scripts/verify-package-candidate.sh",
	} {
		assertContains(script,
			"docs/install-and-try.md",
			"docs/validation.md",
			"docs/controlled-mutation-support.md",
			"docs/current-capabilities.md",
			"docs/v0.5.0-public-baseline-freeze.yaml",
			"docs/milestone-38-field-adoption-freeze.yaml",
			"docs/reports/milestone-38-live-runs.json",
		)
	}
}

func TestPublicDocumentationRelativeLinksResolve(t *testing.T) {
	paths := []string{
		"README.md",
		"docs/cli.md",
		"docs/configuration.md",
		"docs/install-and-try.md",
		"docs/controlled-mutation-support.md",
		"docs/current-capabilities.md",
		"docs/compatibility.md",
		"docs/installation.md",
		"docs/known-issues.md",
		"docs/packaging-candidate.md",
		"docs/quick-start.md",
		"docs/security-model.md",
		"docs/troubleshooting.md",
		"docs/identity.md",
		"docs/vision.md",
		"docs/philosophy.md",
		"docs/releases/v0.5.0.md",
		"docs/milestone-39-documentation-onboarding-public-trial-readiness-plan.md",
		"docs/reports/milestone-39-final.md",
		"docs/milestone-41-installation-onboarding-simplification-plan.md",
	}
	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(data), -1) {
			target := strings.SplitN(match[1], "#", 2)[0]
			if target == "" || strings.Contains(target, "://") {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if _, err := os.Stat(resolved); err != nil {
				t.Errorf("%s has unresolved link %q: %v", path, match[1], err)
			}
		}
	}
}
