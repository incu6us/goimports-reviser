package astutil

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	buildTagPrefix           = "//go:build"
	deprecatedBuildTagPrefix = "//+build"
)

// PackageImports is map of imports with their package names
type PackageImports map[string]string

// UsesImport is for analyze if the import dependency is in use
func UsesImport(f *ast.File, packageImports PackageImports, importPath string) bool {
	importIdentNames := make(map[string]struct{}, len(f.Imports))

	var importSpec *ast.ImportSpec
	for _, spec := range f.Imports {
		name := spec.Name.String()
		switch name {
		case "<nil>":
			pkgName := packageImports[importPath]
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
	ast.Walk(
		visitFn(
			func(node ast.Node) {
				sel, ok := node.(*ast.SelectorExpr)
				if ok {
					ident, ok := sel.X.(*ast.Ident)
					if ok {
						if _, ok := importIdentNames[ident.Name]; ok {
							pkg := packageImports[importPath]
							if (ident.Name == pkg || ident.Name == importSpec.Name.String()) && ident.Obj == nil {
								used = true
								return
							}
						}
					}
				}
			},
		), f,
	)

	return used
}

// LoadPackageDependencies will return all package's imports with it names:
//
//	key - package(ex.: github/pkg/errors), value - name(ex.: errors)
func LoadPackageDependencies(dir, buildTag string) (PackageImports, error) {
	cfg := &packages.Config{
		Dir:   dir,
		Tests: true,
		Mode:  packages.NeedName | packages.NeedImports,
	}

	if buildTag != "" {
		cfg.BuildFlags = []string{fmt.Sprintf(`-tags=%s`, buildTag)}
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		return PackageImports{}, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return PackageImports{}, errors.New("package has an errors")
	}

	result := PackageImports{}

	for _, pkg := range pkgs {
		for imprt, pkg := range pkg.Imports {
			result[imprt] = pkg.Name
		}
	}

	return result, nil
}

// ParseBuildTag parse `//+build ...` or `//go:build ` on a first line of *ast.File
func ParseBuildTag(f *ast.File) string {
	for _, g := range f.Comments {
		for _, c := range g.List {
			if !(strings.HasPrefix(c.Text, buildTagPrefix) || strings.HasPrefix(c.Text, deprecatedBuildTagPrefix)) {
				continue
			}
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(c.Text, buildTagPrefix), deprecatedBuildTagPrefix))
		}
	}

	return ""
}

type visitFn func(node ast.Node)

func (f visitFn) Visit(node ast.Node) ast.Visitor {
	f(node)
	return f
}
