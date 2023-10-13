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
	// ProjectImportsOrder is packages that are inside the current project
	ProjectImportsOrder ImportsOrder = "project"
	// GeneralImportsOrder is packages that are outside. In other words it is general purpose libraries
	GeneralImportsOrder ImportsOrder = "general"
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
			imports = importGroups.std
		case GeneralImportsOrder:
			imports = importGroups.general
		case CompanyImportsOrder:
			imports = importGroups.company
		case ProjectImportsOrder:
			imports = importGroups.project
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
