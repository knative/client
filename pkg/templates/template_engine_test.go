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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/test"
)

type testData struct {
	cmd      *cobra.Command
	validate func(*testing.T, string, *cobra.Command)
}

func TestUsageFunc(t *testing.T) {
	rootCmd, engine := newTestTemplateEngine()
	subCmdWithSubs, _, _ := rootCmd.Find([]string{"g1.1"})
	subCmd, _, _ := rootCmd.Find([]string{"g2.1"})

	data := []testData{
		{
			rootCmd,
			func(t *testing.T, out string, command *cobra.Command) {
				validateRootUsageOutput(t, out)
			},
		},
		{
			subCmd,
			func(t *testing.T, out string, command *cobra.Command) {
				validateSubUsageOutput(t, out, command)
			},
		},
		{
			subCmdWithSubs,
			func(t *testing.T, out string, command *cobra.Command) {
				validateSubUsageOutput(t, out, command)
				subsub := command.Commands()[0]
				assert.Assert(t, util.ContainsAll(out, subsub.Name(), subsub.Short, "Available Commands:"))
			},
		},
	}
	for _, d := range data {
		capture := test.CaptureOutput(t)
		err := (engine.usageFunc())(d.cmd)
		assert.NilError(t, err)
		stdOut, stdErr := capture.Close()

		assert.Equal(t, stdErr, "")
		d.validate(t, stdOut, d.cmd)
	}
}

func TestHelpFunc(t *testing.T) {
	rootCmd, engine := newTestTemplateEngine()
	subCmd := rootCmd.Commands()[0]

	data := []testData{
		{
			rootCmd,
			func(t *testing.T, out string, command *cobra.Command) {
				validateRootUsageOutput(t, out)
				assert.Assert(t, strings.Contains(out, command.Long))
			},
		},
		{
			subCmd,
			func(t *testing.T, out string, command *cobra.Command) {
				validateSubUsageOutput(t, out, command)
				assert.Assert(t, strings.Contains(out, command.Long))
			},
		},
	}
	for _, d := range data {
		capture := test.CaptureOutput(t)
		(engine.helpFunc())(d.cmd, []string{})
		stdOut, stdErr := capture.Close()

		assert.Equal(t, stdErr, "")
		d.validate(t, stdOut, d.cmd)
	}
}

func TestOptionsFunc(t *testing.T) {
	rootCmd, _ := newTestTemplateEngine()
	subCmd := rootCmd.Commands()[0]
	capture := test.CaptureOutput(t)
	err := NewGlobalOptionsFunc()(subCmd)
	assert.NilError(t, err)
	stdOut, stdErr := capture.Close()

	assert.Equal(t, stdErr, "")
	assert.Assert(t, util.ContainsAll(stdOut, "options", "any command", "--global-opt", "global option"))
}

func TestUsageFlags(t *testing.T) {
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	f.StringP("test", "t", "default", "test-option")
	usage := flagsUsagesKubectl(f)
	assert.Equal(t, usage, "  -t, --test='default': test-option\n")
	usage = flagsUsagesCobra(f)
	assert.Equal(t, usage, "  -t, --test string   test-option (default \"default\")\n")

}
func validateRootUsageOutput(t *testing.T, stdOut string) {
	assert.Assert(t, util.ContainsAll(stdOut, "root"))
	assert.Assert(t, util.ContainsAll(stdOut, "header-1", "g1.1", "desc-g1.1", "g1.2", "desc-g1.2"))
	assert.Assert(t, util.ContainsAll(stdOut, "header-2", "g2.1", "desc-g2.1", "g2.2", "desc-g2.2", "g2.3", "desc-g2.3"))
	assert.Assert(t, util.ContainsAll(stdOut, "Use", "root", "--help"))
	assert.Assert(t, util.ContainsAll(stdOut, "Use", "root", "options", "global"))
}

func validateSubUsageOutput(t *testing.T, stdOut string, cmd *cobra.Command) {
	assert.Assert(t, util.ContainsAll(stdOut, "Usage", cmd.CommandPath()+" [options]"))
	assert.Assert(t, util.ContainsAll(stdOut, "Options", "--local-opt", "local option"))
	assert.Assert(t, util.ContainsAll(stdOut, "Use", "root", "options", "global"))
	assert.Assert(t, util.ContainsAll(stdOut, "Aliases", "alias"))
}

func newTestTemplateEngine() (*cobra.Command, templateEngine) {
	rootCmd := &cobra.Command{Use: "root", Short: "desc-root", Long: "longdesc-root"}
	rootCmd.PersistentFlags().String("global-opt", "", "global option")
	cmdGroups := CommandGroups{
		{
			"header-1",
			[]*cobra.Command{newCmd("g1.1"), newCmd("g1.2")},
		},
		{
			"header-2",
			[]*cobra.Command{newCmd("g2.1"), newCmd("g2.2"), newCmd("g2.3")},
		},
	}
	engine := newTemplateEngine(rootCmd, cmdGroups, getTestFuncMap())
	cmdGroups.AddTo(rootCmd)

	// Add a sub-command to first command
	cmd, _, _ := rootCmd.Find([]string{"g1.1"})
	cmd.AddCommand(newCmd("g1.1.1"))

	rootCmd.SetUsageFunc(engine.usageFunc())

	return rootCmd, engine
}

func newCmd(name string) *cobra.Command {
	ret := &cobra.Command{
		Use:     name,
		Short:   "desc-" + name,
		Long:    "longdesc-" + name,
		Run:     func(cmd *cobra.Command, args []string) {},
		Aliases: []string{"alias"},
	}
	ret.Flags().String("local-opt", "", "local option")
	return ret
}
