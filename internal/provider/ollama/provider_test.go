package ollama

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
		{value: "http://localhost:11434", want: "http://localhost:11434"},
		{value: "http://localhost:11434/", want: "http://localhost:11434"},
		{value: "https://ollama.example", want: "https://ollama.example"},
	}

	for _, test := range valid {
		t.Run(test.value, func(t *testing.T) {
			provider, err := New(test.value, "", client)
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
		"localhost:11434",
		"ftp://localhost:11434",
		"http:///api",
		"http://localhost:11434/api",
		"http://user:pass@localhost:11434",
		"http://localhost:11434?debug=true",
		"http://localhost:11434#fragment",
	}

	for _, value := range invalid {
		t.Run("invalid "+value, func(t *testing.T) {
			if _, err := New(value, "", client); err == nil {
				t.Fatal("expected invalid base URL error")
			}
		})
	}
}

func TestNewRejectsInvalidServicesAndDefaults(t *testing.T) {
	if _, err := New("http://localhost:11434", "", nil); err == nil {
		t.Fatal("expected nil HTTP client error")
	}

	if _, err := New(
		"http://localhost:11434",
		" model",
		&http.Client{},
	); err == nil {
		t.Fatal("expected invalid default model error")
	}
}

func TestProviderIdentityAndModelResolution(t *testing.T) {
	provider, err := New(
		"http://localhost:11434",
		"qwen3",
		&http.Client{},
	)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	if provider.ID() != "ollama" {
		t.Fatalf("expected ollama ID, got %q", provider.ID())
	}

	if got, err := provider.model("gemma4"); err != nil || got != "gemma4" {
		t.Fatalf("expected explicit model, got %q, %v", got, err)
	}

	if got, err := provider.model(""); err != nil || got != "qwen3" {
		t.Fatalf("expected default model, got %q, %v", got, err)
	}

	provider.defaultModel = ""
	if _, err := provider.model(""); err == nil {
		t.Fatal("expected missing model error")
	} else if !errors.Is(err, pkgProvider.ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest, got %v", err)
	}
}
