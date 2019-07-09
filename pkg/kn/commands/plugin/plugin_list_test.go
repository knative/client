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

package plugin

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestPluginList(t *testing.T) {
	var (
		rootCmd, pluginCmd, pluginListCmd *cobra.Command
		tmpPathDir                        string
		err                               error
	)

	setup := func(t *testing.T) {
		knParams := &commands.KnParams{}
		pluginCmd = NewPluginCommand(knParams)
		assert.Assert(t, pluginCmd != nil)

		rootCmd, _, _ = commands.CreateTestKnCommand(pluginCmd, knParams)
		assert.Assert(t, rootCmd != nil)

		pluginListCmd = FindSubCommand(t, pluginCmd, "list")
		assert.Assert(t, pluginListCmd != nil)

		tmpPathDir, err = ioutil.TempDir("", "plugin_list")
		assert.Assert(t, err == nil)
	}

	cleanup := func(t *testing.T) {
		err = os.RemoveAll(tmpPathDir)
		assert.Assert(t, err == nil)
	}

	t.Run("creates a new cobra.Command", func(t *testing.T) {
		setup(t)
		defer cleanup(t)

		assert.Assert(t, pluginListCmd != nil)
		assert.Assert(t, pluginListCmd.Use == "list")
		assert.Assert(t, pluginListCmd.Short == "List all visible plugin executables")
		assert.Assert(t, strings.Contains(pluginListCmd.Long, "List all visible plugin executables"))
		assert.Assert(t, pluginListCmd.RunE != nil)
	})

	t.Run("when using $PATH as plugin location", func(t *testing.T) {
		var pluginPath string

		beforeEach := func(t *testing.T) {
			err = os.Setenv("PATH", tmpPathDir)
			assert.Assert(t, err == nil)
		}

		t.Run("no plugins installed", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			beforeEach(t)

			t.Run("warns user that no plugins found", func(t *testing.T) {
				rootCmd.SetArgs([]string{"plugin", "list"})
				err = rootCmd.Execute()
				assert.Assert(t, err != nil)
				assert.Assert(t, strings.Contains(err.Error(), "warning: unable to find any kn plugins in your pluginDir"))
			})
		})

		t.Run("plugins installed", func(t *testing.T) {
			t.Run("with valid plugin in $PATH", func(t *testing.T) {
				beforeEach := func(t *testing.T) {
					pluginPath = CreateTestPluginInPath(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
					assert.Assert(t, pluginPath != "")

					err = os.Setenv("PATH", tmpPathDir)
					assert.Assert(t, err == nil)
				}

				t.Run("list plugins in $PATH", func(t *testing.T) {
					setup(t)
					defer cleanup(t)
					beforeEach(t)

					rootCmd.SetArgs([]string{"plugin", "list"})
					err = rootCmd.Execute()
					assert.Assert(t, err == nil)

					//TODO: test output to contain the plugin
				})
			})

			t.Run("with non-executable plugin", func(t *testing.T) {
				beforeEach := func(t *testing.T) {
					pluginPath = CreateTestPluginInPath(t, KnTestPluginName, KnTestPluginScript, FileModeReadable, tmpPathDir)
					assert.Assert(t, pluginPath != "")
				}

				t.Run("warns user plugin invalid", func(t *testing.T) {
					setup(t)
					defer cleanup(t)
					beforeEach(t)

					rootCmd.SetArgs([]string{"plugin", "list"})
					err = rootCmd.Execute()
					assert.Assert(t, err != nil)
					assert.Assert(t, strings.Contains(err.Error(), "warning: unable to find any kn plugins in your pluginDir"))
				})
			})

			t.Run("with plugins with same name", func(t *testing.T) {
				var tmpPathDir2 string

				beforeEach := func(t *testing.T) {
					pluginPath = CreateTestPluginInPath(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
					assert.Assert(t, pluginPath != "")

					tmpPathDir2, err = ioutil.TempDir("", "plugin_list")
					assert.Assert(t, err == nil)

					err = os.Setenv("PATH", tmpPathDir+":"+tmpPathDir2)
					assert.Assert(t, err == nil)

					pluginPath = CreateTestPluginInPath(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir2)
					assert.Assert(t, pluginPath != "")
				}

				afterEach := func(t *testing.T) {
					err = os.RemoveAll(tmpPathDir)
					assert.Assert(t, err == nil)

					err = os.RemoveAll(tmpPathDir2)
					assert.Assert(t, err == nil)
				}

				t.Run("warns user about second (in $PATH) plugin shadowing first", func(t *testing.T) {
					setup(t)
					defer cleanup(t)
					beforeEach(t)
					defer afterEach(t)

					rootCmd.SetArgs([]string{"plugin", "list"})
					err = rootCmd.Execute()
					assert.Assert(t, err != nil)
					assert.Assert(t, strings.Contains(err.Error(), "error: one plugin warning was found"))
				})
			})

			t.Run("with plugins with name of existing command", func(t *testing.T) {
				var fakeCmd *cobra.Command

				beforeEach := func(t *testing.T) {
					fakeCmd = &cobra.Command{
						Use: "fake",
					}
					rootCmd.AddCommand(fakeCmd)

					pluginPath = CreateTestPluginInPath(t, "kn-fake", KnTestPluginScript, FileModeExecutable, tmpPathDir)
					assert.Assert(t, pluginPath != "")

					err = os.Setenv("PATH", tmpPathDir)
					assert.Assert(t, err == nil)
				}

				afterEach := func(t *testing.T) {
					rootCmd.RemoveCommand(fakeCmd)
				}

				t.Run("warns user about overwritting exising command", func(t *testing.T) {
					setup(t)
					defer cleanup(t)
					beforeEach(t)
					defer afterEach(t)

					rootCmd.SetArgs([]string{"plugin", "list"})
					err = rootCmd.Execute()
					assert.Assert(t, err != nil)
					assert.Assert(t, strings.Contains(err.Error(), "error: one plugin warning was found"))
				})
			})
		})
	})

	t.Run("when using pluginDir config variable", func(t *testing.T) {
		beforeEach := func(t *testing.T) {
			pluginPath = CreateTestPluginInPath(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
			assert.Assert(t, pluginPath != "")

			err = os.Setenv("PATH", "")
			assert.Assert(t, err == nil)
		}

		t.Run("list plugins in --plugin-dir", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			beforeEach(t)

			rootCmd.SetArgs([]string{"plugin", "list", "--plugin-dir", tmpPathDir})
			err = rootCmd.Execute()
			assert.Assert(t, err == nil)

			//TODO: test output to contain the plugin
		})

		t.Run("no plugins installed", func(t *testing.T) {
			setup(t)
			defer cleanup(t)

			rootCmd.SetArgs([]string{"plugin", "list", "--plugin-dir", tmpPathDir})
			err = rootCmd.Execute()
			assert.Assert(t, err != nil)

			//TODO: test output is no plugin found
		})
	})
}
