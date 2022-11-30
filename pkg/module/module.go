package module

import (
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

const goModFilename = "go.mod"

// Name reads module value from ./go.mod
func Name(goModRootPath string) (string, error) {
	goModFile := filepath.Join(goModRootPath, goModFilename)

	data, err := os.ReadFile(goModFile)
	if err != nil {
		return "", err
	}

	f, err := modfile.Parse(goModFile, data, nil)
	if err != nil {
		return "", err
	}

	if f.Module != nil {
		return f.Module.Mod.Path, nil
	}

	return "", &UndefinedModuleError{}
}

// GoModRootPath in case of any directory or file of the project will return root dir of the project where go.mod file
// is exist
func GoModRootPath(path string) (string, error) {
	if path == "" {
		return "", &PathIsNotSetError{}
	}

	path = filepath.Clean(path)

	for {
		if fi, err := os.Stat(filepath.Join(path, goModFilename)); err == nil && !fi.IsDir() {
			return path, nil
		}

		d := filepath.Dir(path)
		if d == path {
			break
		}

		path = d
	}

	return "", nil
}

func DetermineProjectName(projectName, filePath string) (string, error) {
	if projectName == "" {
		projectRootPath, err := GoModRootPath(filePath)
		if err != nil {
			return "", err
		}

		moduleName, err := Name(projectRootPath)
		if err != nil {
			return "", err
		}

		return moduleName, nil
	}

	return projectName, nil
}
