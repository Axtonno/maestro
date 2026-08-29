package directchat

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFileLoaderRejectsPhysicalPathTypes(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "real"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "real", "file.php"), []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "internal")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, logical := range []string{"internal/file.php", "escape/file.php", "pipe"} {
		t.Run(logical, func(t *testing.T) {
			if _, err := loadFile(t.Context(), root, logical, 1024); !errors.Is(err, ErrFileNotAllowed) {
				t.Fatalf("%q was accepted: %v", logical, err)
			}
		})
	}

	linkRoot := filepath.Join(t.TempDir(), "workspace")
	if err := os.Symlink(root, linkRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFile(t.Context(), linkRoot, "real/file.php", 1024); !errors.Is(err, ErrFileNotAllowed) {
		t.Fatalf("symlink workspace root was accepted: %v", err)
	}
}

func TestFileLoaderDetectsMutationAndReplacement(t *testing.T) {
	for _, testCase := range []struct {
		name string
		hook func(string, string) func()
	}{
		{name: "content mutation", hook: func(root, logical string) func() {
			return func() {
				if err := os.WriteFile(filepath.Join(root, logical), []byte("changed!"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{name: "file replacement", hook: func(root, logical string) func() {
			return func() {
				original := filepath.Join(root, logical)
				if err := os.Rename(original, original+".old"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(original, []byte("original"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{name: "mode mutation", hook: func(root, logical string) func() {
			return func() {
				if err := os.Chmod(filepath.Join(root, logical), 0o400); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{name: "symlink replacement", hook: func(root, logical string) func() {
			return func() {
				original := filepath.Join(root, logical)
				if err := os.Rename(original, original+".old"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(original+".old", original); err != nil {
					t.Fatal(err)
				}
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			logical := "file.php"
			if err := os.WriteFile(filepath.Join(root, logical), []byte("original"), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := loadFileWithHooks(t.Context(), root, logical, 1024, fileLoadHooks{afterFirstRead: testCase.hook(root, logical)})
			if !errors.Is(err, ErrFileNotAllowed) {
				t.Fatalf("unstable file was accepted: %v", err)
			}
		})
	}
}

func TestFileLoaderLogicalPathAndContentPolicy(t *testing.T) {
	root := t.TempDir()
	for _, logical := range []string{
		"", ".", "..", "./file.php", "dir//file.php", "dir/../file.php",
		"../file.php", "/file.php", `dir\file.php`, "line\nbreak.php",
		"tab\tfile.php", "bidi\u202efile.php", "nul\x00file.php",
	} {
		if _, err := loadFile(t.Context(), root, logical, 16); !errors.Is(err, ErrFileNotAllowed) {
			t.Errorf("unsafe logical path %q accepted: %v", logical, err)
		}
	}

	cases := []struct {
		name    string
		content []byte
	}{
		{name: "empty", content: nil},
		{name: "utf8 bom", content: []byte{0xef, 0xbb, 0xbf, 'x'}},
		{name: "exact maximum", content: bytes.Repeat([]byte{'x'}, 16)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			logical := testCase.name + ".php"
			if err := os.WriteFile(filepath.Join(root, logical), testCase.content, 0o600); err != nil {
				t.Fatal(err)
			}
			content, err := loadFile(t.Context(), root, logical, 16)
			if err != nil || !bytes.Equal([]byte(content), testCase.content) {
				t.Fatalf("content=%x err=%v", []byte(content), err)
			}
		})
	}
	if err := os.WriteFile(filepath.Join(root, "over.php"), bytes.Repeat([]byte{'x'}, 17), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFile(t.Context(), root, "over.php", 16); !errors.Is(err, ErrFileNotAllowed) {
		t.Fatalf("over-limit file accepted: %v", err)
	}
}

func TestFileLoaderDetectsWorkspaceRootReplacement(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "file.php"), []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	hook := func() {
		if err := os.Rename(root, root+".old"); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "file.php"), []byte("replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := loadFileWithHooks(t.Context(), root, "file.php", 1024, fileLoadHooks{afterFirstRead: hook}); !errors.Is(err, ErrFileNotAllowed) {
		t.Fatalf("replaced workspace root was accepted: %v", err)
	}
}

func TestFileLoaderDetectsParentDirectoryReplacement(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "routes")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "api.php"), []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	hook := func() {
		if err := os.Rename(directory, directory+".old"); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "api.php"), []byte("replacement"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := loadFileWithHooks(t.Context(), root, "routes/api.php", 1024, fileLoadHooks{afterFirstRead: hook}); !errors.Is(err, ErrFileNotAllowed) {
		t.Fatalf("replaced parent directory was accepted: %v", err)
	}
}

func TestFileLoaderPreservesCancellationReason(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.php"), []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loadFile(ctx, root, "file.php", 1024); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was collapsed: %v", err)
	}
}
