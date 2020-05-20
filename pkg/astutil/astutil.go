package astutil

import (
	"go/ast"
	"path"
	"strconv"
	"strings"
)

// UsesImport is a similar to astutil.UsesImport but with skipping version in the import path
func UsesImport(f *ast.File, importPath string) bool {
	importIdentNames := make(map[string]struct{}, len(f.Imports))

	var importSpec *ast.ImportSpec
	for _, spec := range f.Imports {
		name := spec.Name.String()
		switch name {
		case "<nil>":
			pkgName, _ := PackageNameFromImportPath(importPath)
			importIdentNames[pkgName] = struct{}{}
		case "_", ".":
			return true
		default:
			importIdentNames[name] = struct{}{}
		}

		if importPath == strings.Trim(spec.Path.Value, `"`) {
			importSpec = spec
		}
	}

	var used bool
	ast.Walk(visitFn(func(node ast.Node) {
		sel, ok := node.(*ast.SelectorExpr)
		if ok {
			ident, ok := sel.X.(*ast.Ident)
			if ok {
				if _, ok := importIdentNames[ident.Name]; ok {
					pkg, _ := PackageNameFromImportPath(importPath)
					if (ident.Name == pkg || ident.Name == importSpec.Name.String()) && ident.Obj == nil {
						used = true
						return
					}
				}
			}
		}
	}), f)

	return used
}

// PackageNameFromImportPath will return package alias name
// and true if it has a version suffix in the end of the path (ex.: github.com/go-pg/pg/v9)
func PackageNameFromImportPath(importPath string) (string, bool) {
	var hasVersionSuffix bool

	base := path.Base(importPath)
	if strings.HasPrefix(base, "v") {
		if _, err := strconv.Atoi(base[1:]); err == nil {
			hasVersionSuffix = true
			dir := path.Dir(importPath)
			if dir != "." {
				base = path.Base(dir)
			}
		}
	}

	return base, hasVersionSuffix
}

type visitFn func(node ast.Node)

func (f visitFn) Visit(node ast.Node) ast.Visitor {
	f(node)
	return f
}
