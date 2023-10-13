package reviser

type groupsImports struct {
	*common
	blanked []string
	dotted  []string
}

type common struct {
	std     []string
	general []string
	company []string
	project []string
}

func (c *common) defaultSorting() [][]string {
	return [][]string{c.std, c.general, c.company, c.project}
}
