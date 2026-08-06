package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestPullModelMapsRequestAndProgress(t *testing.T) {
	body := &trackingBody{reader: strings.NewReader(
		"{\"status\":\"pulling manifest\"}\n" +
			"{\"status\":\"pulling abc\",\"digest\":\"sha256:abc\",\"total\":100,\"completed\":40}\n" +
			"{\"status\":\"verifying sha256 digest\",\"digest\":\"sha256:abc\"}\n" +
			"{\"status\":\"success\"}\n",
	)}
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/pull" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}

		decoded := modelPullRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode pull request: %v", err)
		}
		if decoded.Model != "gemma4" || !decoded.Stream {
			t.Errorf("unexpected pull request: %#v", decoded)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://ollama.test", "gemma4", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{},
	)
	if err != nil {
		t.Fatalf("pull model: %v", err)
	}

	wantStages := []pkgProvider.ModelPullStage{
		pkgProvider.ModelPullStageResolving,
		pkgProvider.ModelPullStageDownloading,
		pkgProvider.ModelPullStageVerifying,
		pkgProvider.ModelPullStageCompleted,
	}
	for index, wantStage := range wantStages {
		progress, err := stream.Recv()
		if err != nil {
			t.Fatalf("receive progress %d: %v", index, err)
		}
		if progress.Model != "gemma4" || progress.Stage != wantStage {
			t.Fatalf("unexpected progress %d: %#v", index, progress)
		}
		if index == 1 && (progress.Digest != "sha256:abc" ||
			progress.TotalBytes != 100 || progress.CompletedBytes != 40) {
			t.Fatalf("unexpected byte progress: %#v", progress)
		}
	}

	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after completion, got %v", err)
	}
	if body.closes() != 1 {
		t.Fatalf("expected body closed once, got %d", body.closes())
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close completed stream: %v", err)
	}
}

func TestOllamaPullStageMapsKnownAndFutureStatuses(t *testing.T) {
	tests := []struct {
		status string
		stage  pkgProvider.ModelPullStage
	}{
		{status: "pulling manifest", stage: pkgProvider.ModelPullStageResolving},
		{status: "pulling sha256:abc", stage: pkgProvider.ModelPullStageDownloading},
		{status: "verifying sha256 digest", stage: pkgProvider.ModelPullStageVerifying},
		{status: "writing manifest", stage: pkgProvider.ModelPullStageFinalizing},
		{status: "removing any unused layers", stage: pkgProvider.ModelPullStageFinalizing},
		{status: "success", stage: pkgProvider.ModelPullStageCompleted},
		{status: "future status", stage: pkgProvider.ModelPullStageUnknown},
	}

	for _, test := range tests {
		if got := ollamaPullStage(test.status); got != test.stage {
			t.Errorf("status %q: expected %q, got %q", test.status, test.stage, got)
		}
	}
}

func TestPullModelRejectsInvalidStreams(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "premature EOF", body: "{\"status\":\"pulling manifest\"}\n"},
		{name: "malformed JSON", body: "not-json\n"},
		{name: "missing status", body: "{}\n"},
		{name: "negative total", body: "{\"status\":\"pulling layer\",\"total\":-1}\n"},
		{name: "completed over total", body: "{\"status\":\"pulling layer\",\"total\":10,\"completed\":11}\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := newTestProvider(t, "model", func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				_, _ = writer.Write([]byte(test.body))
			})
			stream, err := provider.PullModel(
				context.Background(),
				pkgProvider.ModelPullRequest{},
			)
			if err != nil {
				t.Fatalf("open pull stream: %v", err)
			}

			var receiveError error
			for receiveError == nil {
				_, receiveError = stream.Recv()
			}
			if !errors.Is(receiveError, pkgProvider.ErrInvalidResponse) {
				t.Fatalf("expected ErrInvalidResponse, got %v", receiveError)
			}
		})
	}
}

func TestPullModelPropagatesStreamErrorAndCancellation(t *testing.T) {
	provider := newTestProvider(t, "model", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, _ = writer.Write([]byte("{\"error\":\"pull failed\"}\n"))
	})
	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{},
	)
	if err != nil {
		t.Fatalf("open pull stream: %v", err)
	}
	if _, err := stream.Recv(); err == nil {
		t.Fatal("expected stream API error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancellable, err := provider.PullModel(ctx, pkgProvider.ModelPullRequest{})
	if err != nil {
		t.Fatalf("open cancellable pull stream: %v", err)
	}
	cancel()
	if _, err := cancellable.Recv(); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if err := cancellable.Close(); err != nil {
		t.Fatalf("close canceled stream: %v", err)
	}
}

func TestPullModelCloseIsIdempotent(t *testing.T) {
	provider := newTestProvider(t, "model", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, _ = writer.Write([]byte("{\"status\":\"pulling manifest\"}\n"))
	})
	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{},
	)
	if err != nil {
		t.Fatalf("open pull stream: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close stream: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close stream again: %v", err)
	}
	if _, err := stream.Recv(); !errors.Is(err, pkgProvider.ErrInvalidStream) {
		t.Fatalf("expected ErrInvalidStream, got %v", err)
	}
}

func TestRemoveModelMapsRequest(t *testing.T) {
	provider := newTestProvider(t, "default-model", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodDelete || request.URL.Path != "/api/delete" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		decoded := modelRemoveRequest{}
		if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
			t.Errorf("decode remove request: %v", err)
		}
		if decoded.Model != "explicit-model" {
			t.Errorf("unexpected model %q", decoded.Model)
		}
		writer.WriteHeader(http.StatusOK)
	})

	if err := provider.RemoveModel(
		context.Background(),
		pkgProvider.ModelRemoveRequest{Model: "explicit-model"},
	); err != nil {
		t.Fatalf("remove model: %v", err)
	}
}

func TestRemoveModelPropagatesHTTPError(t *testing.T) {
	provider := newTestProvider(t, "model", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"error":"model not found"}`))
	})

	if err := provider.RemoveModel(
		context.Background(),
		pkgProvider.ModelRemoveRequest{},
	); err == nil {
		t.Fatal("expected removal error")
	}
}
