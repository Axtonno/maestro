package maestro_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
)

func TestReleaseV2CompatibilityAndCurrentV3ContractAreFrozen(t *testing.T) {
	v2Path := filepath.FromSlash("configs/maestro.v0.3.0-compat.yaml")
	v2Bytes, err := os.ReadFile(v2Path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(v2Bytes)); got != "1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee" {
		t.Fatalf("v0.3.0 public configuration drifted: %s", got)
	}
	v2Before := append([]byte(nil), v2Bytes...)
	v2, err := productconfig.LoadChat(v2Path)
	if err != nil {
		t.Fatalf("load v2 compatibility profile: %v", err)
	}
	v2Chat, ok := v2.ChatProfile()
	if !ok || v2.Version != productconfig.CandidateVersion {
		t.Fatalf("v2 profile was not preserved: %#v", v2)
	}
	if v2Chat.NumPredict != 0 || v2Chat.Residency.Duration != 0 ||
		v2Chat.GenerationOptions().MaxTokens != 0 {
		t.Fatalf("v2 gained implicit generation settings: %#v", v2Chat)
	}
	v2After, err := os.ReadFile(v2Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(v2Before) != string(v2After) {
		t.Fatal("loading v2 rewrote the public configuration")
	}

	v3, err := productconfig.LoadChat(filepath.FromSlash("configs/maestro.chat.example.yaml"))
	if err != nil {
		t.Fatalf("load v3 release profile: %v", err)
	}
	v3Chat, ok := v3.ChatProfile()
	if !ok || v3.Version != productconfig.QualificationVersion ||
		v3Chat.NumPredict != 1024 || v3Chat.Residency.Duration != 5*time.Minute ||
		v3Chat.GenerationOptions().MaxTokens != 1024 {
		t.Fatalf("v3 operational contract drifted: %#v", v3Chat)
	}
}

func TestMilestone23UnknownAndInvalidConfigurationsFailClosed(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		content string
		kind    productconfig.DiagnosticKind
		field   string
	}{
		{name: "unknown version", content: "version: 99\n", kind: productconfig.DiagnosticInvalidValue, field: "version"},
		{name: "unknown field", content: "version: 3\nsecret-sentinel: value-sentinel\n", kind: productconfig.DiagnosticUnknownField},
		{name: "malformed", content: "version: [\npath-sentinel\n", kind: productconfig.DiagnosticYAMLInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config-path-sentinel.yaml")
			if err := os.WriteFile(path, []byte(testCase.content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := productconfig.LoadChat(path)
			if !errors.Is(err, productconfig.ErrInvalid) {
				t.Fatalf("configuration did not fail closed: %v", err)
			}
			diagnostic := productconfig.Diagnose(err)
			if diagnostic.Kind != testCase.kind || diagnostic.Field != testCase.field {
				t.Fatalf("diagnostic=%#v", diagnostic)
			}
			if containsAny(diagnostic.Field, "value-sentinel", "path-sentinel", path) {
				t.Fatalf("public diagnostic leaked sensitive input: %#v", diagnostic)
			}
		})
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && len(value) >= len(needle) {
			for index := 0; index+len(needle) <= len(value); index++ {
				if value[index:index+len(needle)] == needle {
					return true
				}
			}
		}
	}
	return false
}
