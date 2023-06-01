//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

const subPackage = "."

type testData struct {
	Year      int
	Functions []string
}

var t = template.Must(template.New("allTests").Parse(`//go:build testing

package tests

/*
 * SPDX-FileCopyrightText: {{.Year}} Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

var AllTests = []PersistenceTest{
	{{ range $element := .Functions -}}
	{{ $element }},
	{{ end -}}
}
`))

func main() {
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, subPackage, nil, 0)
	if err != nil {
		fmt.Println("Failed to parse package:", err)
		os.Exit(1)
	}

	fns := make([]string, 0, 32)
	for _, pack := range packs {
		for _, f := range pack.Files {
			for _, d := range f.Decls {
				if fn, isFn := d.(*ast.FuncDecl); isFn {
					if strings.HasPrefix(fn.Name.Name, "Test") {
						fns = append(fns, fn.Name.Name)
					}
				}
			}
		}
	}

	sort.Strings(fns)
	now := time.Now()
	data := testData{
		Year:      now.Year(),
		Functions: fns,
	}
	f, err := os.OpenFile("all.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o444)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := t.Execute(f, data); err != nil {
		log.Fatal(err)
	}
}
