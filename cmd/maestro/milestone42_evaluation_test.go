//go:build maestro_m42_evaluation

package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestM42CLIChatMutationChatUsesOneModelWithoutHandoff(t *testing.T) {
	configPath, file := newCLIV4Config(t)
	encoded, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	encoded = bytes.Replace(encoded, []byte(productconfig.QualifiedDirectChatModel), []byte(productconfig.SingleModelEvaluationModel), 1)
	encoded = bytes.Replace(encoded, []byte(productconfig.QualifiedDirectChatDigest), []byte(productconfig.SingleModelEvaluationDigest), 1)
	if err := os.WriteFile(configPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	chatResponse := func(content string) pkgProvider.CompletionResponse {
		return pkgProvider.CompletionResponse{
			Model:        productconfig.SingleModelEvaluationModel,
			Message:      pkgProvider.Message{Role: pkgProvider.RoleAssistant, Content: content},
			FinishReason: pkgProvider.FinishReasonStop,
		}
	}
	provider := &cliProvider{
		id: "ollama",
		responses: []pkgProvider.CompletionResponse{
			chatResponse("Observed facts\nReady\n\nPossible inferences\nNone\n\nInformation not determinable\nNone"),
			chatResponse(`{"decision":"propose","new_text":"$workers = 8;"}`),
			chatResponse("Observed facts\nStill ready\n\nPossible inferences\nNone\n\nInformation not determinable\nNone"),
		},
		discovered: []pkgProvider.ModelInfo{{
			Model:  pkgProvider.Model{ID: productconfig.SingleModelEvaluationModel},
			Digest: productconfig.SingleModelEvaluationDigest,
			State:  pkgProvider.ModelStateLoaded,
		}},
	}
	dependencies := cliTestDependencies(provider)
	dependencies.isTerminal = func(io.Reader) bool { return true }

	run := func(arguments []string) (int, string, string) {
		t.Helper()
		var stdout, stderr bytes.Buffer
		code := runWithIO(arguments, strings.NewReader(""), &stdout, &stderr, dependencies)
		return code, stdout.String(), stderr.String()
	}
	for index, arguments := range [][]string{
		{"chat", "--config", configPath, "Are you ready?"},
		{"mutate", "--preview", "--config", configPath, "--file", "app/Worker.php", "--lines", "2:2", "set workers to 8"},
		{"chat", "--config", configPath, "Are you still ready?"},
	} {
		code, stdout, stderr := run(arguments)
		if code != 0 {
			t.Fatalf("step %d: code=%d stdout=%q stderr=%q", index+1, code, stdout, stderr)
		}
		if !strings.Contains(stdout, "model\t"+productconfig.SingleModelEvaluationModel) {
			t.Fatalf("step %d used wrong model: %q", index+1, stdout)
		}
	}
	assertFile(t, file, "<?php\n$workers = 4;\nreturn $workers;\n")
	if len(provider.requests) != 3 || len(provider.unloaded) != 0 {
		t.Fatalf("requests=%#v unloaded=%v", provider.requests, provider.unloaded)
	}
	for _, request := range provider.requests {
		if request.Model != productconfig.SingleModelEvaluationModel {
			t.Fatalf("routing drift: %#v", provider.requests)
		}
	}

	code, stdout, stderr := run([]string{"doctor", "--mode", "mutation", "--config", configPath})
	if code != 0 || stderr != "" || !strings.Contains(stdout, "pass\tconfiguration\tschema_v4_single_model_evaluation") {
		t.Fatalf("doctor: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}
