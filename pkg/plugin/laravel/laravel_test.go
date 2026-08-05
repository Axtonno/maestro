package laravel_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/antonio-cafeo/maestro"
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

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	if got := laravelPlugin.FrameworkVersion(); got != "^12.0" {
		t.Fatalf("expected Laravel version ^12.0, got %q", got)
	}
	if state := runtime.StateManager().Get(loaded).State; state != pkgRuntime.StateRunning {
		t.Fatalf("expected Laravel plugin running, got %d", state)
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
