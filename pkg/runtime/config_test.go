package runtime

import "testing"

func TestNewConfigCopiesInputAndUsesExactKeys(t *testing.T) {
	values := map[string]any{
		"providers.default": "ollama",
	}
	config := NewConfig(values)

	values["providers.default"] = "llama.cpp"

	if got := config.Get("providers.default"); got != "ollama" {
		t.Fatalf("expected copied value ollama, got %#v", got)
	}

	if got := config.Get("Providers.Default"); got != nil {
		t.Fatalf("expected exact key lookup, got %#v", got)
	}

	if got := config.Get("missing"); got != nil {
		t.Fatalf("expected nil missing value, got %#v", got)
	}
}

func TestNewConfigAcceptsNilValues(t *testing.T) {
	config := NewConfig(nil)

	if got := config.Get("missing"); got != nil {
		t.Fatalf("expected nil missing value, got %#v", got)
	}
}
