package directchat

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

func loadFile(ctx context.Context, workspaceRoot, logical string, maximum int) (string, error) {
	return loadFileWithHooks(ctx, workspaceRoot, logical, maximum, fileLoadHooks{})
}

type fileLoadHooks struct {
	afterFirstRead func()
}

func loadFileWithHooks(ctx context.Context, workspaceRoot, logical string, maximum int, hooks fileLoadHooks) (string, error) {
	if ctx == nil || maximum < 1 || pkgContext.DocumentPath(logical).Validate() != nil ||
		strings.ContainsRune(logical, '\x00') || strings.Contains(logical, `\`) {
		return "", ErrFileNotAllowed
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	rootInfo, err := os.Lstat(workspaceRoot)
	if err != nil || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return "", ErrFileNotAllowed
	}
	root, err := os.OpenRoot(workspaceRoot)
	if err != nil {
		return "", ErrFileNotAllowed
	}
	defer root.Close()

	if err := validatePhysicalPath(root, logical); err != nil {
		return "", ErrFileNotAllowed
	}
	pathBefore, err := root.Lstat(logical)
	if err != nil || pathBefore.Mode()&os.ModeSymlink != 0 || !pathBefore.Mode().IsRegular() {
		return "", ErrFileNotAllowed
	}
	file, err := root.Open(logical)
	if err != nil {
		return "", ErrFileNotAllowed
	}
	defer file.Close()
	openedBefore, err := file.Stat()
	if err != nil || !openedBefore.Mode().IsRegular() || !os.SameFile(pathBefore, openedBefore) ||
		openedBefore.Size() > int64(maximum) {
		return "", ErrFileNotAllowed
	}

	first, err := readBounded(ctx, file, maximum)
	if err != nil {
		return "", err
	}
	if hooks.afterFirstRead != nil {
		hooks.afterFirstRead()
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", ErrFileNotAllowed
	}
	second, err := readBounded(ctx, file, maximum)
	if err != nil {
		return "", err
	}
	openedAfter, err := file.Stat()
	pathAfter, pathErr := root.Lstat(logical)
	rootAfter, rootErr := os.Lstat(workspaceRoot)
	if err != nil || pathErr != nil || pathAfter.Mode()&os.ModeSymlink != 0 ||
		rootErr != nil || rootAfter.Mode()&os.ModeSymlink != 0 || !os.SameFile(rootInfo, rootAfter) ||
		!stableFile(openedBefore, openedAfter) || !os.SameFile(openedAfter, pathAfter) ||
		!bytes.Equal(first, second) || !utf8.Valid(first) || bytes.IndexByte(first, 0) >= 0 {
		return "", ErrFileNotAllowed
	}
	return string(first), nil
}

func validatePhysicalPath(root *os.Root, logical string) error {
	parts := strings.Split(logical, "/")
	for index := range parts {
		candidate := path.Join(parts[:index+1]...)
		info, err := root.Lstat(candidate)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fs.ErrInvalid
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fs.ErrInvalid
		}
	}
	return nil
}

func readBounded(ctx context.Context, reader io.Reader, maximum int) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(reader, int64(maximum)+1))
	if err != nil {
		return nil, ErrFileNotAllowed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(content) > maximum {
		return nil, ErrFileNotAllowed
	}
	return content, nil
}

func stableFile(before, after fs.FileInfo) bool {
	return before != nil && after != nil && os.SameFile(before, after) &&
		before.Mode() == after.Mode() && before.Size() == after.Size() &&
		before.ModTime().Equal(after.ModTime())
}

func fileError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return ErrFileNotAllowed
}
