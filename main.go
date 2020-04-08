package main

import (
	"fmt"
	"log"

	"github.com/incu6us/goimport-reviser/reviser"
)

const (
	projectName = "goimport-reviser"
	filePath    = "./testdata/example.go"
)

func main() {
	out, err := reviser.Execute(projectName, filePath)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	fmt.Println(string(out))
}
