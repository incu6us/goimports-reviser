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
	"path"
	"sort"
	"strings"

	"github.com/incu6us/goimports-reviser/v2/pkg/astutil"
	"github.com/incu6us/goimports-reviser/v2/pkg/std"
)

const (
	stringValueSeparator = ","
)

// Option is an int alias for options
type Option int

const (
	// OptionRemoveUnusedImports is an option to remove unused imports
	OptionRemoveUnusedImports Option = iota + 1

	// OptionUseAliasForVersionSuffix is an option to set explicit package name in imports
	OptionUseAliasForVersionSuffix
)

// Options is a slice of executing options
type Options []Option

func (o Options) shouldRemoveUnusedImports() bool {
	for _, option := range o {
		if option == OptionRemoveUnusedImports {
			return true
		}
	}

	return false
}

func (o Options) shouldUseAliasForVersionSuffix() bool {
	for _, option := range o {
		if option == OptionUseAliasForVersionSuffix {
			return true
		}
	}

	return false
}

// Execute is for revise imports and format the code
func Execute(projectName, filePath, localPkgPrefixes string, options ...Option) ([]byte, bool, error) {
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	fset := token.NewFileSet()

	pf, err := parser.ParseFile(fset, "", originalContent, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	importsWithMetadata, err := parseImports(pf, filePath, options)
	if err != nil {
		return nil, false, err
	}

	stdImports, generalImports, projectLocalPkgs, projectImports := groupImports(
		projectName,
		localPkgPrefixes,
		importsWithMetadata,
	)

	fixImports(pf, stdImports, generalImports, projectLocalPkgs, projectImports, importsWithMetadata)

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

func groupImports(
	projectName string,
	localPkgPrefixes string,
	importsWithMetadata map[string]*commentsMetadata,
) ([]string, []string, []string, []string) {
	var (
		stdImports       []string
		projectImports   []string
		projectLocalPkgs []string
		generalImports   []string
	)

	localPackagePrefixes := commaValueToSlice(localPkgPrefixes)

	for imprt := range importsWithMetadata {
		pkgWithoutAlias := skipPackageAlias(imprt)

		if _, ok := std.StdPackages[pkgWithoutAlias]; ok {
			stdImports = append(stdImports, imprt)
			continue
		}

		var isLocalPackageFound bool
		for _, localPackagePrefix := range localPackagePrefixes {
			if strings.HasPrefix(pkgWithoutAlias, localPackagePrefix) {
				projectLocalPkgs = append(projectLocalPkgs, imprt)
				isLocalPackageFound = true
				break
			}
		}

		if isLocalPackageFound {
			continue
		}

		if strings.Contains(pkgWithoutAlias, projectName) {
			projectImports = append(projectImports, imprt)
			continue
		}

		generalImports = append(generalImports, imprt)
	}

	sort.Strings(stdImports)
	sort.Strings(generalImports)
	sort.Strings(projectLocalPkgs)
	sort.Strings(projectImports)

	return stdImports, generalImports, projectLocalPkgs, projectImports
}

func commaValueToSlice(s string) []string {
	values := strings.Split(s, stringValueSeparator)
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		result = append(result, value)
	}

	return result
}

func skipPackageAlias(pkg string) string {
	values := strings.Split(pkg, " ")
	if len(values) > 1 {
		return strings.Trim(values[1], `"`)
	}

	return strings.Trim(pkg, `"`)
}

func generateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// TODO: fix gocyclo
func fixImports(
	f *ast.File,
	stdImports, generalImports, projectLocalPkgs, projectImports []string,
	commentsMetadata map[string]*commentsMetadata,
) {
	var importsPositions []*importPosition

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				importsPositions = append(
					importsPositions, &importPosition{
						Start: dd.Pos(),
						End:   dd.End(),
					},
				)

				var specs []ast.Spec

				linesCounter := len(stdImports)
				for _, stdImport := range stdImports {
					spec := &ast.ImportSpec{
						Path: &ast.BasicLit{Value: importWithComment(stdImport, commentsMetadata), Kind: dd.Tok},
					}
					specs = append(specs, spec)

					linesCounter--

					if linesCounter == 0 && (len(generalImports) > 0 || len(projectLocalPkgs) > 0 || len(projectImports) > 0) {
						spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

						specs = append(specs, spec)
					}
				}

				linesCounter = len(generalImports)
				for _, generalImport := range generalImports {
					spec := &ast.ImportSpec{
						Path: &ast.BasicLit{Value: importWithComment(generalImport, commentsMetadata), Kind: dd.Tok},
					}
					specs = append(specs, spec)

					linesCounter--

					if linesCounter == 0 && (len(projectLocalPkgs) > 0 || len(projectImports) > 0) {
						spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

						specs = append(specs, spec)
					}
				}

				linesCounter = len(projectLocalPkgs)
				for _, projectLocalPkg := range projectLocalPkgs {
					spec := &ast.ImportSpec{
						Path: &ast.BasicLit{Value: importWithComment(projectLocalPkg, commentsMetadata), Kind: dd.Tok},
					}
					specs = append(specs, spec)

					linesCounter--

					if linesCounter == 0 && len(projectImports) > 0 {
						spec = &ast.ImportSpec{Path: &ast.BasicLit{Value: "", Kind: token.STRING}}

						specs = append(specs, spec)
					}
				}

				for _, projectImport := range projectImports {
					spec := &ast.ImportSpec{
						Path: &ast.BasicLit{Value: importWithComment(projectImport, commentsMetadata), Kind: dd.Tok},
					}
					specs = append(specs, spec)
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

	if len(f.Imports) > 0 {
		f.Comments = importsComments
	}
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

func parseImports(f *ast.File, filePath string, options Options) (map[string]*commentsMetadata, error) {
	importsWithMetadata := map[string]*commentsMetadata{}

	shouldRemoveUnusedImports := options.shouldRemoveUnusedImports()
	shouldUseAliasForVersionSuffix := options.shouldUseAliasForVersionSuffix()

	var packageImports map[string]string
	var err error

	if shouldRemoveUnusedImports || shouldUseAliasForVersionSuffix {
		packageImports, err = astutil.LoadPackageDependencies(path.Dir(filePath), astutil.ParseBuildTag(f))
		if err != nil {
			return nil, err
		}
	}

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			dd := decl.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				for _, spec := range dd.Specs {
					var importSpecStr string
					importSpec := spec.(*ast.ImportSpec)

					if shouldRemoveUnusedImports && !astutil.UsesImport(
						f, packageImports, strings.Trim(importSpec.Path.Value, `"`),
					) {
						continue
					}

					if importSpec.Name != nil {
						importSpecStr = strings.Join([]string{importSpec.Name.String(), importSpec.Path.Value}, " ")
					} else {
						if shouldUseAliasForVersionSuffix {
							importSpecStr = setAliasForVersionedImportSpec(importSpec, packageImports)
						} else {
							importSpecStr = importSpec.Path.Value
						}
					}

					importsWithMetadata[importSpecStr] = &commentsMetadata{
						Doc:     importSpec.Doc,
						Comment: importSpec.Comment,
					}
				}
			}
		}
	}

	return importsWithMetadata, nil
}

func setAliasForVersionedImportSpec(importSpec *ast.ImportSpec, packageImports map[string]string) string {
	var importSpecStr string

	imprt := strings.Trim(importSpec.Path.Value, `"`)
	aliasName := packageImports[imprt]

	importSuffix := path.Base(imprt)
	if importSuffix != aliasName {
		importSpecStr = fmt.Sprintf("%s %s", aliasName, importSpec.Path.Value)
	} else {
		importSpecStr = importSpec.Path.Value
	}

	return importSpecStr
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
	if p.Start <= comment.Pos() && comment.Pos() <= p.End {
		return true
	}

	return false
}
