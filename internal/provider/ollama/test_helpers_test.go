package ollama

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func newTestProvider(
	t *testing.T,
	defaultModel string,
	handler http.HandlerFunc,
) *Provider {
	t.Helper()

	client := &http.Client{Transport: roundTripFunc(func(
		request *http.Request,
	) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		response := recorder.Result()
		response.Request = request

		return response, nil
	})}

	provider, err := New("http://ollama.test", defaultModel, client)
	if err != nil {
		t.Fatalf("create test provider: %v", err)
	}

	return provider
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatalf("encode test response: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type trackingBody struct {
	reader io.Reader

	mu         sync.Mutex
	closeCount int
}

func (b *trackingBody) Read(buffer []byte) (int, error) {
	return b.reader.Read(buffer)
}

func (b *trackingBody) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closeCount++

	return nil
}

func (b *trackingBody) closes() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.closeCount
}
