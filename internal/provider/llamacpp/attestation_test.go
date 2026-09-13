package llamacpp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverModelsAttestsExactLocalGGUF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model.gguf")
	content := []byte("GGUFfixture-model")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		switch request.URL.Path {
		case "/models":
			writeJSON(t, recorder, modelsResponse{Data: []modelData{{ID: "fixture-model"}}})
		case "/props":
			props := serverProps{ModelAlias: "fixture-model", ModelPath: path, BuildInfo: "build-1"}
			props.Settings.ContextWindow = 4096
			writeJSON(t, recorder, props)
		default:
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		response := recorder.Result()
		response.Request = request
		return response, nil
	})}
	provider, err := New("http://127.0.0.1:8080", "fixture-model", "", client, Attestation{
		ModelPath: path, ModelDigest: digest, ServerBuild: "build-1", ContextWindow: 4096,
	})
	if err != nil {
		t.Fatal(err)
	}
	models, err := provider.DiscoverModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].Digest != digest || models[0].Format != "gguf" ||
		models[0].ContextLength != 4096 || models[0].State != "loaded" || models[0].SizeBytes != int64(len(content)) {
		t.Fatalf("unexpected attested model: %#v", models)
	}
}

func TestAttestationRejectsRemoteServerAndIdentityDrift(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model.gguf")
	content := []byte("GGUFfixture-model")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	expected := Attestation{ModelPath: path, ModelDigest: digest, ServerBuild: "build-1", ContextWindow: 4096}
	if _, err := New("https://llama.example", "fixture-model", "", http.DefaultClient, expected); err == nil {
		t.Fatal("remote attestation was accepted")
	}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		if request.URL.Path == "/models" {
			writeJSON(t, recorder, modelsResponse{Data: []modelData{{ID: "fixture-model"}}})
		} else {
			props := serverProps{ModelAlias: "fixture-model", ModelPath: path, BuildInfo: "wrong-build"}
			props.Settings.ContextWindow = 4096
			writeJSON(t, recorder, props)
		}
		response := recorder.Result()
		response.Request = request
		return response, nil
	})}
	provider, err := New("http://127.0.0.1:8080", "fixture-model", "", client, expected)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.DiscoverModels(context.Background()); err == nil {
		t.Fatal("server build drift was accepted")
	}
}
