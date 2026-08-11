package plugin

import (
	"context"
	"testing"

	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

var _ Plugin = (*typeTestPlugin)(nil)
var _ Loader = LoaderFunc(nil)
var _ pkgRuntime.Event = Event{}

type typeTestPlugin struct{}

func (p *typeTestPlugin) Metadata() pkgRuntime.Metadata {
	return pkgRuntime.Metadata{ID: "test", Name: "Test", Version: "1.0.0"}
}

func (p *typeTestPlugin) Manifest() Manifest {
	return Manifest{RuntimeAPIVersion: RuntimeAPIVersion}
}

func TestLoaderFunc(t *testing.T) {
	want := &typeTestPlugin{}
	loader := LoaderFunc(func(context.Context) (Plugin, error) {
		return want, nil
	})

	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load plugin: %v", err)
	}
	if got != want {
		t.Fatal("loader function returned an unexpected plugin")
	}
}

func TestRuntimeAPIVersionIsStableAndNonEmpty(t *testing.T) {
	if RuntimeAPIVersion != "1" {
		t.Fatalf("expected Runtime API version 1, got %q", RuntimeAPIVersion)
	}
}

func TestWorkspaceDetectionCapabilityIsNamespaced(t *testing.T) {
	if CapabilityWorkspaceDetection != "plugin.workspace-detection" {
		t.Fatalf(
			"unexpected workspace detection capability %q",
			CapabilityWorkspaceDetection,
		)
	}
}

func TestEvent(t *testing.T) {
	registered := &typeTestPlugin{}
	payload := EventPayload{ID: "test", Plugin: registered}
	event := Event{Topic: EventRegistered, Data: payload}

	if event.Name() != EventRegistered {
		t.Fatalf("expected topic %q, got %q", EventRegistered, event.Name())
	}
	if got := event.Payload(); got != payload {
		t.Fatalf("expected payload %#v, got %#v", payload, got)
	}
}
