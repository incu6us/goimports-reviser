package reviser

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/incu6us/goimports-reviser/helper"
)

// Revise imports and format the code
func Execute(projectName, filePath string) ([]byte, bool, error) {
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	fset := token.NewFileSet()

	pf, err := parser.ParseFile(fset, "", originalContent, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	imports, commentsGroup := combineAllImportsWithMetadata(pf)

	stdImports, generalImports, projectImports := groupImports(projectName, imports)

	fixImports(pf, stdImports, generalImports, projectImports, commentsGroup)

	fixedImportsContent, err := generateFile(fset, pf)
	if err != nil {
		return nil, false, err
	}

	formattedContent, err := format.Source(fixedImportsContent)
	if err != nil {
		return nil, false, err
	}

	return formattedContent, !bytes.Equal(originalContent, formattedContent), nil
}

func groupImports(projectName string, imports []string) ([]string, []string, []string) {
	var (
		stdImports     []string
		projectImports []string
		generalImports []string
	)

	sort.Strings(imports)

	for _, imprt := range imports {
		pkgWithoutAlias := skipPackageAlias(imprt)

		if _, ok := helper.StdPackages[pkgWithoutAlias]; ok {
			stdImports = append(stdImports, imprt)
			continue
		}

		if strings.Contains(pkgWithoutAlias, projectName) {
			projectImports = append(projectImports, imprt)
			continue
		}

		generalImports = append(generalImports, imprt)
	}

	return stdImports, generalImports, projectImports
}

func skipPackageAlias(pkg string) string {
	values := strings.Split(pkg, " ")
	if len(values) > 1 {
		return values[1]
	}

	return pkg
}

func generateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func fixImports(f *ast.File, stdImports []string, generalImports []string, projectImports []string, commentsGroup map[string]*ast.CommentGroup) {
	importCommentsPos := make(map[token.Pos]struct{}, len(f.Comments))

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				combineCommentsPos(dd, importCommentsPos)

				var specs []ast.Spec

				linesCounter := len(stdImports)
				for _, stdImport := range stdImports {
					iSpec := &ast.ImportSpec{
						Doc:  commentGroup(stdImport, commentsGroup),
						Path: &ast.BasicLit{Value: stdImport, Kind: dd.Tok},
					}
					specs = append(specs, iSpec)

					linesCounter--

					if linesCounter == 0 && (len(generalImports) > 0 || len(projectImports) > 0) {
						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

						specs = append(specs, iSpec)
					}
				}

				linesCounter = len(generalImports)
				for _, generalImport := range generalImports {
					iSpec := &ast.ImportSpec{
						Doc:  commentGroup(generalImport, commentsGroup),
						Path: &ast.BasicLit{Value: generalImport, Kind: dd.Tok},
					}
					specs = append(specs, iSpec)

					linesCounter--

					if linesCounter == 0 && len(projectImports) > 0 {
						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

						specs = append(specs, iSpec)
					}
				}

				for _, projectImport := range projectImports {
					iSpec := &ast.ImportSpec{
						Doc:  commentGroup(projectImport, commentsGroup),
						Path: &ast.BasicLit{Value: projectImport, Kind: dd.Tok},
					}
					specs = append(specs, iSpec)
				}

				dd.Specs = specs
			}
		}
	}

	clearImportComments(f, importCommentsPos)
}

func clearImportComments(f *ast.File, importCommentsPos map[token.Pos]struct{}) {
	importsComments := make([]*ast.CommentGroup, 0, len(f.Comments))

	for _, comment := range f.Comments {
		if _, ok := importCommentsPos[comment.Pos()]; !ok {
			importsComments = append(importsComments, comment)
		}
	}

	f.Comments = importsComments
}

func combineCommentsPos(dd *ast.GenDecl, importCommentsPos map[token.Pos]struct{}) {
	for _, spec := range dd.Specs {
		doc := spec.(*ast.ImportSpec).Doc
		if doc != nil {
			importCommentsPos[doc.Pos()] = struct{}{}
		}
	}
}

func commentGroup(imprt string, commentsGroup map[string]*ast.CommentGroup) *ast.CommentGroup {
	commentGroup, ok := commentsGroup[imprt]
	if ok {
		if commentGroup != nil && len(commentGroup.List) > 0 {
			return commentGroup
		}
	}

	return nil
}

func combineAllImportsWithMetadata(f *ast.File) ([]string, map[string]*ast.CommentGroup) {
	var imports []string
	commentsGroup := map[string]*ast.CommentGroup{}

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
					commentsGroup[importSpecStr] = importSpec.Doc
				}
			}
		}
	}

	return imports, commentsGroup
}
