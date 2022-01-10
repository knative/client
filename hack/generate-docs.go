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

	"github.com/spf13/cobra/doc"
	"knative.dev/client/pkg/kn/root"
)

func main() {
	rootCmd, err := root.NewRootCommand(nil)
	if err != nil {
		log.Panicf("can not create root command: %v", err)
	}

	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	err = doc.GenMarkdownTree(rootCmd, dir+"/docs/cmd/")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func emptyString(filename string) string {
	return ""
}
