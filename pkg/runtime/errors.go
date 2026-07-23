package runtime

import "errors"

var (
	ErrAlreadyRegistered = errors.New("component already registered")

	ErrNotFound = errors.New("component not found")

	ErrInvalidState = errors.New("invalid lifecycle state")

	ErrAlreadyStarted = errors.New("runtime already started")

	ErrAlreadyStopped = errors.New("runtime already stopped")
)