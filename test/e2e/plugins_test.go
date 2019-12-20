// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

const (
	TestPluginCode string = `#!/bin/bash

echo "Hello Knative, I'm a Kn plugin"
echo "  My plugin file is $0"
echo "  I recieved arguments: $1 $2 $3 $4"`

	KnConfigDefault string = `plugins-dir: %s
lookup-plugins: %s`

	FileModeReadable   = 0644
	FileModeReadWrite  = 0666
	FileModeExecutable = 0777
)

func TestPluginWorkflow(t *testing.T) {
	var (
		knConfigDir, knPluginsDir, knPluginsDir2  string
		knConfigPath, knPluginPath, knPluginPath2 string
		lookupPlugins                             bool
		err                                       error
	)

	t.Parallel()
	test := NewE2eTest(t)
	setup := func(t *testing.T) {
		test.createNamespaceOnSetup = false
		test.Setup(t)
		defer test.Teardown(t)

		knConfigDir, err = ioutil.TempDir("", "kn-config")
		assert.Assert(t, err == nil)

		knPluginsDir = filepath.Join(knConfigDir, "plugins")
		err = os.MkdirAll(knPluginsDir, FileModeExecutable)
		assert.Assert(t, err == nil)

		knPluginsDir2 = filepath.Join(knConfigDir, "plugins2")
		err = os.MkdirAll(knPluginsDir2, FileModeExecutable)
		assert.Assert(t, err == nil)

		knConfigPath, err = test.createConfig(t, "config.yaml", "", knConfigDir)
		assert.Assert(t, err == nil)
		assert.Assert(t, knConfigPath != "")

		knPluginPath, err = test.createPlugin(t, "kn-helloe2e", TestPluginCode, knPluginsDir)
		assert.Assert(t, err == nil)
		assert.Assert(t, knPluginPath != "")

		knPluginPath2, err = test.createPlugin(t, "kn-hello2e2e", TestPluginCode, knPluginsDir2)
		assert.Assert(t, err == nil)
		assert.Assert(t, knPluginPath2 != "")
	}

	teardown := func(t *testing.T) {
		err = os.RemoveAll(knConfigDir)
		assert.Assert(t, err == nil)
	}

	t.Run("when kn config is empty", func(t *testing.T) {
		t.Run("when using --plugin-dir", func(t *testing.T) {
			t.Run("when --plugins-dir has a plugin", func(t *testing.T) {
				t.Run("when --lookup-plugins is false", func(t *testing.T) {
					lookupPlugins = false

					setup(t)
					defer teardown(t)

					knFlags := []string{fmt.Sprintf("--plugins-dir=%s", knPluginsDir), "--lookup-plugins=false"}
					assert.Assert(t, lookupPlugins == false)

					t.Run("list plugin in --plugins-dir", func(t *testing.T) {
						test.listPlugin(t, knFlags, []string{knPluginPath}, []string{})
					})

					t.Run("execute plugin in --plugins-dir", func(t *testing.T) {
						test.runPlugin(t, knFlags, "helloe2e", []string{"e2e", "test"}, []string{"Hello Knative, I'm a Kn plugin", "I recieved arguments: e2e"})
					})

					t.Run("does not list any other plugin in $PATH", func(t *testing.T) {
						test.listPlugin(t, knFlags, []string{knPluginPath}, []string{knPluginPath2})
					})
				})

				t.Run("with --lookup-plugins is true", func(t *testing.T) {
					lookupPlugins = true

					setup(t)
					defer teardown(t)

					knFlags := []string{fmt.Sprintf("--plugins-dir=%s", knPluginsDir), "--lookup-plugins=true"}
					assert.Assert(t, lookupPlugins == true)

					t.Run("list plugin in --plugins-dir", func(t *testing.T) {
						test.listPlugin(t, knFlags, []string{knPluginPath}, []string{knPluginPath2})
					})

					t.Run("execute plugin in --plugins-dir", func(t *testing.T) {
						test.runPlugin(t, knFlags, "helloe2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
					})

					t.Run("when other plugins are in $PATH", func(t *testing.T) {
						var oldPath string

						setupPath := func(t *testing.T) {
							oldPath = os.Getenv("PATH")
							err := os.Setenv("PATH", fmt.Sprintf("%s:%s", oldPath, knPluginsDir2))
							assert.Assert(t, err == nil)
						}

						tearDownPath := func(t *testing.T) {
							err = os.Setenv("PATH", oldPath)
							assert.Assert(t, err == nil)
						}

						t.Run("list plugin in $PATH", func(t *testing.T) {
							setupPath(t)
							defer tearDownPath(t)
							test.listPlugin(t, knFlags, []string{knPluginPath, knPluginPath2}, []string{})
						})

						t.Run("execute plugin in $PATH", func(t *testing.T) {
							setupPath(t)
							defer tearDownPath(t)
							test.runPlugin(t, knFlags, "hello2e2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
						})
					})
				})
			})
		})
	})
}

// Private

func (test *e2eTest) createConfig(t *testing.T, configName, configCode, configPath string) (string, error) {
	configFilename := filepath.Join(configPath, configName)
	err := ioutil.WriteFile(configFilename, []byte(configCode), FileModeReadWrite)
	if err != nil {
		return "", err
	}
	return configFilename, nil
}

func (test *e2eTest) createPlugin(t *testing.T, pluginName, pluginCode, pluginPath string) (string, error) {
	pluginFilename := filepath.Join(pluginPath, pluginName)
	err := ioutil.WriteFile(pluginFilename, []byte(pluginCode), FileModeExecutable)
	if err != nil {
		return "", err
	}
	return pluginFilename, nil
}

func (test *e2eTest) listPlugin(t *testing.T, knFlags []string, expectedPlugins []string, unexpectedPlugins []string) {
	knArgs := addArgs([]string{}, knFlags...)
	knArgs = addArgs(knArgs, []string{"plugin", "list"}...)

	out, err := test.kn.RunWithOpts(knArgs, runOpts{NoNamespace: true})
	assert.Assert(t, err == nil)
	assert.Check(t, util.ContainsAll(out, expectedPlugins...))
	assert.Check(t, util.ContainsNone(out, unexpectedPlugins...))
}

func (test *e2eTest) runPlugin(t *testing.T, knFlags []string, pluginName string, args []string, expectedOutput []string) {
	knArgs := addArgs([]string{}, knFlags...)
	knArgs = append(knArgs, pluginName)
	knArgs = addArgs(knArgs, args...)

	out, err := test.kn.RunWithOpts(knArgs, runOpts{NoNamespace: true})
	assert.Assert(t, err == nil)
	for _, output := range expectedOutput {
		assert.Check(t, util.ContainsAll(out, output))
	}
}

func addArgs(cmdArgs []string, args ...string) []string {
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg)
	}
	return cmdArgs
}
