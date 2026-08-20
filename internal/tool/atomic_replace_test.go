//go:build linux

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

var errInjectedAtomicFault = errors.New("injected atomic filesystem fault")

func TestAtomicReplaceCommitsNewInodePreservesModeAndNeverExposesPartialContent(t *testing.T) {
	rootPath, logical, original := atomicFixture(t)
	target := filepath.Join(rootPath, filepath.FromSlash(logical))
	if err := os.Chmod(target, 0o640); err != nil {
		t.Fatal(err)
	}
	beforeInfo, _ := os.Stat(target)
	original = "<?php\nclass Order {}\n" + strings.Repeat("// unchanged padding\n", 4096)
	proposed := "<?php\nfinal class Order {}\n" + strings.Repeat("// unchanged padding\n", 4096)
	if err := os.WriteFile(target, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	readerErrors := make(chan error, 1)
	var readers sync.WaitGroup
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				content, err := os.ReadFile(target)
				if err != nil {
					select {
					case readerErrors <- err:
					default:
					}
					return
				}
				if string(content) != original && string(content) != proposed {
					select {
					case readerErrors <- errors.New("reader observed partial content"):
					default:
					}
					return
				}
			}
		}()
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := replacePhysicalFileAtomically(
		t.Context(), root, logical, digest(original), "class Order", "final class Order",
		proposed, 2<<20, defaultAtomicFileOps(),
	)
	_ = root.Close()
	close(stop)
	readers.Wait()
	select {
	case err := <-readerErrors:
		t.Fatal(err)
	default:
	}
	if err != nil || !outcome.matched || !outcome.committed || !outcome.durable {
		t.Fatalf("outcome=%#v err=%v", outcome, err)
	}
	after, _ := os.ReadFile(target)
	afterInfo, _ := os.Stat(target)
	if string(after) != proposed || afterInfo.Mode().Perm() != 0o640 || os.SameFile(beforeInfo, afterInfo) {
		t.Fatalf("atomic replacement mismatch: mode=%o same_inode=%t", afterInfo.Mode().Perm(), os.SameFile(beforeInfo, afterInfo))
	}
	assertNoAtomicTemps(t, rootPath)
}

func TestAtomicReplaceFaultsBeforeRenameLeaveTargetByteIdenticalAndClean(t *testing.T) {
	for _, stage := range []string{
		"open_parent", "open_target", "read", "create_temp", "write", "chmod",
		"sync_file", "open_target_recheck", "read_recheck", "rename",
	} {
		t.Run(stage, func(t *testing.T) {
			rootPath, logical, original := atomicFixture(t)
			root, err := os.OpenRoot(rootPath)
			if err != nil {
				t.Fatal(err)
			}
			ops := &faultAtomicFileOps{delegate: defaultAtomicFileOps(), failStage: stage}
			outcome, replaceErr := replacePhysicalFileAtomically(
				t.Context(), root, logical, digest(original), "class Order", "final class Order",
				"<?php\nfinal class Order {}\n", 2<<20, ops,
			)
			_ = root.Close()
			if !errors.Is(replaceErr, errInjectedAtomicFault) || outcome.committed {
				t.Fatalf("stage=%s outcome=%#v err=%v", stage, outcome, replaceErr)
			}
			content, _ := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(logical)))
			if string(content) != original {
				t.Fatalf("stage %s changed target: %q", stage, content)
			}
			assertNoAtomicTemps(t, rootPath)
		})
	}
}

func TestAtomicReplaceDetectsConcurrentTargetReplacementAtCommitBoundary(t *testing.T) {
	rootPath, logical, original := atomicFixture(t)
	concurrent := "<?php\nclass Concurrent {}\n"
	root, _ := os.OpenRoot(rootPath)
	ops := &faultAtomicFileOps{
		delegate: defaultAtomicFileOps(),
		beforeRecheck: func() error {
			return os.WriteFile(filepath.Join(rootPath, filepath.FromSlash(logical)), []byte(concurrent), 0o644)
		},
	}
	outcome, err := replacePhysicalFileAtomically(
		t.Context(), root, logical, digest(original), "class Order", "final class Order",
		"<?php\nfinal class Order {}\n", 2<<20, ops,
	)
	_ = root.Close()
	content, _ := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(logical)))
	if err != nil || outcome.matched || outcome.committed || string(content) != concurrent {
		t.Fatalf("outcome=%#v content=%q err=%v", outcome, content, err)
	}
	assertNoAtomicTemps(t, rootPath)
}

