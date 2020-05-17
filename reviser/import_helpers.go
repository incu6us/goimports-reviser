package reviser

import (
	"go/ast"
	"path"
	"strconv"
	"strings"
)

// UsesImport is a similar to astutil.UsesImport but with skipping version in the import path
func UsesImport(f *ast.File, importPath string) bool {
	importIdentNames := make(map[string]struct{}, len(f.Imports))

	for _, spec := range f.Imports {
		name := spec.Name.String()
		switch name {
		case "<nil>":
			importIdentNames[AliasFromImportPath(importPath)] = struct{}{}
		case "_", ".":
			return true
		default:
			importIdentNames[name] = struct{}{}
		}
	}

	var used bool
	ast.Walk(visitFn(func(node ast.Node) {
		sel, ok := node.(*ast.SelectorExpr)
		if ok {
			ident, ok := sel.X.(*ast.Ident)
			if ok {
				if _, ok := importIdentNames[ident.Name]; ok {
					used = true
					return
				}
			}
		}
	}), f)

	return used
}

// AliasFromImportPath will get package alias if it has a version in the end of the path (ex.: github.com/go-pg/pg/v9)
func AliasFromImportPath(importPath string) string {
	base := path.Base(importPath)
	if strings.HasPrefix(base, "v") {
		if _, err := strconv.Atoi(base[1:]); err == nil {
			dir := path.Dir(importPath)
			if dir != "." {
				base = path.Base(dir)
			}
		}
	}

	return base
}

type visitFn func(node ast.Node)

func (f visitFn) Visit(node ast.Node) ast.Visitor {
	f(node)
	return f
}
