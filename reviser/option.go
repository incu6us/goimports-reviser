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

// WithLocalPackagePrefix option for 4th group, like inter-org package prefixes
func WithLocalPackagePrefix(s string) Option {
	return func(i *SourceFile) error {
		prefixes := strings.Split(s, stringValueSeparator)
		for _, prefix := range prefixes {
			i.localPackagePrefixes = append(i.localPackagePrefixes, strings.TrimSpace(prefix))
		}
		return nil
	}
}
