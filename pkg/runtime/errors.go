package runtime

import "errors"

var (
	ErrAlreadyRegistered = errors.New("component already registered")
	ErrNotFound           = errors.New("component not found")
	ErrInvalidMetadata    = errors.New("invalid component metadata")
	ErrCyclicDependency   = errors.New("cyclic dependency")
	ErrInvalidState       = errors.New("invalid state")
	ErrAlreadyStarted     = errors.New("runtime already started")
	ErrAlreadyStopped     = errors.New("runtime already stopped")
)