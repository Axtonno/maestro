package controlledmutation

import "errors"

var (
	ErrInvalidRequest        = errors.New("controlled mutation request invalid")
	ErrProfileRequired       = errors.New("controlled mutation profile required")
	ErrTTYRequired           = errors.New("controlled mutation requires a TTY")
	ErrProviderUnavailable   = errors.New("controlled mutation provider unavailable")
	ErrCapabilityUnsupported = errors.New("controlled mutation capability unsupported")
	ErrModelIdentity         = errors.New("controlled mutation model identity mismatch")
	ErrResponseInvalid       = errors.New("controlled mutation response invalid")
	ErrInsufficientInfo      = errors.New("controlled mutation insufficient information")
	ErrApprovalRejected      = errors.New("controlled mutation approval rejected")
	ErrStaleSource           = errors.New("controlled mutation source stale")
	ErrExecutionFailed       = errors.New("controlled mutation execution failed")
)
