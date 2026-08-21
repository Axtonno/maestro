package mutation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/antonio-cafeo/maestro/internal/benchmark/developer"
)

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type WorkspaceSnapshot struct {
	Digest string       `json:"digest"`
	Files  []FileDigest `json:"files"`
}

type MaterializedFixture struct {
	Root    string
	Initial WorkspaceSnapshot
	cleanup func() error
}

func MaterializeFixture(ctx context.Context, profile Profile) (*MaterializedFixture, error) {
	if ctx == nil {
		return nil, errors.New("materialize mutation fixture: nil context")
	}
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	dataset, err := developer.LoadDataset()
	if err != nil {
		return nil, err
	}
	content, err := dataset.File(profile.Fixture.Target)
	if err != nil {
		return nil, err
	}
	if digestBytes(content) != profile.Fixture.InitialSHA256 ||
		strings.Count(string(content), profile.Fixture.Old) != 1 {
		return nil, errors.New("embedded mutation fixture differs from the frozen profile")
	}
	proposed := strings.Replace(string(content), profile.Fixture.Old, profile.Fixture.Replacement, 1)
	if digestBytes([]byte(proposed)) != profile.Fixture.ExpectedSHA256 {
		return nil, errors.New("mutation fixture expected digest differs from the frozen replacement")
	}
	root, cleanup, err := dataset.Materialize()
	if err != nil {
		return nil, err
	}
	snapshot, err := SnapshotWorkspace(ctx, root)
	if err != nil {
		_ = cleanup()
		return nil, err
	}
	return &MaterializedFixture{Root: root, Initial: snapshot, cleanup: cleanup}, nil
}

func (fixture *MaterializedFixture) Cleanup() error {
	if fixture == nil || fixture.cleanup == nil {
		return nil
	}
	err := fixture.cleanup()
	fixture.cleanup = nil
	return err
}

func SnapshotWorkspace(ctx context.Context, root string) (WorkspaceSnapshot, error) {
	if ctx == nil || !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return WorkspaceSnapshot{}, errors.New("snapshot workspace root is invalid")
	}
	files := make([]FileDigest, 0)
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if name == root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace snapshot encountered symlink: %s", filepath.Base(name))
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("workspace snapshot encountered non-regular entry: %s", filepath.Base(name))
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		files = append(files, FileDigest{Path: filepath.ToSlash(relative), SHA256: digestBytes(content)})
		return nil
	})
	if err != nil {
		return WorkspaceSnapshot{}, err
	}
	slices.SortFunc(files, func(left, right FileDigest) int { return strings.Compare(left.Path, right.Path) })
	hash := sha256.New()
	for _, file := range files {
		_, _ = hash.Write([]byte(file.Path))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(file.SHA256))
		_, _ = hash.Write([]byte{0})
	}
	return WorkspaceSnapshot{Digest: hex.EncodeToString(hash.Sum(nil)), Files: files}, nil
}

func (snapshot WorkspaceSnapshot) File(path string) (string, bool) {
	for _, file := range snapshot.Files {
		if file.Path == path {
			return file.SHA256, true
		}
	}
	return "", false
}

func (snapshot WorkspaceSnapshot) Validate() error {
	if !validSHA256(snapshot.Digest) || len(snapshot.Files) == 0 {
		return errors.New("workspace snapshot identity is invalid")
	}
	previous := ""
	for _, file := range snapshot.Files {
		if file.Path == "" || file.Path <= previous || filepath.IsAbs(file.Path) ||
			filepath.ToSlash(filepath.Clean(file.Path)) != file.Path || !validSHA256(file.SHA256) {
			return errors.New("workspace snapshot file is invalid")
		}
		previous = file.Path
	}
	return nil
}

func digestBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
