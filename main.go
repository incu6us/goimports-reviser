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
	projectNameKey = "project-name"
	filePathKey    = "file-path"
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
}

var usage = func() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if err := validateInputs(projectName, filePath); err != nil {
		fmt.Printf("%s\n\n", err)
		usage()
		os.Exit(1)
	}

	formattedOutput, hasChange, err := reviser.Execute(projectName, filePath)
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
	var errMesages []string

	if projectName == "" {
		errMesages = append(errMesages, fmt.Sprintf("-%s should be set", projectNameKey))
	}

	if filePath == "" {
		errMesages = append(errMesages, fmt.Sprintf("-%s should be set", filePathKey))
	}

	if len(errMesages) > 0 {
		return errors.New(strings.Join(errMesages, "\n"))
	}

	return nil
}
