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
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/config"
	"knative.dev/client/pkg/util"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestPluginListBasic(t *testing.T) {

	pluginListCmd := NewPluginListCommand(&commands.KnParams{})
	assert.Assert(t, pluginListCmd != nil)

	assert.Assert(t, pluginListCmd != nil)
	assert.Assert(t, pluginListCmd.Use == "list")
	assert.Assert(t, pluginListCmd.Short == "List plugins")
	assert.Assert(t, strings.Contains(pluginListCmd.Long, "List all installed plugins"))
	assert.Assert(t, pluginListCmd.RunE != nil)
}

func TestPluginListOutput(t *testing.T) {
	pluginDir, cleanupFunc := prepareTestSetup(t, "kn-test1", 0777, "kn-test2", 0644)
	defer cleanupFunc()

	for _, verbose := range []bool{false, true} {
		outBuf := bytes.Buffer{}
		testCmd := cobra.Command{
			Use: "kn",
		}
		testCmd.SetOut(&outBuf)
		testCmd.AddCommand(&cobra.Command{Use: "children"})

		err := listPlugins(&testCmd, pluginListFlags{verbose: verbose})
		assert.NilError(t, err)

		out := outBuf.String()

		assert.Assert(t, util.ContainsAll(out, "kn-test1", "kn-test2"))

		if verbose {
			assert.Assert(t, util.ContainsAll(out, pluginDir))
		}

		if runtime.GOOS != "windows" {
			assert.Assert(t, util.ContainsAll(out, "WARNING", "not executable"))
		}
	}
}

func TestPluginListNoPlugins(t *testing.T) {
	pluginDir, cleanupFunc := prepareTestSetup(t)
	defer cleanupFunc()

	for _, verbose := range []bool{false, true} {
		outBuf := bytes.Buffer{}
		testCmd := cobra.Command{}
		testCmd.SetOut(&outBuf)

		err := listPlugins(&testCmd, pluginListFlags{verbose: verbose})
		assert.NilError(t, err)

		out := outBuf.String()
		assert.Assert(t, util.ContainsAll(out, "No", "found"))
		if verbose {
			assert.Assert(t, util.ContainsAll(out, pluginDir))
		}
	}
}

func TestPluginListOverridingBuiltinCommand(t *testing.T) {
	pluginDir, cleanupFunc := prepareTestSetup(t, "kn-existing", 0777)
	defer cleanupFunc()

	outBuf := bytes.Buffer{}
	testCmd := cobra.Command{
		Use: "kn",
	}
	testCmd.AddCommand(&cobra.Command{Use: "existing"})
	testCmd.SetOut(&outBuf)
	err := listPlugins(&testCmd, pluginListFlags{verbose: false})
	assert.Assert(t, err != nil)

	out := outBuf.String()
	assert.Assert(t, util.ContainsAll(out, "ERROR", "'existing'", "overwrites", pluginDir, "kn-existing"))
}

func TestPluginListExtendingBuiltinCommandGroup(t *testing.T) {
	_, cleanupFunc := prepareTestSetup(t, "kn-existing-addon", 0777)
	defer cleanupFunc()

	outBuf := bytes.Buffer{}
	testCmd := cobra.Command{
		Use: "kn",
	}
	testGroup := &cobra.Command{Use: "existing"}
	testGroup.AddCommand(&cobra.Command{Use: "builtin"})
	testCmd.AddCommand(testGroup)
	testCmd.SetOut(&outBuf)
	err := listPlugins(&testCmd, pluginListFlags{verbose: false})
	assert.NilError(t, err)

	out := outBuf.String()
	assert.Assert(t, util.ContainsAll(out, "kn-existing-addon"))
	assert.Assert(t, !strings.Contains(out, "ERROR"))
}

// Private

func prepareTestSetup(t *testing.T, args ...interface{}) (string, func()) {
	tmpPathDir, err := ioutil.TempDir("", "plugin_list")
	assert.NilError(t, err)

	// Prepare configuration to for our test
	oldConfig := config.GlobalConfig
	config.GlobalConfig = &config.TestConfig{
		TestPluginsDir:          tmpPathDir,
		TestLookupPluginsInPath: false,
	}

	for i := 0; i < len(args); i += 2 {
		name := args[i].(string)
		perm := args[i+1].(int)
		createTestPlugin(t, name, tmpPathDir, os.FileMode(perm))
	}

	return tmpPathDir, func() {
		config.GlobalConfig = oldConfig
		os.RemoveAll(tmpPathDir)
	}
}

// CreateTestPluginInPath with name, path, script, and fileMode and return the tmp random path
func createTestPlugin(t *testing.T, name string, dir string, perm os.FileMode) string {
	var nameExt string
	if runtime.GOOS == "windows" {
		nameExt = name + ".bat"
	} else {
		nameExt = name
	}
	fullPath := filepath.Join(dir, nameExt)
	err := ioutil.WriteFile(fullPath, []byte{}, perm)
	assert.NilError(t, err)
	return fullPath
}
