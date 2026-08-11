package contextengine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

func TestFilesystemSourceProducesDeterministicSnapshot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\r\n\r\nfunc main() {}\r\n")
	writeFile(t, root, "app/Service.php", "<?php\nclass Service {}\n")
	writeFile(t, root, ".secret", "not indexed")
	writeFile(t, root, "vendor/dependency.php", "<?php")
	writeBytes(t, root, "asset.bin", []byte{0, 1, 2})
	if err := os.Symlink(filepath.Join(root, "main.go"), filepath.Join(root, "linked.go")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	engine := New()
	snapshot, err := engine.Index(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, pkgContext.DefaultScanPolicy()))
	if err != nil {
		t.Fatalf("index workspace: %v", err)
	}
	documents := snapshot.Documents()
	if len(documents) != 2 || documents[0].Path() != "app/Service.php" || documents[1].Path() != "main.go" {
		t.Fatalf("unexpected documents: %#v", documents)
	}
	if documents[1].Content() != "package main\n\nfunc main() {}\n" {
		t.Fatalf("line endings were not normalized: %q", documents[1].Content())
	}
	if documents[0].Language() != "php" || documents[1].Language() != "go" {
		t.Fatalf("unexpected language classification: %q %q", documents[0].Language(), documents[1].Language())
	}
	metadata := snapshot.Metadata()
	if metadata.Generation != 1 || metadata.DocumentCount != 2 {
		t.Fatalf("unexpected snapshot metadata: %#v", metadata)
	}
}

func TestFilesystemSourceHonorsIncludeHiddenAndBinary(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n")
	writeFile(t, root, "nested/other.go", "package nested\n")
	writeFile(t, root, ".config", "enabled=true\n")
	writeBytes(t, root, "asset.bin", []byte{0, 1, 2})
	policy := pkgContext.DefaultScanPolicy()
	policy.Include = []string{"*.go", ".config", "*.bin"}
	policy.IncludeHidden = true
	policy.IncludeBinary = true

	result, err := NewFilesystemSource().Scan(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, policy))
	if err != nil {
		t.Fatalf("scan workspace: %v", err)
	}
	if len(result.Documents) != 4 {
		t.Fatalf("expected four documents, got %d", len(result.Documents))
	}
	var binaryFound bool
	for _, document := range result.Documents {
		if document.Path() == "asset.bin" {
			binaryFound = document.MediaType() == "application/octet-stream" && document.Content() == string([]byte{0, 1, 2})
		}
	}
	if !binaryFound {
		t.Fatal("opaque binary document was not indexed")
	}
}

func TestFilesystemSourceEnforcesLimits(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.txt", "1234")
	writeFile(t, root, "b.txt", "5678")
	tests := []struct {
		name   string
		policy pkgContext.ScanPolicy
	}{
		{name: "file count", policy: pkgContext.ScanPolicy{MaxFiles: 1, MaxFileBytes: 8, MaxTotalBytes: 8}},
		{name: "file size", policy: pkgContext.ScanPolicy{MaxFiles: 2, MaxFileBytes: 3, MaxTotalBytes: 8}},
		{name: "total size", policy: pkgContext.ScanPolicy{MaxFiles: 2, MaxFileBytes: 4, MaxTotalBytes: 7}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFilesystemSource().Scan(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, test.policy))
			if !errors.Is(err, pkgContext.ErrLimitExceeded) {
				t.Fatalf("expected limit error, got %v", err)
			}
		})
	}
}

func TestFilesystemSourceRejectsFileReplacementDuringRead(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package old\n")
	source := &FilesystemSource{beforeOpen: func(filename string) {
		replacement := filepath.Join(root, "replacement")
		if err := os.WriteFile(replacement, []byte("package new\n"), 0o600); err != nil {
			t.Errorf("write replacement: %v", err)
			return
		}
		if err := os.Rename(replacement, filename); err != nil {
			t.Errorf("replace fixture: %v", err)
		}
	}}
	_, err := source.Scan(context.Background(), workspace(t, root, pkgContext.SourceFilesystem, pkgContext.DefaultScanPolicy()))
	if !errors.Is(err, pkgContext.ErrSourceFailure) {
		t.Fatalf("expected mutation failure, got %v", err)
	}
}

func TestIndexFailurePreservesPublishedSnapshot(t *testing.T) {
	root := t.TempDir()
	document := document(t, "main.go", "package main\n")
	source := &mutableSource{id: "context.mutable", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document}}}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	workspace := workspace(t, root, source.id, pkgContext.DefaultScanPolicy())
	first, err := engine.Index(context.Background(), workspace)
	if err != nil {
		t.Fatalf("initial index: %v", err)
	}
	source.set(pkgContext.ScanResult{}, errors.New("scan failed"))
	if _, err := engine.Index(context.Background(), workspace); !errors.Is(err, pkgContext.ErrSourceFailure) {
		t.Fatalf("expected source failure, got %v", err)
	}
	current, found := engine.Snapshot("workspace")
	if !found || current.Metadata().Generation != first.Metadata().Generation || current.Documents()[0].Digest() != document.Digest() {
		t.Fatalf("failed refresh replaced snapshot: %#v %v", current.Metadata(), found)
	}
}

