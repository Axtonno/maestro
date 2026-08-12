package tool

import (
	"context"
	"fmt"
	"sync/atomic"

	pkgTool "github.com/antonio-cafeo/maestro/pkg/tool"
)

type permit struct {
	issuer     *Runtime
	run        pkgTool.RunID
	permission pkgTool.Fingerprint
	prepared   pkgTool.Fingerprint
	consumed   atomic.Bool
}

func (runtime *Runtime) issue(request pkgTool.PermissionRequest) *permit {
	prepared, _ := request.Prepared()
	return &permit{
		issuer: runtime, run: request.Run(), permission: request.Fingerprint(),
		prepared: prepared.Fingerprint(),
	}
}

func (runtime *Runtime) consume(
	prepared pkgTool.PreparedInvocation,
	permission pkgTool.Fingerprint,
	candidate *permit,
) error {
	if candidate == nil || candidate.issuer != runtime ||
		candidate.run != prepared.Invocation().Run() ||
		candidate.permission != permission ||
		candidate.prepared != prepared.Fingerprint() ||
		!candidate.consumed.CompareAndSwap(false, true) {
		return fmt.Errorf("permit is absent, mismatched, or already consumed: %w", pkgTool.ErrPermissionDenied)
	}
	return nil
}

func (runtime *Runtime) execute(
	ctx context.Context,
	candidate pkgTool.Tool,
	prepared pkgTool.PreparedInvocation,
	permission pkgTool.Fingerprint,
	authority *permit,
) (result pkgTool.Result, err error) {
	if err := runtime.consume(prepared, permission, authority); err != nil {
		return pkgTool.Result{}, err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			result = pkgTool.Result{}
			err = fmt.Errorf("tool execute panicked: %w", pkgTool.ErrExecutionFailed)
		}
	}()
	return candidate.Execute(ctx, prepared)
}
