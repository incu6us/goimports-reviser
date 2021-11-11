package grouporder

import (
	"fmt"
	"strings"
)

// ImportGroupOrder represents order of import groups
type ImportGroupOrder []Group

// Group represents the name of import group
type Group string

var (
	// GroupStd is std libs, e.g. fmt, errors, strings...
	GroupStd Group = "std"
	// GroupOrganization is packages that belong to the same organization
	GroupOrganization Group = "org"
	// GroupProject is packages that are inside the current project
	GroupProject Group = "prj"
	// GroupExternal is packages that are outside for example from github
	GroupExternal Group = "ext"
)

const groupsCount = 4

// NewImportGroupOrder creates new ImportGroupOrder from given string. String example: "std,ext,org,prj".
func NewImportGroupOrder(s string) (ImportGroupOrder, error) {
	groups := strings.Split(s, ",")
	if len(groups) != groupsCount {
		return nil, fmt.Errorf("the order list should contain %d items", groupsCount)
	}

	result := make([]Group, 0, groupsCount)
	for _, g := range groups {
		group := Group(g)
		switch group {
		case GroupStd, GroupOrganization, GroupProject, GroupExternal:
		default:
			return nil, fmt.Errorf("unknown group type '%s'", group)
		}

		result = append(result, group)
	}

	return result, nil
}

// GroupedImports returns ordered groups of imports.
func (o ImportGroupOrder) GroupedImports(std, organization, project, external []string) [][]string {
	result := make([][]string, 0, groupsCount)
	for _, group := range o {
		var imports []string
		switch group {
		case GroupStd:
			imports = std
		case GroupOrganization:
			imports = organization
		case GroupProject:
			imports = project
		case GroupExternal:
			imports = external
		}

		result = append(result, imports)
	}

	return result
}
