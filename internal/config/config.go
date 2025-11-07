package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/incu6us/goimports-reviser/v3/reviser"
)

// Config holds all application configuration
// Following SRP: single source of truth for configuration
type Config struct {
	// File/Directory paths
	OriginPaths []string

	// Project configuration
	ProjectName        string
	CompanyPkgPrefixes string
	ImportsOrder       string
	Excludes           string

	// Feature flags
	ShouldRemoveUnusedImports   bool
	ShouldSetAlias              bool
	ShouldFormat                bool
	ShouldApplyToGeneratedFiles bool
	ShouldSeparateNamedImports  bool
	IsRecursive                 bool
	IsUseCache                  bool

	// Output configuration
	Output        string
	ListFileName  bool
	SetExitStatus bool

	// Version flags
	ShowVersion     bool
	ShowVersionOnly bool

	// Deprecated
	LocalPkgPrefixes string
	FilePath         string
}

// DeprecationMessages holds messages about deprecated flags
type DeprecationMessages struct {
	messages []string
}

// Add adds a deprecation message
func (d *DeprecationMessages) Add(msg string) {
	d.messages = append(d.messages, msg)
}

// HasMessages returns true if there are deprecation messages
func (d *DeprecationMessages) HasMessages() bool {
	return len(d.messages) > 0
}

// Messages returns all deprecation messages
func (d *DeprecationMessages) Messages() []string {
	return d.messages
}

// Validate validates the configuration
// Following SRP: validation logic is separate from parsing
func (c *Config) Validate() error {
	if len(c.OriginPaths) == 0 {
		return errors.New("no file(s) or directory(ies) specified on input")
	}

	if len(c.OriginPaths) == 1 && c.OriginPaths[0] == reviser.StandardInput {
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeNamedPipe == 0 {
			return errors.New("no data on stdin")
		}
	}

	return nil
}

// ToReviserOptions converts config to reviser options
// Following SRP: conversion logic is isolated
func (c *Config) ToReviserOptions() (reviser.SourceFileOptions, error) {
	var options reviser.SourceFileOptions

	if c.ShouldRemoveUnusedImports {
		options = append(options, reviser.WithRemovingUnusedImports)
	}

	if c.ShouldSetAlias {
		options = append(options, reviser.WithUsingAliasForVersionSuffix)
	}

	if c.ShouldFormat {
		options = append(options, reviser.WithCodeFormatting)
	}

	if !c.ShouldApplyToGeneratedFiles {
		options = append(options, reviser.WithSkipGeneratedFile)
	}

	if c.ShouldSeparateNamedImports {
		options = append(options, reviser.WithSeparatedNamedImports)
	}

	if c.CompanyPkgPrefixes != "" {
		options = append(options, reviser.WithCompanyPackagePrefixes(c.CompanyPkgPrefixes))
	}

	if c.ImportsOrder != "" {
		order, err := reviser.StringToImportsOrders(c.ImportsOrder)
		if err != nil {
			return nil, err
		}
		options = append(options, reviser.WithImportsOrder(order))
	}

	return options, nil
}

// ConfigParser parses command-line flags into Config
// Following SRP: parsing responsibility is isolated
type ConfigParser struct {
	flagSet *flag.FlagSet
	config  *Config
}