func TestAtomicReplaceReportsPostCommitSyncFailureAndCancellation(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		configure   func(*faultAtomicFileOps, context.CancelFunc)
		wantError   error
		wantDurable bool
	}{
		{name: "directory sync", configure: func(ops *faultAtomicFileOps, _ context.CancelFunc) { ops.failStage = "sync_directory" }, wantError: errInjectedAtomicFault},
		{name: "canceled after rename", configure: func(ops *faultAtomicFileOps, cancel context.CancelFunc) { ops.cancelAfterRename = cancel }, wantError: context.Canceled, wantDurable: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			rootPath, logical, original := atomicFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ops := &faultAtomicFileOps{delegate: defaultAtomicFileOps()}
			testCase.configure(ops, cancel)
			root, _ := os.OpenRoot(rootPath)
			outcome, err := replacePhysicalFileAtomically(
				ctx, root, logical, digest(original), "class Order", "final class Order",
				"<?php\nfinal class Order {}\n", 2<<20, ops,
			)
			_ = root.Close()
			content, _ := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(logical)))
			if !errors.Is(err, testCase.wantError) || !outcome.committed || outcome.durable != testCase.wantDurable || string(content) != "<?php\nfinal class Order {}\n" {
				t.Fatalf("outcome=%#v content=%q err=%v", outcome, content, err)
			}
			assertNoAtomicTemps(t, rootPath)
		})
	}
}

func TestAtomicReplaceAttemptsCleanupAndSurfacesCleanupFailure(t *testing.T) {
	rootPath, logical, original := atomicFixture(t)
	root, _ := os.OpenRoot(rootPath)
	ops := &faultAtomicFileOps{delegate: defaultAtomicFileOps(), failStage: "write", failCleanup: true}
	outcome, err := replacePhysicalFileAtomically(
		t.Context(), root, logical, digest(original), "class Order", "final class Order",
		"<?php\nfinal class Order {}\n", 2<<20, ops,
	)
	_ = root.Close()
	if !errors.Is(err, errInjectedAtomicFault) || outcome.committed || ops.cleanupCalls != 1 {
		t.Fatalf("outcome=%#v cleanup_calls=%d err=%v", outcome, ops.cleanupCalls, err)
	}
	content, _ := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(logical)))
	if string(content) != original {
		t.Fatalf("cleanup failure changed target: %q", content)
	}
}

func TestWorkspacePatchReportsCommittedButUnsyncedDirectory(t *testing.T) {
	rootPath, logical, original := atomicFixture(t)
	workspace, _ := pkgContext.NewWorkspace("workspace", rootPath, pkgContext.WorkspaceOptions{
		Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy(),
	})
	registry := NewWorkspaceRegistry()
	run := pkgTool.RunID("run-post-commit")
	_ = registry.Bind(run, workspace)
	tools, _ := NewWorkspaceTools(registry)
	var patch *workspaceTool
	for _, candidate := range tools {
		if candidate.Descriptor().ID() == WorkspacePatchID {
			patch = candidate.(*workspaceTool)
		}
	}
	patch.atomicOps = &faultAtomicFileOps{delegate: defaultAtomicFileOps(), failStage: "sync_directory"}
	arguments, _ := json.Marshal(map[string]string{
		"path": logical, "old": "class Order", "new": "final class Order", "expected_digest": digest(original),
	})
	invocation, _ := pkgTool.NewInvocation(WorkspacePatchID, "call-post-commit", run, arguments)
	prepared, err := patch.Prepare(t.Context(), invocation)
	if err != nil {
		t.Fatal(err)
	}
	result, err := patch.Execute(t.Context(), prepared)
	if err != nil || result.Outcome() != pkgTool.ResultFailed || result.Reason() != "post_commit_sync_failed" ||
		!strings.Contains(result.Content(), `"applied":true`) || !strings.Contains(result.Content(), `"durable":false`) {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	content, _ := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(logical)))
	if string(content) != "<?php\nfinal class Order {}\n" {
		t.Fatalf("committed patch was not retained: %q", content)
	}
}

func atomicFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	logical := "app/Order.php"
	original := "<?php\nclass Order {}\n"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(logical)), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, logical, original
}

func assertNoAtomicTemps(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "app"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), atomicTempPrefix) {
			t.Fatalf("atomic temporary was not cleaned: %s", entry.Name())
		}
	}
}

type faultAtomicFileOps struct {
	delegate          atomicFileOps
	failStage         string
	failCleanup       bool
	beforeRecheck     func() error
	cancelAfterRename context.CancelFunc
	openTargetCalls   int
	readCalls         int
	cleanupCalls      int
}

func (ops *faultAtomicFileOps) failure(stage string) error {
	if ops.failStage == stage {
		return errInjectedAtomicFault
	}
	return nil
}

func (ops *faultAtomicFileOps) openParent(root *os.Root, logical string) (*os.File, error) {
	if err := ops.failure("open_parent"); err != nil {
		return nil, err
	}
	return ops.delegate.openParent(root, logical)
}

func (ops *faultAtomicFileOps) openTarget(parent *os.File, name string) (*os.File, error) {
	ops.openTargetCalls++
	stage := "open_target"
	if ops.openTargetCalls == 2 {
		stage = "open_target_recheck"
		if ops.beforeRecheck != nil {
			if err := ops.beforeRecheck(); err != nil {
				return nil, err
			}
		}
	}
	if err := ops.failure(stage); err != nil {
		return nil, err
	}
	return ops.delegate.openTarget(parent, name)
}

func (ops *faultAtomicFileOps) createTemp(parent *os.File, mode os.FileMode) (*os.File, string, error) {
	if err := ops.failure("create_temp"); err != nil {
		return nil, "", err
	}
	return ops.delegate.createTemp(parent, mode)
}

func (ops *faultAtomicFileOps) read(file *os.File, limit int64) ([]byte, error) {
	ops.readCalls++
	stage := "read"
	if ops.readCalls == 2 {
		stage = "read_recheck"
	}
	if err := ops.failure(stage); err != nil {
		return nil, err
	}
	return ops.delegate.read(file, limit)
}

func (ops *faultAtomicFileOps) write(file *os.File, content []byte) error {
	if err := ops.failure("write"); err != nil {
		return err
	}
	return ops.delegate.write(file, content)
}

func (ops *faultAtomicFileOps) chmod(file *os.File, mode os.FileMode) error {
	if err := ops.failure("chmod"); err != nil {
		return err
	}
	return ops.delegate.chmod(file, mode)
}

func (ops *faultAtomicFileOps) syncFile(file *os.File) error {
	if err := ops.failure("sync_file"); err != nil {
		return err
	}
	return ops.delegate.syncFile(file)
}

func (ops *faultAtomicFileOps) rename(parent *os.File, source, target string) error {
	if err := ops.failure("rename"); err != nil {
		return err
	}
	if err := ops.delegate.rename(parent, source, target); err != nil {
		return err
	}
	if ops.cancelAfterRename != nil {
		ops.cancelAfterRename()
	}
	return nil
}

func (ops *faultAtomicFileOps) syncDirectory(directory *os.File) error {
	if err := ops.failure("sync_directory"); err != nil {
		return err
	}
	return ops.delegate.syncDirectory(directory)
}

func (ops *faultAtomicFileOps) remove(parent *os.File, name string) error {
	ops.cleanupCalls++
	if ops.failCleanup {
		return errInjectedAtomicFault
	}
	return ops.delegate.remove(parent, name)
}
