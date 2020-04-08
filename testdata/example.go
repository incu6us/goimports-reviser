package testdata

import (
	"log"

	"github.com/incu6us/goimport-reviser/testdata/innderpkg"

	"bytes"

	"github.com/pkg/errors"
)

func Example(s string) string {
	b := bytes.NewBufferString(s)
	result, err := innderpkg.InnerFn(b.String())
	if err != nil {
		log.Fatal(errors.WithStack(err))
		return ""
	}

	return result
}
