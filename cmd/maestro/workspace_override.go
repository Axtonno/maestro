package main

import (
	"fmt"
	"path/filepath"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
)

type configValidator func(productconfig.Config) error

func applyCurrentWorkspace(config productconfig.Config, enabled bool, dependencies commandDependencies, validate configValidator) (productconfig.Config, error) {
	if !enabled {
		return config, nil
	}
	if dependencies.workingDirectory == nil {
		return productconfig.Config{}, fmt.Errorf("resolve current workspace: %w", productconfig.ErrInvalid)
	}
	root, err := dependencies.workingDirectory()
	if err != nil {
		return productconfig.Config{}, fmt.Errorf("resolve current workspace: %w: %w", err, productconfig.ErrInvalid)
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return productconfig.Config{}, fmt.Errorf("normalize current workspace: %w: %w", err, productconfig.ErrInvalid)
	}
	config.Workspace.Root = filepath.Clean(absolute)
	if validate != nil {
		if err := validate(config); err != nil {
			return productconfig.Config{}, err
		}
	}
	return config, nil
}
