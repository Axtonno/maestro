package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

const maxApprovalInputBytes = 64

// TerminalApprover is a fail-closed, run-local approval adapter. It renders
// only prepared actions; invocation arguments and workspace/model content are
// deliberately unavailable to the renderer.
type TerminalApprover struct {
	input       *bufio.Reader
	output      io.Writer
	interactive bool
	mu          sync.Mutex
}

func NewTerminalApprover(input io.Reader, output io.Writer, interactive bool) *TerminalApprover {
	if input == nil {
		input = strings.NewReader("")
	}
	if output == nil {
		output = io.Discard
	}
	return &TerminalApprover{
		input: bufio.NewReaderSize(input, maxApprovalInputBytes), output: output,
		interactive: interactive,
	}
}

func (approver *TerminalApprover) Approve(ctx context.Context, request pkgTool.PermissionRequest) (pkgTool.Approval, error) {
	if approver == nil || ctx == nil || request.Validate() != nil {
		return pkgTool.Approval{}, pkgTool.ErrInvalidPermissionRequest
	}
	if err := ctx.Err(); err != nil {
		return pkgTool.Approval{}, err
	}
	approver.mu.Lock()
	defer approver.mu.Unlock()
	if !approver.interactive {
		return denyApproval("non_interactive")
	}

	approver.render(request)
	line, err := approver.readLine(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return pkgTool.Approval{}, err
		}
		return denyApproval("input_unavailable")
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "o", "once":
		return pkgTool.NewApproval(pkgTool.ApprovalAllow, "terminal_allow_once", "", pkgTool.GrantOneShot)
	case "r", "run":
		return pkgTool.NewApproval(pkgTool.ApprovalAllow, "terminal_allow_run", "", pkgTool.GrantRun)
	case "", "d", "deny":
		return denyApproval("terminal_deny")
	default:
		return denyApproval("input_invalid")
	}
}

func (approver *TerminalApprover) render(request pkgTool.PermissionRequest) {
	safeWrite := func(format string, values ...any) {
		defer func() { _ = recover() }()
		_, _ = fmt.Fprintf(approver.output, format, values...)
	}
	safeWrite("approval required\n")
	safeWrite("  subject: %s\n", request.Subject())
	if target, ok := request.ModelTarget(); ok {
		safeWrite("  model: %s/%s\n", target.Provider(), target.Model())
	}
	if prepared, ok := request.Prepared(); ok {
		safeWrite("  tool: %s\n", prepared.Invocation().Tool())
	}
	for index, action := range request.Actions() {
		if action.Effect() == pkgTool.EffectModelDisclose {
			if disclosure, ok := request.Disclosure(); ok {
				safeWrite("  action %d: %s workspace=%s sections=%d tokens=%d bytes=%d\n",
					index+1, action.Effect(), disclosure.Workspace(), disclosure.Sections(),
					disclosure.Tokens(), disclosure.Bytes())
			}
			continue
		}
		if action.Workspace() != "" {
			safeWrite("  action %d: %s resource=%s workspace=%s\n", index+1, action.Effect(), action.Resource(), action.Workspace())
		} else {
			safeWrite("  action %d: %s resource=%s\n", index+1, action.Effect(), action.Resource())
		}
	}
	safeWrite("allow? [d]eny/[o]nce/[r]un (default deny): ")
}

func (approver *TerminalApprover) readLine(ctx context.Context) (string, error) {
	type result struct {
		line string
		err  error
	}
	completed := make(chan result, 1)
	go func() {
		line, err := approver.input.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) || len(line) > maxApprovalInputBytes {
			completed <- result{err: errors.New("approval input exceeds limit")}
			return
		}
		completed <- result{line: string(line), err: err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case value := <-completed:
		return value.line, value.err
	}
}

func denyApproval(reason string) (pkgTool.Approval, error) {
	return pkgTool.NewApproval(pkgTool.ApprovalDeny, reason, pkgTool.DenyTerminal, "")
}
