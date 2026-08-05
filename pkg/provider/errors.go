package provider

import "errors"

var (
	ErrInvalidProvider       = errors.New("invalid provider")
	ErrAlreadyRegistered     = errors.New("provider already registered")
	ErrNotFound              = errors.New("provider not found")
	ErrDefaultNotConfigured  = errors.New("default provider not configured")
	ErrUnsupportedCapability = errors.New("provider capability not supported")
	ErrInvalidStream         = errors.New("invalid provider stream")
)
