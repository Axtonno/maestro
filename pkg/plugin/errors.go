package plugin

import "errors"

var (
	ErrInvalidPlugin     = errors.New("invalid plugin")
	ErrAlreadyRegistered = errors.New("plugin already registered")
	ErrNotFound          = errors.New("plugin not found")
)
