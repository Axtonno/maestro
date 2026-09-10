//go:build maestro_m42_evaluation

package productconfig

import (
	"errors"
	"strings"
	"testing"
)

func TestM42BuildAcceptsOnlyFrozenSingleModelEvaluationIdentity(t *testing.T) {
	path := writeV4(t, singleModelV4(t.TempDir()))
	config, err := LoadMutation(path)
	if err != nil {
		t.Fatalf("load frozen M42 profile: %v", err)
	}
	if config.ProductProfile() != ProductProfileSingleModelEvaluation ||
		config.DirectChat.Model != SingleModelEvaluationModel ||
		config.ControlledMutation.Model != SingleModelEvaluationModel {
		t.Fatalf("unexpected M42 profile: %#v", config)
	}
	if _, err := LoadChat(path); err != nil {
		t.Fatalf("chat rejected frozen M42 profile: %v", err)
	}

	drifted := strings.Replace(singleModelV4(t.TempDir()), SingleModelEvaluationDigest, strings.Repeat("0", 64), 1)
	if _, err := LoadMutation(writeV4(t, drifted)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("M42 build accepted identity drift: %v", err)
	}
}
