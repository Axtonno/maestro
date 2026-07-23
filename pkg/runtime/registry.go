package runtime

type Registry interface {
	Register(Service) error

	Resolve(name string) (Service, error)

	Has(name string) bool
}