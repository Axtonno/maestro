package runtime

type Registry interface {
	Register(Component) error

	Resolve(ComponentID) (Component, error)

	Has(ComponentID) bool
}
