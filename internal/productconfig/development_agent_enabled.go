//go:build maestro_development

package productconfig

func supportedAgentID(id string) bool {
	return id == "agent.reference" || id == "agent.progressive-reference"
}
