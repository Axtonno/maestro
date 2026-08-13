package productconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const maxConfigBytes = 1 << 20

func Load(path string) (Config, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read configuration: %w: %w", err, ErrNotFound)
		}
		return Config{}, fmt.Errorf("read configuration: %w: %w", err, ErrInvalid)
	}
	if len(encoded) == 0 || len(encoded) > maxConfigBytes {
		return Config{}, fmt.Errorf("configuration must contain 1..%d bytes: %w", maxConfigBytes, ErrInvalid)
	}
	if err := rejectAliases(encoded); err != nil {
		return Config{}, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	decoder.KnownFields(true)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %v: %w", err, ErrInvalid)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("configuration contains multiple YAML documents: %w", ErrInvalid)
		}
		return Config{}, fmt.Errorf("decode trailing configuration: %v: %w", err, ErrInvalid)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("resolve configuration path: %w: %w", err, ErrInvalid)
	}
	config.path = filepath.Clean(absolute)
	if !filepath.IsAbs(config.Workspace.Root) && config.Workspace.Root != "" {
		config.Workspace.Root = filepath.Clean(filepath.Join(filepath.Dir(config.path), config.Workspace.Root))
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func rejectAliases(encoded []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	for {
		var document yaml.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse configuration: %v: %w", err, ErrInvalid)
		}
		if containsAlias(&document) {
			return fmt.Errorf("configuration aliases are not supported: %w", ErrInvalid)
		}
	}
}

func containsAlias(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return true
	}
	for _, child := range node.Content {
		if containsAlias(child) {
			return true
		}
	}
	return false
}
