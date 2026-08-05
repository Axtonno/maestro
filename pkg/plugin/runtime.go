package plugin

import "context"

// Loader constructs a plugin on demand. Implementations must not register or
// start the result; long-lived resources belong in lifecycle capabilities.
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
	Registered() []ID

	RegisterLoader(ID, Loader) error
	Load(context.Context, ID) (Plugin, error)
	Available() []ID
}
