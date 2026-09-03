package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgLaravel "github.com/antonio-cafeo/maestro/pkg/plugin/laravel"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestWorkspaceToolsListReadSearchWriteAndPatch(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "app", "Order.php"), []byte("<?php\nclass Order {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runtime, _, run := workspaceRuntime(t, root)

	listed := invokeWorkspace(t, runtime, run, WorkspaceListID, json.RawMessage(`{"path":"src","max_entries":10}`))
	if listed.Outcome() != pkgTool.ResultSuccess || !strings.Contains(listed.Content(), `"path":"src/main.go"`) {
		t.Fatalf("unexpected listing: %#v", listed)
	}
	read := invokeWorkspace(t, runtime, run, WorkspaceReadID, json.RawMessage(`{"path":"src/main.go"}`))
	if !strings.Contains(read.Content(), `package main`) || !strings.Contains(read.Content(), `"digest"`) {
		t.Fatalf("unexpected read: %s", read.Content())
	}
	searched := invokeWorkspace(t, runtime, run, WorkspaceSearchID, json.RawMessage(`{"path":"src","query":"func main","max_results":5}`))
	if searched.ItemCount() != 1 || !strings.Contains(searched.Content(), `"line":2`) {
		t.Fatalf("unexpected search: %#v", searched)
	}

	written := invokeWorkspace(t, runtime, run, WorkspaceWriteID, json.RawMessage(`{"path":"src/new.go","content":"package new\n","expected_digest":"absent"}`))
	if written.Outcome() != pkgTool.ResultSuccess {
		t.Fatalf("write failed: %#v", written)
	}
	existing := invokeWorkspace(t, runtime, run, WorkspaceReadID, json.RawMessage(`{"path":"app/Order.php"}`))
	digest := resultDigest(t, existing.Content())
	patchedArguments, _ := json.Marshal(map[string]string{
		"path": "app/Order.php", "old": "class Order {}", "new": "final class Order {}", "expected_digest": digest,
	})
	patched := invokeWorkspace(t, runtime, run, WorkspacePatchID, patchedArguments)
	if patched.Outcome() != pkgTool.ResultSuccess {
		t.Fatalf("patch failed: %#v", patched)
	}
	content, err := os.ReadFile(filepath.Join(root, "app", "Order.php"))
	if err != nil || string(content) != "<?php\nfinal class Order {}\n" {
		t.Fatalf("unexpected patched file: %q %v", content, err)
	}

	conflict := invokeWorkspace(t, runtime, run, WorkspaceWriteID, json.RawMessage(`{"path":"src/new.go","content":"overwrite","expected_digest":"absent"}`))
	if conflict.Outcome() != pkgTool.ResultFailed || conflict.Reason() != "precondition_failed" {
		t.Fatalf("expected content precondition conflict: %#v", conflict)
	}
}

func TestWorkspacePatchPreparesAuthoritativePreviewWithoutMutation(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	original := "<?php\nclass Order\n{\n    public function total(): int { return 1; }\n}\n"
	logical := "app/Order.php"
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(logical)), []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}
	workspace, err := pkgContext.NewWorkspace("workspace", root, pkgContext.WorkspaceOptions{
		Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy(),
	})
	if err != nil {
		t.Fatal(err)
	}
	registry := NewWorkspaceRegistry()
	run := pkgTool.RunID("run-preview")
	if err := registry.Bind(run, workspace); err != nil {
		t.Fatal(err)
	}
	tools, err := NewWorkspaceTools(registry)
	if err != nil {
		t.Fatal(err)
	}
	var patchTool pkgTool.Tool
	for _, candidate := range tools {
		if candidate.Descriptor().ID() == WorkspacePatchID {
			patchTool = candidate
		}
	}
	arguments, _ := json.Marshal(map[string]string{
		"path": logical, "old": "return 1", "new": "return 2", "expected_digest": digest(original),
	})
	invocation, _ := pkgTool.NewInvocation(WorkspacePatchID, "call-preview", run, arguments)
	prepared, err := patchTool.Prepare(context.Background(), invocation)
	if err != nil {
		t.Fatal(err)
	}
	preview, ok := prepared.Preview()
	if !ok || preview.MediaType() != "text/x-diff" || !strings.Contains(preview.Body(), "-    public function total(): int { return 1; }") ||
		!strings.Contains(preview.Body(), "+    public function total(): int { return 2; }") {
		t.Fatalf("unexpected patch preview: %#v", preview)
	}
	if strings.Contains(preview.Body(), root) || strings.Contains(preview.Summary(), root) {
		t.Fatalf("physical root leaked in preview: %#v", preview)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
	if err != nil || string(content) != original {
		t.Fatalf("prepare mutated workspace: %q %v", content, err)
	}
}

