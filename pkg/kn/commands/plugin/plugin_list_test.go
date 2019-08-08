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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/util"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

type testContext struct {
	pluginsDir string
	pathDir    string
	rootCmd    *cobra.Command
	out        *bytes.Buffer
	origPath   string
}

func (ctx *testContext) execute(args ...string) error {
	ctx.rootCmd.SetArgs(append(args, fmt.Sprintf("--plugins-dir=%s", ctx.pluginsDir)))
	return ctx.rootCmd.Execute()
}

func (ctx *testContext) output() string {
	return ctx.out.String()
}

func (ctx *testContext) cleanup() {
	os.RemoveAll(ctx.pluginsDir)
	os.RemoveAll(ctx.pathDir)
	os.Setenv("PATH", ctx.origPath)
}

func (ctx *testContext) createTestPlugin(pluginName string, fileMode os.FileMode, inPath bool) error {
	path := ctx.pluginsDir
	if inPath {
		path = ctx.pathDir
	}
	return ctx.createTestPluginWithPath(pluginName, fileMode, path)
}

func (ctx *testContext) createTestPluginWithPath(pluginName string, fileMode os.FileMode, path string) error {
	if runtime.GOOS == "windows" {
		pluginName += ".bat"
	}
	fullPath := filepath.Join(path, pluginName)
	return ioutil.WriteFile(fullPath, []byte(KnTestPluginScript), fileMode)
}

func TestPluginList(t *testing.T) {

	setup := func(t *testing.T) *testContext {
		knParams := &commands.KnParams{}
		pluginCmd := NewPluginCommand(knParams)

		rootCmd, _, out := commands.CreateTestKnCommand(pluginCmd, knParams)
		pluginsDir, err := ioutil.TempDir("", "plugin-list-plugindir")
		assert.NilError(t, err)
		pathDir, err := ioutil.TempDir("", "plugin-list-pathdir")
		assert.NilError(t, err)

		origPath := os.Getenv("PATH")
		assert.NilError(t, os.Setenv("PATH", pathDir))

		return &testContext{
			rootCmd:    rootCmd,
			out:        out,
			pluginsDir: pluginsDir,
			pathDir:    pathDir,
			origPath:   origPath,
		}
	}

	t.Run("creates a new cobra.Command", func(t *testing.T) {
		pluginCmd := NewPluginCommand(&commands.KnParams{})
		pluginListCmd := FindSubCommand(t, pluginCmd, "list")
		assert.Assert(t, pluginListCmd != nil)

		assert.Assert(t, pluginListCmd != nil)
		assert.Assert(t, pluginListCmd.Use == "list")
		assert.Assert(t, pluginListCmd.Short == "List plugins")
		assert.Assert(t, strings.Contains(pluginListCmd.Long, "List all installed plugins"))
		assert.Assert(t, pluginListCmd.Flags().Lookup("plugins-dir") != nil)
		assert.Assert(t, pluginListCmd.RunE != nil)
	})

	t.Run("when pluginsDir does not include any plugins", func(t *testing.T) {
		t.Run("when --lookup-plugins-in-path is true", func(t *testing.T) {
			t.Run("no plugins installed", func(t *testing.T) {

				t.Run("warns user that no plugins found in verbose mode", func(t *testing.T) {
					ctx := setup(t)
					defer ctx.cleanup()
					err := ctx.execute("plugin", "list", "--lookup-plugins-in-path=true", "--verbose")
					assert.NilError(t, err)
					assert.Assert(t, util.ContainsAll(ctx.output(), "No plugins found"))
				})

				t.Run("no output when no plugins found", func(t *testing.T) {
					ctx := setup(t)
					defer ctx.cleanup()
					err := ctx.execute("plugin", "list", "--lookup-plugins-in-path=true")
					assert.NilError(t, err)
					assert.Equal(t, ctx.output(), "")
				})
			})

			t.Run("plugins installed", func(t *testing.T) {
				t.Run("with valid plugin in $PATH", func(t *testing.T) {

					t.Run("list plugins in $PATH", func(t *testing.T) {
						ctx := setup(t)
						defer ctx.cleanup()

						err := ctx.createTestPlugin(KnTestPluginName, FileModeExecutable, true)
						assert.NilError(t, err)

						err = ctx.execute("plugin", "list", "--lookup-plugins-in-path=true")
						assert.NilError(t, err)
						assert.Assert(t, util.ContainsAll(ctx.output(), KnTestPluginName))
					})
				})

				t.Run("with non-executable plugin", func(t *testing.T) {
					t.Run("warns user plugin invalid", func(t *testing.T) {
						ctx := setup(t)
						defer ctx.cleanup()

						err := ctx.createTestPlugin(KnTestPluginName, FileModeReadable, false)
						assert.NilError(t, err)

						err = ctx.execute("plugin", "list", "--lookup-plugins-in-path=false")
						assert.NilError(t, err)
						assert.Assert(t, util.ContainsAll(ctx.output(), "WARNING", "not executable", "current user"))
					})
				})

				t.Run("with plugins with same name", func(t *testing.T) {

					t.Run("warns user about second (in $PATH) plugin shadowing first", func(t *testing.T) {
						ctx := setup(t)
						defer ctx.cleanup()

						err := ctx.createTestPlugin(KnTestPluginName, FileModeExecutable, true)
						assert.NilError(t, err)

						tmpPathDir2, err := ioutil.TempDir("", "plugins_list")
						assert.NilError(t, err)
						defer os.RemoveAll(tmpPathDir2)

						err = os.Setenv("PATH", ctx.pathDir+string(os.PathListSeparator)+tmpPathDir2)
						assert.NilError(t, err)

						err = ctx.createTestPluginWithPath(KnTestPluginName, FileModeExecutable, tmpPathDir2)
						assert.NilError(t, err)

						err = ctx.execute("plugin", "list", "--lookup-plugins-in-path=true")
						assert.NilError(t, err)
						assert.Assert(t, util.ContainsAll(ctx.output(), "WARNING", "shadowed", "ignored"))
					})
				})

				t.Run("with plugins with name of existing command", func(t *testing.T) {
					t.Run("warns user about overwriting existing command", func(t *testing.T) {
						ctx := setup(t)
						defer ctx.cleanup()

						fakeCmd := &cobra.Command{
							Use: "fake",
						}
						ctx.rootCmd.AddCommand(fakeCmd)
						defer ctx.rootCmd.RemoveCommand(fakeCmd)

						err := ctx.createTestPlugin("kn-fake", FileModeExecutable, true)
						assert.NilError(t, err)

						err = ctx.execute("plugin", "list", "--lookup-plugins-in-path=true")
						assert.ErrorContains(t, err, "overwrite", "built-in")
						assert.Assert(t, util.ContainsAll(ctx.output(), "ERROR", "overwrite", "built-in"))
					})
				})
			})
		})
	})

	t.Run("when pluginsDir has plugins", func(t *testing.T) {
		t.Run("list plugins in --plugins-dir", func(t *testing.T) {
			ctx := setup(t)
			defer ctx.cleanup()

			err := ctx.createTestPlugin(KnTestPluginName, FileModeExecutable, false)

			err = ctx.execute("plugin", "list")
			assert.NilError(t, err)
			assert.Assert(t, util.ContainsAll(ctx.output(), KnTestPluginName))
		})

		t.Run("no plugins installed", func(t *testing.T) {
			ctx := setup(t)
			defer ctx.cleanup()

			err := ctx.execute("plugin", "list")
			assert.NilError(t, err)
			assert.Equal(t, ctx.output(), "")
		})
	})
}
