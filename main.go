package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/incu6us/goimports-reviser/reviser"
)

const (
	projectNameKey      = "project-name"
	filePathKey         = "file-path"
	versionKey          = "version"
	removeUnusedImports = "rm-unused"
)

// Project build specific vars
var (
	Version string
	Commit  string

	shouldShowVersion         *bool
	shouldRemoveUnusedImports *bool
)

var projectName, filePath string

func init() {
	flag.StringVar(
		&projectName,
		projectNameKey,
		"",
		"your project name(ex.: github.com/incu6us/goimport-reviser)",
	)

	flag.StringVar(
		&filePath,
		filePathKey,
		"",
		"file path to fix imports(ex.: ./reviser/reviser.go)",
	)

	shouldRemoveUnusedImports = flag.Bool(
		removeUnusedImports,
		false,
		"remove unused imports",
	)

	if Version != "" {
		shouldShowVersion = flag.Bool(
			versionKey,
			false,
			"to show the version",
		)
	}
}

var usage = func() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if shouldShowVersion != nil && *shouldShowVersion {
		fmt.Printf("version: %s (%s)\n", Version, Commit)
		return
	}

	if err := validateInputs(projectName, filePath); err != nil {
		fmt.Printf("%s\n\n", err)
		usage()
		os.Exit(1)
	}

	var options reviser.Options
	if shouldRemoveUnusedImports != nil && *shouldRemoveUnusedImports {
		options = append(options, reviser.OptionRemoveUnusedImports)
	}

	formattedOutput, hasChange, err := reviser.Execute(projectName, filePath, options...)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if !hasChange {
		return
	}

	if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
		log.Fatalf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))
	}
}

func validateInputs(projectName, filePath string) error {
	var errMessages []string

	if projectName == "" {
		errMessages = append(errMessages, fmt.Sprintf("-%s should be set", projectNameKey))
	}

	if filePath == "" {
		errMessages = append(errMessages, fmt.Sprintf("-%s should be set", filePathKey))
	}

	if len(errMessages) > 0 {
		return errors.New(strings.Join(errMessages, "\n"))
	}

	return nil
}
