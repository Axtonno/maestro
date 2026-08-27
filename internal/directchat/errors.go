package directchat

import "errors"

var (
	ErrInvalidRequest        = errors.New("invalid direct chat request")
	ErrProfileRequired       = errors.New("direct chat profile required")
	ErrFileNotAllowed        = errors.New("direct chat file not allowed")
	ErrProviderUnavailable   = errors.New("direct chat provider unavailable")
	ErrCapabilityUnsupported = errors.New("direct chat capability unsupported")
	ErrResponseInvalid       = errors.New("direct chat response invalid")
	ErrLimitExceeded         = errors.New("direct chat limit exceeded")
)
