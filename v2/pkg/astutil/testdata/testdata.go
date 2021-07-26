// +build test

package testdata

import (
	"fmt"

	"github.com/pkg/errors"
)

func main() {
	fmt.Printf("%s", errors.New("some error here"))
}
