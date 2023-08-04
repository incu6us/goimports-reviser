package reviser

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	goExtension   = ".go"
	recursivePath = "./..."
)

var (
	currentPaths = []string{".", "." + string(filepath.Separator)}
)

var (
	ErrPathIsNotDir = errors.New("path is not a directory")
)

// SourceDir to validate and fix import
type SourceDir struct {
	projectName     string
	dir             string
	isRecursive     bool
	excludePatterns []string // see filepath.Match
	hasExcludes     bool
}

func NewSourceDir(projectName string, path string, isRecursive bool, excludes string) *SourceDir {
	if path == recursivePath {
		isRecursive = true
	}
	absPath, err := filepath.Abs(path)
	patterns := strings.Split(excludes, ",")
	if err == nil {
		for i := 0; i < len(patterns); i++ {
			patterns[i] = strings.TrimSpace(patterns[i])
			if !filepath.IsAbs(patterns[i]) {
				patterns[i] = filepath.Join(absPath, patterns[i])
			}
		}
	}

	return &SourceDir{
		projectName:     projectName,
		dir:             path,
		isRecursive:     isRecursive,
		excludePatterns: patterns,
		hasExcludes:     len(patterns) > 0,
	}
}

func (d *SourceDir) Fix(options ...SourceFileOption) error {
	var ok bool
	d.dir, ok = IsDir(d.dir)
	if !ok {
		return ErrPathIsNotDir
	}
	err := filepath.WalkDir(d.dir, d.walk(options...))
	if err != nil {
		return fmt.Errorf("failed to walk dif: %w", err)
	}

	return nil
}

func (d *SourceDir) walk(options ...SourceFileOption) fs.WalkDirFunc {
	return func(path string, dirEntry fs.DirEntry, err error) error {
		if !d.isRecursive && dirEntry.IsDir() && filepath.Base(d.dir) != dirEntry.Name() {
			return filepath.SkipDir
		}
		if dirEntry.IsDir() && d.isExcluded(path) {
			return filepath.SkipDir
		}
		if isGoFile(path) && !dirEntry.IsDir() && !d.isExcluded(path) {
			content, hasChange, err := NewSourceFile(d.projectName, path).Fix(options...)
			if err != nil {
				return fmt.Errorf("failed to fix: %w", err)
			}
			if hasChange {
				if err := os.WriteFile(path, content, 0644); err != nil {
					log.Fatalf("failed to write fixed result to file(%s): %+v\n", path, err)
				}
			}
		}
		return nil
	}
}

func (d *SourceDir) isExcluded(path string) bool {
	if d.hasExcludes {
		for _, pattern := range d.excludePatterns {
			matched, err := filepath.Match(pattern, path)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

func IsDir(path string) (string, bool) {
	if path == recursivePath || slices.Contains(currentPaths, path) {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return path, false
		}
	}

	dir, err := os.Open(path)
	if err != nil {
		return path, false
	}

	dirStat, err := dir.Stat()
	if err != nil {
		return path, false
	}

	return path, dirStat.IsDir()
}

func isGoFile(path string) bool {
	return filepath.Ext(path) == goExtension
}
