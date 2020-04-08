// +build gen

package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

//go:generate go run -tags gen ./...

const (
	fileName = "../package_list.go"

	fileTemplate = `package helper

var StdPackages = map[string]struct{}{
{{- range $index, $element := .}}
	"\"{{$element}}\"": {},
{{- end}}
}

`
)

func main() {
	w := bytes.NewBufferString("")

	tpl := template.New("tpl")
	tpl, err := tpl.Parse(fileTemplate)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
		return
	}

	packageList, err := packages.Load(nil, "std")
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
		return
	}

	if err := tpl.Execute(w, packageList); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
		return
	}

	if err := ioutil.WriteFile(fileName, w.Bytes(), 0644); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}
}
