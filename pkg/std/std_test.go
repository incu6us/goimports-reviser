package std

import "testing"

func TestIsStd(t *testing.T) {
	for pkg, want := range map[string]bool{
		"uuid":                       true, // added in Go 1.27, not in the generated list
		"encoding/json":              true,
		"golang.org/x/exp/slices":    false,
		"github.com/incu6us/reviser": false,
	} {
		if got := IsStd(pkg); got != want {
			t.Errorf("IsStd(%q) = %v, want %v", pkg, got, want)
		}
	}
}
