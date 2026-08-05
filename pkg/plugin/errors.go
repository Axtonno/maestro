package plugin

import "errors"

var (
	ErrInvalidPlugin           = errors.New("invalid plugin")
	ErrInvalidManifest         = errors.New("invalid plugin manifest")
	ErrIncompatible            = errors.New("incompatible plugin")
	ErrAlreadyRegistered       = errors.New("plugin already registered")
	ErrNotFound                = errors.New("plugin not found")
	ErrInvalidLoader           = errors.New("invalid plugin loader")
	ErrLoaderAlreadyRegistered = errors.New("plugin loader already registered")
	ErrLoaderNotFound          = errors.New("plugin loader not found")
	ErrLoadFailed              = errors.New("plugin load failed")
)
