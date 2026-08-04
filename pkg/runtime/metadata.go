package runtime

type Metadata struct {
	ID ComponentID

	Name string

	Version string

	Description string

	Dependencies []Dependency

	Capabilities []Capability
}