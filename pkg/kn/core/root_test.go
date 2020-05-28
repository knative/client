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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/plugin"
)

func TestNewDefaultKnCommand(t *testing.T) {
	var rootCmd *cobra.Command

	setup := func(t *testing.T) {
		rootCmd, _ = NewDefaultKnCommand()
	}

	t.Run("returns a valid root command", func(t *testing.T) {
		setup(t)

		checkRootCmd(t, rootCmd)
	})
}

func TestNewDefaultKnCommandWithArgs(t *testing.T) {
	var (
		rootCmd       *cobra.Command
		pluginHandler plugin.PluginHandler
		args          []string
	)

	setup := func(t *testing.T) {
		rootCmd, _ = NewDefaultKnCommandWithArgs(NewKnCommand(), pluginHandler, args, os.Stdin, os.Stdout, os.Stderr)
	}

	t.Run("when pluginHandler is nil", func(t *testing.T) {
		args = []string{}
		setup(t)

		t.Run("returns a valid root command", func(t *testing.T) {
			checkRootCmd(t, rootCmd)
		})
	})

	t.Run("when pluginHandler is not nil", func(t *testing.T) {
		t.Run("when args empty", func(t *testing.T) {
			args = []string{}
			setup(t)

			t.Run("returns a valid root command", func(t *testing.T) {
				checkRootCmd(t, rootCmd)
			})
		})

		t.Run("when args not empty", func(t *testing.T) {
			var (
				pluginName, pluginPath, tmpPathDir string
				err                                error
			)

			pluginName = "fake-plugin-name"

			beforeEach := func(t *testing.T) {
				tmpPathDir, err = ioutil.TempDir("", "plugin_list")
				assert.Assert(t, err == nil)

				pluginPath = plugin.CreateTestPluginInPath(t, "kn-"+pluginName, plugin.KnTestPluginScript, plugin.FileModeExecutable, tmpPathDir)
			}

			afterEach := func(t *testing.T) {
				err = os.RemoveAll(tmpPathDir)
				assert.Assert(t, err == nil)
			}

			t.Run("when -h or --help option is present for plugin, return valid root command", func(t *testing.T) {
				helpOptions := []string{"-h", "--help"}
				for _, helpOption := range helpOptions {
					beforeEach(t)
					args = []string{pluginPath, pluginName, helpOption}
					setup(t)
					defer afterEach(t)

					checkRootCmd(t, rootCmd)
				}
			})

			t.Run("when --help option is present for normal command, return valid root command", func(t *testing.T) {
				helpOptions := []string{"-h", "--help"}
				for _, helpOption := range helpOptions {
					beforeEach(t)
					args = []string{"service", helpOption}
					setup(t)
					defer afterEach(t)

					checkRootCmd(t, rootCmd)
				}
			})

			t.Run("tries to handle args[1:] as plugin and return valid root command", func(t *testing.T) {
				beforeEach(t)
				args = []string{pluginPath, pluginName}
				setup(t)
				defer afterEach(t)

				checkRootCmd(t, rootCmd)
			})

			t.Run("when plugin extends an existing command group it return valid root command", func(t *testing.T) {
				pluginName = "service-fakecmd"
				beforeEach(t)
				args = []string{pluginPath, pluginName}
				setup(t)
				defer afterEach(t)

				checkRootCmd(t, rootCmd)
			})

			t.Run("when plugin extends and shadows an existing command group it fails", func(t *testing.T) {
				pluginName = "service-create"
				beforeEach(t)
				args = []string{pluginPath, pluginName, "test"}
				setup(t)
				defer afterEach(t)

				checkRootCmd(t, rootCmd)
			})
		})
	})
}

func TestNewKnCommand(t *testing.T) {
	var rootCmd *cobra.Command

	setup := func(t *testing.T) {
		rootCmd = NewKnCommand(commands.KnParams{})
	}

	t.Run("returns a valid root command", func(t *testing.T) {
		setup(t)
		checkRootCmd(t, rootCmd)
	})

	t.Run("sets the output params", func(t *testing.T) {
		setup(t)
		assert.Assert(t, rootCmd.OutOrStdout() != nil)
	})

	t.Run("sets the config and kubeconfig global flags", func(t *testing.T) {
		setup(t)
		assert.Assert(t, rootCmd.PersistentFlags().Lookup("config") != nil)
		assert.Assert(t, rootCmd.PersistentFlags().Lookup("kubeconfig") != nil)
	})

	t.Run("adds the top level commands: version and completion", func(t *testing.T) {
		setup(t)
		checkCommand(t, "version", rootCmd)
		checkCommand(t, "completion", rootCmd)
	})

	t.Run("adds the top level group commands", func(t *testing.T) {
		setup(t)
		checkCommandGroup(t, "service", rootCmd)
		checkCommandGroup(t, "revision", rootCmd)
	})
}

func TestEmptyAndUnknownSubCommands(t *testing.T) {
	var rootCmd, fakeCmd, fakeSubCmd *cobra.Command

	setup := func(t *testing.T) {
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

	t.Run("deals with empty and unknown sub-commands for all group commands", func(t *testing.T) {
		setup(t)
		EmptyAndUnknownSubCommands(rootCmd)
		checkCommand(t, "fake-sub-cmd-name", fakeCmd)
		checkCommandGroup(t, "fake-cmd-name", rootCmd)
	})
}

// Private

func checkRootCmd(t *testing.T, rootCmd *cobra.Command) {
	assert.Assert(t, rootCmd != nil)

	assert.Equal(t, rootCmd.Name(), "kn")
	assert.Equal(t, rootCmd.Short, "Knative client")
	assert.Assert(t, strings.Contains(rootCmd.Long, "Manage your Knative building blocks:"))

	assert.Assert(t, rootCmd.DisableAutoGenTag)
	assert.Assert(t, rootCmd.SilenceUsage)
	assert.Assert(t, rootCmd.SilenceErrors)

	assert.Assert(t, rootCmd.Flags().Lookup("plugins-dir") != nil)
	assert.Assert(t, rootCmd.Flags().Lookup("lookup-plugins") != nil)

	assert.Assert(t, rootCmd.RunE == nil)
}

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
