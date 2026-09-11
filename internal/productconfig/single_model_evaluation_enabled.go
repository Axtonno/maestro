//go:build maestro_m42_evaluation

package productconfig

// This build mode reproduces the rejected M42 candidate. Build it with
// `go build -tags maestro_m42_evaluation ./cmd/maestro`. It must not be used
// for a public artifact or interpreted as setup/profile support.
const singleModelEvaluationEnabled = true
