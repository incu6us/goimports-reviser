package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/incu6us/goimports-reviser/v2/pkg/grouporder"
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
	groupsOrderArg         = "groups-order"
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
)

var projectName, filePath, localPkgPrefixes, output, groupsOrder string

func init() {
	flag.StringVar(
		&filePath,
		filePathArg,
		"",
		"File path to fix imports(ex.: ./reviser/reviser.go). Required parameter.",
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
		`Can be "file" or "stdout". Whether to write the formatted content back to the file or to stdout. Optional parameter.`,
	)

	flag.StringVar(
		&groupsOrder,
		groupsOrderArg,
		"std,ext,org,prj",
		`Sets the order of import groups. 
Groups names: std - std libs, ext - external libs, org - organization libs, prj - project libs.
Examples: 
	"std,ext,org,prj"
	"std,prj,org,ext"
`,
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

	if err := validateRequiredParam(filePath); err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	projectName, err := determineProjectName(projectName, filePath)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	groupOrder, err := grouporder.NewImportGroupOrder(groupsOrder)
	if err != nil {
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

	formattedOutput, hasChange, err := reviser.Execute(projectName, filePath, localPkgPrefixes, groupOrder, options...)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if output == "stdout" {
		fmt.Print(string(formattedOutput))
	} else if output == "file" {
		if !hasChange {
			return
		}

		if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))
		}
	} else {
		log.Fatalf(`invalid output "%s" specified`, output)
	}
}

func determineProjectName(projectName, filePath string) (string, error) {
	if projectName == "" {
		projectRootPath, err := module.GoModRootPath(filePath)
		if err != nil {
			return "", err
		}

		moduleName, err := module.Name(projectRootPath)
		if err != nil {
			return "", err
		}

		return moduleName, nil
	}

	return projectName, nil
}

func validateRequiredParam(filePath string) error {
	if filePath == "" {
		return errors.Errorf("-%s should be set", filePathArg)
	}

	return nil
}
