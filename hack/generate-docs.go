// Copyright © 2018 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra/doc"
	"knative.dev/client/pkg/kn/core"
)

func main() {
	rootCmd := core.NewKnCommand()

	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	var withFrontMatter bool
	var err error
	if len(os.Args) > 2 {
		withFrontMatter, err = strconv.ParseBool(os.Args[2])
		if err != nil {
			log.Panicf("Invalid argument %s, has to be boolean to switch on/off generation of frontmatter", os.Args[2])
		}
	}
	prependFunc := emptyString
	if withFrontMatter {
		prependFunc = addFrontMatter
	}
	err = doc.GenMarkdownTreeCustom(rootCmd, dir+"/docs/cmd/", prependFunc, identity)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func emptyString(filename string) string {
	return ""
}

func addFrontMatter(fileName string) string {
	// Convert to a title
	title := filepath.Base(fileName)
	title = title[0 : len(title)-len(filepath.Ext(title))]
	title = strings.ReplaceAll(title, "_", " ")
	ret := `
---
title: "%s"
#linkTitle: "OPTIONAL_ALTERNATE_NAV_TITLE"
weight: 5
type: "docs"
---
`
	return fmt.Sprintf(ret, title)
}

func identity(s string) string {
	return s
}
