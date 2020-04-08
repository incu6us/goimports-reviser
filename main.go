package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"strings"

	"github.com/incu6us/goimport-reviser/helper"
)

const testFilePath = "./testdata/example.go"

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, testFilePath, nil, 0)
	if err != nil {
		log.Println(err)
		return
	}

	imports := combineImports(f)

	var stdImports []string
	var otherImports []string

	for _, generalImport := range imports {
		if _, ok := helper.StdPackages[generalImport]; ok {
			stdImports = append(stdImports, generalImport)
			continue
		}

		otherImports = append(otherImports, generalImport)
	}

	fixImports(f, stdImports, otherImports)

	out, err := GenerateFile(fset, f)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(out))
}

func GenerateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func fixImports(f *ast.File, stdImports []string, generalImports []string) {
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

					if linesCounter == 0 && len(generalImports) > 0 {
						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: ""}}

						specs = append(specs, iSpec)
					}
				}

				for _, generalImport := range generalImports {
					iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: generalImport}}
					specs = append(specs, iSpec)
				}

				dd.Specs = specs
			}
		}
	}
}

func combineImports(f *ast.File) []string {
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
