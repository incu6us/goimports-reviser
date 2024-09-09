package reviser

import (
	"fmt"
	"strings"
)

// ImportsOrder represents the name of import order
type ImportsOrder string

const (
	// StdImportsOrder is std libs, e.g. fmt, errors, strings...
	StdImportsOrder      ImportsOrder = "std"
	NamedStdImportsOrder ImportsOrder = "namedStd"
	// CompanyImportsOrder is packages that belong to the same organization
	CompanyImportsOrder      ImportsOrder = "company"
	NamedCompanyImportsOrder ImportsOrder = "namedCompany"
	// ProjectImportsOrder is packages that are inside the current project
	ProjectImportsOrder      ImportsOrder = "project"
	NamedProjectImportsOrder ImportsOrder = "namedProject"
	// GeneralImportsOrder is packages that are outside. In other words it is general purpose libraries
	GeneralImportsOrder      ImportsOrder = "general"
	NamedGeneralImportsOrder ImportsOrder = "namedGeneral"
	// BlankedImportsOrder is separate group for "_" imports
	BlankedImportsOrder ImportsOrder = "blanked"
	// DottedImportsOrder is separate group for "." imports
	DottedImportsOrder ImportsOrder = "dotted"
)

const (
	defaultImportsOrder = "std,general,company,project"
)

// ImportsOrders alias to []ImportsOrder
type ImportsOrders []ImportsOrder

func (o ImportsOrders) sortImportsByOrder(importGroups *groupsImports) [][]string {
	if len(o) == 0 {
		return importGroups.defaultSorting()
	}

	var result [][]string
	for _, group := range o {
		var imports []string
		switch group {
		case StdImportsOrder:
			imports = appendGroups(importGroups.std, importGroups.namedStd)
		case GeneralImportsOrder:
			imports = appendGroups(importGroups.general, importGroups.namedGeneral)
		case CompanyImportsOrder:
			imports = appendGroups(importGroups.company, importGroups.namedCompany)
		case ProjectImportsOrder:
			imports = appendGroups(importGroups.project, importGroups.namedProject)
		case BlankedImportsOrder:
			imports = importGroups.blanked
		case DottedImportsOrder:
			imports = importGroups.dotted
		}

		result = append(result, imports)
	}

	return result
}

func (o ImportsOrders) hasBlankedImportOrder() bool {
	for _, order := range o {
		if order == BlankedImportsOrder {
			return true
		}
	}
	return false
}

func (o ImportsOrders) hasDottedImportOrder() bool {
	for _, order := range o {
		if order == DottedImportsOrder {
			return true
		}
	}
	return false
}

func (o ImportsOrders) hasRequiredGroups() bool {
	var (
		hasStd     bool
		hasCompany bool
		hasGeneral bool
		hasProject bool
	)
	for _, order := range o {
		switch order {
		case StdImportsOrder:
			hasStd = true
		case CompanyImportsOrder:
			hasCompany = true
		case GeneralImportsOrder:
			hasGeneral = true
		case ProjectImportsOrder:
			hasProject = true
		}
	}
	return hasStd && hasCompany && hasGeneral && hasProject
}

// StringToImportsOrders will convert string, like "std,general,company,project" to ImportsOrder array type.
// Default value for empty string is "std,general,company,project"
func StringToImportsOrders(s string) (ImportsOrders, error) {
	if strings.TrimSpace(s) == "" {
		s = defaultImportsOrder
	}

	groups := unique(strings.Split(s, ","))

	var groupOrder []ImportsOrder
	for _, g := range groups {
		group := ImportsOrder(strings.TrimSpace(g))
		switch group {
		case StdImportsOrder, CompanyImportsOrder, ProjectImportsOrder,
			GeneralImportsOrder, BlankedImportsOrder, DottedImportsOrder:
		default:
			return nil, fmt.Errorf(`unknown order group type: %q`, group)
		}

		groupOrder = append(groupOrder, group)
	}

	if !ImportsOrders(groupOrder).hasRequiredGroups() {
		return nil, fmt.Errorf(`use default at least 4 parameters to sort groups of your imports: %q`, defaultImportsOrder)
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

func appendGroups(input ...[]string) []string {
	switch len(input) {
	case 0:
		return []string{}
	case 1:
		return input[0]
	default:
		break
	}
	separator := []string{"\n", "\n"}

	var output []string

	for idx, block := range input {
		if idx == 0 {
			output = append(output, block...)
			continue
		}
		if len(block) > 0 {
			output = append(output, append(separator, block...)...)
		}
	}

	return output
}
