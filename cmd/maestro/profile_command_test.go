package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestProfileReportsEffectiveQualifiedIdentityWithoutProviderIO(t *testing.T) {
	configPath, _ := newCLIV4Config(t)
	workspace := t.TempDir()
	provider := newCLIMutationProvider(`{"decision":"abstain"}`)
	dependencies := cliTestDependencies(provider)
	dependencies.workingDirectory = func() (string, error) { return workspace, nil }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"profile", "--config", configPath, "--workspace-current"},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	var identity profileIdentity
	if err := json.Unmarshal(stdout.Bytes(), &identity); err != nil {
		t.Fatalf("decode identity: %v", err)
	}
	if identity.SchemaVersion != 1 || identity.Profile != "recommended" ||
		identity.Provider != "ollama" || identity.ChatModel != productconfig.QualifiedDirectChatModel ||
		identity.MutationModel != productconfig.QualifiedMutationModel {
		t.Fatalf("unexpected identity: %#v", identity)
	}
	if len(provider.requests) != 0 || provider.inspectCalls != 0 {
		t.Fatalf("profile performed provider IO: %#v", provider)
	}
	if strings.Contains(stdout.String(), workspace) || strings.Contains(stdout.String(), configPath) {
		t.Fatalf("identity leaked a physical path: %q", stdout.String())
	}
}

func TestChatWorkspaceCurrentUsesProcessWorkingDirectory(t *testing.T) {
	configPath, _ := newCLIInteractionConfig(t, false)
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "routes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "routes", "api.php"), []byte("<?php\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{
		Model: "chat-model", Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: "Current workspace used."},
		FinishReason: pkgProvider.FinishReasonStop,
	}}}
	dependencies := cliTestDependencies(provider)
	dependencies.workingDirectory = func() (string, error) { return workspace, nil }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"chat", "--config", configPath, "--workspace-current", "--file", "routes/api.php", "Question"},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "result\nCurrent workspace used.") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestWorkspaceCurrentFailsClosedWhenWorkingDirectoryIsUnavailable(t *testing.T) {
	configPath, _ := newCLIV4Config(t)
	dependencies := cliTestDependencies(newCLIMutationProvider(`{"decision":"abstain"}`))
	dependencies.workingDirectory = func() (string, error) { return "", os.ErrNotExist }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"profile", "--config", configPath, "--workspace-current"},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "profile failed: invalid_request") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
