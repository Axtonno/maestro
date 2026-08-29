package directchat

import (
	"context"
	"errors"
	"os"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
)

type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckFail CheckStatus = "fail"
	CheckSkip CheckStatus = "skip"
)

type Check struct {
	Name   string
	Status CheckStatus
	Detail string
}

// Doctor validates and probes only the direct-chat graph. It never performs a
// completion and never composes agent, retrieval, tool or plugin services.
func Doctor(ctx context.Context, config productconfig.Config, dependencies Dependencies) []Check {
	checks := []Check{{Name: "config", Status: CheckPass, Detail: "schema_v2_chat_valid"}}
	if ctx == nil || config.ValidateChatExecutionProfile() != nil {
		return append(checks[:0],
			Check{Name: "config", Status: CheckFail, Detail: "configuration_invalid"},
			Check{Name: "workspace", Status: CheckSkip, Detail: "configuration_invalid"},
			Check{Name: "composition", Status: CheckSkip, Detail: "configuration_invalid"},
			Check{Name: "model", Status: CheckSkip, Detail: "configuration_invalid"},
			Check{Name: "generation", Status: CheckSkip, Detail: "configuration_invalid"},
		)
	}
	info, err := os.Lstat(config.Workspace.Root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return append(checks,
			Check{Name: "workspace", Status: CheckFail, Detail: "root_unavailable"},
			Check{Name: "composition", Status: CheckSkip, Detail: "workspace_unavailable"},
			Check{Name: "model", Status: CheckSkip, Detail: "workspace_unavailable"},
			Check{Name: "generation", Status: CheckSkip, Detail: "workspace_unavailable"},
		)
	}
	checks = append(checks, Check{Name: "workspace", Status: CheckPass, Detail: "root_available"})
	service, err := Build(config, dependencies)
	if err != nil {
		detail := "composition_failed"
		if errors.Is(err, productconfig.ErrSecretMissing) {
			detail = "secret_environment_missing"
		}
		return append(checks,
			Check{Name: "composition", Status: CheckFail, Detail: detail},
			Check{Name: "model", Status: CheckSkip, Detail: "composition_failed"},
			Check{Name: "generation", Status: CheckSkip, Detail: "composition_failed"},
		)
	}
	checks = append(checks, Check{Name: "composition", Status: CheckPass, Detail: "direct_chat_provider"})
	if err := service.preflight(ctx, service.profile.Streaming); err != nil {
		return append(checks,
			Check{Name: "model", Status: CheckFail, Detail: "required_capability_unavailable"},
			Check{Name: "generation", Status: CheckSkip, Detail: "model_unavailable"},
		)
	}
	return append(checks,
		Check{Name: "model", Status: CheckPass, Detail: "completion_capabilities_available"},
		Check{Name: "generation", Status: CheckPass, Detail: "generation_controls_available"},
	)
}
