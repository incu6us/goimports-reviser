package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/incu6us/goimports-reviser/v3/helper"
	"github.com/incu6us/goimports-reviser/v3/internal/cache"
	"github.com/incu6us/goimports-reviser/v3/internal/output"
	"github.com/incu6us/goimports-reviser/v3/reviser"
)

// Processor handles file processing operations
// Following SRP: single responsibility for processing files
// Following DIP: depends on abstractions (interfaces)
type Processor struct {
	cacheManager  cache.Manager
	outputHandler output.Handler
}

// NewProcessor creates a new file processor
func NewProcessor(cacheManager cache.Manager, outputHandler output.Handler) *Processor {
	return &Processor{
		cacheManager:  cacheManager,
		outputHandler: outputHandler,
	}
}

// ProcessResult holds the result of processing
type ProcessResult struct {
	HasChange bool
	Error     error
}

// ProcessFile processes a single file
// Following SRP: focused on single file processing
func (p *Processor) ProcessFile(projectName, filePath string, options reviser.SourceFileOptions) ProcessResult {
	// Make path absolute unless it's standard input
	if filePath != reviser.StandardInput {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return ProcessResult{Error: fmt.Errorf("failed to get abs path: %w", err)}
		}
		filePath = absPath
	}

	// Check cache first
	if filePath != reviser.StandardInput {
		fileContent, err := os.ReadFile(filePath)
		if err == nil && p.cacheManager.IsCached(filePath, fileContent) {
			return ProcessResult{HasChange: false}
		}
	}

	// Process the file
	formattedOutput, _, hasChange, err := reviser.NewSourceFile(projectName, filePath).Fix(options...)
	if err != nil {
		return ProcessResult{Error: fmt.Errorf("failed to fix file: %w", err)}
	}

	// Update cache
	if filePath != reviser.StandardInput {
		if err := p.cacheManager.UpdateCache(filePath, formattedOutput); err != nil {
			return ProcessResult{Error: fmt.Errorf("failed to update cache: %w", err)}
		}
	}

	// Write output
	if err := p.outputHandler.Write(filePath, formattedOutput, hasChange); err != nil {
		return ProcessResult{Error: err}
	}

	return ProcessResult{HasChange: hasChange}
}

// ProcessDirectory processes a directory
// Following SRP: focused on directory processing
func (p *Processor) ProcessDirectory(projectName, dirPath string, recursive bool, excludes string, listDiff bool, options reviser.SourceFileOptions) ProcessResult {
	if listDiff {
		unformattedFiles, err := reviser.NewSourceDir(projectName, dirPath, recursive, excludes).Find(options...)
		if err != nil {
			return ProcessResult{Error: fmt.Errorf("failed to find unformatted files %s: %w", dirPath, err)}
		}

		if unformattedFiles != nil {
			fmt.Printf("%s\n", unformattedFiles.String())
			return ProcessResult{HasChange: true}
		}

		return ProcessResult{HasChange: false}
	}

	err := reviser.NewSourceDir(projectName, dirPath, recursive, excludes).Fix(options...)
	if err != nil {
		return ProcessResult{Error: fmt.Errorf("failed to fix directory %s: %w", dirPath, err)}
	}

	return ProcessResult{HasChange: false}
}

// ProcessPaths processes multiple paths (files or directories)
// Following SRP: orchestrates processing of multiple paths
func (p *Processor) ProcessPaths(paths []string, projectNameFlag string, recursive bool, excludes string, listDiff bool, options reviser.SourceFileOptions) (bool, error) {
	var hasChange bool

	for _, originPath := range paths {
		projectName, err := helper.DetermineProjectName(projectNameFlag, originPath, helper.OSGetwdOption)
		if err != nil {
			return false, fmt.Errorf("could not determine project name for path %s: %w", originPath, err)
		}

		// Check if path is a directory
		if _, ok := reviser.IsDir(originPath); ok {
			result := p.ProcessDirectory(projectName, originPath, recursive, excludes, listDiff, options)
			if result.Error != nil {
				return false, result.Error
			}
			if result.HasChange {
				hasChange = true
			}
			continue
		}

		// Process as file
		result := p.ProcessFile(projectName, originPath, options)
		if result.Error != nil {
			return false, result.Error
		}
		if result.HasChange {
			hasChange = true
		}
	}

	return hasChange, nil
}
