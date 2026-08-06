package llamacpp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestPullModelMapsRequestAndAggregatesProgress(t *testing.T) {
	var mu sync.Mutex
	requests := make([]string, 0, 3)
	provider := newTestProvider(t, "", "secret", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing bearer authentication")
		}
		mu.Lock()
		requests = append(requests, request.Method+" "+request.URL.Path)
		mu.Unlock()

		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/models/sse":
			writer.Header().Set("Content-Type", "text/event-stream")
			_, _ = writer.Write([]byte(
				"data: {\"model\":\"other\",\"event\":\"download_progress\",\"data\":{\"ignored\":{\"done\":1,\"total\":1}}}\n\n" +
					"data: {\"model\":\"repo/model:Q4\",\"event\":\"download_progress\",\"data\":{\"first\":{\"done\":40,\"total\":100},\"second\":{\"done\":5,\"total\":20}}}\n\n" +
					"data: {\"model\":\"repo/model:Q4\",\"event\":\"download_finished\"}\n\n",
			))
		case request.Method == http.MethodPost && request.URL.Path == "/models":
			decoded := modelLifecycleRequest{}
			if err := json.NewDecoder(request.Body).Decode(&decoded); err != nil {
				t.Errorf("decode model pull request: %v", err)
			}
			if decoded.Model != "repo/model:Q4" {
				t.Errorf("unexpected model %q", decoded.Model)
			}
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		case request.Method == http.MethodGet && request.URL.Path == "/models":
			writeJSON(t, writer, modelsResponse{Data: []modelData{{
				ID: "repo/model:Q4", Status: modelStatusData{Value: "unloaded"},
			}}})
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		}
	})

	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{Model: "repo/model:Q4"},
	)
	if err != nil {
		t.Fatalf("pull model: %v", err)
	}

	progress, err := stream.Recv()
	if err != nil {
		t.Fatalf("receive progress: %v", err)
	}
	want := pkgProvider.ModelPullProgress{
		Model:          "repo/model:Q4",
		Stage:          pkgProvider.ModelPullStageDownloading,
		Detail:         "download_progress",
		TotalBytes:     120,
		CompletedBytes: 45,
	}
	if !reflect.DeepEqual(progress, want) {
		t.Fatalf("expected %#v, got %#v", want, progress)
	}

	completed, err := stream.Recv()
	if err != nil {
		t.Fatalf("receive completion: %v", err)
	}
	if completed.Model != "repo/model:Q4" ||
		completed.Stage != pkgProvider.ModelPullStageCompleted {
		t.Fatalf("unexpected completion: %#v", completed)
	}
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close completed stream: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	wantRequests := []string{"POST /models", "GET /models/sse", "GET /models"}
	if !reflect.DeepEqual(requests, wantRequests) {
		t.Fatalf("expected requests %#v, got %#v", wantRequests, requests)
	}
}

func TestPullModelRejectsInvalidEvents(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "premature EOF", body: ": keepalive\n\n"},
		{name: "malformed JSON", body: "data: not-json\n\n"},
		{
			name: "empty progress",
			body: "data: {\"model\":\"model\",\"event\":\"download_progress\",\"data\":{}}\n\n",
		},
		{
			name: "invalid byte counts",
			body: "data: {\"model\":\"model\",\"event\":\"download_progress\",\"data\":{\"file\":{\"done\":11,\"total\":10}}}\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := llamaPullTestProvider(t, test.body, nil)
			stream, err := provider.PullModel(
				context.Background(),
				pkgProvider.ModelPullRequest{Model: "model"},
			)
			if err != nil {
				t.Fatalf("open pull stream: %v", err)
			}
			if _, err := stream.Recv(); !errors.Is(
				err,
				pkgProvider.ErrInvalidResponse,
			) {
				t.Fatalf("expected ErrInvalidResponse, got %v", err)
			}
		})
	}
}

func TestPullModelPropagatesFailureEvent(t *testing.T) {
	provider := llamaPullTestProvider(
		t,
		"data: {\"model\":\"model\",\"event\":\"download_failed\",\"data\":{\"message\":\"download denied\"}}\n\n",
		nil,
	)
	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{Model: "model"},
	)
	if err != nil {
		t.Fatalf("open pull stream: %v", err)
	}
	if _, err := stream.Recv(); err == nil || !strings.Contains(
		err.Error(),
		"download denied",
	) {
		t.Fatalf("expected download failure, got %v", err)
	}
}

