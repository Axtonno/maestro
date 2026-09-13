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

	// LocalModelPath enables local GGUF attestation. When set, model
	// discovery verifies that llama-server loaded this exact regular file.
	LocalModelPath string

	// ModelDigest is the lowercase SHA-256 expected for LocalModelPath.
	ModelDigest string

	// ServerBuild is the exact llama-server build_info expected from /props.
	ServerBuild string

	// ContextWindow is the effective server-side context window expected from
	// /props. A zero value does not constrain it.
	ContextWindow int

	// APIKey enables Bearer authentication when it is not empty.
	APIKey string

	// HTTPClient, when provided, is used unchanged and Timeout is ignored.
	HTTPClient *http.Client
}
