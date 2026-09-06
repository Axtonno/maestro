package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestWorkspaceReplaceRequiresTTYBeforeProviderIO(t *testing.T) {
	config, file := newCLIV4Config(t)
	provider := newCLIMutationProvider(`{"decision":"propose","new_text":"$workers = 8;"}`)
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return false }
	var stdout, stderr bytes.Buffer
	code := runWithIO([]string{"workspace", "replace", "--config", config, "--file", "app/Worker.php", "--lines", "2:2", "set workers to 8"}, strings.NewReader("o\n"), &stdout, &stderr, dependencies)
	if code != 3 || stdout.Len() != 0 || stderr.String() != "mutation failed: tty_required\n" || len(provider.requests) != 0 || len(provider.unloaded) != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	assertFile(t, file, "<?php\n$workers = 4;\nreturn $workers;\n")
}

func TestWorkspaceReplaceRejectsRangeOutsideAppAndSymlinkBeforeGeneration(t *testing.T) {
	for _, testCase := range []struct {
		name, file, lines string
		symlink           bool
	}{
		{"range", "app/Worker.php", "0:2", false},
		{"outside app", "Worker.php", "1:1", false},
		{"symlink", "app/Link.php", "1:1", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			config, file := newCLIV4Config(t)
			if testCase.name == "outside app" {
				if err := os.WriteFile(filepath.Join(filepath.Dir(filepath.Dir(file)), "Worker.php"), []byte("<?php\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if testCase.symlink {
				if err := os.Symlink(file, filepath.Join(filepath.Dir(file), "Link.php")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			provider := newCLIMutationProvider(`{"decision":"abstain"}`)
			dependencies := cliTestDependencies(provider)
			dependencies.isTerminal = func(io.Reader) bool { return true }
			var stdout, stderr bytes.Buffer
			code := runWithIO([]string{"workspace", "replace", "--config", config, "--file", testCase.file, "--lines", testCase.lines, "change it"}, strings.NewReader("d\n"), &stdout, &stderr, dependencies)
			if code != 2 || len(provider.requests) != 0 || provider.inspectCalls != 0 || len(provider.unloaded) != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q provider=%#v", code, stdout.String(), stderr.String(), provider)
			}
		})
	}
}

func TestWorkspaceReplaceAllowDenyStaleAndAbstain(t *testing.T) {
	tests := []struct {
		name, input, raw, wantTerminal string
		code                           int
		stale                          bool
	}{
		{"allow", "o\n", `{"decision":"propose","new_text":"$workers = 8;"}`, "terminal\tapplied", 0, false},
		{"deny", "d\n", `{"decision":"propose","new_text":"$workers = 8;"}`, "mutation failed: approval_rejected", 3, false},
		{"empty deny", "\n", `{"decision":"propose","new_text":"$workers = 8;"}`, "mutation failed: approval_rejected", 3, false},
		{"stale", "o\n", `{"decision":"propose","new_text":"$workers = 8;"}`, "mutation failed: stale_source", 3, true},
		{"abstain", "o\n", `{"decision":"abstain"}`, "mutation failed: insufficient_information", 3, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			config, file := newCLIV4Config(t)
			provider := newCLIMutationProvider(testCase.raw)
			dependencies := cliTestDependencies(provider)
			dependencies.isTerminal = func(io.Reader) bool { return true }
			if testCase.stale {
				dependencies.mutationAfterPreview = func() {
					if err := os.WriteFile(file, []byte("<?php\n$concurrent = true;\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			var stdout, stderr bytes.Buffer
			code := runWithIO([]string{"workspace", "replace", "--config", config, "--file", "app/Worker.php", "--lines", "2:2", "set workers to 8"}, strings.NewReader(testCase.input), &stdout, &stderr, dependencies)
			combined := stdout.String() + stderr.String()
			if code != testCase.code || !strings.Contains(combined, testCase.wantTerminal) {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if testCase.name != "abstain" && (!strings.Contains(stderr.String(), "single_file: true") || !strings.Contains(stderr.String(), "preview (text/x-diff):")) {
				t.Fatalf("preview missing: %q", stderr.String())
			}
			want := "<?php\n$workers = 4;\nreturn $workers;\n"
			if testCase.name == "allow" {
				want = "<?php\n$workers = 8;\nreturn $workers;\n"
			}
			if testCase.stale {
				want = "<?php\n$concurrent = true;\n"
			}
			assertFile(t, file, want)
		})
	}
}

func TestMutationDoctorChecksSeparatedIdentityAndTTYWithoutCompletion(t *testing.T) {
	config, _ := newCLIV4Config(t)
	provider := newCLIMutationProvider(`{"decision":"abstain"}`)
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return true }
	var stdout, stderr bytes.Buffer
	code := runWithIO([]string{"doctor", "--mode", "mutation", "--config", config}, strings.NewReader(""), &stdout, &stderr, dependencies)
	for _, expected := range []string{"pass\tconfiguration\tschema_v4_profiles_separated", "pass\tdirect_chat_model\tqualified_model_digest", "pass\tcontrolled_mutation_model\tqualified_model_digest", "pass\tmutation_prompt\tqualified_prompt_digest", "pass\tmutation_schema\tqualified_schema_digest", "pass\ttty\tinteractive_terminal", "pass\tcapability\thost_bound_mutation_available"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("missing %q in %q", expected, stdout.String())
		}
	}
	if code != 0 || stderr.Len() != 0 || len(provider.requests) != 0 || len(provider.unloaded) != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestDoctorAllAggregatesChatAndMutationWithoutCompletion(t *testing.T) {
	config, _ := newCLIV4Config(t)
	provider := newCLIMutationProvider(`{"decision":"abstain"}`)
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return true }
	var stdout, stderr bytes.Buffer
	code := runWithIO([]string{"doctor", "--mode", "all", "--config", config}, strings.NewReader(""), &stdout, &stderr, dependencies)
	for _, expected := range []string{
		"pass\tchat_config\tschema_v4_chat_valid",
		"pass\tchat_model\tcompletion_capabilities_available",
		"pass\tmutation_configuration\tschema_v4_profiles_separated",
		"pass\tmutation_direct_chat_model\tqualified_model_digest",
		"pass\tmutation_controlled_mutation_model\tqualified_model_digest",
		"pass\tmutation_capability\thost_bound_mutation_available",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("missing %q in %q", expected, stdout.String())
		}
	}
	if code != 0 || stderr.Len() != 0 || len(provider.requests) != 0 || len(provider.unloaded) != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func newCLIMutationProvider(raw string) *cliProvider {
	return &cliProvider{id: "ollama", responses: []pkgProvider.CompletionResponse{{Model: productconfig.QualifiedMutationModel, Message: pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: raw}, FinishReason: pkgProvider.FinishReasonStop}}, discovered: []pkgProvider.ModelInfo{{Model: pkgProvider.Model{ID: productconfig.QualifiedDirectChatModel}, Digest: productconfig.QualifiedDirectChatDigest, State: pkgProvider.ModelStateLoaded}, {Model: pkgProvider.Model{ID: productconfig.QualifiedMutationModel}, Digest: productconfig.QualifiedMutationDigest, State: pkgProvider.ModelStateAvailable}}}
}

func newCLIV4Config(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	app := filepath.Join(root, "app")
	if err := os.Mkdir(app, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(app, "Worker.php")
	if err := os.WriteFile(file, []byte("<?php\n$workers = 4;\nreturn $workers;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf(`version: 4
provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 5m
  api_key_env: ""
workspace:
  id: laravel
  root: %s
  framework: laravel
direct_chat:
  model: %s
  digest: %s
  timeout: 5m
  streaming: false
  num_ctx: 4096
  num_predict: 1024
  thinking: "false"
  residency: 5m
  max_file_bytes: 1048576
  max_output_bytes: 1048576
controlled_mutation:
  enabled: true
  model: %s
  digest: %s
  timeout: 5m
  num_ctx: 4096
  num_predict: 1024
  thinking: "false"
  residency: 5m
  prompt: %s
  prompt_sha256: %s
  schema: %s
  schema_sha256: %s
  max_output_bytes: 1048576
`, root, productconfig.QualifiedDirectChatModel, productconfig.QualifiedDirectChatDigest, productconfig.QualifiedMutationModel, productconfig.QualifiedMutationDigest, productconfig.MutationPromptID, productconfig.MutationPromptSHA256, productconfig.MutationSchemaID, productconfig.MutationSchemaSHA256)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, file
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file=%q err=%v", got, err)
	}
}
