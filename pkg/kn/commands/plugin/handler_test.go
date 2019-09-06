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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

func TestPluginHandler(t *testing.T) {
	var (
		pluginHandler, tPluginHandler                  PluginHandler
		pluginPath, pluginName, tmpPathDir, pluginsDir string
		lookupPluginsInPath                            bool
		err                                            error
	)

	setup := func(t *testing.T) {
		tmpPathDir, err = ioutil.TempDir("", "plugin_list")
		assert.Assert(t, err == nil)
		pluginsDir = tmpPathDir
	}

	cleanup := func(t *testing.T) {
		err = os.RemoveAll(tmpPathDir)
		assert.Assert(t, err == nil)
	}

	beforeEach := func(t *testing.T) {
		pluginName = "fake"
		pluginPath = CreateTestPluginInPath(t, "kn-"+pluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
		assert.Assert(t, pluginPath != "")

		pluginHandler = &DefaultPluginHandler{
			ValidPrefixes:       []string{"kn"},
			PluginsDir:          pluginsDir,
			LookupPluginsInPath: lookupPluginsInPath,
		}
		assert.Assert(t, pluginHandler != nil)

		tPluginHandler = NewTestPluginHandler(pluginHandler)
		assert.Assert(t, tPluginHandler != nil)
	}

	t.Run("#NewDefaultPluginHandler", func(t *testing.T) {
		setup(t)
		defer cleanup(t)

		pHandler := NewDefaultPluginHandler([]string{"kn"}, pluginPath, false)
		assert.Assert(t, pHandler != nil)
	})

	t.Run("#Lookup", func(t *testing.T) {
		t.Run("when plugin in pluginsDir", func(t *testing.T) {
			t.Run("returns the first filepath matching prefix", func(t *testing.T) {
				setup(t)
				defer cleanup(t)
				beforeEach(t)

				path, exists := pluginHandler.Lookup(pluginName)
				assert.Assert(t, path != "", fmt.Sprintf("no path when Lookup(%s)", pluginName))
				assert.Assert(t, exists == true, fmt.Sprintf("could not Lookup(%s)", pluginName))
			})

			t.Run("returns empty filepath when no matching prefix found", func(t *testing.T) {
				setup(t)
				defer cleanup(t)

				path, exists := pluginHandler.Lookup("bogus-plugin-name")
				assert.Assert(t, path == "", fmt.Sprintf("unexpected plugin: kn-bogus-plugin-name"))
				assert.Assert(t, exists == false, fmt.Sprintf("unexpected plugin: kn-bogus-plugin-name"))
			})
		})

		t.Run("when plugin is in $PATH", func(t *testing.T) {
			t.Run("--lookup-plugins-in-path=true", func(t *testing.T) {
				setup(t)
				defer cleanup(t)

				pluginsDir = filepath.Join(tmpPathDir, "bogus")
				err = os.Setenv("PATH", tmpPathDir)
				assert.Assert(t, err == nil)
				lookupPluginsInPath = true

				beforeEach(t)

				path, exists := pluginHandler.Lookup(pluginName)
				assert.Assert(t, path != "", fmt.Sprintf("no path when Lookup(%s)", pluginName))
				assert.Assert(t, exists == true, fmt.Sprintf("could not Lookup(%s)", pluginName))
			})

			t.Run("--lookup-plugins-in-path=false", func(t *testing.T) {
				setup(t)
				defer cleanup(t)

				pluginsDir = filepath.Join(tmpPathDir, "bogus")
				err = os.Setenv("PATH", tmpPathDir)
				assert.Assert(t, err == nil)
				lookupPluginsInPath = false

				beforeEach(t)

				path, exists := pluginHandler.Lookup(pluginName)
				assert.Assert(t, path == "")
				assert.Assert(t, exists == false)
			})
		})
	})

	t.Run("#Execute", func(t *testing.T) {
		t.Run("fails executing bogus plugin name", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			beforeEach(t)

			bogusPath := filepath.Join(filepath.Dir(pluginPath), "kn-bogus-plugin-name")
			err = pluginHandler.Execute(bogusPath, []string{bogusPath}, os.Environ())
			assert.Assert(t, err != nil, fmt.Sprintf("bogus plugin in path %s unexpectedly executed OK", bogusPath))
		})
	})

	t.Run("HandlePluginCommand", func(t *testing.T) {
		t.Run("success handling", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			beforeEach(t)

			err = HandlePluginCommand(tPluginHandler, []string{pluginName})
			assert.Assert(t, err == nil, fmt.Sprintf("test plugin %s failed executing", fmt.Sprintf("kn-%s", pluginName)))
		})

		t.Run("fails handling", func(t *testing.T) {
			setup(t)
			defer cleanup(t)

			err = HandlePluginCommand(tPluginHandler, []string{"bogus"})
			assert.Assert(t, err != nil, fmt.Sprintf("test plugin %s expected to fail executing", "bogus"))
		})
	})
}

// TestPluginHandler - needed to mock Execute() call

type testPluginHandler struct {
	pluginHandler PluginHandler
}

func NewTestPluginHandler(pluginHandler PluginHandler) PluginHandler {
	return &testPluginHandler{
		pluginHandler: pluginHandler,
	}
}

func (tHandler *testPluginHandler) Lookup(name string) (string, bool) {
	return tHandler.pluginHandler.Lookup(name)
}

func (tHandler *testPluginHandler) Execute(executablePath string, cmdArgs, environment []string) error {
	// Always success (avoids doing syscall.Exec which exits tests framework)
	return nil
}
