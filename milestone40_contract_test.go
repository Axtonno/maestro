package maestro_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMilestone40PrototypePreservesCLIControlBoundary(t *testing.T) {
	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	var manifest struct {
		Main             string   `json:"main"`
		ExtensionKind    []string `json:"extensionKind"`
		ActivationEvents []string `json:"activationEvents"`
		Capabilities     struct {
			Untrusted struct {
				Supported *bool `json:"supported"`
			} `json:"untrustedWorkspaces"`
			Virtual struct {
				Supported *bool `json:"supported"`
			} `json:"virtualWorkspaces"`
		} `json:"capabilities"`
		Contributes struct {
			Commands []struct {
				Command string `json:"command"`
			} `json:"commands"`
		} `json:"contributes"`
	}
	if err := json.Unmarshal([]byte(read("editors/vscode-maestro/package.json")), &manifest); err != nil {
		t.Fatal(err)
	}
	wantCommands := []string{
		"maestro.askActiveFile",
		"maestro.replaceSelection",
		"maestro.doctor",
		"maestro.version",
	}
	if manifest.Main != "./extension.js" || len(manifest.ExtensionKind) != 1 || manifest.ExtensionKind[0] != "workspace" || manifest.Capabilities.Untrusted.Supported == nil || *manifest.Capabilities.Untrusted.Supported || manifest.Capabilities.Virtual.Supported == nil || *manifest.Capabilities.Virtual.Supported || len(manifest.Contributes.Commands) != len(wantCommands) || len(manifest.ActivationEvents) != len(wantCommands) {
		t.Fatalf("invalid M40 extension manifest: %#v", manifest)
	}
	for index, want := range wantCommands {
		if manifest.Contributes.Commands[index].Command != want || manifest.ActivationEvents[index] != "onCommand:"+want {
			t.Fatalf("M40 command %d is not registered consistently", index)
		}
	}

	var launch struct {
		Configurations []struct {
			Type    string   `json:"type"`
			Request string   `json:"request"`
			Args    []string `json:"args"`
		} `json:"configurations"`
	}
	if err := json.Unmarshal([]byte(read("editors/vscode-maestro/.vscode/launch.json")), &launch); err != nil {
		t.Fatal(err)
	}
	if len(launch.Configurations) != 1 || launch.Configurations[0].Type != "extensionHost" || launch.Configurations[0].Request != "launch" || len(launch.Configurations[0].Args) != 1 || launch.Configurations[0].Args[0] != "--extensionDevelopmentPath=${workspaceFolder}" {
		t.Fatalf("invalid M40 development host launch: %#v", launch)
	}

	var matrix struct {
		Status      string
		Verdict     string
		ClaimPolicy struct {
			PublicUnchanged     bool `yaml:"public_v0_5_0_unchanged"`
			MarketplaceAllowed  bool `yaml:"marketplace_publication_authorized"`
			ReleaseAllowed      bool `yaml:"release_packaging_authorized"`
			NewRuntimeAuthority bool `yaml:"new_runtime_authority"`
		} `yaml:"claim_policy"`
		SecurityGates struct {
			TTYRequired       bool `yaml:"cli_tty_approval_required"`
			AutoApproval      bool `yaml:"extension_auto_approval"`
			ExtensionWrites   int  `yaml:"extension_file_writes"`
			SingleSelection   bool `yaml:"mutation_single_selection_required"`
			CompleteLines     bool `yaml:"mutation_complete_lines_required"`
			MutationPHPInApp  bool `yaml:"mutation_php_below_app_required"`
			ShellTokensQuoted bool `yaml:"shell_tokens_quoted"`
			OptionTerminator  bool `yaml:"option_terminator_before_user_prompt"`
			WorkspaceTrust    bool `yaml:"workspace_trust_required"`
		} `yaml:"security_gates"`
		OfflineCases map[string]string `yaml:"offline_cases"`
		LiveCases    map[string]string `yaml:"live_cases"`
	}
	if err := yaml.Unmarshal([]byte(read("docs/milestone-40-vscode-prototype-matrix.yaml")), &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.Status != "validation_pending" || matrix.Verdict != "vscode_prototype_validation_pending" || !matrix.ClaimPolicy.PublicUnchanged || matrix.ClaimPolicy.MarketplaceAllowed || matrix.ClaimPolicy.ReleaseAllowed || matrix.ClaimPolicy.NewRuntimeAuthority {
		t.Fatalf("invalid M40 claim boundary: %#v", matrix)
	}
	if !matrix.SecurityGates.TTYRequired || matrix.SecurityGates.AutoApproval || matrix.SecurityGates.ExtensionWrites != 0 || !matrix.SecurityGates.SingleSelection || !matrix.SecurityGates.CompleteLines || !matrix.SecurityGates.MutationPHPInApp || !matrix.SecurityGates.ShellTokensQuoted || !matrix.SecurityGates.OptionTerminator || !matrix.SecurityGates.WorkspaceTrust {
		t.Fatalf("invalid M40 security gates: %#v", matrix.SecurityGates)
	}
	wantOffline := map[string]string{
		"V01_manifest_and_activation":            "passed",
		"V02_command_registration":               "passed",
		"V03_shell_token_escaping":               "not_run",
		"V04_workspace_containment":              "not_run",
		"V05_selection_coordinate_mapping":       "not_run",
		"V06_mutation_scope_narrowing":           "not_run",
		"V07_dirty_and_multiselection_rejection": "passed",
		"V08_no_write_or_auto_approval_api":      "passed",
	}
	for name, want := range wantOffline {
		if matrix.OfflineCases[name] != want {
			t.Fatalf("M40 offline case %s is %s, want %s", name, matrix.OfflineCases[name], want)
		}
	}
	for name, status := range matrix.LiveCases {
		if status != "not_run" {
			t.Fatalf("M40 live case %s must remain not_run, got %s", name, status)
		}
	}

	extension := read("editors/vscode-maestro/extension.js")
	builder := read("editors/vscode-maestro/command-builder.js")
	for _, required := range []string{
		"createTerminal({ name, cwd: folder.uri, shellPath: '/bin/sh' })",
		"terminal.sendText(command, true)",
		"vscode.workspace.isTrusted",
		"editor.document.isDirty",
		"editor.selections.length !== 1",
		"requireWholeLineSelection(target.editor)",
		"mutationLogicalPath",
		"inclusiveSelectedLines",
	} {
		if !strings.Contains(extension, required) {
			t.Fatalf("extension is missing control %q", required)
		}
	}
	for _, forbidden := range []string{
		"workspace.fs.writeFile",
		"new vscode.WorkspaceEdit",
		"editor.edit(",
		"child_process",
		"spawn(",
		"exec(",
	} {
		if strings.Contains(extension, forbidden) || strings.Contains(builder, forbidden) {
			t.Fatalf("prototype contains forbidden authority %q", forbidden)
		}
	}
	for _, required := range []string{
		"FORBIDDEN_TOKEN_BYTES",
		"value.replace(/'/g",
		"logical.startsWith('../')",
		"logical.startsWith('app/')",
		"path.posix.extname(logical).toLowerCase() !== '.php'",
	} {
		if !strings.Contains(builder, required) {
			t.Fatalf("command builder is missing control %q", required)
		}
	}
}
