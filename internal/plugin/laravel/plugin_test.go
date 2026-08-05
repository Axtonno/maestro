package laravel

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestNewValidatesAndNormalizesWorkspaceRoot(t *testing.T) {
	if _, err := New(""); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
	if _, err := New("   "); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}

	workspace := t.TempDir()
	plugin, err := New(workspace)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}

	if !filepath.IsAbs(plugin.Root()) {
		t.Fatalf("expected absolute root, got %q", plugin.Root())
	}
	if plugin.FrameworkVersion() != "" {
		t.Fatalf("expected no version before initialization, got %q", plugin.FrameworkVersion())
	}
}

func TestPluginDeclaresIdentityAndCompatibility(t *testing.T) {
	plugin, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}

	metadata := plugin.Metadata()
	if metadata.ID != ID || metadata.Name != "Laravel" || metadata.Version != Version {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if got, want := metadata.Capabilities, []pkgRuntime.Capability{
		pkgRuntime.CapabilityInitialize,
		pkgRuntime.CapabilityHealth,
	}; !equalCapabilities(got, want) {
		t.Fatalf("expected capabilities %v, got %v", want, got)
	}
	if plugin.Manifest().RuntimeAPIVersion != pkgPlugin.RuntimeAPIVersion {
		t.Fatalf("unexpected plugin manifest: %#v", plugin.Manifest())
	}
}

func TestPluginDetectsLaravelWorkspace(t *testing.T) {
	workspace := newLaravelWorkspace(t, "^12.0")
	plugin, err := New(workspace)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}

	if err := plugin.(pkgRuntime.Initializer).Initialize(nil); err != nil {
		t.Fatalf("initialize plugin: %v", err)
	}
	if got := plugin.FrameworkVersion(); got != "^12.0" {
		t.Fatalf("expected framework version ^12.0, got %q", got)
	}
	if err := plugin.(pkgRuntime.HealthChecker).Health(nil); err != nil {
		t.Fatalf("check plugin health: %v", err)
	}

	if err := os.Remove(filepath.Join(workspace, "artisan")); err != nil {
		t.Fatalf("remove artisan: %v", err)
	}
	if err := plugin.(pkgRuntime.HealthChecker).Health(nil); !errors.Is(
		err,
		ErrNotDetected,
	) {
		t.Fatalf("expected ErrNotDetected after artisan removal, got %v", err)
	}
}

func TestPluginRejectsNonLaravelWorkspace(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, string)
		want    error
	}{
		{
			name: "missing artisan",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "composer.json"), `{"require":{"laravel/framework":"^12.0"}}`)
			},
			want: ErrNotDetected,
		},
		{
			name: "artisan directory",
			prepare: func(t *testing.T, root string) {
				if err := os.Mkdir(filepath.Join(root, "artisan"), 0o755); err != nil {
					t.Fatalf("create artisan directory: %v", err)
				}
			},
			want: ErrNotDetected,
		},
		{
			name: "missing composer manifest",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
			},
			want: ErrNotDetected,
		},
		{
			name: "invalid composer manifest",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
				writeTestFile(t, filepath.Join(root, "composer.json"), `{`)
			},
			want: ErrInvalidComposerManifest,
		},
		{
			name: "missing framework requirement",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
				writeTestFile(t, filepath.Join(root, "composer.json"), `{"require":{"php":"^8.2"}}`)
			},
			want: ErrNotDetected,
		},
		{
			name: "oversized composer manifest",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
				content := bytes.Repeat([]byte(" "), maxComposerManifestBytes+1)
				if err := os.WriteFile(filepath.Join(root, "composer.json"), content, 0o600); err != nil {
					t.Fatalf("write oversized composer manifest: %v", err)
				}
			},
			want: ErrInvalidComposerManifest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			test.prepare(t, workspace)

			plugin, err := New(workspace)
			if err != nil {
				t.Fatalf("create plugin: %v", err)
			}

			err = plugin.(pkgRuntime.Initializer).Initialize(nil)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestLoaderHonorsContextAndConstructsPlugin(t *testing.T) {
	workspace := t.TempDir()
	loader := NewLoader(workspace)

	loaded, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load plugin: %v", err)
	}
	if loaded.Metadata().ID != ID {
		t.Fatalf("expected plugin ID %q, got %q", ID, loaded.Metadata().ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loader.Load(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if _, err := loader.Load(nil); !errors.Is(err, pkgPlugin.ErrInvalidLoader) {
		t.Fatalf("expected ErrInvalidLoader, got %v", err)
	}
}

func newLaravelWorkspace(t *testing.T, version string) string {
	t.Helper()

	workspace := t.TempDir()
	writeTestFile(t, filepath.Join(workspace, "artisan"), "#!/usr/bin/env php")
	writeTestFile(
		t,
		filepath.Join(workspace, "composer.json"),
		`{"require":{"php":"^8.2","laravel/framework":"`+version+`"}}`,
	)

	return workspace
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file %q: %v", path, err)
	}
}

func equalCapabilities(left, right []pkgRuntime.Capability) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
