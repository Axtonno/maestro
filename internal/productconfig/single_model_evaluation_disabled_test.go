//go:build !maestro_m42_evaluation

package productconfig

import (
	"errors"
	"testing"
)

func TestDefaultBuildRejectsExactSingleModelEvaluationProfile(t *testing.T) {
	path := writeV4(t, singleModelV4(t.TempDir()))
	if _, err := LoadChat(path); !errors.Is(err, ErrInvalid) {
		t.Fatalf("default build accepted M42 evaluation profile for chat: %v", err)
	}
	if _, err := LoadMutation(path); !errors.Is(err, ErrInvalid) {
		t.Fatalf("default build accepted M42 evaluation profile for mutation: %v", err)
	}
}
