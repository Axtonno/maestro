package plugin

// Runtime registers plugins in the Runtime Core and maintains their dedicated
// index. Only plugins registered through this interface are resolved here.
type Runtime interface {
	Register(Plugin) error
	Resolve(ID) (Plugin, error)
	Has(ID) bool
}