// NewConfigParser creates a new config parser
func NewConfigParser() *ConfigParser {
	cfg := &Config{}
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// File/Directory flags
	fs.StringVar(&cfg.FilePath, "file-path", "", "Deprecated. Put file name as an argument(last item) of command line.")
	fs.StringVar(&cfg.Excludes, "excludes", "", "Exclude files or dirs, example: '.git/,proto/*.go'.")

	// Project flags
	fs.StringVar(&cfg.ProjectName, "project-name", "", "Your project name(ex.: github.com/incu6us/goimports-reviser). Optional parameter.")
	fs.StringVar(&cfg.CompanyPkgPrefixes, "company-prefixes", "", "Company package prefixes which will be placed after 3rd-party group by default(if defined). Values should be comma-separated. Optional parameters.")
	fs.StringVar(&cfg.LocalPkgPrefixes, "local", "", "Deprecated")
	fs.StringVar(&cfg.Output, "output", "file", `Can be "file", "write" or "stdout". Whether to write the formatted content back to the file or to stdout. When "write" together with "-list-diff" will list the file name and write back to the file. Optional parameter.`)
	fs.StringVar(&cfg.ImportsOrder, "imports-order", "std,general,company,project", `Your imports groups can be sorted in your way.
std - std import group;
general - libs for general purpose;
company - inter-org or your company libs(if you set '-company-prefixes'-option, then 4th group will be split separately. In other case, it will be the part of general purpose libs);
project - your local project dependencies;
blanked - imports with "_" alias;
dotted - imports with "." alias.
Optional parameter.`)

	// Feature flags
	fs.BoolVar(&cfg.ShouldRemoveUnusedImports, "rm-unused", false, "Remove unused imports. Optional parameter.")
	fs.BoolVar(&cfg.ShouldSetAlias, "set-alias", false, "Set alias for versioned package names, like 'github.com/go-pg/pg/v9'. In this case import will be set as 'pg \"github.com/go-pg/pg/v9\"'. Optional parameter.")
	fs.BoolVar(&cfg.ShouldFormat, "format", false, "Option will perform additional formatting. Optional parameter.")
	fs.BoolVar(&cfg.ShouldSeparateNamedImports, "separate-named", false, "Option will separate named imports from the rest of the imports, per group. Optional parameter.")
	fs.BoolVar(&cfg.IsRecursive, "recursive", false, "Apply rules recursively if target is a directory. In case of ./... execution will be recursively applied by default. Optional parameter.")
	fs.BoolVar(&cfg.IsUseCache, "use-cache", false, "Use cache to improve performance. Optional parameter.")
	fs.BoolVar(&cfg.ShouldApplyToGeneratedFiles, "apply-to-generated-files", false, "Apply imports sorting and formatting(if the option is set) to generated files. Generated file is a file with first comment which starts with comment '// Code generated'. Optional parameter.")

	// Output flags
	fs.BoolVar(&cfg.ListFileName, "list-diff", false, "Option will list files whose formatting differs from goimports-reviser. Optional parameter.")
	fs.BoolVar(&cfg.SetExitStatus, "set-exit-status", false, "set the exit status to 1 if a change is needed/made. Optional parameter.")

	// Version flags
	fs.BoolVar(&cfg.ShowVersion, "version", false, "Show version information")
	fs.BoolVar(&cfg.ShowVersionOnly, "version-only", false, "Show only the version string")

	return &ConfigParser{
		flagSet: fs,
		config:  cfg,
	}
}

// Parse parses command-line arguments
func (p *ConfigParser) Parse(args []string) (*Config, *DeprecationMessages, error) {
	deprecations := &DeprecationMessages{}

	if err := p.flagSet.Parse(args); err != nil {
		return nil, nil, err
	}

	// Handle deprecated file-path flag
	if p.config.FilePath != "" {
		deprecations.Add(fmt.Sprintf("-file-path is deprecated. Put file name(s) as last argument to the command(Example: goimports-reviser -rm-unused -set-alias -format goimports-reviser/main.go)"))
		p.config.OriginPaths = append(p.config.OriginPaths, p.config.FilePath)
	}

	// Handle deprecated local flag
	if p.config.LocalPkgPrefixes != "" {
		if p.config.CompanyPkgPrefixes == "" {
			p.config.CompanyPkgPrefixes = p.config.LocalPkgPrefixes
		}
		deprecations.Add("-local is deprecated and will be removed soon. Use -company-prefixes instead.")
	}

	// Get positional arguments
	p.config.OriginPaths = append(p.config.OriginPaths, p.flagSet.Args()...)

	// Handle stdin
	if len(p.config.OriginPaths) == 1 && p.config.OriginPaths[0] == "-" {
		p.config.OriginPaths[0] = reviser.StandardInput
	}

	return p.config, deprecations, nil
}

// PrintUsage prints the usage information
func (p *ConfigParser) PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	p.flagSet.PrintDefaults()
}
