// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ingore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/scanner"
)

type Type struct {
	TypeName string
	FileName string
	TypeList []string
	MapCode  string
}

func main() {
	var types = []Type{
		Type{
			TypeName: "TiffType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "ImageType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "CompressType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "DataType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "TagType",
			FileName: "tiff_types.go",
		},

		Type{
			TypeName: "TagValue_PhotometricType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "TagValue_PredictorType",
			FileName: "tiff_types.go",
		},
		Type{
			TypeName: "TagValue_ResolutionUnitType",
			FileName: "tiff_types.go",
		},
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `
// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Auto generated by gen_helper.go, DO NOT EDIT!!!

package tiff

import (
	"fmt"
)

`[1:])

	for _, v := range types {
		v.Init()
		v.GenMapCode()

		fmt.Fprintf(&buf, "%s\n", v.MapCode)
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("z_tiff_types_string.go", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *Type) Init() {
	f, err := os.Open(p.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var s scanner.Scanner
	var typeMap = make(map[string]bool)

	s.Init(f)
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if tok&scanner.ScanIdents != 0 {
			if strings.HasPrefix(s.TokenText(), p.TypeName+"_") {
				if _, ok := typeMap[s.TokenText()]; !ok {
					p.TypeList = append(p.TypeList, s.TokenText())
					typeMap[s.TokenText()] = true
				}
			}
		}
	}
}

func (p *Type) GenMapCode() {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "var _%sTable = map[%s]string {\n", p.TypeName, p.TypeName)
	for _, s := range p.TypeList {
		fmt.Fprintf(&buf, "\t%s: `%s`,\n", s, s)
	}
	fmt.Fprintf(&buf, "}\n")

	fmt.Fprintf(&buf, `
func (p %s) String() string {
	if name, ok := _%sTable[p]; ok {
		return name
	}
	return fmt.Sprintf("%s_Unknown(%%d)", uint16(p))
}
`,
		p.TypeName,
		p.TypeName,
		p.TypeName,
	)

	p.MapCode = buf.String()
}
