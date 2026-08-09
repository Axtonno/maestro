package gestor

import "context"

// Source discovers declarations from one authoritative Maestro subsystem.
type Source interface {
	ID() SourceID
	Discover(context.Context) ([]Descriptor, error)
}

// Registry owns discovery sources and publishes all-or-nothing snapshots.
type Registry interface {
	RegisterSource(Source) error
	Refresh(context.Context) error
	Invalidate()
	Snapshot() Snapshot
}

// Resolver reads a single snapshot and applies only explicit query filters and
// preferences. It does not execute discovery or resolved capabilities.
type Resolver interface {
	Candidates(Query) ([]Descriptor, error)
	Resolve(Query) (Resolution, error)
}
