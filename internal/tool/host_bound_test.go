package tool

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/mutation"
	ce "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pt "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestHostBoundPhysicalSpliceAndStale(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("atomic commit requires Linux")
	}
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "apply", true: "stale"}[stale], func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "app"), 0700); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(root, "app", "Test.php")
			before := "<?php\r\n// à\r\n// à\r\n"
			if err := os.WriteFile(file, []byte(before), 0600); err != nil {
				t.Fatal(err)
			}
			w, err := ce.NewWorkspace("host-test", root, ce.WorkspaceOptions{Source: ce.SourceFilesystem, Policy: ce.DefaultScanPolicy()})
			if err != nil {
				t.Fatal(err)
			}
			registry := NewWorkspaceRegistry()
			if err := registry.Bind("run", w); err != nil {
				t.Fatal(err)
			}
			h, err := NewHostBoundMutation(registry)
			if err != nil {
				t.Fatal(err)
			}
			s, err := h.Capture(context.Background(), "run", []string{"app/Test.php"}, 3, 3)
			if err != nil {
				t.Fatal(err)
			}
			p, err := h.Prepare(context.Background(), s, "call", []byte(`{"decision":"propose","new_text":"// β"}`))
			if err != nil {
				t.Fatal(err)
			}
			var args preparedReplaceArguments
			if err := json.Unmarshal(p.Arguments(), &args); err != nil {
				t.Fatal(err)
			}
			preview, _ := p.Preview()
			if args.Fingerprint != s.Target().Fingerprint("// β", preview.Body()) {
				t.Fatal("fingerprint mismatch")
			}
			want := "<?php\r\n// à\r\n// β\r\n"
			if stale {
				want = before + "// concurrent\r\n"
				if err := os.WriteFile(file, []byte(want), 0600); err != nil {
					t.Fatal(err)
				}
			}
			result, err := h.Execute(context.Background(), p)
			if err != nil {
				t.Fatal(err)
			}
			if (result.Effect() == pt.EffectApplied) == stale {
				t.Fatal("incorrect effect")
			}
			got, err := os.ReadFile(file)
			if err != nil || string(got) != want {
				t.Fatalf("bytes %q %v", got, err)
			}
			if _, err := h.Capture(context.Background(), "run", []string{"app/Test.php", "app/Other.php"}, 1, 1); !errors.Is(err, ErrRequestOutOfScope) {
				t.Fatal(err)
			}
			if _, err := h.Capture(context.Background(), "run", []string{"app/secrets/Test.php"}, 1, 1); !errors.Is(err, mutation.ErrSensitiveTarget) {
				t.Fatal(err)
			}
			if err := os.Symlink(file, filepath.Join(root, "app", "Link.php")); err != nil {
				t.Fatal(err)
			}
			if _, err := h.Capture(context.Background(), "run", []string{"app/Link.php"}, 1, 1); err == nil {
				t.Fatal("symlink accepted")
			}
		})
	}
}
