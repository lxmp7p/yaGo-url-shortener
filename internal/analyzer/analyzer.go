package analyzer

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "projectlint",
	Doc:  "checks forbidden panic, log.Fatal and os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			filename := pass.Fset.File(file.Pos()).Name()
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}

			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			isMain := file.Name.Name == "main" && fn.Name.Name == "main"

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				checkPanic(call, pass)
				checkExit(call, pass, isMain)

				return true
			})
		}
	}
	return nil, nil
}

func checkPanic(call *ast.CallExpr, pass *analysis.Pass) {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "panic" {
			pass.Reportf(call.Pos(), "forbiden panic")
		}
	}
}

func checkExit(call *ast.CallExpr, pass *analysis.Pass, isMain bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}

	obj, ok := pass.TypesInfo.Uses[pkg]
	if !ok {
		return
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return
	}

	switch {
	case pkgName.Imported().Path() == "os" && sel.Sel.Name == "Exit" && !isMain:
		pass.Reportf(call.Pos(), "os exit")

	case pkgName.Imported().Path() == "log" && sel.Sel.Name == "Fatal" && !isMain:
		pass.Reportf(call.Pos(), "log.Fatal is forbidden outside main")
	}
}
