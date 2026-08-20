package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type atomicFileOps interface {
	openParent(*os.Root, string) (*os.File, error)
	openTarget(*os.File, string) (*os.File, error)
	createTemp(*os.File, os.FileMode) (*os.File, string, error)
	read(*os.File, int64) ([]byte, error)
	write(*os.File, []byte) error
	chmod(*os.File, os.FileMode) error
	syncFile(*os.File) error
	rename(*os.File, string, string) error
	syncDirectory(*os.File) error
	remove(*os.File, string) error
}

type atomicReplaceOutcome struct {
	matched   bool
	committed bool
	durable   bool
}

func replacePhysicalFileAtomically(
	ctx context.Context,
	root *os.Root,
	logical, expected, old, replacement, proposed string,
	maxBytes int64,
	ops atomicFileOps,
) (outcome atomicReplaceOutcome, err error) {
	if err := ctx.Err(); err != nil {
		return outcome, err
	}
	if ops == nil {
		return outcome, fmt.Errorf("atomic filesystem operations are unavailable: %w", pkgTool.ErrExecutionFailed)
	}
	parentName := path.Dir(logical)
	baseName := path.Base(logical)
	parent, err := ops.openParent(root, parentName)
	if err != nil {
		return outcome, err
	}
	defer func() { err = errors.Join(err, parent.Close()) }()

	source, sourceInfo, sourceData, err := readAtomicSource(ctx, ops, parent, baseName, maxBytes)
	if err != nil {
		return outcome, err
	}
	if digest(string(sourceData)) != expected || strings.Count(string(sourceData), old) != 1 ||
		strings.Replace(string(sourceData), old, replacement, 1) != proposed {
		_ = source.Close()
		return outcome, nil
	}
	if err := source.Close(); err != nil {
		return outcome, err
	}
	outcome.matched = true

	temporary, temporaryName, err := ops.createTemp(parent, sourceInfo.Mode().Perm())
	if err != nil {
		return outcome, err
	}
	cleanup := true
	defer func() {
		if temporary != nil {
			err = errors.Join(err, temporary.Close())
		}
		if cleanup {
			err = errors.Join(err, ops.remove(parent, temporaryName))
		}
	}()

	if err := ops.write(temporary, []byte(proposed)); err != nil {
		return outcome, err
	}
	if err := ctx.Err(); err != nil {
		return outcome, err
	}
	if err := ops.chmod(temporary, sourceInfo.Mode().Perm()); err != nil {
		return outcome, err
	}
	if err := ops.syncFile(temporary); err != nil {
		return outcome, err
	}
	if err := temporary.Close(); err != nil {
		temporary = nil
		return outcome, err
	}
	temporary = nil

	recheck, recheckInfo, recheckData, err := readAtomicSource(ctx, ops, parent, baseName, maxBytes)
	if err != nil {
		return outcome, err
	}
	if !os.SameFile(sourceInfo, recheckInfo) || digest(string(recheckData)) != expected ||
		strings.Count(string(recheckData), old) != 1 || strings.Replace(string(recheckData), old, replacement, 1) != proposed {
		_ = recheck.Close()
		outcome.matched = false
		return outcome, nil
	}
	if err := recheck.Close(); err != nil {
		return outcome, err
	}
	if err := ctx.Err(); err != nil {
		return outcome, err
	}
	if err := ops.rename(parent, temporaryName, baseName); err != nil {
		return outcome, err
	}
	cleanup = false
	outcome.committed = true
	if syncErr := ops.syncDirectory(parent); syncErr != nil {
		return outcome, syncErr
	}
	outcome.durable = true
	if err := ctx.Err(); err != nil {
		return outcome, err
	}
	return outcome, nil
}

func readAtomicSource(
	ctx context.Context,
	ops atomicFileOps,
	parent *os.File,
	baseName string,
	maxBytes int64,
) (*os.File, os.FileInfo, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	file, err := ops.openTarget(parent, baseName)
	if err != nil {
		return nil, nil, nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, nil, nil, errors.Join(err, pkgContext.ErrInvalidPath)
	}
	if info.Size() > maxBytes {
		_ = file.Close()
		return nil, nil, nil, pkgTool.ErrLimitExceeded
	}
	data, err := ops.read(file, maxBytes+1)
	if err != nil {
		_ = file.Close()
		return nil, nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		_ = file.Close()
		return nil, nil, nil, err
	}
	if int64(len(data)) > maxBytes || !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		_ = file.Close()
		return nil, nil, nil, pkgTool.ErrLimitExceeded
	}
	return file, info, data, nil
}

type platformAtomicFileOps struct{}

func (platformAtomicFileOps) openParent(root *os.Root, logical string) (*os.File, error) {
	entryInfo, err := root.Lstat(logical)
	if err != nil || entryInfo.Mode()&os.ModeSymlink != 0 || !entryInfo.IsDir() {
		return nil, errors.Join(err, pkgContext.ErrInvalidPath)
	}
	parent, err := root.Open(logical)
	if err != nil {
		return nil, err
	}
	info, err := parent.Stat()
	if err != nil || !info.IsDir() || !os.SameFile(entryInfo, info) {
		_ = parent.Close()
		return nil, errors.Join(err, pkgContext.ErrInvalidPath)
	}
	recheck, err := root.Lstat(logical)
	if err != nil || recheck.Mode()&os.ModeSymlink != 0 || !os.SameFile(info, recheck) {
		_ = parent.Close()
		return nil, errors.Join(err, pkgContext.ErrInvalidPath)
	}
	return parent, nil
}

func (platformAtomicFileOps) read(file *os.File, limit int64) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return io.ReadAll(io.LimitReader(file, limit))
}

func (platformAtomicFileOps) write(file *os.File, content []byte) error {
	written, err := file.Write(content)
	if err != nil {
		return err
	}
	if written != len(content) {
		return io.ErrShortWrite
	}
	return nil
}

func (platformAtomicFileOps) chmod(file *os.File, mode os.FileMode) error { return file.Chmod(mode) }
func (platformAtomicFileOps) syncFile(file *os.File) error                { return file.Sync() }
func (platformAtomicFileOps) syncDirectory(directory *os.File) error      { return directory.Sync() }

func atomicResultContent(logical, content string, applied, durable bool) string {
	encoded, _ := json.Marshal(struct {
		Path    string `json:"path"`
		Digest  string `json:"digest"`
		Applied bool   `json:"applied"`
		Durable bool   `json:"durable"`
	}{Path: logical, Digest: digest(content), Applied: applied, Durable: durable})
	return string(encoded)
}
