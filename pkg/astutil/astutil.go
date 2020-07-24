package astutil

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	srcPathPrefix         = "src"
	pathSeparator         = string(os.PathSeparator)
	goFileExtensionSuffix = ".go"
)

// UsesImport is a similar to astutil.UsesImport but with skipping version in the import path
func UsesImport(f *ast.File, gopath, importPath string) bool {
	importIdentNames := make(map[string]struct{}, len(f.Imports))

	var importSpec *ast.ImportSpec
	for _, spec := range f.Imports {
		name := spec.Name.String()
		switch name {
		case "<nil>":
			pkgName, _ := PackageNameFromImportPath(gopath, importPath)
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
					pkg, _ := PackageNameFromImportPath(gopath, importPath)
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

// PackageNameFromImportPath will return package name
// and true if import base suffix is different from its package name
func PackageNameFromImportPath(gopath, importPath string) (string, bool) {
	pkgNameFromPath := path.Base(importPath)

	if strings.HasPrefix(pkgNameFromPath, "v") {
		if _, err := strconv.Atoi(pkgNameFromPath[1:]); err == nil {
			dir := path.Dir(importPath)
			if dir != "." {
				pkgNameFromPath = path.Base(dir)
			}

			return pkgNameFromPath, true
		}
	}

	pkgNameFromFS, err := resolvePackageName(gopath, importPath)
	if err != nil {
		if os.IsNotExist(err) {
			return pkgNameFromPath, false
		}

		panic(err)
	}

	if pkgNameFromFS != pkgNameFromPath {
		return pkgNameFromFS, true
	}

	return pkgNameFromFS, false
}

type visitFn func(node ast.Node)

func (f visitFn) Visit(node ast.Node) ast.Visitor {
	f(node)
	return f
}

// resolvePackageName resolves import to package name token(on local FS)
// Input:
//		1 - GOPATH value
//		2 - import package name(like: github.com/pkg/errors)
// Output:
//		1 - package (like: errors)
//		2 - error
func resolvePackageName(gopath string, pkg string) (string, error) {
	srcPath := strings.Join([]string{gopath, srcPathPrefix}, pathSeparator)

	pkgPath := strings.Join([]string{srcPath, pkg}, pathSeparator)

	fileInfos, err := ioutil.ReadDir(pkgPath)
	if err != nil {
		return "", err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		if filepath.Ext(fileInfo.Name()) != goFileExtensionSuffix {
			continue
		}

		relativePathToFile := strings.Join([]string{pkgPath, fileInfo.Name()}, pathSeparator)

		pf, err := parser.ParseFile(token.NewFileSet(), relativePathToFile, nil, parser.PackageClauseOnly)
		if err != nil {
			return "", err
		}

		if pf.Name != nil {
			return pf.Name.String(), nil
		}
	}

	return "", nil
}
