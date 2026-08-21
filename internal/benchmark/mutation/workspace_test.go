package mutation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMaterializedFixtureMatchesFrozenDigestsAndCleansUp(t *testing.T) {
	profile, err := LoadProfile("../../../docs/mutation-qualification-profile.yaml")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := MaterializeFixture(context.Background(), profile)
	if err != nil {
		t.Fatal(err)
	}
	root := fixture.Root
	if err := fixture.Initial.Validate(); err != nil {
		t.Fatal(err)
	}
	if digest, ok := fixture.Initial.File(profile.Fixture.Target); !ok || digest != profile.Fixture.InitialSHA256 {
		t.Fatalf("unexpected target digest: %q, %v", digest, ok)
	}
	if err := fixture.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("fixture root survived cleanup: %v", err)
	}
}

func TestSnapshotRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target"), []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := SnapshotWorkspace(context.Background(), root); err == nil {
		t.Fatal("workspace snapshot accepted a symlink")
	}
}
