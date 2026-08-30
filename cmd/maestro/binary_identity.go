package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type binaryIdentity struct {
	Executable string
	SHA256     string
}

func currentBinaryIdentity() (binaryIdentity, error) {
	executable, err := os.Executable()
	if err != nil {
		return binaryIdentity{}, fmt.Errorf("resolve executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return binaryIdentity{}, fmt.Errorf("resolve executable path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return binaryIdentity{}, fmt.Errorf("resolve executable symlinks: %w", err)
	}
	executable = filepath.Clean(resolved)
	file, err := os.Open(executable)
	if err != nil {
		return binaryIdentity{}, fmt.Errorf("open executable: %w", err)
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return binaryIdentity{}, fmt.Errorf("hash executable: %w", err)
	}
	return binaryIdentity{Executable: executable, SHA256: fmt.Sprintf("%x", digest.Sum(nil))}, nil
}
