package laravel

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
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

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("read working directory: %v", err)
	}
	relativeRoot, err := filepath.Rel(workingDirectory, workspace)
	if err != nil {
		t.Fatalf("construct relative workspace root: %v", err)
	}
	relativePlugin, err := New(relativeRoot)
	if err != nil {
		t.Fatalf("create plugin from relative root: %v", err)
	}
	if relativePlugin.Root() != plugin.Root() {
		t.Fatalf(
			"expected normalized root %q, got %q",
			plugin.Root(),
			relativePlugin.Root(),
		)
	}

	missingPlugin, err := New(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatalf("create plugin for missing root: %v", err)
	}
	if err := missingPlugin.(pkgRuntime.Initializer).Initialize(nil); !errors.Is(
		err,
		ErrNotDetected,
	) {
		t.Fatalf("expected deferred missing-root detection, got %v", err)
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
		pkgPlugin.CapabilityWorkspaceDetection,
		pkgContext.CapabilityWorkspaceProvider,
	}; !equalCapabilities(got, want) {
		t.Fatalf("expected capabilities %v, got %v", want, got)
	}
	if plugin.Manifest().RuntimeAPIVersion != pkgPlugin.RuntimeAPIVersion {
		t.Fatalf("unexpected plugin manifest: %#v", plugin.Manifest())
	}
}

func TestPluginProvidesFrameworkNeutralWorkspace(t *testing.T) {
	root := newLaravelWorkspace(t, "^12.0")
	plugin, err := New(root)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}
	if _, err := plugin.Workspace(nil); !errors.Is(err, pkgContext.ErrInvalidWorkspace) {
		t.Fatalf("expected nil context rejection, got %v", err)
	}
	before, err := plugin.Workspace(context.Background())
	if err != nil {
		t.Fatalf("describe workspace before initialization: %v", err)
	}
	if before.ID() != pkgContext.WorkspaceID(ID) || before.Root() != filepath.Clean(root) || before.Source() != pkgContext.SourceFilesystem || before.Metadata()["framework"] != "laravel" {
		t.Fatalf("unexpected generic workspace: id=%q root=%q source=%q metadata=%#v", before.ID(), before.Root(), before.Source(), before.Metadata())
	}
	if _, exists := before.Metadata()["framework_version"]; exists {
		t.Fatal("workspace exposed a version before initialization")
	}
	policy := before.Policy()
	if policy.MaxFileBytes != maxLaravelSourceBytes {
		t.Fatalf("unexpected Laravel source file limit: %d", policy.MaxFileBytes)
	}
	for _, required := range []string{"app/**", "resources/views/**", "routes/**", "README.md", "composer.json", "dataset.json"} {
		if !slices.Contains(policy.Include, required) {
			t.Fatalf("Laravel source policy does not include %q: %#v", required, policy.Include)
		}
	}
	if slices.Contains(policy.Include, "public/**") || slices.Contains(policy.Include, "storage/**") {
		t.Fatalf("Laravel source policy includes generated/runtime paths: %#v", policy.Include)
	}
	if err := plugin.(pkgRuntime.Initializer).Initialize(nil); err != nil {
		t.Fatalf("initialize plugin: %v", err)
	}
	after, err := plugin.Workspace(context.Background())
	if err != nil {
		t.Fatalf("describe initialized workspace: %v", err)
	}
	if after.Metadata()["framework_version"] != "^12.0" {
		t.Fatalf("unexpected framework-neutral metadata: %#v", after.Metadata())
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
	if got := plugin.FrameworkVersion(); got != "^12.0" {
		t.Fatalf("failed health check changed version snapshot to %q", got)
	}

	writeTestFile(t, filepath.Join(workspace, "artisan"), "#!/usr/bin/env php")
	writeTestFile(
		t,
		filepath.Join(workspace, "composer.json"),
		`{"require":{"laravel/framework":"^13.0"}}`,
	)
	if err := plugin.(pkgRuntime.HealthChecker).Health(nil); err != nil {
		t.Fatalf("check updated workspace health: %v", err)
	}
	if got := plugin.FrameworkVersion(); got != "^12.0" {
		t.Fatalf("health check mutated initialized snapshot to %q", got)
	}
	if err := plugin.(pkgRuntime.Initializer).Initialize(nil); err != nil {
		t.Fatalf("reinitialize updated workspace: %v", err)
	}
	if got := plugin.FrameworkVersion(); got != "^13.0" {
		t.Fatalf("expected refreshed version ^13.0, got %q", got)
	}

	writeTestFile(t, filepath.Join(workspace, "composer.json"), `{`)
	if err := plugin.(pkgRuntime.Initializer).Initialize(nil); !errors.Is(
		err,
		ErrInvalidComposerManifest,
	) {
		t.Fatalf("expected failed reinitialization, got %v", err)
	}
	if got := plugin.FrameworkVersion(); got != "^13.0" {
		t.Fatalf("failed initialization corrupted version snapshot to %q", got)
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
			name: "empty framework requirement",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
				writeTestFile(t, filepath.Join(root, "composer.json"), `{"require":{"laravel/framework":"   "}}`)
			},
			want: ErrNotDetected,
		},
		{
			name: "invalid require section",
			prepare: func(t *testing.T, root string) {
				writeTestFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php")
				writeTestFile(t, filepath.Join(root, "composer.json"), `{"require":[]}`)
			},
			want: ErrInvalidComposerManifest,
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

func TestPluginWorkspaceViewIsConcurrent(t *testing.T) {
	workspace := newLaravelWorkspace(t, "^12.0")
	plugin, err := New(workspace)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}
	initializer := plugin.(pkgRuntime.Initializer)
	health := plugin.(pkgRuntime.HealthChecker)
	if err := initializer.Initialize(nil); err != nil {
		t.Fatalf("initialize plugin: %v", err)
	}

	const operationCount = 100
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, operationCount*3)
	for range operationCount {
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			if err := initializer.Initialize(nil); err != nil {
				errorsChannel <- err
			}
		}()
		go func() {
			defer waitGroup.Done()
			if err := health.Health(nil); err != nil {
				errorsChannel <- err
				return
			}
			if plugin.Root() != filepath.Clean(workspace) {
				errorsChannel <- errors.New("workspace root changed")
			}
			if plugin.FrameworkVersion() != "^12.0" {
				errorsChannel <- errors.New("framework version changed")
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent workspace operation: %v", err)
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
