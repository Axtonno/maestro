package plugin

import "context"

// Loader constructs a plugin on demand. Implementations must not register or
// start the result; long-lived resources belong in lifecycle capabilities.
// Load may be invoked concurrently when callers request the same catalog ID,
// so implementations must protect any mutable factory-local state.
type Loader interface {
	Load(context.Context) (Plugin, error)
}

// LoaderFunc adapts a function to Loader.
type LoaderFunc func(context.Context) (Plugin, error)

func (f LoaderFunc) Load(ctx context.Context) (Plugin, error) {
	return f(ctx)
}

// Runtime registers plugins in the Runtime Core and maintains their dedicated
// index. Only plugins registered through this interface are resolved here.
type Runtime interface {
	Register(Plugin) error
	Resolve(ID) (Plugin, error)
	Has(ID) bool
	// Registered returns a defensive snapshot in successful registration
	// order. The relative order of concurrent registrations is unspecified.
	Registered() []ID

	RegisterLoader(ID, Loader) error
	// Load performs an independent factory and registration attempt. Concurrent
	// calls for the same ID may invoke the loader more than once, but at most one
	// registration can succeed.
	Load(context.Context, ID) (Plugin, error)
	// Available returns a defensive snapshot in successful loader registration
	// order. The relative order of concurrent registrations is unspecified.
	Available() []ID
}
