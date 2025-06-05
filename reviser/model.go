package reviser

type groupsImports struct {
	*common
	blanked []string
	dotted  []string
}

type common struct {
	std          []string
	namedStd     []string
	general      []string
	namedGeneral []string
	company      []string
	namedCompany []string
	project      []string
	namedProject []string
}

func (c *common) defaultSorting() [][]string {
	return [][]string{
		c.std, c.namedStd, c.general, c.namedGeneral, c.company, c.namedCompany, c.project, c.namedProject,
	}
}
