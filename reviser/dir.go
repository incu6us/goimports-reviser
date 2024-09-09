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

type walkCallbackFunc = func(hasChanged bool, path string, content []byte) error

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
}

func NewSourceDir(projectName string, path string, isRecursive bool, excludes string) *SourceDir {
	patterns := make([]string, 0)

	// get the absolute path
	absPath, err := filepath.Abs(path)

	// if path is recursive, then we need to remove the "/..." suffix
	if path == recursivePath {
		isRecursive = true
		absPath = strings.TrimSuffix(absPath, "/...")
	}

	if err == nil {
		segs := strings.Split(excludes, ",")
		for _, seg := range segs {
			p := strings.TrimSpace(seg)
			if p != "" {
				if !filepath.IsAbs(p) {
					// resolve the absolute path
					p = filepath.Join(absPath, p)
				}
				// Check pattern is well-formed.
				if _, err = filepath.Match(p, ""); err == nil {
					patterns = append(patterns, p)
				}
			}
		}
	}
	return &SourceDir{
		projectName:     projectName,
		dir:             absPath,
		isRecursive:     isRecursive,
		excludePatterns: patterns,
	}
}

func (d *SourceDir) Fix(options ...SourceFileOption) error {
	var ok bool
	d.dir, ok = IsDir(d.dir)
	if !ok {
		return ErrPathIsNotDir
	}
	err := filepath.WalkDir(d.dir, d.walk(
		func(hasChanged bool, path string, content []byte) error {
			if !hasChanged {
				return nil
			}
			if err := os.WriteFile(path, content, 0o644); err != nil {
				log.Fatalf("failed to write fixed result to file(%s): %+v\n", path, err)
				return err
			}
			return nil
		},
		options...,
	))
	if err != nil {
		return fmt.Errorf("failed to walk dif: %w", err)
	}

	return nil
}

// Find collection of bad formatted paths
func (d *SourceDir) Find(options ...SourceFileOption) (*UnformattedCollection, error) {
	var (
		ok                     bool
		badFormattedCollection []string
	)
	d.dir, ok = IsDir(d.dir)
	if !ok {
		return nil, ErrPathIsNotDir
	}
	err := filepath.WalkDir(d.dir, d.walk(
		func(hasChanged bool, path string, content []byte) error {
			if !hasChanged {
				return nil
			}
			badFormattedCollection = append(badFormattedCollection, path)
			return nil
		},
		options...,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to walk dif: %w", err)
	}

	if len(badFormattedCollection) == 0 {
		return nil, nil
	}

	return newUnformattedCollection(badFormattedCollection), nil
}

func (d *SourceDir) walk(callback walkCallbackFunc, options ...SourceFileOption) fs.WalkDirFunc {
	return func(path string, dirEntry fs.DirEntry, err error) error {
		if !d.isRecursive && dirEntry.IsDir() && filepath.Base(d.dir) != dirEntry.Name() {
			return filepath.SkipDir
		}
		if dirEntry.IsDir() && d.isExcluded(path) {
			return filepath.SkipDir
		}
		if isGoFile(path) && !dirEntry.IsDir() && !d.isExcluded(path) {
			content, _, hasChange, err := NewSourceFile(d.projectName, path).Fix(options...)
			if err != nil {
				return fmt.Errorf("failed to fix %s: %w", path, err)
			}
			return callback(hasChange, path, content)
		}
		return nil
	}
}

func (d *SourceDir) isExcluded(path string) bool {
	var absPath string
	if filepath.IsAbs(path) {
		absPath = path
	} else {
		absPath = filepath.Join(d.dir, path)
	}
	for _, pattern := range d.excludePatterns {
		matched, err := filepath.Match(pattern, absPath)
		if err == nil && matched {
			return true
		}
	}
	return false
}

type UnformattedCollection struct {
	list []string
}

func newUnformattedCollection(list []string) *UnformattedCollection {
	return &UnformattedCollection{
		list: list,
	}
}

func (c *UnformattedCollection) List() []string {
	list := make([]string, len(c.list))
	copy(list, c.list)
	return list
}

func (c *UnformattedCollection) String() string {
	if c == nil {
		return ""
	}

	var builder strings.Builder
	for i, file := range c.list {
		builder.WriteString(file)
		if len(c.list)-1 > i {
			builder.WriteString("\n")
		}
	}
	return builder.String()
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
