package developer

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DatasetID      = "maestro-laravel-mini"
	DatasetVersion = "1.0.0"
	datasetRoot    = "testdata/laravel-v1"
)

//go:embed testdata/laravel-v1
var datasetFiles embed.FS

type Criterion struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	AnyTerms    []string `json:"any_terms"`
	AllTerms    []string `json:"all_terms"`
}

type Task struct {
	ID            string      `json:"id"`
	Kind          string      `json:"kind"`
	ModelRole     string      `json:"model_role"`
	Instruction   string      `json:"instruction"`
	Files         []string    `json:"files"`
	RelevantFiles []string    `json:"relevant_files,omitempty"`
	Criteria      []Criterion `json:"criteria,omitempty"`
}

type Dataset struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Framework string `json:"framework"`
	Tasks     []Task `json:"tasks"`

	taskIndex map[string]Task
}

func LoadDataset() (Dataset, error) {
	encoded, err := datasetFiles.ReadFile(datasetRoot + "/dataset.json")
	if err != nil {
		return Dataset{}, fmt.Errorf("read embedded developer dataset: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	var dataset Dataset
	if err := decoder.Decode(&dataset); err != nil {
		return Dataset{}, fmt.Errorf("decode embedded developer dataset: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Dataset{}, err
	}
	if err := dataset.validate(); err != nil {
		return Dataset{}, err
	}
	dataset.taskIndex = make(map[string]Task, len(dataset.Tasks))
	for _, task := range dataset.Tasks {
		dataset.taskIndex[task.ID] = task
	}
	return dataset, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return errors.New("developer dataset contains multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("read developer dataset trailing content: %w", err)
	}
	return nil
}

func (d Dataset) validate() error {
	if d.ID != DatasetID || d.Version != DatasetVersion || d.Framework != "laravel" {
		return errors.New("developer dataset identity is invalid")
	}
	if len(d.Tasks) != 6 {
		return fmt.Errorf("developer dataset must contain six tasks, got %d", len(d.Tasks))
	}
	seen := make(map[string]struct{}, len(d.Tasks))
	for _, task := range d.Tasks {
		if strings.TrimSpace(task.ID) == "" || strings.TrimSpace(task.Instruction) == "" ||
			strings.TrimSpace(task.ModelRole) == "" || len(task.Files) == 0 {
			return fmt.Errorf("developer dataset task %q is incomplete", task.ID)
		}
		if _, exists := seen[task.ID]; exists {
			return fmt.Errorf("developer dataset task %q is duplicated", task.ID)
		}
		seen[task.ID] = struct{}{}
		for _, name := range append(append([]string(nil), task.Files...), task.RelevantFiles...) {
			if !safeDatasetPath(name) {
				return fmt.Errorf("developer dataset task %q has unsafe file %q", task.ID, name)
			}
			if _, err := datasetFiles.ReadFile(path.Join(datasetRoot, name)); err != nil {
				return fmt.Errorf("developer dataset task %q references %q: %w", task.ID, name, err)
			}
		}
		switch task.Kind {
		case "generation":
			if task.ModelRole != "chat" || len(task.Criteria) != 3 || len(task.RelevantFiles) != 0 {
				return fmt.Errorf("developer generation task %q has an invalid rubric", task.ID)
			}
			for _, criterion := range task.Criteria {
				if strings.TrimSpace(criterion.ID) == "" || strings.TrimSpace(criterion.Description) == "" ||
					(len(criterion.AnyTerms) == 0) == (len(criterion.AllTerms) == 0) {
					return fmt.Errorf("developer task %q has an incomplete criterion", task.ID)
				}
			}
		case "retrieval":
			if task.ModelRole != "embedding" || len(task.RelevantFiles) == 0 || len(task.Criteria) != 0 {
				return fmt.Errorf("developer retrieval task %q is invalid", task.ID)
			}
		default:
			return fmt.Errorf("developer dataset task %q has unknown kind %q", task.ID, task.Kind)
		}
	}
	return nil
}

func safeDatasetPath(name string) bool {
	return name != "" && name == path.Clean(name) && !path.IsAbs(name) && name != ".." && !strings.HasPrefix(name, "../")
}

func (d Dataset) Task(id string) (Task, bool) {
	task, exists := d.taskIndex[id]
	return task, exists
}

func (d Dataset) File(name string) ([]byte, error) {
	if !safeDatasetPath(name) {
		return nil, fmt.Errorf("unsafe developer dataset file %q", name)
	}
	content, err := datasetFiles.ReadFile(path.Join(datasetRoot, name))
	if err != nil {
		return nil, fmt.Errorf("read developer dataset file %q: %w", name, err)
	}
	return append([]byte(nil), content...), nil
}

func (d Dataset) Prompt(task Task) (string, error) {
	var prompt strings.Builder
	prompt.WriteString(task.Instruction)
	prompt.WriteString("\n\nUse only the following versioned fixture files:\n")
	for _, name := range task.Files {
		content, err := d.File(name)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&prompt, "\n--- %s ---\n%s\n", name, content)
	}
	return prompt.String(), nil
}

func (d Dataset) Materialize() (string, func() error, error) {
	root, err := os.MkdirTemp("", "maestro-laravel-benchmark-")
	if err != nil {
		return "", nil, fmt.Errorf("create developer benchmark workspace: %w", err)
	}
	cleanup := func() error { return os.RemoveAll(root) }
	err = fs.WalkDir(datasetFiles, datasetRoot, func(name string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if name == datasetRoot {
			return nil
		}
		relative := strings.TrimPrefix(name, datasetRoot+"/")
		target := filepath.Join(root, filepath.FromSlash(relative))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		content, err := datasetFiles.ReadFile(name)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o600)
	})
	if err != nil {
		return "", cleanup, fmt.Errorf("materialize developer dataset: %w", err)
	}
	return root, cleanup, nil
}
