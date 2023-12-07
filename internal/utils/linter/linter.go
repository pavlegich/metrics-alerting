// Пакет linter содержит собственный анализатор для multicheker
package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitChecker = &analysis.Analyzer{
	Name: "exitcheck",
	Doc: `check and prohibit use of a direct 
	os.Exit call in main()`,
	Run: run,
}

func isExitExpr(s *ast.SelectorExpr) bool {
	i, ok := s.X.(*ast.Ident)
	return ok && i.Name == "os" && s.Sel.Name == "Exit"
}

func findMainFunc(file *ast.File) (*ast.FuncDecl, bool) {
	if file.Name.Name != "main" {
		return nil, false
	}

	for _, d := range file.Decls {
		f, ok := d.(*ast.FuncDecl)
		if ok && f.Name.Name == "main" {
			return f, true
		}
	}
	return nil, false
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExpr := func(x *ast.CallExpr) {
		if s, ok := x.Fun.(*ast.SelectorExpr); ok {
			if isExitExpr(s) {
				pass.Reportf(s.Pos(), "os.Exit calls are prohibited in main()")
			}
		}
	}

	for _, file := range pass.Files {
		if mainFunc, ok := findMainFunc(file); ok {
			ast.Inspect(mainFunc, func(n ast.Node) bool {
				if x, ok := n.(*ast.CallExpr); ok {
					callExpr(x)
				}
				return true
			})
		}
	}
	return nil, nil
}
