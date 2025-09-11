package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

const (
	rootDir      = "testdata"
	testCaseName = "case.yaml"
)

func TestIntegration(t *testing.T) {
	err := filepath.WalkDir(rootDir, func(path string, entry fs.DirEntry, err error) error {
		if !strings.HasSuffix(path, testCaseName) {
			return nil
		}

		scen, err := scenario(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		t.Run(dir+"|"+scen.Name, func(t *testing.T) {
			err = scen.run(t, dir)
			require.NoError(t, err)
		})
		return nil
	})
	if err != nil {
		t.Errorf("walk dir error: %s", err)
		return
	}
}

func scenario(path string) (*testCase, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var testScenario testCase
	err = yaml.Unmarshal(fileContent, &testScenario)
	if err != nil {
		return nil, err
	}
	return &testScenario, nil
}

type testCase struct {
	Name     string  `yaml:"name"`
	Args     string  `yaml:"args"`
	IsStdin  bool    `yaml:"is_stdin"`
	Actual   entries `yaml:"actual"` // required if IsStdin is set
	Expected entries `yaml:"expected"`
}

type entry struct {
	Filename string `yaml:"filename"`
	Status   int    `yaml:"status"`
	Content  string `yaml:"content"`
}

type entries []entry

func (e entries) byFilename(name string) (entry, error) {
	for _, content := range e {
		if content.Filename == name {
			return content, nil
		}
	}
	return entry{}, fmt.Errorf("content not found by filename: %s", name)
}

func (c *testCase) run(t *testing.T, rootDir string) error {
	err := c.validate()
	if err != nil {
		return err
	}
	for _, actual := range c.Actual {
		file := filepath.Join(rootDir, actual.Filename)
		err := os.WriteFile(file, []byte(actual.Content), 0o777)
		if err != nil {
			return err
		}
		params := strings.TrimSpace(c.Args)
		if !c.IsStdin {
			params += " " + file
		}
		_, exitStatus, err := execByParams(params)
		if err != nil {
			_ = os.Remove(file)
			return err
		}
		actualContent, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		expected, err := c.Expected.byFilename(actual.Filename)
		if err != nil {
			return err
		}
		expectedContent, err := format.Source([]byte(expected.Content))
		if err != nil {
			return err
		}
		assert.Equal(t, expected.Status, exitStatus)
		assert.Equal(t, string(expectedContent), string(actualContent))
		_ = os.Remove(file)
	}
	return nil
}

func (c *testCase) validate() error {
	if len(c.Actual) != len(c.Expected) {
		return fmt.Errorf("file names should match each other")
	}
	var isFound bool
	for _, actual := range c.Actual {
		for _, expected := range c.Expected {
			if actual.Filename == expected.Filename {
				isFound = true
				break
			}
		}
	}
	if !isFound {
		return fmt.Errorf("file names should match each other")
	}
	return nil
}

func execByParams(args string) (string, int, error) {
	cmd := exec.Command("go", append([]string{"run", "main.go"}, strings.Split(args, " ")...)...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Start()
	if err != nil {
		return "", -1, fmt.Errorf("execution error: %s", err)
	}
	if err = cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return errBuf.String(), exitErr.ExitCode(), nil
		}
	}
	return outBuf.String(), 0, nil
}
