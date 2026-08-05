package laravel

import (
	"context"
	"fmt"

	pkgPlugin "github.com/antonio-cafeo/maestro/pkg/plugin"
)

var _ pkgPlugin.Loader = (*loader)(nil)

type loader struct {
	root string
}

func NewLoader(root string) pkgPlugin.Loader {
	return &loader{root: root}
}

func (l *loader) Load(ctx context.Context) (pkgPlugin.Plugin, error) {
	if ctx == nil {
		return nil, fmt.Errorf(
			"load Laravel plugin: context is nil: %w",
			pkgPlugin.ErrInvalidLoader,
		)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return New(l.root)
}
