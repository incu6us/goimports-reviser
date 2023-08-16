package helper

import (
	"os"

	"github.com/incu6us/goimports-reviser/v3/pkg/module"
	"github.com/incu6us/goimports-reviser/v3/reviser"
)

type Option func() (string, error)

func OSGetwdOption() (string, error) {
	return os.Getwd()
}

func DetermineProjectName(projectName, filePath string, option Option) (string, error) {
	if filePath == reviser.StandardInput {
		var err error
		filePath, err = option()
		if err != nil {
			return "", err
		}
	}

	return module.DetermineProjectName(projectName, filePath)
}
