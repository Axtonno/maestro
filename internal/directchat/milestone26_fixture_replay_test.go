package directchat

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type milestone26ValidatorFixture struct {
	ID           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	ToolCall     bool   `json:"tool_call"`
	FinishReason string `json:"finish_reason"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Expected     string `json:"expected"`
}

// TestMilestone26OfflineValidatorFixtureReplay is deliberately test-only. It
// freezes the normalized inputs for every completion validator predicate so
// diagnosis can be repeated without a provider call or a new generation.
func TestMilestone26OfflineValidatorFixtureReplay(t *testing.T) {
	encoded, err := os.ReadFile("testdata/milestone26-validator-fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []milestone26ValidatorFixture
	if err := json.Unmarshal(encoded, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.ID, func(t *testing.T) {
			toolCalls := []pkgProvider.ToolCall(nil)
			if fixture.ToolCall {
				toolCalls = []pkgProvider.ToolCall{{Name: "synthetic_tool"}}
			}
			provider := providerWithResponse(pkgProvider.CompletionResponse{
				Model: fixture.Model,
				Message: pkgProvider.Message{
					Role: pkgProvider.Role(fixture.Role), Content: fixture.Content, ToolCalls: toolCalls,
				},
				FinishReason: fixture.FinishReason,
				Usage:        pkgProvider.Usage{InputTokens: fixture.InputTokens, OutputTokens: fixture.OutputTokens},
			})
			service := buildService(t, directConfig(t.TempDir()), provider)
			_, err := service.Execute(t.Context(), Request{Question: "Synthetic replay"})
			if fixture.Expected == "accepted" && err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
			if fixture.Expected == "response_invalid" && !errors.Is(err, ErrResponseInvalid) {
				t.Fatalf("expected response_invalid, got %v", err)
			}
		})
	}
}

func TestMilestone26OfflineValidatorRejectsInvalidUTF8AndNegativeDuration(t *testing.T) {
	t.Run("invalid_utf8", func(t *testing.T) {
		provider := validProvider()
		provider.response.Message.Content = string([]byte{0xff})
		service := buildService(t, directConfig(t.TempDir()), provider)
		if _, err := service.Execute(t.Context(), Request{Question: "Synthetic replay"}); !errors.Is(err, ErrResponseInvalid) {
			t.Fatalf("expected response_invalid, got %v", err)
		}
	})

	t.Run("negative_duration", func(t *testing.T) {
		provider := validProvider()
		times := []time.Time{time.Unix(2, 0), time.Unix(1, 0)}
		service, err := Build(directConfig(t.TempDir()), Dependencies{
			ProviderFactory: fixtureFactory(provider),
			Now: func() time.Time {
				value := times[0]
				times = times[1:]
				return value
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.Execute(t.Context(), Request{Question: "Synthetic replay"}); !errors.Is(err, ErrResponseInvalid) {
			t.Fatalf("expected response_invalid, got %v", err)
		}
	})
}
