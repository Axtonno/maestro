//go:build maestro_m42_evaluation

package directchat

import (
	"context"
	"testing"

	"github.com/antonio-cafeo/maestro/internal/productconfig"
	pkgProvider "github.com/antonio-cafeo/maestro/pkg/provider"
)

func TestM42ChatKeepsSharedModelResident(t *testing.T) {
	config := v4ChatConfig(t.TempDir())
	config.DirectChat.Model = productconfig.SingleModelEvaluationModel
	config.DirectChat.Digest = productconfig.SingleModelEvaluationDigest
	provider := &v4Provider{models: []pkgProvider.ModelInfo{{
		Model:  pkgProvider.Model{ID: productconfig.SingleModelEvaluationModel},
		Digest: productconfig.SingleModelEvaluationDigest,
		State:  pkgProvider.ModelStateLoaded,
	}}}
	service, err := Build(config, Dependencies{ProviderFactory: func(productconfig.Config, string) (pkgProvider.Provider, error) {
		return provider, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Execute(context.Background(), Request{Question: "Are you ready?"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Model != productconfig.SingleModelEvaluationModel || len(provider.requests) != 1 ||
		provider.requests[0].Model != productconfig.SingleModelEvaluationModel || len(provider.unloads) != 0 {
		t.Fatalf("single-model routing drift: result=%#v requests=%#v unloads=%v", result, provider.requests, provider.unloads)
	}
}
