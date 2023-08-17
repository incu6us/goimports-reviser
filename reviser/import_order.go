package reviser

import (
	"fmt"
	"strings"
)

// ImportsOrder represents the name of import order
type ImportsOrder string

const (
	// StdImportsOrder is std libs, e.g. fmt, errors, strings...
	StdImportsOrder ImportsOrder = "std"
	// CompanyImportsOrder is packages that belong to the same organization
	CompanyImportsOrder ImportsOrder = "company"
	// NamedImportsOrder is packages that are inside the current project
	NamedImportsOrder ImportsOrder = "named"
	// ProjectImportsOrder is packages that are inside the current project
	ProjectImportsOrder ImportsOrder = "project"
	// GeneralImportsOrder is packages that are outside. In other words it is general purpose libraries
	GeneralImportsOrder ImportsOrder = "general"
)

const (
	defaultImportsOrder = "std,general,company,named,project"
)

// ImportsOrders alias to []ImportsOrder
type ImportsOrders []ImportsOrder

func (o ImportsOrders) sortImportsByOrder(
	std []string,
	general []string,
	named []string,
	company []string,
	project []string,
) ([]string, []string, []string, []string, []string) {
	if len(o) == 0 {
		return std, general, named, company, project
	}

	var result [][]string
	for _, group := range o {
		var imports []string
		switch group {
		case StdImportsOrder:
			imports = std
		case GeneralImportsOrder:
			imports = general
		case CompanyImportsOrder:
			imports = company
		case NamedImportsOrder:
			imports = named
		case ProjectImportsOrder:
			imports = project
		}

		fmt.Printf("%v %v\n", group, imports)

		result = append(result, imports)
	}

	return result[0], result[1], result[2], result[3], result[4]
}

// StringToImportsOrders will convert string, like "std,general,company,named,project" to ImportsOrder array type.
// Default value for empty string is "std,general,company,named,project"
func StringToImportsOrders(s string) (ImportsOrders, error) {
	if len(strings.TrimSpace(s)) == 0 {
		s = defaultImportsOrder
	}

	groups := unique(strings.Split(s, ","))
	if len(groups) != 5 {
		return nil, fmt.Errorf(`use this parameters to sort all groups of your imports: "%s"`, defaultImportsOrder)
	}

	var groupOrder []ImportsOrder
	for _, g := range groups {
		group := ImportsOrder(strings.TrimSpace(g))
		switch group {
		case StdImportsOrder, CompanyImportsOrder, NamedImportsOrder, ProjectImportsOrder, GeneralImportsOrder:
		default:
			return nil, fmt.Errorf(`unknown order group type: "%s"`, group)
		}

		groupOrder = append(groupOrder, group)
	}

	return groupOrder, nil
}

func unique(s []string) []string {
	keys := make(map[string]struct{})
	var list []string
	for _, entry := range s {
		if _, ok := keys[entry]; !ok {
			list = append(list, entry)
		}
	}
	return list
}
