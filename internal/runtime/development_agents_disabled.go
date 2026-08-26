//go:build !maestro_development

package runtime

import internalAgent "github.com/antonio-cafeo/maestro/internal/agent"

func registerDevelopmentAgents(*internalAgent.Runtime) error { return nil }
