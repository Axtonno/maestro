package contextengine

import (
	"context"

	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

type embeddingRuntime interface {
	Embed(context.Context, pkgProvider.ID, pkgProvider.EmbeddingRequest) (pkgProvider.EmbeddingResponse, error)
}
