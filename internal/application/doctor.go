package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckWarn CheckStatus = "warn"
	CheckFail CheckStatus = "fail"
	CheckSkip CheckStatus = "skip"
)

type Check struct {
	Name   string
	Status CheckStatus
	Detail string
}

func Doctor(ctx context.Context, config productconfig.Config, dependencies Dependencies) []Check {
	checks := []Check{{Name: "config", Status: CheckPass, Detail: fmt.Sprintf("schema_v%d_valid", config.Version)}}
	if info, err := os.Stat(config.Workspace.Root); err != nil || !info.IsDir() {
		checks = append(checks, Check{Name: "workspace", Status: CheckFail, Detail: "root_unavailable"})
	} else {
		checks = append(checks, Check{Name: "workspace", Status: CheckPass, Detail: "root_available"})
	}
	application, err := Build(config, dependencies)
	if err != nil {
		detail := "composition_failed"
		switch {
		case errors.Is(err, productconfig.ErrSecretMissing):
			detail = "secret_environment_missing"
		case errors.Is(err, productconfig.ErrInvalid):
			detail = "configuration_invalid"
		}
		checks = append(checks,
			Check{Name: "composition", Status: CheckFail, Detail: detail},
			Check{Name: "provider", Status: CheckSkip, Detail: "composition_failed"},
			Check{Name: "model", Status: CheckSkip, Detail: "composition_failed"},
			Check{Name: "laravel", Status: CheckSkip, Detail: "composition_failed"},
		)
		return checks
	}
	defer func() {
		closeContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = application.Close(closeContext)
	}()
	checks = append(checks, Check{Name: "composition", Status: CheckPass, Detail: "targets_registered"})
	agentFound := false
	for _, descriptor := range application.runtime.Agents().Descriptors() {
		agentFound = agentFound || string(descriptor.ID()) == config.Agent.ID
	}
	if agentFound {
		checks = append(checks, Check{Name: "agent", Status: CheckPass, Detail: "configured_agent_registered"})
	} else {
		checks = append(checks, Check{Name: "agent", Status: CheckFail, Detail: "configured_agent_missing"})
	}
	registeredTools := make(map[string]struct{})
	for _, descriptor := range application.runtime.Tools().Descriptors() {
		registeredTools[string(descriptor.ID())] = struct{}{}
	}
	toolsFound := true
	for _, configured := range config.Agent.Tools {
		_, exists := registeredTools[configured]
		toolsFound = toolsFound && exists
	}
	if toolsFound {
		checks = append(checks, Check{Name: "tools", Status: CheckPass, Detail: "configured_tools_registered"})
	} else {
		checks = append(checks, Check{Name: "tools", Status: CheckFail, Detail: "configured_tool_missing"})
	}
	policyFound := false
	for _, id := range application.runtime.Tools().Policies() {
		policyFound = policyFound || string(id) == config.Policy.ID
	}
	if policyFound {
		checks = append(checks, Check{Name: "policy", Status: CheckPass, Detail: "configured_policy_registered"})
	} else {
		checks = append(checks, Check{Name: "policy", Status: CheckFail, Detail: "configured_policy_missing"})
	}

	providerID := pkgProvider.ID(config.Provider.ID)
	instance, instanceErr := application.runtime.Providers().Capabilities(ctx, providerID, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetInstance})
	if instanceErr != nil {
		checks = append(checks, Check{Name: "provider", Status: CheckFail, Detail: "instance_probe_failed"})
	} else {
		checks = append(checks, Check{Name: "provider", Status: CheckPass, Detail: fmt.Sprintf("capabilities_%d", len(instance.Capabilities))})
	}
	agentProfile := config.AgentProfile()
	model, modelErr := application.runtime.Providers().Capabilities(ctx, providerID, pkgProvider.CapabilityRequest{Target: pkgProvider.CapabilityTargetModel, Model: agentProfile.Model})
	if modelErr != nil {
		checks = append(checks, Check{Name: "model", Status: CheckFail, Detail: "model_probe_failed"})
	} else if !requiredModelCapabilities(model, agentProfile.Streaming) {
		checks = append(checks, Check{Name: "model", Status: CheckFail, Detail: "required_capability_unavailable"})
	} else if config.Version == productconfig.CandidateVersion &&
		pkgProvider.ValidateGenerationCapabilities(model, agentProfile.GenerationOptions()) != nil {
		checks = append(checks, Check{Name: "generation", Status: CheckFail, Detail: "generation_control_unavailable"})
	} else {
		checks = append(checks, Check{Name: "model", Status: CheckPass, Detail: "required_capabilities_available"})
		if config.Version == productconfig.CandidateVersion {
			checks = append(checks, Check{Name: "generation", Status: CheckPass, Detail: "generation_controls_available"})
		}
	}
	if err := application.Start(ctx); err != nil {
		checks = append(checks, Check{Name: "laravel", Status: CheckFail, Detail: "workspace_detection_failed"})
	} else {
		checks = append(checks, Check{Name: "laravel", Status: CheckPass, Detail: "workspace_detected"})
	}
	return checks
}

func requiredModelCapabilities(report pkgProvider.CapabilityReport, streaming bool) bool {
	required := map[pkgProvider.Capability]bool{
		pkgProvider.CapabilityCompletion:  false,
		pkgProvider.CapabilityToolCalling: false,
	}
	if streaming {
		required[pkgProvider.CapabilityStreaming] = false
	}
	for _, descriptor := range report.Capabilities {
		if _, exists := required[descriptor.Capability]; exists && descriptor.Support == pkgProvider.CapabilitySupported && descriptor.Availability == pkgProvider.CapabilityAvailabilityAvailable {
			required[descriptor.Capability] = true
		}
	}
	for _, available := range required {
		if !available {
			return false
		}
	}
	return true
}
