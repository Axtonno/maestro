package ollama

import (
	"errors"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "http://localhost:11434"
	DefaultTimeout = 30 * time.Second
)

var ErrInvalidConfig = errors.New("invalid Ollama provider configuration")

type Config struct {
	// BaseURL is the Ollama server origin and must not include /api.
	BaseURL string

	// Timeout configures the client created by New when HTTPClient is nil.
	// A zero value selects DefaultTimeout.
	Timeout time.Duration

	DefaultModel string

	// HTTPClient, when provided, is used unchanged and Timeout is ignored.
	HTTPClient *http.Client
}
