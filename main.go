package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

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
)

var (
	projectName, filePath, companyPkgPrefixes, output, importsOrder string

	// Deprecated
	localPkgPrefixes string
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

	originFilePath := flag.Arg(0)
	if filePath != "" {
		deprecatedMessagesCh <- fmt.Sprintf("-%s is deprecated. Put file name as last argument to the command(Example: goimports-reviser -rm-unused -set-alias -format goimports-reviser/main.go)\n", filePathArg)
		originFilePath = filePath
	}

	if originFilePath == "" {
		originFilePath = reviser.StandardInput
	}

	if err := validateRequiredParam(originFilePath); err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	var options reviser.Options
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
		deprecatedMessagesCh <- fmt.Sprintf(`-%s is deprecated and will be removed soon. Use -%s instead.\n`, localArg, companyPkgPrefixesArg)
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

	originProjectName, err := helper.DetermineProjectName(projectName, originFilePath)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	close(deprecatedMessagesCh)

	formattedOutput, hasChange, err := reviser.NewSourceFile(originProjectName, originFilePath).Fix(options...)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if !hasChange && *listFileName {
		printDeprecations(deprecatedMessagesCh)
		return
	}
	if hasChange && *listFileName && output != "write" {
		fmt.Println(originFilePath)
	} else if output == "stdout" || originFilePath == reviser.StandardInput {
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

	if hasChange && *setExitStatus {
		os.Exit(1)
	}

	printDeprecations(deprecatedMessagesCh)
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

func printDeprecations(deprecatedMessagesCh chan string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var hasDeprecations bool
		for deprecatedMessage := range deprecatedMessagesCh {
			hasDeprecations = true
			fmt.Printf("%s\n", deprecatedMessage)
		}
		if hasDeprecations {
			fmt.Printf("All changes to file are applied, but command-line syntax should be fixed\n")
			os.Exit(1)
		}
		wg.Done()
	}()
	wg.Wait()
}
