package exit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check for using os.Exit in main package in main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.File:
				if n.Name.Name != "main" {
					return false
				}
			case *ast.FuncDecl:
				if n.Name.Name != "main" {
					return false
				}
			case *ast.SelectorExpr:
				pack, ok := n.X.(*ast.Ident)
				if !ok {
					return false
				}
				if pack.Name == "os" && n.Sel.Name == "Exit" {
					pass.Reportf(n.Pos(), "using os.Exit() is not allowed in main function of main package")
				}
				return false
			}
			return true
		})
	}

	return nil, nil
}
