package contextengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

type FilesystemSource struct {
	beforeOpen func(string)
}

func NewFilesystemSource() *FilesystemSource { return &FilesystemSource{} }

func (*FilesystemSource) ID() pkgContext.SourceID { return pkgContext.SourceFilesystem }

func (source *FilesystemSource) Scan(ctx context.Context, workspace pkgContext.Workspace) (pkgContext.ScanResult, error) {
	if ctx == nil {
		return pkgContext.ScanResult{}, fmt.Errorf("scan filesystem with nil context: %w", pkgContext.ErrInvalidSource)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.ScanResult{}, err
	}
	if err := workspace.Validate(); err != nil {
		return pkgContext.ScanResult{}, err
	}
	rootInfo, err := os.Lstat(workspace.Root())
	if err != nil {
		return pkgContext.ScanResult{}, fmt.Errorf("inspect workspace root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return pkgContext.ScanResult{}, fmt.Errorf("workspace root is not a physical directory: %w", pkgContext.ErrInvalidWorkspace)
	}

	policy := workspace.Policy()
	documents := make([]pkgContext.Document, 0)
	var totalBytes int64
	err = filepath.WalkDir(workspace.Root(), func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if current == workspace.Root() {
			return nil
		}
		logical, err := logicalPath(workspace.Root(), current)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if hiddenPath(logical) && !policy.IncludeHidden {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if excluded(string(logical), policy.Exclude) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() || !included(string(logical), policy.Include) {
			return nil
		}
		if len(documents) >= policy.MaxFiles {
			return fmt.Errorf("workspace exceeds %d files: %w", policy.MaxFiles, pkgContext.ErrLimitExceeded)
		}
		content, mediaType, language, err := readStableFile(ctx, current, policy.MaxFileBytes, policy.IncludeBinary, source.beforeOpen)
		if errors.Is(err, errSkippedBinary) {
			return nil
		}
		if err != nil {
			return err
		}
		if totalBytes+int64(len(content)) > policy.MaxTotalBytes {
			return fmt.Errorf("workspace exceeds %d bytes: %w", policy.MaxTotalBytes, pkgContext.ErrLimitExceeded)
		}
		document, err := pkgContext.NewDocument(logical, mediaType, language, content)
		if err != nil {
			return err
		}
		documents = append(documents, document)
		totalBytes += int64(len(content))
		return nil
	})
	if err != nil {
		return pkgContext.ScanResult{}, err
	}
	slices.SortFunc(documents, func(left, right pkgContext.Document) int {
		return left.Path().Compare(right.Path())
	})
	return pkgContext.ScanResult{Documents: documents}, nil
}

var errSkippedBinary = errors.New("binary file skipped")

func readStableFile(ctx context.Context, filename string, maxBytes int64, includeBinary bool, beforeOpen func(string)) (string, string, pkgContext.Language, error) {
	before, err := os.Lstat(filename)
	if err != nil {
		return "", "", "", err
	}
	if !before.Mode().IsRegular() {
		return "", "", "", fmt.Errorf("file changed type during scan: %w", pkgContext.ErrSourceFailure)
	}
	if before.Size() > maxBytes {
		return "", "", "", fmt.Errorf("file exceeds %d bytes: %w", maxBytes, pkgContext.ErrLimitExceeded)
	}
	if beforeOpen != nil {
		beforeOpen(filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", "", "", err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return "", "", "", err
	}
	if !os.SameFile(before, opened) {
		return "", "", "", fmt.Errorf("file changed before read: %w", pkgContext.ErrSourceFailure)
	}
	data, err := io.ReadAll(io.LimitReader(&contextReader{ctx: ctx, reader: file}, maxBytes+1))
	if err != nil {
		return "", "", "", err
	}
	if int64(len(data)) > maxBytes {
		return "", "", "", fmt.Errorf("file exceeds %d bytes: %w", maxBytes, pkgContext.ErrLimitExceeded)
	}
	after, err := os.Lstat(filename)
	if err != nil {
		return "", "", "", err
	}
	if !os.SameFile(opened, after) || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
		return "", "", "", fmt.Errorf("file changed during read: %w", pkgContext.ErrSourceFailure)
	}
	if binary(data) {
		if !includeBinary {
			return "", "", "", errSkippedBinary
		}
		return string(data), "application/octet-stream", "", nil
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	mediaType, language := classifyText(filename)
	return content, mediaType, language, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader *contextReader) Read(buffer []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.reader.Read(buffer)
}

func logicalPath(root, filename string) (pkgContext.DocumentPath, error) {
	relative, err := filepath.Rel(root, filename)
	if err != nil {
		return "", err
	}
	logical := pkgContext.DocumentPath(filepath.ToSlash(relative))
	if err := logical.Validate(); err != nil {
		return "", err
	}
	return logical, nil
}

func hiddenPath(logical pkgContext.DocumentPath) bool {
	for _, part := range strings.Split(string(logical), "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func included(logical string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	return matchesAny(logical, patterns)
}

func excluded(logical string, patterns []string) bool { return matchesAny(logical, patterns) }

func matchesAny(logical string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if logical == prefix || strings.HasPrefix(logical, prefix+"/") {
				return true
			}
		}
		if matched, _ := path.Match(pattern, logical); matched {
			return true
		}
		if !strings.Contains(pattern, "/") {
			if matched, _ := path.Match(pattern, path.Base(logical)); matched {
				return true
			}
		}
	}
	return false
}

func binary(data []byte) bool {
	return bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data)
}

func classifyText(filename string) (string, pkgContext.Language) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".go":
		return "text/x-go", "go"
	case ".php":
		return "text/x-php", "php"
	case ".js", ".mjs", ".cjs":
		return "text/javascript", "javascript"
	case ".ts", ".tsx":
		return "text/typescript", "typescript"
	case ".json":
		return "application/json", "json"
	case ".md", ".markdown":
		return "text/markdown", "markdown"
	case ".yaml", ".yml":
		return "application/yaml", "yaml"
	default:
		return "text/plain", ""
	}
}
