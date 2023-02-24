// Tool for Golang to sort goimports by 3-4 groups: std, general, local(which is optional) and project dependencies.
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
//		"golang.org/x/exp/slices"
// 	)
//
// Output:
//
//	 import (
//		"bytes"
//		"log"
//
//		"golang.org/x/exp/slices"
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
