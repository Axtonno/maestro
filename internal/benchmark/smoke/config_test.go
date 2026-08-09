package smoke

import (
	"testing"
	"time"
)

func TestConfigFromEnvironmentBuildsExplicitProviderProfile(t *testing.T) {
	manifest := loadTestManifest(t)
	environment := map[string]string{
		"MAESTRO_LLAMACPP_BASE_URL":          "http://localhost:8080",
		"MAESTRO_LLAMACPP_API_KEY":           "secret-key",
		"MAESTRO_LLAMACPP_CHAT_MODEL":        "chat.gguf",
		"MAESTRO_LLAMACPP_EMBED_MODEL":       "embed.gguf",
		"MAESTRO_LLAMACPP_LIFECYCLE_MODEL":   "lifecycle.gguf",
		"MAESTRO_LLAMACPP_ACQUISITION_MODEL": "temporary.gguf",
		"MAESTRO_ALLOW_CATALOG_MUTATION":     "true",
	}

	config, err := configFromEnvironment(
		manifest,
		ProviderLlamaCPP,
		time.Minute,
		func(name string) string { return environment[name] },
	)
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	if !config.Enabled || config.BaseURL != "http://localhost:8080" ||
		config.APIKey != "secret-key" || !config.AllowCatalogMutation ||
		config.Models["chat"] != "chat.gguf" ||
		config.Models["acquisition_fixture"] != "temporary.gguf" {
		t.Fatalf("unexpected configuration: %#v", config)
	}
	profile := config.ConfigurationProfile()
	if profile.Provider.ID != ProviderLlamaCPP ||
		profile.Models["embedding"].ID != "embed.gguf" {
		t.Fatalf("unexpected report profile: %#v", profile)
	}
}

func TestConfigFromEnvironmentDisablesMissingProviderWithoutDefaultingLocalhost(t *testing.T) {
	manifest := loadTestManifest(t)

	config, err := configFromEnvironment(
		manifest,
		ProviderOllama,
		time.Minute,
		func(string) string { return "" },
	)
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	if config.Enabled || config.BaseURL != "" ||
		len(config.MissingEnvironment) != 1 ||
		config.MissingEnvironment[0] != "MAESTRO_OLLAMA_BASE_URL" {
		t.Fatalf("unexpected disabled configuration: %#v", config)
	}
}

func TestConfigFromEnvironmentRejectsAmbiguousMutationGuard(t *testing.T) {
	manifest := loadTestManifest(t)

	_, err := configFromEnvironment(
		manifest,
		ProviderOllama,
		time.Minute,
		func(name string) string {
			if name == "MAESTRO_ALLOW_CATALOG_MUTATION" {
				return "yes"
			}
			return ""
		},
	)
	if err == nil {
		t.Fatal("expected invalid mutation guard")
	}
}
