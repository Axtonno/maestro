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
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
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
	digest := resultDigest(t, written.Content())
	patchedArguments, _ := json.Marshal(map[string]string{
		"path": "src/new.go", "old": "package new", "new": "package changed", "expected_digest": digest,
	})
	patched := invokeWorkspace(t, runtime, run, WorkspacePatchID, patchedArguments)
	if patched.Outcome() != pkgTool.ResultSuccess {
		t.Fatalf("patch failed: %#v", patched)
	}
	content, err := os.ReadFile(filepath.Join(root, "src", "new.go"))
	if err != nil || string(content) != "package changed\n" {
		t.Fatalf("unexpected patched file: %q %v", content, err)
	}

	conflict := invokeWorkspace(t, runtime, run, WorkspaceWriteID, json.RawMessage(`{"path":"src/new.go","content":"overwrite","expected_digest":"absent"}`))
	if conflict.Outcome() != pkgTool.ResultFailed || conflict.Reason() != "precondition_failed" {
		t.Fatalf("expected content precondition conflict: %#v", conflict)
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
