// Tool for Golang to sort goimports by 3 groups: std, general and project dependencies.
// It will help you to keep your code cleaner.
//
// Example:
//	goimports-reviser -project-name github.com/incu6us/goimports-reviser -file-path ./reviser/reviser.go -rm-unused
//
// Input:
// 	import (
//		"log"
//
//		"github.com/incu6us/goimports-reviser/testdata/innderpkg"
//
//		"bytes"
//
//		"github.com/pkg/errors"
// 	)
//
// Output:
//
//	 import (
//		"bytes"
//		"log"
//
//		"github.com/pkg/errors"
//
//		"github.com/incu6us/goimports-reviser/testdata/innderpkg"
//	 )
//
// If you need to set package names explicitly(in import declaration), you can use additional option `-set-alias`.
//
// More:
//
// 	goimports-reviser -h
//
package main
