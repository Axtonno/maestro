package runtime

type Registry interface {
	Register(Service) error

	Resolve(ComponentID) (Service, error)

	Has(ComponentID) bool
}