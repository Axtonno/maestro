package contextengine

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	pkgContext "github.com/antonio-cafeo/maestro/pkg/contextengine"
)

const (
	GoAnalyzerID      pkgContext.AnalyzerID = "context.go-ast"
	GoAnalyzerVersion                       = "1"
)

type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer { return &GoAnalyzer{} }

func (*GoAnalyzer) ID() pkgContext.AnalyzerID { return GoAnalyzerID }
func (*GoAnalyzer) Version() string           { return GoAnalyzerVersion }

func (*GoAnalyzer) Supports(document pkgContext.Document) bool {
	return document.Language() == "go" && document.MediaType() == "text/x-go"
}

func (analyzer *GoAnalyzer) Analyze(ctx context.Context, document pkgContext.Document) (pkgContext.Analysis, error) {
	if ctx == nil {
		return pkgContext.Analysis{}, fmt.Errorf("analyze Go document with nil context: %w", pkgContext.ErrInvalidAnalyzer)
	}
	if err := ctx.Err(); err != nil {
		return pkgContext.Analysis{}, err
	}
	if !analyzer.Supports(document) {
		return pkgContext.Analysis{}, pkgContext.ErrUnsupported
	}

	fileSet := token.NewFileSet()
	file, parseErr := parser.ParseFile(fileSet, string(document.Path()), document.Content(), parser.AllErrors)
	diagnostics := make([]pkgContext.Diagnostic, 0, 1)
	if parseErr != nil {
		diagnostics = append(diagnostics, pkgContext.Diagnostic{
			Path: document.Path(), Severity: pkgContext.DiagnosticError, Code: "go_parse_error",
		})
	}
	if file == nil {
		return pkgContext.NewAnalysis(document, analyzer.ID(), analyzer.Version(), nil, nil, nil, diagnostics)
	}

	symbols := make([]pkgContext.Symbol, 0)
	relations := make([]pkgContext.Relation, 0)
	chunks := make([]pkgContext.Chunk, 0, len(file.Decls))
	nextSymbol := func(name string, kind pkgContext.SymbolKind, sourceRange pkgContext.SourceRange, container string) string {
		id := fmt.Sprintf("symbol_%d", len(symbols))
		symbols = append(symbols, pkgContext.Symbol{ID: id, Name: name, Kind: kind, Range: sourceRange, Container: container})
		return id
	}

	packageRange := nodeRange(fileSet, file.Name, document.SizeBytes())
	packageID := nextSymbol(file.Name.Name, pkgContext.SymbolPackage, packageRange, "")
	typeIDs := make(map[string]string)

	for declarationIndex, declaration := range file.Decls {
		if err := ctx.Err(); err != nil {
			return pkgContext.Analysis{}, err
		}
		declarationRange := nodeRange(fileSet, declaration, document.SizeBytes())
		chunks = append(chunks, pkgContext.Chunk{
			ID: fmt.Sprintf("chunk_%d", declarationIndex), Kind: "declaration", Range: declarationRange,
		})
		switch typed := declaration.(type) {
		case *ast.GenDecl:
			for _, specification := range typed.Specs {
				switch value := specification.(type) {
				case *ast.ImportSpec:
					name := strings.Trim(value.Path.Value, "\"")
					if name == "" {
						continue
					}
					id := nextSymbol(name, pkgContext.SymbolImport, nodeRange(fileSet, value, document.SizeBytes()), packageID)
					relations = append(relations,
						pkgContext.Relation{From: packageID, To: id, Kind: pkgContext.RelationContains},
						pkgContext.Relation{From: packageID, To: name, Kind: pkgContext.RelationImports},
					)
				case *ast.TypeSpec:
					id := nextSymbol(value.Name.Name, pkgContext.SymbolType, nodeRange(fileSet, value, document.SizeBytes()), packageID)
					typeIDs[value.Name.Name] = id
					relations = append(relations, pkgContext.Relation{From: packageID, To: id, Kind: pkgContext.RelationContains})
					if structure, ok := value.Type.(*ast.StructType); ok {
						for _, field := range structure.Fields.List {
							for _, name := range field.Names {
								fieldID := nextSymbol(name.Name, pkgContext.SymbolField, nodeRange(fileSet, name, document.SizeBytes()), id)
								relations = append(relations, pkgContext.Relation{From: id, To: fieldID, Kind: pkgContext.RelationContains})
							}
						}
					}
				case *ast.ValueSpec:
					kind := pkgContext.SymbolVariable
					if typed.Tok == token.CONST {
						kind = pkgContext.SymbolConstant
					}
					for _, name := range value.Names {
						id := nextSymbol(name.Name, kind, nodeRange(fileSet, name, document.SizeBytes()), packageID)
						relations = append(relations, pkgContext.Relation{From: packageID, To: id, Kind: pkgContext.RelationContains})
					}
				}
			}
		case *ast.FuncDecl:
			kind := pkgContext.SymbolFunction
			container := packageID
			if typed.Recv != nil && len(typed.Recv.List) > 0 {
				kind = pkgContext.SymbolMethod
				if receiver := receiverName(typed.Recv.List[0].Type); receiver != "" {
					if id, exists := typeIDs[receiver]; exists {
						container = id
					}
				}
			}
			id := nextSymbol(typed.Name.Name, kind, declarationRange, container)
			relations = append(relations, pkgContext.Relation{From: container, To: id, Kind: pkgContext.RelationContains})
		}
	}

	return pkgContext.NewAnalysis(document, analyzer.ID(), analyzer.Version(), symbols, relations, chunks, diagnostics)
}

func nodeRange(fileSet *token.FileSet, node ast.Node, size int) pkgContext.SourceRange {
	start := fileSet.PositionFor(node.Pos(), false).Offset
	end := fileSet.PositionFor(node.End(), false).Offset
	if start < 0 {
		start = 0
	}
	if end > size {
		end = size
	}
	if end <= start {
		end = start + 1
		if end > size {
			start, end = 0, size
		}
	}
	return pkgContext.SourceRange{Start: start, End: end}
}

func receiverName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverName(typed.X)
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	default:
		return ""
	}
}
