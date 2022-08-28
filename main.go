package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/incu6us/goimports-reviser/v3/helper"
	"github.com/incu6us/goimports-reviser/v3/reviser"
)

const (
	projectNameArg         = "project-name"
	versionArg             = "version"
	removeUnusedImportsArg = "rm-unused"
	setAliasArg            = "set-alias"
	companyPkgPrefixesArg  = "company-prefixes"
	outputArg              = "output"
	importsOrderArg        = "imports-order"
	formatArg              = "format"
	listDiffFileNameArg    = "list-diff"
	setExitStatusArg       = "set-exit-status"
	recursiveArg           = "recursive"

	// Deprecated options
	localArg    = "local"
	filePathArg = "file-path"
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
	isRecursive               *bool
)

var (
	projectName, companyPkgPrefixes, output, importsOrder string

	// Deprecated
	localPkgPrefixes, filePath string
)

func init() {
	flag.StringVar(
		&filePath,
		filePathArg,
		"",
		"Deprecated. Put file name as an argument(last item) of command line.",
	)

	flag.StringVar(
		&projectName,
		projectNameArg,
		"",
		"Your project name(ex.: github.com/incu6us/goimports-reviser). Optional parameter.",
	)

	flag.StringVar(
		&companyPkgPrefixes,
		companyPkgPrefixesArg,
		"",
		"Company package prefixes which will be placed after 3rd-party group by default(if defined). Values should be comma-separated. Optional parameters.",
	)

	flag.StringVar(
		&localPkgPrefixes,
		localArg,
		"",
		"Deprecated",
	)

	flag.StringVar(
		&output,
		outputArg,
		"file",
		`Can be "file", "write" or "stdout". Whether to write the formatted content back to the file or to stdout. When "write" together with "-list" will list the file name and write back to the file. Optional parameter.`,
	)

	flag.StringVar(
		&importsOrder,
		importsOrderArg,
		"std,general,company,project",
		`Your imports groups can be sorted in your way. 
std - std import group; 
general - libs for general purpose; 
company - inter-org libs(if you set '-local'-option, then 4th group will be split separately. In other case, it will be the part of general purpose libs); 
project - your local project dependencies. 
Optional parameter.`,
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

	isRecursive = flag.Bool(
		recursiveArg,
		false,
		"Apply rules recursively if target is a directory. In case of ./... execution will be recursively applied by default. Optional parameter.",
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
	deprecatedMessagesCh := make(chan string, 10)
	flag.Parse()

	if shouldShowVersion != nil && *shouldShowVersion {
		printVersion()
		return
	}

	originPaths := flag.Args()
	if filePath != "" {
		deprecatedMessagesCh <- fmt.Sprintf("-%s is deprecated. Put file names as arguments to the command(Example: goimports-reviser -rm-unused -set-alias -format goimports-reviser/main.go goimports-reviser/file.go)", filePathArg)
		originPaths = append(originPaths, filePath)
	}

	if len(originPaths) == 0 {
		fmt.Print("files to revise should be set")
		printUsage()
		os.Exit(1)
	}

	var options reviser.SourceFileOptions
	if shouldRemoveUnusedImports != nil && *shouldRemoveUnusedImports {
		options = append(options, reviser.WithRemovingUnusedImports)
	}

	if shouldSetAlias != nil && *shouldSetAlias {
		options = append(options, reviser.WithUsingAliasForVersionSuffix)
	}

	if shouldFormat != nil && *shouldFormat {
		options = append(options, reviser.WithCodeFormatting)
	}

	if localPkgPrefixes != "" {
		if companyPkgPrefixes != "" {
			companyPkgPrefixes = localPkgPrefixes
		}
		deprecatedMessagesCh <- fmt.Sprintf(`-%s is deprecated and will be removed soon. Use -%s instead.`, localArg, companyPkgPrefixesArg)
	}

	if companyPkgPrefixes != "" {
		options = append(options, reviser.WithCompanyPackagePrefixes(companyPkgPrefixes))
	}

	if importsOrder != "" {
		order, err := reviser.StringToImportsOrders(importsOrder)
		if err != nil {
			fmt.Printf("%s\n\n", err)
			printUsage()
			os.Exit(1)
		}
		options = append(options, reviser.WithImportsOrder(order))
	}

	close(deprecatedMessagesCh)

	fileChangedFlag := false
	for _, originPath := range originPaths {
		originProjectName, err := helper.DetermineProjectName(projectName, originPath)
		if err != nil {
			fmt.Printf("%s\n\n", err)
			printUsage()
			os.Exit(1)
		}

		if _, ok := reviser.IsDir(originPath); ok {
			err := reviser.NewSourceDir(originProjectName, originPath, *isRecursive).Fix(options...)
			if err != nil {
				log.Fatalf("%+v", errors.WithStack(err))
			}
			continue
		}

		formattedOutput, hasChange, err := reviser.NewSourceFile(originProjectName, originPath).Fix(options...)
		if hasChange {
			fileChangedFlag = true
		}
		if err != nil {
			log.Fatalf("%+v", errors.WithStack(err))
		}

		resultPostProcess(hasChange, deprecatedMessagesCh, originPath, formattedOutput)
	}

	if fileChangedFlag && *setExitStatus {
		os.Exit(1)
	}
}

func resultPostProcess(hasChange bool, deprecatedMessagesCh chan string, originFilePath string, formattedOutput []byte) {
	if !hasChange && *listFileName {
		printDeprecations(deprecatedMessagesCh)
		return
	}
	if hasChange && *listFileName && output != "write" {
		fmt.Println(originFilePath)
	} else if output == "stdout" {
		fmt.Print(string(formattedOutput))
	} else if output == "file" || output == "write" {
		if !hasChange {
			printDeprecations(deprecatedMessagesCh)
			return
		}

		if err := ioutil.WriteFile(originFilePath, formattedOutput, 0644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v", originFilePath, errors.WithStack(err))
		}
		if *listFileName {
			fmt.Println(originFilePath)
		}
	} else {
		log.Fatalf(`invalid output "%s" specified`, output)
	}

	printDeprecations(deprecatedMessagesCh)
}

func printDeprecations(deprecatedMessagesCh chan string) {
	var hasDeprecations bool
	for deprecatedMessage := range deprecatedMessagesCh {
		hasDeprecations = true
		fmt.Printf("%s\n", deprecatedMessage)
	}
	if hasDeprecations {
		fmt.Printf("All changes to file are applied, but command-line syntax should be fixed\n")
		os.Exit(1)
	}
}
