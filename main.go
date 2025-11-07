package main

import (
	"os"

	"github.com/incu6us/goimports-reviser/v3/internal/app"
)

// Project build specific vars
// These are set during build time via -ldflags
var (
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string
)

// main is now clean and follows SOLID principles
// Single Responsibility: only creates app and runs it
// Dependency Inversion: depends on app abstraction
func main() {
	application := app.New(Tag, Commit, SourceURL, GoVersion)
	exitCode := application.Run(os.Args[1:])
	os.Exit(exitCode)
}
