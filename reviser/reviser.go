package reviser

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"github.com/incu6us/goimports-reviser/helper"
)

func Execute(projectName, filePath string) ([]byte, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	imports := combineAllImports(f)

	stdImports, generalImports, projectImports := groupImports(projectName, imports)

	fixImports(f, stdImports, generalImports, projectImports)

	formattedContent, err := generateFile(fset, f)
	if err != nil {
		return nil, err
	}

	ok, err := hasDiff(formattedContent, filePath)
	if err != nil {
		return nil, err
	}

	if ok {
		return formattedContent, nil
	}

	return nil, nil
}

func hasDiff(formattedContent []byte, filePath string) (bool, error) {
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(originalContent, formattedContent), nil
}

func groupImports(projectName string, imports []string) ([]string, []string, []string) {
	var (
		stdImports     []string
		projectImports []string
		generalImports []string
	)

	sort.Strings(imports)

	for _, imprt := range imports {
		if _, ok := helper.StdPackages[imprt]; ok {
			stdImports = append(stdImports, imprt)
			continue
		}

		if strings.Contains(imprt, projectName) {
			projectImports = append(projectImports, imprt)
			continue
		}

		generalImports = append(generalImports, imprt)
	}

	return stdImports, generalImports, projectImports
}

func generateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func fixImports(f *ast.File, stdImports []string, generalImports []string, projectImports []string) {
	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				var specs []ast.Spec

				linesCounter := len(stdImports)
				for _, stdImport := range stdImports {
					iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: stdImport}}
					specs = append(specs, iSpec)

					linesCounter--

					if linesCounter == 0 && (len(generalImports) > 0 || len(projectImports) > 0) {
						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: ""}}

						specs = append(specs, iSpec)
					}
				}

				linesCounter = len(generalImports)
				for _, generalImport := range generalImports {
					iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: generalImport}}
					specs = append(specs, iSpec)

					linesCounter--

					if linesCounter == 0 && len(projectImports) > 0 {
						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: ""}}

						specs = append(specs, iSpec)
					}
				}

				for _, projectImport := range projectImports {
					iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: projectImport}}
					specs = append(specs, iSpec)
				}

				dd.Specs = specs
			}
		}
	}
}

func combineAllImports(f *ast.File) []string {
	var imports []string

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				for _, spec := range dd.Specs {
					var importSpecStr string
					importSpec := spec.(*ast.ImportSpec)

					if importSpec.Name != nil {
						importSpecStr = strings.Join([]string{importSpec.Name.String(), importSpec.Path.Value}, " ")
					} else {
						importSpecStr = importSpec.Path.Value
					}

					imports = append(imports, importSpecStr)
				}
			}
		}
	}

	return imports
}
