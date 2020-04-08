package testdata

import (
	"bytes"
	"goimport-validator/testdata/innderpkg"
	"log"

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
