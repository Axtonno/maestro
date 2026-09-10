//go:build maestro_m42_evaluation

package productconfig

// This build mode exists only to collect M42 qualification evidence. Build it
// with `go build -tags maestro_m42_evaluation ./cmd/maestro`. It must not be
// used for a public artifact or interpreted as setup/profile support.
const singleModelEvaluationEnabled = true
