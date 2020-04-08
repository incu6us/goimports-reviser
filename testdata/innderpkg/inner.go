package innderpkg

import "github.com/pkg/errors"

func InnerFn(s string) (string, error) {
	return s, errors.New("some error")
}
