package module

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

const goModFilename = "go.mod"

func Name(goModRootPath string) (string, error) {
	goModFile := filepath.Join(goModRootPath, goModFilename)

	data, err := ioutil.ReadFile(goModFile)
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
