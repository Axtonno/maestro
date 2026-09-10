package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type setupCLIProvider struct {
	*cliProvider
	pulls []string
}

func (provider *setupCLIProvider) PullModel(_ context.Context, request pkgProvider.ModelPullRequest) (pkgProvider.ModelPullStream, error) {
	provider.pulls = append(provider.pulls, request.Model)
	digest := ""
	switch request.Model {
	case productconfig.QualifiedDirectChatModel:
		digest = productconfig.QualifiedDirectChatDigest
	case productconfig.QualifiedMutationModel:
		digest = productconfig.QualifiedMutationDigest
	default:
		return nil, errors.New("unexpected model pull")
	}
	provider.discovered = append(provider.discovered, pkgProvider.ModelInfo{Model: pkgProvider.Model{ID: request.Model}, Digest: digest})
	return &setupPullStream{}, nil
}

type setupPullStream struct{ completed bool }

func (stream *setupPullStream) Recv() (pkgProvider.ModelPullProgress, error) {
	if stream.completed {
		return pkgProvider.ModelPullProgress{}, io.EOF
	}
	stream.completed = true
	return pkgProvider.ModelPullProgress{Stage: pkgProvider.ModelPullStageCompleted}, nil
}

func (*setupPullStream) Close() error { return nil }

func TestSetupCreatesPrivateIdempotentConfiguration(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "maestro", "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{
		id: "ollama",
		discovered: []pkgProvider.ModelInfo{
			{Model: pkgProvider.Model{ID: productconfig.QualifiedDirectChatModel}, Digest: productconfig.QualifiedDirectChatDigest},
			{Model: pkgProvider.Model{ID: productconfig.QualifiedMutationModel}, Digest: productconfig.QualifiedMutationDigest},
		},
	}}
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return false }

	for attempt := 0; attempt < 2; attempt++ {
		var stdout, stderr bytes.Buffer
		code := runWithIO(
			[]string{"setup", "--config", configPath, "--workspace", workspace},
			strings.NewReader(""), &stdout, &stderr, dependencies,
		)
		if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "✓ Ollama detected and responding") || !strings.Contains(stdout.String(), "maestro chat \"Come puoi aiutarmi?\"") {
			t.Fatalf("attempt=%d code=%d stdout=%q stderr=%q", attempt, code, stdout.String(), stderr.String())
		}
		wantConfigMessage := "Configuration created"
		if attempt == 1 {
			wantConfigMessage = "Configuration already valid"
		}
		if !strings.Contains(stdout.String(), wantConfigMessage) {
			t.Fatalf("attempt=%d output=%q", attempt, stdout.String())
		}
	}

	config, err := productconfig.LoadMutation(configPath)
	if err != nil || config.Workspace.Root != workspace || config.ProductProfile() != productconfig.ProductProfileRecommended {
		t.Fatalf("config=%#v err=%v", config, err)
	}
	info, err := os.Stat(configPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%v err=%v", info.Mode(), err)
	}
}

func TestSetupReportsMissingModelsWithoutPullInNonInteractiveUse(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama", discovered: []pkgProvider.ModelInfo{}}}
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return false }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--config", configPath, "--workspace", workspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 4 || stderr.Len() != 0 || !strings.Contains(stdout.String(), productconfig.QualifiedDirectChatModel+" is missing") || !strings.Contains(stdout.String(), "maestro setup --pull") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := productconfig.LoadMutation(configPath); err != nil {
		t.Fatalf("configuration was not preserved for retry: %v", err)
	}
}

func TestSetupPullsAndVerifiesMissingModelsWhenExplicitlyAuthorized(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama", discovered: []pkgProvider.ModelInfo{}}}
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return false }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--pull", "--config", configPath, "--workspace", workspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 0 || len(provider.pulls) != 2 || !strings.Contains(stdout.String(), "✓ "+productconfig.QualifiedDirectChatModel+" available for chat") || !strings.Contains(stdout.String(), "✓ Workspace configured") || !strings.Contains(stderr.String(), "stage=completed") {
		t.Fatalf("code=%d pulls=%v stdout=%q stderr=%q", code, provider.pulls, stdout.String(), stderr.String())
	}
}

func TestSetupPromptsBeforePullingMissingModels(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama", discovered: []pkgProvider.ModelInfo{}}}
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return true }
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--config", configPath, "--workspace", workspace},
		strings.NewReader("y\n"), &stdout, &stderr, dependencies,
	)
	if code != 0 || len(provider.pulls) != 2 || !strings.Contains(stderr.String(), "Download the missing recommended models") {
		t.Fatalf("code=%d pulls=%v stdout=%q stderr=%q", code, provider.pulls, stdout.String(), stderr.String())
	}
}

func TestSetupReportsUnreachableOllamaWithoutRemovingConfiguration(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama", discoverErr: errors.New("connection refused")}}
	dependencies := cliTestDependencies(provider)
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--config", configPath, "--workspace", workspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 4 || !strings.Contains(stderr.String(), "ollama_unreachable") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := productconfig.LoadMutation(configPath); err != nil {
		t.Fatalf("configuration was not preserved for retry: %v", err)
	}
}

func TestSetupRejectsUnexpectedModelDigestWithoutPull(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	provider := &setupCLIProvider{cliProvider: &cliProvider{
		id: "ollama",
		discovered: []pkgProvider.ModelInfo{{
			Model:  pkgProvider.Model{ID: productconfig.QualifiedDirectChatModel},
			Digest: strings.Repeat("f", 64),
		}},
	}}
	dependencies := cliTestDependencies(provider)
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--pull", "--config", configPath, "--workspace", workspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	if code != 4 || len(provider.pulls) != 0 || !strings.Contains(stderr.String(), "model_digest_mismatch") {
		t.Fatalf("code=%d pulls=%v stdout=%q stderr=%q", code, provider.pulls, stdout.String(), stderr.String())
	}
}

func TestSetupRejectsExistingInvalidConfigurationWithoutOverwrite(t *testing.T) {
	workspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	original := []byte("version: 4\nunknown: true\n")
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama"}}
	dependencies := cliTestDependencies(provider)
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--config", configPath, "--workspace", workspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	got, err := os.ReadFile(configPath)
	if code != 2 || !strings.Contains(stderr.String(), "configuration_invalid") || err != nil || !bytes.Equal(got, original) {
		t.Fatalf("code=%d stdout=%q stderr=%q content=%q err=%v", code, stdout.String(), stderr.String(), got, err)
	}
}

func TestSetupRejectsSilentReuseForAnotherWorkspace(t *testing.T) {
	firstWorkspace := t.TempDir()
	secondWorkspace := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if _, _, err := ensureSetupConfig(configPath, firstWorkspace); err != nil {
		t.Fatal(err)
	}
	provider := &setupCLIProvider{cliProvider: &cliProvider{id: "ollama"}}
	dependencies := cliTestDependencies(provider)
	var stdout, stderr bytes.Buffer
	code := runWithIO(
		[]string{"setup", "--config", configPath, "--workspace", secondWorkspace},
		strings.NewReader(""), &stdout, &stderr, dependencies,
	)
	config, err := productconfig.LoadMutation(configPath)
	if code != 2 || !strings.Contains(stderr.String(), "workspace_mismatch") || err != nil || config.Workspace.Root != firstWorkspace {
		t.Fatalf("code=%d config=%#v stdout=%q stderr=%q err=%v", code, config, stdout.String(), stderr.String(), err)
	}
}
