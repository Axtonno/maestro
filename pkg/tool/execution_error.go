package tool

import "fmt"

type ErrorKind string

const (
	ErrorInvalid    ErrorKind = "invalid"
	ErrorNotFound   ErrorKind = "not_found"
	ErrorPermission ErrorKind = "permission"
	ErrorLimit      ErrorKind = "limit"
	ErrorCanceled   ErrorKind = "canceled"
	ErrorDeadline   ErrorKind = "deadline_exceeded"
	ErrorExecution  ErrorKind = "execution"
	ErrorInternal   ErrorKind = "internal"
)

type ExecutionError struct {
	Kind   ErrorKind
	Run    RunID
	Tool   ID
	Call   CallID
	Reason string
	cause  error
}

func NewExecutionError(
	kind ErrorKind,
	run RunID,
	tool ID,
	call CallID,
	reason string,
	cause error,
) *ExecutionError {
	if !safeCode(reason) {
		reason = "internal"
	}
	return &ExecutionError{Kind: kind, Run: run, Tool: tool, Call: call, Reason: reason, cause: cause}
}

func (err *ExecutionError) Error() string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"tool %q call %q in run %q failed (%s:%s)",
		err.Tool,
		err.Call,
		err.Run,
		err.Kind,
		err.Reason,
	)
}

func (err *ExecutionError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

func (err *ExecutionError) Is(target error) bool {
	if err == nil {
		return false
	}
	switch err.Kind {
	case ErrorInvalid:
		return target == ErrInvalidInvocation || target == ErrInvalidExecutionRequest
	case ErrorNotFound:
		return target == ErrNotFound
	case ErrorPermission:
		return target == ErrPermissionDenied
	case ErrorLimit:
		return target == ErrLimitExceeded
	case ErrorExecution:
		return target == ErrExecutionFailed
	default:
		return false
	}
}
