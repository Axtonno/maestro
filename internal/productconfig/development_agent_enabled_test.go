//go:build maestro_development

package productconfig

import "testing"

func TestDevelopmentBuildAcceptsProgressiveReferenceAgent(t *testing.T) {
	config := validConfig(t.TempDir())
	config.Agent.ID = "agent.progressive-reference"
	if err := config.Validate(); err != nil {
		t.Fatalf("validate progressive development agent: %v", err)
	}
}
