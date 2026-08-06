package llamacpp

import (
	"errors"
	"net/http"
	"testing"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestNewValidatesAndNormalizesBaseURL(t *testing.T) {
	client := &http.Client{}
	valid := []struct {
		value string
		want  string
	}{
		{value: "http://localhost:8080", want: "http://localhost:8080"},
		{value: "http://localhost:8080/", want: "http://localhost:8080"},
		{value: "https://llama.example", want: "https://llama.example"},
	}

	for _, test := range valid {
		t.Run(test.value, func(t *testing.T) {
			provider, err := New(test.value, "", "", client)
			if err != nil {
				t.Fatalf("create provider: %v", err)
			}
			if provider.baseURL != test.want {
				t.Fatalf("expected base URL %q, got %q", test.want, provider.baseURL)
			}
		})
	}

	invalid := []string{
		"",
		"localhost:8080",
		"ftp://localhost:8080",
		"http:///v1",
		"http://localhost:8080/v1",
		"http://user:pass@localhost:8080",
		"http://localhost:8080?debug=true",
		"http://localhost:8080#fragment",
	}
	for _, value := range invalid {
		t.Run("invalid "+value, func(t *testing.T) {
			if _, err := New(value, "", "", client); err == nil {
				t.Fatal("expected invalid base URL error")
			}
		})
	}
}

func TestNewRejectsInvalidServicesAndDefaults(t *testing.T) {
	if _, err := New("http://localhost:8080", "", "", nil); err == nil {
		t.Fatal("expected nil HTTP client error")
	}
	if _, err := New(
		"http://localhost:8080",
		" model",
		"",
		&http.Client{},
	); err == nil {
		t.Fatal("expected invalid default model error")
	}
	if _, err := New(
		"http://localhost:8080",
		"",
		" secret",
		&http.Client{},
	); err == nil {
		t.Fatal("expected invalid API key error")
	}
}

func TestProviderIdentityAndModelResolution(t *testing.T) {
	provider, err := New(
		"http://localhost:8080",
		"local-default",
		"",
		&http.Client{},
	)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if provider.ID() != "llama.cpp" {
		t.Fatalf("expected llama.cpp ID, got %q", provider.ID())
	}
	if got, err := provider.model("explicit"); err != nil || got != "explicit" {
		t.Fatalf("expected explicit model, got %q, %v", got, err)
	}
	if got, err := provider.model(""); err != nil || got != "local-default" {
		t.Fatalf("expected default model, got %q, %v", got, err)
	}

	provider.defaultModel = ""
	if _, err := provider.model(""); !errors.Is(
		err,
		pkgProvider.ErrInvalidRequest,
	) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}
