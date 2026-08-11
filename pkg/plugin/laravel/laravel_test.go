package laravel_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/antonio-cafeo/maestro"
	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
	pkgGestor "github.com/antonio-cafeo/maestro/pkg/gestor"
	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
	"github.com/antonio-cafeo/maestro/pkg/plugin/laravel"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

func TestLaravelPluginLoadsAndStartsThroughMaestro(t *testing.T) {
	workspace := t.TempDir()
	writeWorkspaceFile(t, workspace, "artisan", "#!/usr/bin/env php")
	writeWorkspaceFile(
		t,
		workspace,
		"composer.json",
		`{"require":{"laravel/framework":"^12.0"}}`,
	)

	runtime := maestro.New()
	if err := runtime.Plugins().RegisterLoader(
		laravel.ID,
		laravel.NewLoader(laravel.Config{Root: workspace}),
	); err != nil {
		t.Fatalf("register Laravel loader: %v", err)
	}
	if got, want := runtime.Plugins().Available(), []pkgPlugin.ID{laravel.ID}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected available plugins %v, got %v", want, got)
	}

	loaded, err := runtime.Plugins().Load(context.Background(), laravel.ID)
	if err != nil {
		t.Fatalf("load Laravel plugin: %v", err)
	}
	laravelPlugin, ok := loaded.(laravel.Plugin)
	if !ok {
		t.Fatalf("loaded unexpected plugin type %T", loaded)
	}
	if laravelPlugin.FrameworkVersion() != "" {
		t.Fatal("Laravel detection ran during Load")
	}
	if err := runtime.Gestor().Refresh(t.Context()); err != nil {
		t.Fatalf("refresh Gestor after Laravel load: %v", err)
	}
	descriptorCount := 0
	contextDescriptorCount := 0
	for _, descriptor := range runtime.Gestor().Snapshot().Descriptors() {
		if descriptor.Capability == pkgGestor.CapabilityID(
			pkgPlugin.CapabilityWorkspaceDetection,
		) && descriptor.Target.ID == string(laravel.ID) {
			descriptorCount++
		}
		if descriptor.Capability == pkgGestor.CapabilityID(pkgContext.CapabilityWorkspaceProvider) && descriptor.Target.ID == string(laravel.ID) {
			contextDescriptorCount++
		}
	}
	if descriptorCount != 1 {
		t.Fatalf("expected one Laravel workspace descriptor, got %d", descriptorCount)
	}
	if contextDescriptorCount != 1 {
		t.Fatalf("expected one generic workspace descriptor, got %d", contextDescriptorCount)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	if got := laravelPlugin.FrameworkVersion(); got != "^12.0" {
		t.Fatalf("expected Laravel version ^12.0, got %q", got)
	}
	if state := runtime.StateManager().Get(loaded).State; state != pkgRuntime.StateRunning {
		t.Fatalf("expected Laravel plugin running, got %d", state)
	}
	contextWorkspace, err := laravelPlugin.Workspace(t.Context())
	if err != nil {
		t.Fatalf("describe Laravel workspace: %v", err)
	}
	snapshot, err := runtime.ContextEngine().Index(t.Context(), contextWorkspace)
	if err != nil {
		t.Fatalf("index Laravel workspace: %v", err)
	}
	if snapshot.Metadata().DocumentCount != 2 || snapshot.Metadata().Workspace != pkgContext.WorkspaceID(laravel.ID) {
		t.Fatalf("unexpected Laravel context snapshot: %#v", snapshot.Metadata())
	}
	query, err := pkgGestor.NewQuery(
		pkgGestor.CapabilityID(pkgPlugin.CapabilityWorkspaceDetection),
		pkgGestor.QueryOptions{TargetKind: pkgGestor.TargetKindComponent},
	)
	if err != nil {
		t.Fatalf("construct Laravel capability query: %v", err)
	}
	resolution, err := runtime.Gestor().Resolve(query)
	if err != nil {
		t.Fatalf("resolve Laravel workspace capability: %v", err)
	}
	if resolution.Descriptor().Target.ID != string(laravel.ID) {
		t.Fatalf("unexpected Laravel capability target: %#v", resolution.Descriptor().Target)
	}
	contextQuery, err := pkgGestor.NewQuery(
		pkgGestor.CapabilityID(pkgContext.CapabilityWorkspaceProvider),
		pkgGestor.QueryOptions{TargetKind: pkgGestor.TargetKindComponent},
	)
	if err != nil {
		t.Fatalf("construct generic workspace query: %v", err)
	}
	contextResolution, err := runtime.Gestor().Resolve(contextQuery)
	if err != nil {
		t.Fatalf("resolve generic workspace capability: %v", err)
	}
	if contextResolution.Descriptor().Target.ID != string(laravel.ID) {
		t.Fatalf("unexpected generic workspace target: %#v", contextResolution.Descriptor().Target)
	}

	if err := runtime.Stop(context.Background()); err != nil {
		t.Fatalf("stop runtime: %v", err)
	}
}

func TestLaravelFacadeValidatesConfiguration(t *testing.T) {
	if _, err := laravel.New(laravel.Config{}); !errors.Is(
		err,
		laravel.ErrInvalidConfig,
	) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func writeWorkspaceFile(
	t *testing.T,
	root string,
	name string,
	content string,
) {
	t.Helper()

	if err := os.WriteFile(
		filepath.Join(root, name),
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf("write workspace file %q: %v", name, err)
	}
}
