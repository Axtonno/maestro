package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/antonio-cafeo/maestro"
	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

func TestTerminalApproverChoicesAreExplicitAndScoped(t *testing.T) {
	request := approvalModelRequest(t)
	for _, testCase := range []struct {
		name        string
		input       string
		interactive bool
		kind        pkgTool.ApprovalKind
		scope       pkgTool.GrantScope
		reason      string
	}{
		{name: "one shot", input: "o\n", interactive: true, kind: pkgTool.ApprovalAllow, scope: pkgTool.GrantOneShot, reason: "terminal_allow_once"},
		{name: "run", input: "run\n", interactive: true, kind: pkgTool.ApprovalAllow, scope: pkgTool.GrantRun, reason: "terminal_allow_run"},
		{name: "default deny", input: "\n", interactive: true, kind: pkgTool.ApprovalDeny, reason: "terminal_deny"},
		{name: "explicit deny", input: "deny\n", interactive: true, kind: pkgTool.ApprovalDeny, reason: "terminal_deny"},
		{name: "invalid", input: "yes\n", interactive: true, kind: pkgTool.ApprovalDeny, reason: "input_invalid"},
		{name: "eof", interactive: true, kind: pkgTool.ApprovalDeny, reason: "input_unavailable"},
		{name: "non interactive", input: "run\n", kind: pkgTool.ApprovalDeny, reason: "non_interactive"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			approval, err := NewTerminalApprover(strings.NewReader(testCase.input), &output, testCase.interactive).Approve(t.Context(), request)
			if err != nil || approval.Kind() != testCase.kind || approval.Scope() != testCase.scope || approval.Reason() != testCase.reason {
				t.Fatalf("approval=%#v err=%v", approval, err)
			}
			if testCase.kind == pkgTool.ApprovalDeny && approval.Disposition() != pkgTool.DenyTerminal {
				t.Fatalf("deny is not terminal: %#v", approval)
			}
			if testCase.interactive != strings.Contains(output.String(), "approval required") {
				t.Fatalf("unexpected output %q", output.String())
			}
		})
	}
}

func TestTerminalApproverRendersActionsWithoutDisclosureFingerprint(t *testing.T) {
	fingerprint := pkgTool.Fingerprint(strings.Repeat("a", 64))
	manifest, err := pkgTool.NewDisclosureManifest("laravel", 3, 2, 150, 600, fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	target, _ := pkgTool.NewModelTarget("ollama", "llama3.1:8b")
	request, _ := pkgTool.NewModelPermissionRequest("policy.release", "run-approval", target, &manifest)
	var output bytes.Buffer
	approval, err := NewTerminalApprover(strings.NewReader("o\n"), &output, true).Approve(t.Context(), request)
	if err != nil || approval.Kind() != pkgTool.ApprovalAllow {
		t.Fatalf("approval=%#v err=%v", approval, err)
	}
	got := output.String()
	for _, expected := range []string{"subject: model", "model: ollama/llama3.1:8b", "model.invoke", "model.disclose", "workspace=laravel", "sections=2", "tokens=150", "bytes=600"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("output %q lacks %q", got, expected)
		}
	}
	if strings.Contains(got, string(fingerprint)) {
		t.Fatalf("disclosure fingerprint leaked: %q", got)
	}
}

func TestTerminalApproverCancellationAndDeadlineInterruptInput(t *testing.T) {
	for _, testCase := range []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		want error
	}{
		{name: "canceled", ctx: func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, want: context.Canceled},
		{name: "deadline", ctx: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Millisecond)
		}, want: context.DeadlineExceeded},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			ctx, cancel := testCase.ctx()
			if testCase.want == context.Canceled {
				go func() {
					time.Sleep(time.Millisecond)
					cancel()
				}()
			} else {
				defer cancel()
			}
			_, err := NewTerminalApprover(reader, io.Discard, true).Approve(ctx, approvalModelRequest(t))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("expected %v, got %v", testCase.want, err)
			}
		})
	}
}

func TestTerminalApproverRunGrantIsReusedAndOneShotIsNot(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		input       string
		policy      pkgTool.PolicyID
		wantPrompts int
	}{
		{name: "run", input: "r\n", policy: "policy.run-approval", wantPrompts: 1},
		{name: "one shot", input: "o\no\n", policy: "policy.once-approval", wantPrompts: 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			config := testConfig(t.TempDir())
			config.Policy.ID = string(testCase.policy)
			config.Policy.Model = "prompt"
			policy, err := NewProductPolicy(config)
			if err != nil {
				t.Fatal(err)
			}
			runtime := maestro.New().Tools()
			if err := runtime.RegisterPolicy(policy); err != nil {
				t.Fatal(err)
			}
			target, _ := pkgTool.NewModelTarget("ollama", "fixture-model")
			request, _ := pkgTool.NewModelPermissionRequest(testCase.policy, "run-approval", target, nil)
			var output bytes.Buffer
			approver := NewTerminalApprover(strings.NewReader(testCase.input), &output, true)
			for range 2 {
				decision, err := runtime.AuthorizeModel(t.Context(), request, approver)
				if err != nil || decision.Kind() != pkgTool.DecisionAllow {
					t.Fatalf("decision=%#v err=%v", decision, err)
				}
			}
			if got := strings.Count(output.String(), "approval required"); got != testCase.wantPrompts {
				t.Fatalf("prompts=%d output=%q", got, output.String())
			}
		})
	}
}

func approvalModelRequest(t *testing.T) pkgTool.PermissionRequest {
	t.Helper()
	target, err := pkgTool.NewModelTarget("ollama", "fixture-model")
	if err != nil {
		t.Fatal(err)
	}
	request, err := pkgTool.NewModelPermissionRequest("policy.release", "run-approval", target, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