func TestPullModelCloseAndCancellationCancelRemoteDownload(t *testing.T) {
	var mu sync.Mutex
	unloadCount := 0
	provider := llamaPullTestProvider(
		t,
		"data: {\"model\":\"model\",\"event\":\"download_progress\",\"data\":{\"file\":{\"done\":1,\"total\":10}}}\n\n",
		func(request *http.Request) {
			if request.Method == http.MethodPost &&
				request.URL.Path == "/models/unload" {
				mu.Lock()
				unloadCount++
				mu.Unlock()
			}
		},
	)

	stream, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{Model: "model"},
	)
	if err != nil {
		t.Fatalf("open pull stream: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close pull stream: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("close pull stream again: %v", err)
	}
	if _, err := stream.Recv(); !errors.Is(err, pkgProvider.ErrInvalidStream) {
		t.Fatalf("expected ErrInvalidStream, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancellable, err := provider.PullModel(
		ctx,
		pkgProvider.ModelPullRequest{Model: "model"},
	)
	if err != nil {
		t.Fatalf("open cancellable stream: %v", err)
	}
	cancel()
	if _, err := cancellable.Recv(); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if unloadCount != 2 {
		t.Fatalf("expected two remote cancellations, got %d", unloadCount)
	}
}

func TestPullModelDoesNotSubscribeWhenStartFails(t *testing.T) {
	subscriptionRequested := false
	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		if request.Method == http.MethodGet {
			subscriptionRequested = true
		}

		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"invalid model"}}`)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}
	provider, err := New("http://llamacpp.test", "model", "", client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if _, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{},
	); err == nil {
		t.Fatal("expected pull start error")
	}
	if subscriptionRequested {
		t.Fatal("subscription was requested after pull start failed")
	}
}

func TestPullModelCancelsDownloadWhenSubscriptionFails(t *testing.T) {
	unloadRequested := false
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/models":
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		case request.Method == http.MethodGet && request.URL.Path == "/models/sse":
			writer.WriteHeader(http.StatusServiceUnavailable)
		case request.Method == http.MethodPost && request.URL.Path == "/models/unload":
			unloadRequested = true
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	})

	if _, err := provider.PullModel(
		context.Background(),
		pkgProvider.ModelPullRequest{Model: "model"},
	); err == nil {
		t.Fatal("expected subscription error")
	}
	if !unloadRequested {
		t.Fatal("remote download was not canceled")
	}
}

func TestRemoveModelMapsQueryRequest(t *testing.T) {
	provider := newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodDelete || request.URL.Path != "/models" {
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		if model := request.URL.Query().Get("model"); model != "repo/model:Q4" {
			t.Errorf("unexpected model query %q", model)
		}
		writeJSON(t, writer, modelLifecycleResponse{Success: true})
	})

	if err := provider.RemoveModel(
		context.Background(),
		pkgProvider.ModelRemoveRequest{Model: "repo/model:Q4"},
	); err != nil {
		t.Fatalf("remove model: %v", err)
	}
}

func TestRemoveModelRejectsUnsuccessfulResponse(t *testing.T) {
	provider := newTestProvider(t, "model", "", func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		writeJSON(t, writer, modelLifecycleResponse{})
	})

	if err := provider.RemoveModel(
		context.Background(),
		pkgProvider.ModelRemoveRequest{},
	); !errors.Is(err, pkgProvider.ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}

func llamaPullTestProvider(
	t *testing.T,
	events string,
	observe func(*http.Request),
) *Provider {
	t.Helper()

	return newTestProvider(t, "", "", func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if observe != nil {
			observe(request)
		}

		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/models/sse":
			writer.Header().Set("Content-Type", "text/event-stream")
			_, _ = writer.Write([]byte(events))
		case request.Method == http.MethodPost &&
			(request.URL.Path == "/models" ||
				request.URL.Path == "/models/unload"):
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		case request.Method == http.MethodGet && request.URL.Path == "/models":
			writeJSON(t, writer, modelsResponse{})
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
			writeJSON(t, writer, modelLifecycleResponse{Success: true})
		}
	})
}
