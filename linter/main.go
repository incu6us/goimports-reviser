// +build linter

package main

import (
	"flag"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/incu6us/goimports-reviser/v2/reviser"
)

func main() {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	var localPackagePrefixes string
	flagSet.StringVar(&localPackagePrefixes, "local", "", "Local package prefixes which will be placed after 3rd-party group(if defined). Values should be comma-separated. Optional parameters.")

	var (
		shouldRemoveUnusedImports bool
		shouldSetAlias            bool
		shouldFormat              bool
	)

	flagSet.BoolVar(&shouldRemoveUnusedImports, "rm-unused", false, "Remove unused imports. Optional parameter.")
	flagSet.BoolVar(&shouldSetAlias, "set-alias", false, "Set alias for versioned package names, like 'github.com/go-pg/pg/v9'. "+
		"In this case import will be set as 'pg \"github.com/go-pg/pg/v9\"'. Optional parameter.")
	flagSet.BoolVar(&shouldFormat, "format", false, "Option will perform additional formatting. Optional parameter.")

	var options reviser.Options
	if shouldRemoveUnusedImports {
		options = append(options, reviser.OptionRemoveUnusedImports)
	}

	if shouldSetAlias {
		options = append(options, reviser.OptionUseAliasForVersionSuffix)
	}

	if shouldFormat {
		options = append(options, reviser.OptionFormat)
	}

	singlechecker.Main(NewAnalyzer(flagSet, localPackagePrefixes, options...))
}
