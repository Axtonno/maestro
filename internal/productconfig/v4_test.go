package productconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV4LoadsSeparatedQualifiedProfiles(t *testing.T) {
	path := writeV4(t, validV4(t.TempDir()))
	config, err := LoadMutation(path)
	if err != nil {
		t.Fatal(err)
	}
	chat, ok := config.ChatProfile()
	if !ok || config.Version != ProductizationVersion || config.ProductProfile() != ProductProfileRecommended || chat.Model != QualifiedDirectChatModel || chat.Digest != QualifiedDirectChatDigest || config.ControlledMutation.Model != QualifiedMutationModel || config.ControlledMutation.Digest != QualifiedMutationDigest {
		t.Fatalf("unexpected v4 profile: %#v", config)
	}
	if _, err := LoadChat(path); err != nil {
		t.Fatalf("chat rejected v4: %v", err)
	}
}

func TestV4RejectsIdentityFallbackAndContractDrift(t *testing.T) {
	valid := validV4(t.TempDir())
	tests := []struct {
		name   string
		mutate func(string) string
	}{
		{"mutation disabled", func(v string) string { return strings.Replace(v, "enabled: true", "enabled: false", 1) }},
		{"chat digest", func(v string) string {
			return strings.Replace(v, QualifiedDirectChatDigest, strings.Repeat("0", 64), 1)
		}},
		{"mutation digest", func(v string) string { return strings.Replace(v, QualifiedMutationDigest, strings.Repeat("1", 64), 1) }},
		{"same model", func(v string) string { return strings.Replace(v, QualifiedMutationModel, QualifiedDirectChatModel, 1) }},
		{"prompt", func(v string) string { return strings.Replace(v, MutationPromptID, "other-prompt", 1) }},
		{"schema", func(v string) string { return strings.Replace(v, MutationSchemaID, "other-schema", 1) }},
		{"fallback field", func(v string) string { return v + "fallback: direct_chat\n" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := LoadMutation(writeV4(t, testCase.mutate(valid)))
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("drift accepted: %v", err)
			}
		})
	}
}

func writeV4(t *testing.T, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func validV4(root string) string {
	return fmt.Sprintf(`version: 4
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
`, root, QualifiedDirectChatModel, QualifiedDirectChatDigest, QualifiedMutationModel, QualifiedMutationDigest, MutationPromptID, MutationPromptSHA256, MutationSchemaID, MutationSchemaSHA256)
}

func singleModelV4(root string) string {
	value := validV4(root)
	value = strings.Replace(value, QualifiedDirectChatModel, SingleModelEvaluationModel, 1)
	return strings.Replace(value, QualifiedDirectChatDigest, SingleModelEvaluationDigest, 1)
}
