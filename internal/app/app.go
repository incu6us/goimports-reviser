package app

import (
	"fmt"
	"log"
	"os"

	"github.com/incu6us/goimports-reviser/v3/internal/cache"
	"github.com/incu6us/goimports-reviser/v3/internal/config"
	"github.com/incu6us/goimports-reviser/v3/internal/output"
	"github.com/incu6us/goimports-reviser/v3/internal/processor"
	"github.com/incu6us/goimports-reviser/v3/internal/version"
)

// Application represents the main application
// Following SRP: orchestrates components without implementing their logic
// Following DIP: depends on abstractions (interfaces)
type Application struct {
	configParser   *config.ConfigParser
	versionManager *version.Manager
}

// New creates a new application instance
func New(tag, commit, sourceURL, goVersion string) *Application {
	return &Application{
		configParser:   config.NewConfigParser(),
		versionManager: version.NewManager(tag, commit, sourceURL, goVersion),
	}
}

// Run runs the application with given arguments
// Following SRP: main orchestration logic
func (a *Application) Run(args []string) int {
	// Parse configuration
	cfg, deprecations, err := a.configParser.Parse(args)
	if err != nil {
		a.configParser.PrintUsage()
		log.Printf("%s\n", err)
		return 1
	}

	// Handle version flags
	if cfg.ShowVersionOnly {
		return a.handleVersionOnly()
	}

	if cfg.ShowVersion {
		return a.handleVersion()
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		a.configParser.PrintUsage()
		log.Printf("%s\n", err)
		return 1
	}

	// Convert config to reviser options
	options, err := cfg.ToReviserOptions()
	if err != nil {
		a.configParser.PrintUsage()
		log.Printf("%s\n", err)
		return 1
	}

	// Create cache manager
	cacheManager := a.createCacheManager(cfg.IsUseCache)

	// Create output handler
	outputHandler, err := a.createOutputHandler(cfg)
	if err != nil {
		log.Printf("%s\n", err)
		return 1
	}

	// Create processor
	proc := processor.NewProcessor(cacheManager, outputHandler)

	// Process files
	log.Printf("Paths: %v\n", cfg.OriginPaths)
	hasChange, err := proc.ProcessPaths(
		cfg.OriginPaths,
		cfg.ProjectName,
		cfg.IsRecursive,
		cfg.Excludes,
		cfg.ListFileName,
		options,
	)
	if err != nil {
		log.Printf("%s\n", err)
		return 1
	}

	// Print deprecation messages
	a.printDeprecations(deprecations)

	// Handle exit status
	if hasChange && cfg.SetExitStatus {
		return 1
	}

	return 0
}

// handleVersionOnly handles the version-only flag
func (a *Application) handleVersionOnly() int {
	versionStr, err := a.versionManager.GetVersionString()
	if err != nil {
		log.Printf("failed to get version: %s\n", err)
		return 1
	}
	fmt.Println(versionStr)
	return 0
}

// handleVersion handles the version flag
func (a *Application) handleVersion() int {
	versionInfo, err := a.versionManager.GetFullVersionInfo()
	if err != nil {
		log.Printf("failed to get version info: %s\n", err)
		return 1
	}
	fmt.Println(versionInfo)
	return 0
}

// createCacheManager creates appropriate cache manager based on configuration
// Following OCP: easy to extend with new cache types
func (a *Application) createCacheManager(useCache bool) cache.Manager {
	if !useCache {
		return cache.NewNoOpCacheManager()
	}

	cacheManager, err := cache.NewFileSystemCacheManager()
	if err != nil {
		log.Printf("Failed to create cache manager, proceeding without cache: %v\n", err)
		return cache.NewNoOpCacheManager()
	}

	return cacheManager
}

// createOutputHandler creates appropriate output handler based on configuration
// Following OCP: easy to extend with new output types
func (a *Application) createOutputHandler(cfg *config.Config) (output.Handler, error) {
	factory := output.NewFactory()

	// Special case: if list-diff is enabled and output is not "write", use DiffListHandler
	if cfg.ListFileName && cfg.Output != "write" {
		return output.NewDiffListHandler(), nil
	}

	// Determine if standard input
	isStandardIn := len(cfg.OriginPaths) == 1 && cfg.OriginPaths[0] == "-"

	outputCfg := output.Config{
		Mode:         output.OutputMode(cfg.Output),
		ListDiff:     cfg.ListFileName,
		IsStandardIn: isStandardIn,
	}

	return factory.Create(outputCfg)
}

// printDeprecations prints deprecation messages
func (a *Application) printDeprecations(deprecations *config.DeprecationMessages) {
	if !deprecations.HasMessages() {
		return
	}

	for _, msg := range deprecations.Messages() {
		log.Printf("%s\n", msg)
	}
	log.Printf("All changes to file are applied, but command-line syntax should be fixed\n")
	os.Exit(1)
}
