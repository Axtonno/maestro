package directchat

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestProductionPackageExcludesAgenticRuntimeImports(t *testing.T) {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test source")
	}
	directory := filepath.Dir(source)
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info fs.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, parsedPackage := range packages {
		for filename, file := range parsedPackage.Files {
			for _, imported := range file.Imports {
				path, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				for _, forbidden := range []string{
					"github.com/antonio-cafeo/maestro/internal/agent",
					"github.com/antonio-cafeo/maestro/internal/application",
					"github.com/antonio-cafeo/maestro/internal/contextengine",
					"github.com/antonio-cafeo/maestro/internal/runtime",
					"github.com/antonio-cafeo/maestro/internal/tool",
					"github.com/antonio-cafeo/maestro/pkg/agent",
					"github.com/antonio-cafeo/maestro/pkg/tool",
				} {
					if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
						t.Errorf("production file %s imports forbidden agentic dependency %s", filepath.Base(filename), path)
					}
				}
			}
		}
	}
}
