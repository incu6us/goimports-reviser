package goanalysis

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/incu6us/goimports-reviser/v3/pkg/module"
	"github.com/incu6us/goimports-reviser/v3/reviser"
)

const errMessage = "imports must be formatted"

func NewAnalyzer(flagSet *flag.FlagSet, localPkgPrefixes string, options ...reviser.SourceFileOption) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "goimportsreviser",
		Doc:   "goimports-reviser linter",
		Run:   run(localPkgPrefixes, options...),
		Flags: *flagSet,
	}
}

func run(localPkgPrefixes string, options ...reviser.SourceFileOption) func(pass *analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := func(formattedFile *ast.File, hasChanged bool) func(node ast.Node) bool {
			return func(node ast.Node) bool {
				file, ok := node.(*ast.File)
				if !ok {
					return true
				}
				if !hasChanged {
					return true
				}

				if len(file.Imports) != len(formattedFile.Imports) {
					pass.Reportf(
						file.Pos(),
						errMessage,
					)
				}

				for i, originalDecl := range file.Decls {
					origDd, ok := originalDecl.(*ast.GenDecl)
					if !ok {
						continue
					}

					if origDd.Tok != token.IMPORT {
						continue
					}

					if origDd != formattedFile.Decls[i] {
						pass.Reportf(
							file.Pos()+origDd.Lparen,
							errMessage,
						)
					}
				}

				return true
			}
		}

		var projectName string

		for _, f := range pass.Files {
			filePath := pass.Fset.File(f.Package).Name()

			if projectName == "" {
				var err error
				projectName, err = module.DetermineProjectName("", filePath)
				if err != nil {
					return nil, err
				}
			}

			formattedFileContent, _, hasChanged, err := reviser.NewSourceFile(projectName, filePath).Fix(options...)
			if err != nil {
				return nil, err
			}

			if !hasChanged {
				continue
			}

			formattedFile, err := parser.ParseFile(token.NewFileSet(), filePath, formattedFileContent, parser.ImportsOnly)
			if err != nil {
				panic(err)
			}

			ast.Inspect(f, inspect(formattedFile, hasChanged))
		}

		return nil, nil
	}
}
