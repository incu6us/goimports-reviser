package reviser

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	goExtension   = ".go"
	recursivePath = "./..."
)

var (
	ErrPathIsNotDir = errors.New("path is not a directory")
)

// SourceDir to validate and fix import
type SourceDir struct {
	projectName string
	dir         string
	isRecursive bool
}

func NewSourceDir(projectName string, path string, isRecursive bool) *SourceDir {
	if path == recursivePath {
		isRecursive = true
	}
	return &SourceDir{projectName: projectName, dir: path, isRecursive: isRecursive}
}

func (d *SourceDir) Fix(options ...SourceFileOption) error {
	var ok bool
	d.dir, ok = IsDir(d.dir)
	if !ok {
		return ErrPathIsNotDir
	}

	err := filepath.WalkDir(d.dir, d.walk(options...))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (d *SourceDir) walk(options ...SourceFileOption) fs.WalkDirFunc {
	return func(path string, dirEntry fs.DirEntry, err error) error {
		if !d.isRecursive && dirEntry.IsDir() && filepath.Base(d.dir) != dirEntry.Name() {
			return filepath.SkipDir
		}
		if isGoFile(path) && !dirEntry.IsDir() {
			_, _, err := NewSourceFile(d.projectName, path).Fix(options...)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}

func IsDir(path string) (string, bool) {
	if path == recursivePath {
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
