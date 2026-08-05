package laravel

import "errors"

var (
	ErrInvalidConfig           = errors.New("invalid Laravel plugin configuration")
	ErrNotDetected             = errors.New("Laravel application not detected")
	ErrInvalidComposerManifest = errors.New("invalid Composer manifest")
)
