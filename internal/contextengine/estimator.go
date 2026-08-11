package contextengine

import (
	"context"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

const (
	UTF8EstimatorID      pkgContext.EstimatorID = "context.utf8-estimator"
	UTF8EstimatorVersion                        = "1"
)

type UTF8Estimator struct{}

func NewUTF8Estimator() *UTF8Estimator            { return &UTF8Estimator{} }
func (*UTF8Estimator) ID() pkgContext.EstimatorID { return UTF8EstimatorID }
func (*UTF8Estimator) Version() string            { return UTF8EstimatorVersion }
func (*UTF8Estimator) Estimate(ctx context.Context, text string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if text == "" {
		return 0, nil
	}
	return (len([]byte(text)) + 2) / 3, nil
}
