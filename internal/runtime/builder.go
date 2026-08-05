package runtime

import "fmt"

type builder struct {
	resolver  *resolver
	validator *validator
}

func newBuilder(
	resolver *resolver,
	validator *validator,
) *builder {
	return &builder{
		resolver:  resolver,
		validator: validator,
	}
}

func (b *builder) Build() (*graph, error) {
	dependencyGraph, err := b.resolver.Resolve()
	if err != nil {
		return nil, fmt.Errorf(
			"build dependency graph: resolve dependencies: %w",
			err,
		)
	}

	if err := b.validator.Validate(dependencyGraph); err != nil {
		return nil, fmt.Errorf(
			"build dependency graph: validate graph: %w",
			err,
		)
	}

	return dependencyGraph, nil
}
