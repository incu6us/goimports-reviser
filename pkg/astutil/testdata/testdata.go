//go:build test

package testdata

import (
	"fmt"

	"golang.org/x/exp/slices"
)

func main() {
	fmt.Println(slices.IsSorted([]int{1, 2, 3}))
}
