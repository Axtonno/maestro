package maestro_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPublicDocumentationDescribesOperationalScope(t *testing.T) {
	assertContains := func(path string, required ...string) {
		t.Helper()
		content := readPublicDoc(t, path)
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
		"docs/current-capabilities.md",
		"docs/controlled-mutation.md",
		"docs/benchmarks.md",
		"CONTRIBUTING.md",
	)
	assertContains("docs/current-capabilities.md",
		"## Cosa funziona oggi",
		"## Sperimentale, non parte del prodotto v0.5.0",
		"## Non supportato oggi",
		"## Hardware consigliato",
		"qwen3.5:9b",
		"qwen2.5-coder:14b",
	)
	assertContains("docs/controlled-mutation.md",
		"Un file PHP regolare e non symlink sotto `app/`",
		"allow-once",
		"`stale_source`",
		"Fuori perimetro oggi",
	)
	assertContains("docs/supported-platforms.md",
		"Linux `amd64`",
		"Windows nativo",
		"macOS",
	)
	assertContains("docs/benchmarks.md",
		"risultati sintetici e riproducibili",
		"14/14 controlli superati",
		"11/12 risposte",
		"NOT_RUN",
	)
	assertContains("docs/roadmap.md",
		"profilo single-model",
		"Windows e macOS",
		"Non promesso",
	)
}

func TestPublicDocumentationAllowlist(t *testing.T) {
	allowedFiles := map[string]bool{
		"architecture.md":                        true,
		"benchmarks.md":                          true,
		"cli.md":                                 true,
		"configuration.md":                       true,
		"controlled-mutation.md":                 true,
		"current-capabilities.md":                true,
		"developer-benchmark-manifest.yaml":      true,
		"installation.md":                        true,
		"known-issues.md":                        true,
		"provider-smoke-benchmark-manifest.yaml": true,
		"quick-start.md":                         true,
		"roadmap.md":                             true,
		"runtime-benchmark-manifest.yaml":        true,
		"security-model.md":                      true,
		"supported-platforms.md":                 true,
		"troubleshooting.md":                     true,
	}
	allowedDirectories := map[string]bool{"releases": true, "schemas": true}

	entries, err := os.ReadDir("docs")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !allowedDirectories[entry.Name()] {
				t.Errorf("docs/%s is outside the public allowlist", entry.Name())
			}
			continue
		}
		if !allowedFiles[entry.Name()] {
			t.Errorf("docs/%s is outside the public allowlist", entry.Name())
		}
	}

	for _, forbidden := range []string{
		"MAESTRO_CONTEXT.md",
		"docs/reports",
		"docs/adr",
		"docs/prompts",
		"scripts",
	} {
		if _, err := os.Stat(forbidden); !os.IsNotExist(err) {
			t.Errorf("internal-only path is present in public tree: %s", forbidden)
		}
	}
}

func TestPublicDocumentationRelativeLinksResolve(t *testing.T) {
	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	paths := []string{"README.md", "CONTRIBUTING.md", "CHANGELOG.md", "SECURITY.md"}
	err := filepath.WalkDir("docs", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".md") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		content := readPublicDoc(t, path)
		for _, match := range linkPattern.FindAllStringSubmatch(content, -1) {
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

func TestPublicMarkdownDoesNotLinkInternalArtifacts(t *testing.T) {
	forbidden := []string{
		"docs/reports/",
		"../reports/",
		"MAESTRO_CONTEXT",
		"controlled-mutation-support.md",
		"compatibility.md",
		"validation.md",
	}
	err := filepath.WalkDir("docs", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		content := readPublicDoc(t, path)
		for _, value := range forbidden {
			if strings.Contains(content, value) {
				t.Errorf("%s references internal or retired path %q", path, value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func readPublicDoc(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
