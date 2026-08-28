package main

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestChatCommandDoesNotImportApplicationComposition(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test source")
	}
	filename := filepath.Join(filepath.Dir(source), "chat_command.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, imported := range parsed.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		if path == "github.com/antonio-cafeo/maestro/internal/application" {
			t.Fatalf("chat command imports the agentic application composition root")
		}
	}
}
