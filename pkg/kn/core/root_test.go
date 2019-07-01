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

package core

import (
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestNewKnCommand(t *testing.T) {
	var rootCmd *cobra.Command

	setup := func() {
		rootCmd = NewKnCommand(commands.KnParams{})
	}

	setup()

	t.Run("returns a valid root command", func(t *testing.T) {
		assert.Assert(t, rootCmd != nil)

		assert.Equal(t, rootCmd.Name(), "kn")
		assert.Equal(t, rootCmd.Short, "Knative client")
		assert.Assert(t, strings.Contains(rootCmd.Long, "Manage your Knative building blocks:"))

		assert.Assert(t, rootCmd.DisableAutoGenTag)
		assert.Assert(t, rootCmd.SilenceUsage)
		assert.Assert(t, rootCmd.SilenceErrors)

		assert.Assert(t, rootCmd.RunE == nil)
	})

	t.Run("sets the output params", func(t *testing.T) {
		assert.Assert(t, rootCmd.OutOrStdout() != nil)
	})

	t.Run("sets the config and kubeconfig global flags", func(t *testing.T) {
		assert.Assert(t, rootCmd.PersistentFlags().Lookup("config") != nil)
		assert.Assert(t, rootCmd.PersistentFlags().Lookup("kubeconfig") != nil)
	})

	t.Run("adds the top level commands: version and completion", func(t *testing.T) {
		checkCommand(t, "version", rootCmd)
		checkCommand(t, "completion", rootCmd)
	})

	t.Run("adds the top level group commands", func(t *testing.T) {
		checkCommandGroup(t, "service", rootCmd)
		checkCommandGroup(t, "revision", rootCmd)
	})
}

func TestEmptyAndUnknownSubCommands(t *testing.T) {
	var rootCmd, fakeCmd, fakeSubCmd *cobra.Command

	setup := func() {
		rootCmd = NewKnCommand(commands.KnParams{})
		fakeCmd = &cobra.Command{
			Use: "fake-cmd-name",
		}
		fakeSubCmd = &cobra.Command{
			Use: "fake-sub-cmd-name",
		}
		fakeCmd.AddCommand(fakeSubCmd)
		rootCmd.AddCommand(fakeCmd)

		assert.Assert(t, fakeCmd.RunE == nil)
		assert.Assert(t, fakeSubCmd.RunE == nil)
	}

	setup()

	t.Run("deals with empty and unknown sub-commands for all group commands", func(t *testing.T) {
		EmptyAndUnknownSubCommands(rootCmd)
		checkCommand(t, "fake-sub-cmd-name", fakeCmd)
		checkCommandGroup(t, "fake-cmd-name", rootCmd)
	})
}

// Private

func checkCommand(t *testing.T, name string, rootCmd *cobra.Command) {
	cmd, _, err := rootCmd.Find([]string{"version"})
	assert.Assert(t, err == nil)
	assert.Assert(t, cmd != nil)
}

func checkCommandGroup(t *testing.T, name string, rootCmd *cobra.Command) {
	cmd, _, err := rootCmd.Find([]string{name})
	assert.Assert(t, err == nil)
	assert.Assert(t, cmd != nil)
	assert.Assert(t, cmd.RunE != nil)
	assert.Assert(t, cmd.HasSubCommands())
}