func TestWorkspacePatchPreparationRejectsUnsupportedAndAmbiguousProposals(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"app/Order.php": "same same\n",
		"app/data.txt":  "same\n",
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	workspace, _ := pkgContext.NewWorkspace("workspace", root, pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	registry := NewWorkspaceRegistry()
	run := pkgTool.RunID("run-reject-preview")
	_ = registry.Bind(run, workspace)
	tools, _ := NewWorkspaceTools(registry)
	var patchTool pkgTool.Tool
	for _, candidate := range tools {
		if candidate.Descriptor().ID() == WorkspacePatchID {
			patchTool = candidate
		}
	}
	for index, test := range []struct {
		path, old, expected string
	}{
		{path: "app/data.txt", old: "same", expected: digest("same\n")},
		{path: "app/Order.php", old: "same", expected: digest("same same\n")},
		{path: "app/Order.php", old: "same same", expected: digest("stale")},
	} {
		arguments, _ := json.Marshal(map[string]string{"path": test.path, "old": test.old, "new": "changed", "expected_digest": test.expected})
		invocation, _ := pkgTool.NewInvocation(WorkspacePatchID, pkgTool.CallID(fmt.Sprintf("call-reject-%d", index)), run, arguments)
		if _, err := patchTool.Prepare(context.Background(), invocation); err == nil {
			t.Fatalf("unsupported proposal accepted: %#v", test)
		}
	}
}

func TestWorkspaceReplaceCompilesStrictV1ProposalIntoBoundPreview(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	original := "package demo\n\nconst answer = 41\n"
	if err := os.WriteFile(filepath.Join(root, "src", "demo.go"), []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}
	workspace, _ := pkgContext.NewWorkspace("workspace", root, pkgContext.WorkspaceOptions{Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy()})
	registry := NewWorkspaceRegistry()
	run := pkgTool.RunID("run-m28-preview")
	_ = registry.Bind(run, workspace)
	replace, err := NewControlledMutationTool(registry)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"version":1,"path":"src/demo.go","operation":"replace","old_text":"answer = 41","new_text":"answer = 42"}`)
	invocation, _ := pkgTool.NewInvocation(WorkspaceReplaceID, "call-m28-preview", run, raw)
	prepared, err := replace.Prepare(t.Context(), invocation)
	if err != nil {
		t.Fatal(err)
	}
	preview, ok := prepared.Preview()
	if !ok || !strings.Contains(preview.Body(), "-const answer = 41") || !strings.Contains(preview.Body(), "+const answer = 42") {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	fields := map[string]string{}
	for _, field := range preview.Fields() {
		fields[field.Label()] = field.Value()
	}
	if fields["path"] != "src/demo.go" || len(fields["fingerprint"]) != 64 || len(fields["before_sha256"]) != 64 || len(fields["after_sha256"]) != 64 {
		t.Fatalf("preview is not fully bound: %#v", fields)
	}
	content, _ := os.ReadFile(filepath.Join(root, "src", "demo.go"))
	if string(content) != original {
		t.Fatal("prepare changed the workspace")
	}
}

func TestWorkspaceToolsRejectTraversalAbsolutePathsAndSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inside.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "external-link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("inside.txt", filepath.Join(root, "internal-link")); err != nil {
		t.Fatal(err)
	}
	runtime, _, run := workspaceRuntime(t, root)

	for index, arguments := range []json.RawMessage{
		json.RawMessage(`{"path":"../outside.txt"}`),
		json.RawMessage(`{"path":"/etc/passwd"}`),
		json.RawMessage(`{"path":"external-link"}`),
		json.RawMessage(`{"path":"internal-link"}`),
	} {
		invocation, _ := pkgTool.NewInvocation(WorkspaceReadID, pkgTool.CallID(fmt.Sprintf("call-unsafe-%d", index)), run, arguments)
		execution, _ := pkgTool.NewExecutionRequest(invocation, "policy.test", nil, workspaceExecutionLimits())
		if _, err := runtime.Invoke(context.Background(), execution); err == nil {
			t.Fatalf("unsafe path was accepted: %s", arguments)
		}
	}

	arguments := json.RawMessage(`{"path":"external-link","content":"changed","expected_digest":"absent"}`)
	invocation, _ := pkgTool.NewInvocation(WorkspaceWriteID, "call-write-link", run, arguments)
	execution, _ := pkgTool.NewExecutionRequest(invocation, "policy.test", nil, workspaceExecutionLimits())
	if _, err := runtime.Invoke(context.Background(), execution); err == nil {
		t.Fatal("write through symlink was accepted")
	}
	content, _ := os.ReadFile(outside)
	if string(content) != "secret" {
		t.Fatalf("outside file changed: %q", content)
	}
}

func TestWorkspaceWriteRechecksDigestAtExecution(t *testing.T) {
	root := t.TempDir()
	filename := filepath.Join(root, "file.txt")
	if err := os.WriteFile(filename, []byte("before"), 0o644); err != nil {
		t.Fatal(err)
	}
	runtime, _, run := workspaceRuntime(t, root)
	expected := digest("before")
	if err := os.WriteFile(filename, []byte("concurrent"), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments, _ := json.Marshal(map[string]string{"path": "file.txt", "content": "after", "expected_digest": expected})
	result := invokeWorkspace(t, runtime, run, WorkspaceWriteID, arguments)
	if result.Outcome() != pkgTool.ResultFailed {
		t.Fatalf("stale precondition accepted: %#v", result)
	}
	content, _ := os.ReadFile(filename)
	if string(content) != "concurrent" {
		t.Fatalf("concurrent content overwritten: %q", content)
	}
}

func TestWorkspaceToolsUseFrameworkNeutralLaravelWorkspace(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "artisan"), []byte("#!/usr/bin/env php\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"require":{"laravel/framework":"^12.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	plugin, err := pkgLaravel.New(pkgLaravel.Config{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := plugin.Workspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	runtime, _, run := runtimeForWorkspace(t, workspace)
	result := invokeWorkspace(t, runtime, run, WorkspaceReadID, json.RawMessage(`{"path":"composer.json"}`))
	if result.Outcome() != pkgTool.ResultSuccess || !strings.Contains(result.Content(), "laravel/framework") {
		t.Fatalf("Laravel WorkspaceProvider path failed: %#v", result)
	}
}

func workspaceRuntime(t *testing.T, root string) (*Runtime, *WorkspaceRegistry, pkgTool.RunID) {
	t.Helper()
	workspace, err := pkgContext.NewWorkspace("workspace", root, pkgContext.WorkspaceOptions{
		Source: pkgContext.SourceFilesystem, Policy: pkgContext.DefaultScanPolicy(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return runtimeForWorkspace(t, workspace)
}

func runtimeForWorkspace(t *testing.T, workspace pkgContext.Workspace) (*Runtime, *WorkspaceRegistry, pkgTool.RunID) {
	t.Helper()
	registry := NewWorkspaceRegistry()
	run := pkgTool.RunID("run-workspace")
	if err := registry.Bind(run, workspace); err != nil {
		t.Fatal(err)
	}
	runtime := NewRuntime()
	tools, err := NewWorkspaceTools(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range tools {
		if err := runtime.Register(candidate); err != nil {
			t.Fatal(err)
		}
	}
	allow, _ := pkgTool.NewDecision(pkgTool.DecisionAllow, "allowed", "", pkgTool.GrantRun)
	if err := runtime.RegisterPolicy(&workspacePolicy{id: "policy.test", decision: allow}); err != nil {
		t.Fatal(err)
	}
	return runtime, registry, run
}

type workspacePolicy struct {
	id       pkgTool.PolicyID
	decision pkgTool.Decision
}

func (policy *workspacePolicy) ID() pkgTool.PolicyID { return policy.id }
func (policy *workspacePolicy) Decide(context.Context, pkgTool.PermissionRequest) (pkgTool.Decision, error) {
	return policy.decision, nil
}

func invokeWorkspace(t *testing.T, runtime *Runtime, run pkgTool.RunID, id pkgTool.ID, arguments json.RawMessage) pkgTool.Result {
	t.Helper()
	invocation, err := pkgTool.NewInvocation(id, pkgTool.CallID("call-"+strings.ReplaceAll(string(id), ".", "-")), run, arguments)
	if err != nil {
		t.Fatal(err)
	}
	execution, err := pkgTool.NewExecutionRequest(invocation, "policy.test", nil, workspaceExecutionLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, err := runtime.Invoke(context.Background(), execution)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func workspaceExecutionLimits() pkgTool.ExecutionLimits {
	return pkgTool.ExecutionLimits{MaxDuration: time.Second, MaxOutputBytes: 1 << 20, MaxItems: 10_000}
}

func resultDigest(t *testing.T, content string) string {
	t.Helper()
	var payload struct {
		Digest string `json:"digest"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Digest
}
