package reviser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/incu6us/goimports-reviser/helper"
)

type Option int

const (
	OptionRemoveUnusedImports Option = iota + 1
)

type Options []Option

func (o Options) shouldRemoveUnusedImports() bool {
	for _, option := range o {
		if option == OptionRemoveUnusedImports {
			return true
		}
	}

	return false
}

// Revise imports and format the code
func Execute(projectName, filePath string, options ...Option) ([]byte, bool, error) {
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	fset := token.NewFileSet()

	pf, err := parser.ParseFile(fset, "", originalContent, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	imports, commentsMetadata := combineAllImportsWithMetadata(pf, options)

	stdImports, generalImports, projectImports := groupImports(projectName, imports)

	fixImports(pf, stdImports, generalImports, projectImports, commentsMetadata)

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

func fixImports(f *ast.File, stdImports []string, generalImports []string, projectImports []string, commentsMetadata map[string]*commentsMetadata) {
	var importsPositions []*importPosition

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				importsPositions = append(importsPositions, &importPosition{
					Start: dd.Pos(),
					End:   dd.End(),
				})

				var specs []ast.Spec

				linesCounter := len(stdImports)
				for _, stdImport := range stdImports {
					iSpec := &ast.ImportSpec{
						Path: &ast.BasicLit{Value: importWithComment(stdImport, commentsMetadata), Kind: dd.Tok},
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
						Path: &ast.BasicLit{Value: importWithComment(generalImport, commentsMetadata), Kind: dd.Tok},
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
						Path: &ast.BasicLit{Value: importWithComment(projectImport, commentsMetadata), Kind: dd.Tok},
					}
					specs = append(specs, iSpec)
				}

				dd.Specs = specs
			}
		}
	}

	clearImportDocs(f, importsPositions)
}

func clearImportDocs(f *ast.File, importsPositions []*importPosition) {
	importsComments := make([]*ast.CommentGroup, 0, len(f.Comments))

	for _, comment := range f.Comments {
		for _, importPosition := range importsPositions {
			if importPosition.IsInRange(comment) {
				continue
			}
			importsComments = append(importsComments, comment)
		}
	}

	f.Comments = importsComments
}

func importWithComment(imprt string, commentsMetadata map[string]*commentsMetadata) string {
	var comment string
	commentGroup, ok := commentsMetadata[imprt]
	if ok {
		if commentGroup != nil && commentGroup.Comment != nil && len(commentGroup.Comment.List) > 0 {
			comment = fmt.Sprintf("// %s", commentGroup.Comment.Text())
		}
	}

	return fmt.Sprintf("%s %s", imprt, comment)
}

func combineAllImportsWithMetadata(f *ast.File, options Options) ([]string, map[string]*commentsMetadata) {
	var imports []string
	importsWithMetadata := map[string]*commentsMetadata{}
	shouldRemoveUnusedImports := options.shouldRemoveUnusedImports()

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				for _, spec := range dd.Specs {
					var importSpecStr string
					importSpec := spec.(*ast.ImportSpec)

					if shouldRemoveUnusedImports && !astutil.UsesImport(f, strings.Trim(importSpec.Path.Value, `"`)) {
						continue
					}

					if importSpec.Name != nil {
						importSpecStr = strings.Join([]string{importSpec.Name.String(), importSpec.Path.Value}, " ")
					} else {
						importSpecStr = importSpec.Path.Value
					}

					imports = append(imports, importSpecStr)
					importsWithMetadata[importSpecStr] = &commentsMetadata{
						Doc:     importSpec.Doc,
						Comment: importSpec.Comment,
					}
				}
			}
		}
	}

	return imports, importsWithMetadata
}

type commentsMetadata struct {
	Doc     *ast.CommentGroup
	Comment *ast.CommentGroup
}

type importPosition struct {
	Start token.Pos
	End   token.Pos
}

func (p *importPosition) IsInRange(comment *ast.CommentGroup) bool {
	if p.Start >= comment.Pos() || comment.Pos() <= p.End {
		return true
	}

	return false
}
