//go:build maestro_development

package runtime

import internalAgent "github.com/antonio-cafeo/maestro/internal/agent"

func registerDevelopmentAgents(runtime *internalAgent.Runtime) error {
	return runtime.Register(internalAgent.NewProgressiveReferenceAgent())
}
