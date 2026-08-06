package llamacpp

import (
	"errors"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "http://localhost:8080"
	DefaultTimeout = 30 * time.Second
)

var ErrInvalidConfig = errors.New("invalid llama.cpp provider configuration")

type Config struct {
	// BaseURL is the llama-server origin and must not include /v1.
	BaseURL string

	// Timeout configures the client created by New when HTTPClient is nil.
	// A zero value selects DefaultTimeout.
	Timeout time.Duration

	DefaultModel string

	// APIKey enables Bearer authentication when it is not empty.
	APIKey string

	// HTTPClient, when provided, is used unchanged and Timeout is ignored.
	HTTPClient *http.Client
}
