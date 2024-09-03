// Copyright © 2020 The Knative Authors
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

package templates

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/test"
)

var groups = CommandGroups{
	{
		"header-1",
		[]*cobra.Command{{Use: "c0"}, {Use: "c1"}},
	},
	{
		"header-2",
		[]*cobra.Command{{Use: "c2"}},
	},
}

func TestAddTo(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	groups.AddTo(rootCmd)

	for idx, cmd := range rootCmd.Commands() {
		assert.Equal(t, cmd.Name(), fmt.Sprintf("c%d", idx))
	}
}

func TestSetUsage(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root", Short: "root", Run: func(cmd *cobra.Command, args []string) {}}
	groups.AddTo(rootCmd)
	groups.SetRootUsage(rootCmd, getTestFuncMap())

	for _, cmd := range rootCmd.Commands() {
		assert.Assert(t, cmd.DisableFlagsInUseLine)
	}

	capture := test.CaptureOutput(t)
	err := (rootCmd.UsageFunc())(rootCmd)
	assert.NilError(t, err)
	stdOut, stdErr := capture.Close()
	assert.Equal(t, stdErr, "")
	assert.Assert(t, util.ContainsAll(stdOut, "header-1", "header-2"))

	capture = test.CaptureOutput(t)
	(rootCmd.HelpFunc())(rootCmd, nil)
	stdOut, stdErr = capture.Close()
	assert.Equal(t, stdErr, "")
	assert.Assert(t, util.ContainsAll(stdOut, "root", "header-1", "header-2"))
}

func getTestFuncMap() *template.FuncMap {
	fMap := template.FuncMap{
		"listPlugins": func(c *cobra.Command) string {
			return ""
		},
	}
	return &fMap
}
