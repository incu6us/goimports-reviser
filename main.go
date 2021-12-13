package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/incu6us/goimports-reviser/v2/pkg/module"
	"github.com/incu6us/goimports-reviser/v2/reviser"
)

const (
	projectNameArg         = "project-name"
	filePathArg            = "file-path"
	versionArg             = "version"
	removeUnusedImportsArg = "rm-unused"
	setAliasArg            = "set-alias"
	localPkgPrefixesArg    = "local"
	outputArg              = "output"
	formatArg              = "format"
	listDiffFileNameArg    = "list-diff"
	setExitStatusArg       = "set-exit-status"
)

// Project build specific vars
var (
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string

	shouldShowVersion         *bool
	shouldRemoveUnusedImports *bool
	shouldSetAlias            *bool
	shouldFormat              *bool
	listFileName              *bool
	setExitStatus             *bool
)

var projectName, filePath, localPkgPrefixes, output string

func init() {
	flag.StringVar(
		&filePath,
		filePathArg,
		"",
		"File path to fix imports(ex.: ./reviser/reviser.go). Optional parameter.",
	)

	flag.StringVar(
		&projectName,
		projectNameArg,
		"",
		"Your project name(ex.: github.com/incu6us/goimports-reviser). Optional parameter.",
	)

	flag.StringVar(
		&localPkgPrefixes,
		localPkgPrefixesArg,
		"",
		"Local package prefixes which will be placed after 3rd-party group(if defined). Values should be comma-separated. Optional parameters.",
	)

	flag.StringVar(
		&output,
		outputArg,
		"file",
		`Can be "file", "write" or "stdout". Whether to write the formatted content back to the file or to stdout. When "write" together with "-list" will list the file name and write back to the file. Optional parameter.`,
	)

	listFileName = flag.Bool(
		listDiffFileNameArg,
		false,
		"Option will list files whose formatting differs from goimports-reviser. Optional parameter.",
	)

	setExitStatus = flag.Bool(
		setExitStatusArg,
		false,
		"set the exit status to 1 if a change is needed/made. Optional parameter.",
	)

	shouldRemoveUnusedImports = flag.Bool(
		removeUnusedImportsArg,
		false,
		"Remove unused imports. Optional parameter.",
	)

	shouldSetAlias = flag.Bool(
		setAliasArg,
		false,
		"Set alias for versioned package names, like 'github.com/go-pg/pg/v9'. "+
			"In this case import will be set as 'pg \"github.com/go-pg/pg/v9\"'. Optional parameter.",
	)

	shouldFormat = flag.Bool(
		formatArg,
		false,
		"Option will perform additional formatting. Optional parameter.",
	)

	if Tag != "" {
		shouldShowVersion = flag.Bool(
			versionArg,
			false,
			"Show version.",
		)
	}
}

func printUsage() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf(
		"version: %s\nbuild with: %s\ntag: %s\ncommit: %s\nsource: %s\n",
		strings.TrimPrefix(Tag, "v"),
		GoVersion,
		Tag,
		Commit,
		SourceURL,
	)
}

func main() {
	flag.Parse()

	if shouldShowVersion != nil && *shouldShowVersion {
		printVersion()
		return
	}

	if filePath == "" {
		filePath = reviser.StandardInput
	}

	if err := validateRequiredParam(filePath); err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	var options reviser.Options
	if shouldRemoveUnusedImports != nil && *shouldRemoveUnusedImports {
		options = append(options, reviser.OptionRemoveUnusedImports)
	}

	if shouldSetAlias != nil && *shouldSetAlias {
		options = append(options, reviser.OptionUseAliasForVersionSuffix)
	}

	if shouldFormat != nil && *shouldFormat {
		options = append(options, reviser.OptionFormat)
	}

	projectName, err := module.DetermineProjectName(projectName, filePath)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	formattedOutput, hasChange, err := reviser.Execute(projectName, filePath, localPkgPrefixes, options...)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if !hasChange && *listFileName {
		return
	}
	if hasChange && *listFileName && output != "write" {
		fmt.Println(filePath)
	} else if output == "stdout" || filePath == reviser.StandardInput {
		fmt.Print(string(formattedOutput))
	} else if output == "file" || output == "write" {
		if !hasChange {
			return
		}

		if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))
		}
		if *listFileName {
			fmt.Println(filePath)
		}
	} else {
		log.Fatalf(`invalid output "%s" specified`, output)
	}

	if hasChange && *setExitStatus {
		os.Exit(1)
	}
}

func validateRequiredParam(filePath string) error {
	if filePath == reviser.StandardInput {
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeNamedPipe == 0 {
			// no data on stdin
			return errors.Errorf("-%s should be set", filePathArg)
		}
	}
	return nil
}
