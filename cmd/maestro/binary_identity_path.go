//go:build !windows

package main

import "path/filepath"

func resolveExecutableForIdentity(executable string) (string, error) {
	return filepath.EvalSymlinks(executable)
}
