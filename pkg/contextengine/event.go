package contextengine

import pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"

const (
	EventIndexStarted   = "context.index.started"
	EventIndexCompleted = "context.index.completed"
	EventIndexFailed    = "context.index.failed"
	EventBuildStarted   = "context.build.started"
	EventBuildCompleted = "context.build.completed"
	EventBuildFailed    = "context.build.failed"
	EventCacheObserved  = "context.cache.observed"
)

const CapabilityWorkspaceProvider pkgRuntime.Capability = "context.workspace-provider"

type EventFailure string

const (
	EventFailureNone      EventFailure = ""
	EventFailureCanceled  EventFailure = "canceled"
	EventFailureDeadline  EventFailure = "deadline_exceeded"
	EventFailureInvalid   EventFailure = "invalid"
	EventFailureNotFound  EventFailure = "not_found"
	EventFailureSource    EventFailure = "source_failure"
	EventFailureAnalyzer  EventFailure = "analyzer_failure"
	EventFailureEmbedding EventFailure = "embedding_failure"
	EventFailureEstimator EventFailure = "estimator_failure"
	EventFailureBudget    EventFailure = "budget_exceeded"
	EventFailureInternal  EventFailure = "internal"
)

// EventPayload is redacted by construction. It contains no root, document
// path, query, text, embedding, provider/model target, or external error text.
type EventPayload struct {
	Workspace     WorkspaceID
	Generation    uint64
	DocumentCount int
	AnalysisCount int
	SectionCount  int
	UsedTokens    int
	Cache         CacheStats
	Failure       EventFailure
}

type Event struct {
	Topic string
	Data  EventPayload
}

func (event Event) Name() string { return event.Topic }
func (event Event) Payload() any { return event.Data }
