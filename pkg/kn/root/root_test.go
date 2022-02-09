// Copyright © 2019 The Knative Authors
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

package root

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"

	"knative.dev/client/pkg/util"
)

func TestNewRootCommand(t *testing.T) {
	os.Args = []string{"kn"}
	rootCmd, err := NewRootCommand(nil)
	assert.NilError(t, err)
	if rootCmd == nil {
		t.Fatal("rootCmd = nil, want not nil")
	}

	assert.Equal(t, rootCmd.Name(), "kn")
	assert.Assert(t, util.ContainsAll(rootCmd.Short, "Knative", "Serving", "Eventing"))
	assert.Assert(t, util.ContainsAll(rootCmd.Long, "Knative", "Serving", "Eventing"))

	assert.Assert(t, rootCmd.DisableAutoGenTag)
	assert.Assert(t, rootCmd.SilenceUsage)
	assert.Assert(t, rootCmd.SilenceErrors)

	assert.Assert(t, rootCmd.OutOrStdout() != nil)

	assert.Assert(t, rootCmd.PersistentFlags().Lookup("config") != nil)
	assert.Assert(t, rootCmd.PersistentFlags().Lookup("kubeconfig") != nil)
	assert.Assert(t, rootCmd.PersistentFlags().Lookup("context") != nil)
	assert.Assert(t, rootCmd.PersistentFlags().Lookup("cluster") != nil)

	assert.Assert(t, rootCmd.RunE == nil)

	fErrorFunc := rootCmd.FlagErrorFunc()
	err = fErrorFunc(rootCmd, errors.New("test"))
	assert.Equal(t, err.Error(), "test for 'kn'")
}

func TestSubCommands(t *testing.T) {
	rootCmd, err := NewRootCommand(nil)
	assert.NilError(t, err)
	checkLeafCommand(t, "version", rootCmd)
}

func TestCommandGroup(t *testing.T) {
	rootCmd, err := NewRootCommand(nil)
	assert.NilError(t, err)
	commandGroups := []string{
		"service", "revision", "plugin", "source", "source apiserver",
		"source sinkbinding", "source ping", "trigger",
	}
	for _, group := range commandGroups {
		cmds := strings.Split(group, " ")
		checkCommandGroup(t, cmds, rootCmd)
	}
}

func TestEmptyAndUnknownSubCommands(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "root",
	}
	fakeGroupCmd := &cobra.Command{
		Use: "fake-group",
	}
	fakeSubCmd := &cobra.Command{
		Use: "fake-subcommand",
	}
	fakeGroupCmd.AddCommand(fakeSubCmd)
	rootCmd.AddCommand(fakeGroupCmd)

	err := validateCommandStructure(rootCmd)
	assert.NilError(t, err)
	checkLeafCommand(t, "fake-subcommand", fakeGroupCmd)
	checkCommandGroup(t, []string{"fake-group"}, rootCmd)
}

func TestCommandGroupWithRunMethod(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "root",
	}
	fakeGroupCmd := &cobra.Command{
		Use: "fake-group",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	fakeSubCmd := &cobra.Command{
		Use: "fake-subcommand",
	}
	fakeGroupCmd.AddCommand(fakeSubCmd)
	rootCmd.AddCommand(fakeGroupCmd)

	err := validateCommandStructure(rootCmd)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), fakeGroupCmd.Name(), "internal", "not enable"))
}

// Private

func checkLeafCommand(t *testing.T, name string, rootCmd *cobra.Command) {
	cmd, _, err := rootCmd.Find([]string{name})
	assert.Assert(t, err == nil)
	if cmd == nil {
		t.Fatal("cmd = nil, want not nil")
	}
	assert.Assert(t, !cmd.HasSubCommands())
}

func checkCommandGroup(t *testing.T, commands []string, rootCmd *cobra.Command) {
	cmd, _, err := rootCmd.Find(commands)
	assert.Assert(t, err == nil)
	if cmd == nil {
		t.Fatal("cmd = nil, want not nil")
	}
	assert.Assert(t, cmd.RunE != nil)
	assert.Assert(t, cmd.HasSubCommands())

	cmd.SetHelpFunc(func(command *cobra.Command, i []string) {}) // Avoid output noise when running the test
	err = cmd.RunE(cmd, []string{})

	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "no", "sub-command", cmd.Name()))

	err = cmd.RunE(cmd, []string{"deeper"})
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "deeper", "unknown", "sub-command", cmd.Name()))
}

func TestRootCommandForBinaryNames(t *testing.T) {
	for _, test := range []struct {
		arg        string
		binaryName string
	}{
		{"kn", "kn"},
		{"kn1", "kn1"},
		{"/usr/bin/mykn", "mykn"},
	} {
		os.Args = []string{test.arg}
		rootCmd, err := NewRootCommand(nil)
		assert.NilError(t, err)
		if rootCmd == nil {
			t.Fatal("rootCmd = nil, want not nil")
		}

		assert.Equal(t, rootCmd.Name(), test.binaryName)
	}

}
