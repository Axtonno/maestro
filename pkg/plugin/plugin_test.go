package plugin_test

import (
	"fmt"

	"github.com/antonio-cafeo/maestro"
	pkgRuntime "github.com/antonio-cafeo/maestro/pkg/runtime"
)

type frameworkPlugin struct{}

func (p *frameworkPlugin) Metadata() pkgRuntime.Metadata {
	return pkgRuntime.Metadata{
		ID:      "laravel",
		Name:    "Laravel",
		Version: "1.0.0",
	}
}

func ExampleRuntime() {
	runtime := maestro.New()

	if err := runtime.Plugins().Register(&frameworkPlugin{}); err != nil {
		panic(err)
	}

	plugin, err := runtime.Plugins().Resolve("laravel")
	if err != nil {
		panic(err)
	}

	fmt.Println(plugin.Metadata().Name)
	// Output: Laravel
}
