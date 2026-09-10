//go:build maestro_m42_evaluation

package controlledmutation

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestM42MutationKeepsSharedModelResidentAndPreservesDeny(t *testing.T) {
	root, _ := mutationWorkspace(t)
	config := fixtureConfig(root)
	config.DirectChat.Model = productconfig.SingleModelEvaluationModel
	config.DirectChat.Digest = productconfig.SingleModelEvaluationDigest
	provider := qualifiedProvider(`{"decision":"propose","new_text":"$workers = 8;"}`)
	service, err := Build(config, Dependencies{
		ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil },
		RunID:           func() (pkgTool.RunID, error) { return "m42-mutation-test", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Execute(context.Background(), Request{
		File: "app/Worker.php", StartLine: 2, EndLine: 2, Instruction: "set workers to 8",
		Approver: approverFunc(func(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
			return pkgTool.NewApproval(pkgTool.ApprovalDeny, "m42_test_deny", pkgTool.DenyTerminal, "")
		}),
	})
	if !errors.Is(err, ErrApprovalRejected) {
		t.Fatalf("mutation result: %v", err)
	}
	if len(provider.requests) != 1 || provider.requests[0].Model != productconfig.SingleModelEvaluationModel || len(provider.unloads) != 0 {
		t.Fatalf("single-model routing drift: requests=%#v unloads=%v", provider.requests, provider.unloads)
	}
}

func TestM42MutationAppliesOnlyAfterAllowOnceWithoutHandoff(t *testing.T) {
	root, file := mutationWorkspace(t)
	config := fixtureConfig(root)
	config.DirectChat.Model = productconfig.SingleModelEvaluationModel
	config.DirectChat.Digest = productconfig.SingleModelEvaluationDigest
	provider := qualifiedProvider(`{"decision":"propose","new_text":"$workers = 8;"}`)
	service, err := Build(config, Dependencies{
		ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) { return provider, nil },
		RunID:           func() (pkgTool.RunID, error) { return "m42-allow-test", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Execute(context.Background(), Request{
		File: "app/Worker.php", StartLine: 2, EndLine: 2, Instruction: "set workers to 8",
		Approver: approverFunc(func(context.Context, pkgTool.PermissionRequest) (pkgTool.Approval, error) {
			return pkgTool.NewApproval(pkgTool.ApprovalAllow, "m42_test_allow", "", pkgTool.GrantOneShot)
		}),
	})
	if err != nil || result.Terminal != "applied" || !result.Durable {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	content, readErr := os.ReadFile(file)
	if readErr != nil || string(content) != "<?php\n$workers = 8;\nreturn $workers;\n" {
		t.Fatalf("content=%q err=%v", content, readErr)
	}
	if len(provider.requests) != 1 || provider.requests[0].Model != productconfig.SingleModelEvaluationModel || len(provider.unloads) != 0 {
		t.Fatalf("single-model routing drift: requests=%#v unloads=%v", provider.requests, provider.unloads)
	}
}
