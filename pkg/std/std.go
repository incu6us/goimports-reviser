package std

import "strings"

// IsStd reports whether pkg is a standard-library import path, using the same
// rule as cmd/go: the first path element contains no dot.
//
// StdPackages is a snapshot of the toolchain that generated it, so it goes
// stale on every Go release that adds a package (e.g. "uuid" in Go 1.27).
// This covers those without regenerating.
func IsStd(pkg string) bool {
	elem, _, _ := strings.Cut(pkg, "/")
	return !strings.Contains(elem, ".")
}
