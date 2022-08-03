package reviser

import "strings"

// Option is an int alias for options
type Option func(f *SourceFile) error

// Options is a slice of executing options
type Options []Option

// WithRemovingUnusedImports is an option to remove unused imports
func WithRemovingUnusedImports(f *SourceFile) error {
	f.shouldRemoveUnusedImports = true
	return nil
}

// WithUsingAliasForVersionSuffix is an option to set explicit package name in imports
func WithUsingAliasForVersionSuffix(f *SourceFile) error {
	f.shouldUseAliasForVersionSuffix = true
	return nil
}

// WithCodeFormatting use to format the code
func WithCodeFormatting(f *SourceFile) error {
	f.shouldFormatCode = true
	return nil
}

// WithCompanyPackagePrefixes option for 3d group(by default), like inter-org or company package prefixes
func WithCompanyPackagePrefixes(s string) Option {
	return func(f *SourceFile) error {
		prefixes := strings.Split(s, stringValueSeparator)
		for _, prefix := range prefixes {
			f.companyPackagePrefixes = append(f.companyPackagePrefixes, strings.TrimSpace(prefix))
		}
		return nil
	}
}

// WithImportsOrder will sort by needed order. Default order is "std,general,company,project"
func WithImportsOrder(orders []ImportsOrder) Option {
	return func(f *SourceFile) error {
		f.importsOrders = orders
		return nil
	}
}
