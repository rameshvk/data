// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// main creates html files from ds_watercolors.json or ds_oil.json
//
//    go run preview.go ../../ds_watercolors.json > preview.html
//    go run preview.go ../../ds_oil.json > preview.html
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	t := template.Must(template.New("preview.t").Funcs(template.FuncMap{
		"splitColors": func(s string) []template.CSS {
			results := strings.Split(s, " ")
			attrs := []template.CSS{}
			for _, result := range results {
				attrs = append(attrs, template.CSS("rgb("+result+")"))
			}
			return attrs
		},
	}).ParseFiles("preview.t"))

	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Failed to read json", os.Args[1], err)
		return
	}

	var data interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Println("Failed to read json", os.Args[1], err)
		return
	}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		fmt.Println("Failed to execute template", err)
	}
}
