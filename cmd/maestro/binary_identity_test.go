package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentBinaryIdentityHashesRunningExecutable(t *testing.T) {
	identity, err := currentBinaryIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(identity.Executable) {
		t.Fatalf("executable is not absolute: %q", identity.Executable)
	}
	file, err := os.Open(identity.Executable)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		t.Fatal(err)
	}
	if expected := fmt.Sprintf("%x", digest.Sum(nil)); identity.SHA256 != expected {
		t.Fatalf("sha256=%s expected=%s", identity.SHA256, expected)
	}
}
