package version

import (
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
)

const modulePathRegex = `^github.com/[\w-]+/goimports-reviser(/v\d+)?@?`

// Info holds version build information
// Following SRP: single responsibility for version data
type Info struct {
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string
}

// Manager handles version operations
// Following SRP: encapsulates all version-related logic
type Manager struct {
	buildInfo       *Info
	modulePathMatch *regexp.Regexp
}

// NewManager creates a new version manager
func NewManager(tag, commit, sourceURL, goVersion string) *Manager {
	return &Manager{
		buildInfo: &Info{
			Tag:       tag,
			Commit:    commit,
			SourceURL: sourceURL,
			GoVersion: goVersion,
		},
		modulePathMatch: regexp.MustCompile(modulePathRegex),
	}
}

// GetVersionString returns the version string
// Following OCP: can be extended without modifying existing code
func (m *Manager) GetVersionString() (string, error) {
	if m.buildInfo.Tag != "" {
		return strings.TrimPrefix(m.buildInfo.Tag, "v"), nil
	}

	bi := m.getBuildInfo()
	myModule, err := m.getMyModuleInfo(bi)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(myModule.Version, "v"), nil
}

// GetFullVersionInfo returns detailed version information
func (m *Manager) GetFullVersionInfo() (string, error) {
	if m.buildInfo.Tag != "" {
		return fmt.Sprintf(
			"version: %s\nbuilt with: %s\ntag: %s\ncommit: %s\nsource: %s",
			strings.TrimPrefix(m.buildInfo.Tag, "v"),
			m.buildInfo.GoVersion,
			m.buildInfo.Tag,
			m.buildInfo.Commit,
			m.buildInfo.SourceURL,
		), nil
	}

	bi := m.getBuildInfo()
	myModule, err := m.getMyModuleInfo(bi)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"version: %s\nbuilt with: %s\ntag: %s\ncommit: %s\nsource: %s",
		strings.TrimPrefix(myModule.Version, "v"),
		bi.GoVersion,
		myModule.Version,
		"n/a",
		myModule.Path,
	), nil
}

// getBuildInfo retrieves build information from the binary
func (m *Manager) getBuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return bi
}

// getMyModuleInfo finds the module information for this application
func (m *Manager) getMyModuleInfo(bi *debug.BuildInfo) (*debug.Module, error) {
	if bi == nil {
		return nil, errors.New("no build info available")
	}

	// depending on the context in which we are called, the main module may not be set
	if bi.Main.Path != "" {
		return &bi.Main, nil
	}

	// if the main module is not set, we need to find the dep that contains our module
	for _, mod := range bi.Deps {
		if m.modulePathMatch.MatchString(mod.Path) {
			return mod, nil
		}
	}

	return nil, errors.New("no matching module found in build info")
}
