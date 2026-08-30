package maestro

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone21QualificationMatrixIsFrozenAndConsistent(t *testing.T) {
	type task struct {
		ID            string   `yaml:"id"`
		Source        string   `yaml:"source"`
		CaptureSHA256 string   `yaml:"capture_sha256"`
		File          string   `yaml:"file"`
		Question      string   `yaml:"question"`
		Required      []string `yaml:"required"`
		Forbidden     []string `yaml:"forbidden"`
	}
	var matrix struct {
		Version int    `yaml:"version"`
		Status  string `yaml:"status"`
		Profile struct {
			Model      string `yaml:"model"`
			NumPredict int    `yaml:"num_predict"`
		} `yaml:"profile"`
		Fixture struct {
			Source string            `yaml:"source"`
			Digest string            `yaml:"digest"`
			Files  map[string]string `yaml:"files"`
		} `yaml:"fixture"`
		SeriesOrder struct {
			Series1                   []string `yaml:"series_1"`
			Series2                   []string `yaml:"series_2"`
			PairedStreamTask          string   `yaml:"paired_stream_task"`
			PairedStreamCaptureSHA256 string   `yaml:"paired_stream_capture_sha256"`
		} `yaml:"series_order"`
		Tasks []task `yaml:"tasks"`
	}

	encoded, err := os.ReadFile("docs/milestone-21-cpu-direct-chat-qualification-matrix.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(encoded, &matrix); err != nil {
		t.Fatalf("decode qualification matrix: %v", err)
	}
	if matrix.Version != 1 || matrix.Status != "task_oracle_freeze" ||
		matrix.Profile.Model != "qwen2.5-coder:7b" || matrix.Profile.NumPredict != 512 {
		t.Fatalf("unexpected frozen header: %#v", matrix)
	}
	if len(matrix.Tasks) != 10 {
		t.Fatalf("task count=%d, want 10", len(matrix.Tasks))
	}
	known := make(map[string]struct{}, len(matrix.Tasks))
	for _, candidate := range matrix.Tasks {
		if candidate.ID == "" || candidate.File == "" || strings.TrimSpace(candidate.Question) == "" ||
			len(candidate.Required) == 0 || len(candidate.Forbidden) == 0 {
			t.Fatalf("incomplete task: %#v", candidate)
		}
		if _, exists := known[candidate.ID]; exists {
			t.Fatalf("duplicate task %q", candidate.ID)
		}
		known[candidate.ID] = struct{}{}
		if strings.HasPrefix(candidate.ID, "Q20-") {
			if candidate.Source != "exact_m20_capture" || !validFrozenSHA256(candidate.CaptureSHA256) {
				t.Fatalf("invalid M20 lineage: %#v", candidate)
			}
		} else if candidate.Source != "conservative_reconstruction" {
			t.Fatalf("invalid M17 lineage: %#v", candidate)
		}
	}
	for name, order := range map[string][]string{"series_1": matrix.SeriesOrder.Series1, "series_2": matrix.SeriesOrder.Series2} {
		if len(order) != len(known) {
			t.Fatalf("%s length=%d, want %d", name, len(order), len(known))
		}
		seen := make(map[string]struct{}, len(order))
		for _, id := range order {
			if _, exists := known[id]; !exists {
				t.Fatalf("%s contains unknown task %q", name, id)
			}
			if _, exists := seen[id]; exists {
				t.Fatalf("%s repeats task %q", name, id)
			}
			seen[id] = struct{}{}
		}
	}
	if matrix.SeriesOrder.PairedStreamTask != "Q20-1" ||
		!validFrozenSHA256(matrix.SeriesOrder.PairedStreamCaptureSHA256) {
		t.Fatalf("invalid paired stream freeze: %#v", matrix.SeriesOrder)
	}

	for logical, expected := range matrix.Fixture.Files {
		content, err := os.ReadFile(filepath.Join(matrix.Fixture.Source, filepath.FromSlash(logical)))
		if err != nil {
			t.Fatalf("read fixture %s: %v", logical, err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(content))
		if actual != expected {
			t.Fatalf("fixture %s digest=%s, want %s", logical, actual, expected)
		}
	}
	if actual := frozenFixtureManifestDigest(t, matrix.Fixture.Source); actual != matrix.Fixture.Digest {
		t.Fatalf("fixture manifest digest=%s, want %s", actual, matrix.Fixture.Digest)
	}
}

func frozenFixtureManifestDigest(t *testing.T, root string) string {
	t.Helper()
	var paths []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type().IsRegular() {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	manifest := sha256.New()
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(manifest, "%x  ./%s\n", sha256.Sum256(content), filepath.ToSlash(relative))
	}
	return fmt.Sprintf("%x", manifest.Sum(nil))
}

func validFrozenSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}
