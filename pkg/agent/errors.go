package agent

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAgentID       = errors.New("invalid agent ID")
	ErrInvalidVersion       = errors.New("invalid agent version")
	ErrInvalidDescriptor    = errors.New("invalid agent descriptor")
	ErrInvalidLimits        = errors.New("invalid agent limits")
	ErrInvalidRequest       = errors.New("invalid agent request")
	ErrInvalidPlan          = errors.New("invalid agent plan")
	ErrInvalidStep          = errors.New("invalid agent plan step")
	ErrInvalidTransition    = errors.New("invalid agent state transition")
	ErrInvalidSession       = errors.New("invalid agent session snapshot")
	ErrInvalidResult        = errors.New("invalid agent result")
	ErrInvalidAgent         = errors.New("invalid agent")
	ErrAlreadyRegistered    = errors.New("agent already registered")
	ErrNotFound             = errors.New("agent not found")
	ErrSessionNotFound      = errors.New("agent session not found")
	ErrSessionActive        = errors.New("agent session already has an active coordinator")
	ErrSessionTerminal      = errors.New("agent session is terminal")
	ErrPlanningFailed       = errors.New("agent planning failed")
	ErrProviderFailed       = errors.New("agent provider operation failed")
	ErrToolFailed           = errors.New("agent tool operation failed")
	ErrMutationFailed       = errors.New("agent workspace mutation failed")
	ErrContextRefreshFailed = errors.New("agent context refresh failed")
	ErrPermissionDenied     = errors.New("agent permission denied")
	ErrLimitExceeded        = errors.New("agent limit exceeded")
	ErrRunCanceled          = errors.New("agent run canceled")
)

type ErrorKind string

const (
	ErrorInvalid    ErrorKind = "invalid"
	ErrorNotFound   ErrorKind = "not_found"
	ErrorPlanning   ErrorKind = "planning"
	ErrorProvider   ErrorKind = "provider"
	ErrorTool       ErrorKind = "tool"
	ErrorPermission ErrorKind = "permission"
	ErrorLimit      ErrorKind = "limit"
	ErrorCanceled   ErrorKind = "canceled"
	ErrorDeadline   ErrorKind = "deadline_exceeded"
	ErrorInternal   ErrorKind = "internal"
)

type RunError struct {
	Kind   ErrorKind
	Run    RunID
	Agent  ID
	Reason string
	cause  error
}

func NewRunError(kind ErrorKind, run RunID, agent ID, reason string, cause error) *RunError {
	if !safeCode(reason) {
		reason = "internal"
	}
	return &RunError{Kind: kind, Run: run, Agent: agent, Reason: reason, cause: cause}
}

func (err *RunError) Error() string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("agent %q run %q failed (%s:%s)", err.Agent, err.Run, err.Kind, err.Reason)
}

func (err *RunError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

func (err *RunError) Is(target error) bool {
	if err == nil {
		return false
	}
	switch err.Kind {
	case ErrorInvalid:
		return target == ErrInvalidRequest
	case ErrorNotFound:
		return target == ErrNotFound
	case ErrorPlanning:
		return target == ErrPlanningFailed
	case ErrorProvider:
		return target == ErrProviderFailed
	case ErrorTool:
		return target == ErrToolFailed
	case ErrorPermission:
		return target == ErrPermissionDenied
	case ErrorLimit:
		return target == ErrLimitExceeded
	case ErrorCanceled, ErrorDeadline:
		return target == ErrRunCanceled
	default:
		return false
	}
}