func TestInvalidSourceResultIsNotPublished(t *testing.T) {
	root := t.TempDir()
	document := document(t, "main.go", "package main\n")
	source := &mutableSource{id: "context.invalid", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document, document}}}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	if _, err := engine.Index(context.Background(), workspace(t, root, source.id, pkgContext.DefaultScanPolicy())); !errors.Is(err, pkgContext.ErrInvalidSnapshot) {
		t.Fatalf("expected invalid snapshot, got %v", err)
	}
	if _, found := engine.Snapshot("workspace"); found {
		t.Fatal("invalid source result was published")
	}
}

func TestSourceRunsWithoutGlobalEngineLock(t *testing.T) {
	root := t.TempDir()
	entered := make(chan struct{})
	release := make(chan struct{})
	blocking := &blockingSource{id: "context.blocking", entered: entered, release: release}
	engine := New()
	if err := engine.RegisterSource(blocking); err != nil {
		t.Fatalf("register blocking source: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := engine.Index(context.Background(), workspace(t, root, blocking.id, pkgContext.DefaultScanPolicy()))
		done <- err
	}()
	<-entered
	registered := make(chan error, 1)
	go func() { registered <- engine.RegisterSource(&mutableSource{id: "context.other"}) }()
	select {
	case err := <-registered:
		if err != nil {
			t.Fatalf("register while scan blocked: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("source callback held the global engine lock")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("blocked index: %v", err)
	}
}

func TestConcurrentIndexGenerationsAreMonotonic(t *testing.T) {
	root := t.TempDir()
	source := &mutableSource{id: "context.concurrent", result: pkgContext.ScanResult{Documents: []pkgContext.Document{document(t, "main.go", "package main\n")}}}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	workspace := workspace(t, root, source.id, pkgContext.DefaultScanPolicy())
	const runs = 20
	errorsChannel := make(chan error, runs)
	var wait sync.WaitGroup
	for range runs {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := engine.Index(context.Background(), workspace)
			errorsChannel <- err
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent index: %v", err)
		}
	}
	snapshot, found := engine.Snapshot("workspace")
	if !found || snapshot.Metadata().Generation != runs {
		t.Fatalf("generation=%d found=%v", snapshot.Metadata().Generation, found)
	}
}

func TestIndexPreservesCancellationAndRejectsTypedNil(t *testing.T) {
	engine := New()
	var source *mutableSource
	if err := engine.RegisterSource(source); !errors.Is(err, pkgContext.ErrInvalidSource) {
		t.Fatalf("expected typed nil rejection, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := engine.Index(ctx, workspace(t, t.TempDir(), pkgContext.SourceFilesystem, pkgContext.DefaultScanPolicy()))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestCancellationDuringSourceDoesNotPublish(t *testing.T) {
	root := t.TempDir()
	entered := make(chan struct{})
	release := make(chan struct{})
	source := &blockingSource{id: "context.cancel", entered: entered, release: release}
	engine := New()
	if err := engine.RegisterSource(source); err != nil {
		t.Fatalf("register source: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := engine.Index(ctx, workspace(t, root, source.id, pkgContext.DefaultScanPolicy()))
		done <- err
	}()
	<-entered
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if _, found := engine.Snapshot("workspace"); found {
		t.Fatal("canceled source published a snapshot")
	}
}

type mutableSource struct {
	mu     sync.Mutex
	id     pkgContext.SourceID
	result pkgContext.ScanResult
	err    error
}

func (source *mutableSource) ID() pkgContext.SourceID { return source.id }
func (source *mutableSource) Scan(context.Context, pkgContext.Workspace) (pkgContext.ScanResult, error) {
	source.mu.Lock()
	defer source.mu.Unlock()
	return source.result, source.err
}
func (source *mutableSource) set(result pkgContext.ScanResult, err error) {
	source.mu.Lock()
	source.result, source.err = result, err
	source.mu.Unlock()
}

type blockingSource struct {
	id      pkgContext.SourceID
	entered chan<- struct{}
	release <-chan struct{}
}

func (source *blockingSource) ID() pkgContext.SourceID { return source.id }
func (source *blockingSource) Scan(ctx context.Context, _ pkgContext.Workspace) (pkgContext.ScanResult, error) {
	close(source.entered)
	select {
	case <-source.release:
		return pkgContext.ScanResult{}, nil
	case <-ctx.Done():
		return pkgContext.ScanResult{}, ctx.Err()
	}
}

func workspace(t *testing.T, root string, source pkgContext.SourceID, policy pkgContext.ScanPolicy) pkgContext.Workspace {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace("workspace", filepath.Clean(root), pkgContext.WorkspaceOptions{Source: source, Policy: policy})
	if err != nil {
		t.Fatalf("construct workspace: %v", err)
	}
	return workspace
}

func document(t *testing.T, path pkgContext.DocumentPath, content string) pkgContext.Document {
	t.Helper()
	document, err := pkgContext.NewDocument(path, "text/x-go", "go", content)
	if err != nil {
		t.Fatalf("construct document: %v", err)
	}
	return document
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	writeBytes(t, root, name, []byte(content))
}

func writeBytes(t *testing.T, root, name string, content []byte) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(filename, content, 0o600); err != nil {
		t.Fatalf("write fixture %q: %v", name, err)
	}
}
