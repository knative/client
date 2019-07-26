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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
)

const (
	TestPluginCode string = `#!/bin/bash

echo "Hello Knative, I'm a Kn plugin"
echo "  My plugin file is $0"
echo "  I recieved arguments: $1 $2 $3 $4"`

	KnConfigDefault string = `pluginDir=~/.kn/plugins
lookuppluginsinpath=true`
)

func TestPluginWorkflow(t *testing.T) {
	var (
		knConfigDir, knPluginsDir  string
		knConfigPath, knPluginPath string
		lookupPluginsInPath        bool
		err                        error
	)

	test := NewE2eTest(t)
	setup := func(t *testing.T) {
		test.Setup(t)
		defer test.Teardown(t)

		knConfigDir, err := ioutil.TempDir("", "kn-config")
		assert.Assert(t, err == nil)

		knPluginsDir = filepath.Join(knConfigDir, "plugins")
		err = os.MkdirAll(knPluginsDir, 0666)
		assert.Assert(t, err == nil)

		knConfigPath, err = test.createConfig(t, "config.yaml", "", knConfigDir)
		assert.Assert(t, err == nil)
		assert.Assert(t, knConfigPath != "")
	}

	teardown := func(t *testing.T) {
		err = os.RemoveAll(knConfigDir)
		assert.Assert(t, err == nil)
	}

	t.Run("when kn config is empty", func(t *testing.T) {
		t.Run("when --plugin-dir", func(t *testing.T) {
			beforeEach := func(t *testing.T) {
				knPluginPath, err = test.createPlugin(t, "kn-hello", TestPluginCode, knPluginsDir)
				assert.Assert(t, err == nil)
				assert.Assert(t, knPluginPath != "")
			}

			t.Run("when --plugins-dir has a plugin", func(t *testing.T) {
				t.Run("when --lookup-plugins-in-path is false", func(t *testing.T) {
					lookupPluginsInPath = false

					setup(t)
					defer teardown(t)
					beforeEach(t)

					assert.Assert(t, lookupPluginsInPath == false)

					t.Run("list plugin in --plugins-dir", func(t *testing.T) {})
					t.Run("execute plugin in --plugins-dir", func(t *testing.T) {})
					t.Run("not list plugin in $PATH", func(t *testing.T) {})
				})

				t.Run("with --lookup-plugins-in-path is true", func(t *testing.T) {
					lookupPluginsInPath = true

					setup(t)
					defer teardown(t)
					beforeEach(t)

					assert.Assert(t, lookupPluginsInPath == true)

					t.Run("list plugin in --plugins-dir", func(t *testing.T) {})
					t.Run("list plugin in $PATH", func(t *testing.T) {})
					t.Run("execute plugin in --plugins-dir", func(t *testing.T) {})
					t.Run("execute plugin in $PATH", func(t *testing.T) {})
				})
			})
		})
	})
}

// Private

func (test *e2eTest) createConfig(t *testing.T, configName, configCode, configPath string) (string, error) {
	//TODO
	return "", nil
}

func (test *e2eTest) createPlugin(t *testing.T, pluginName, pluginCode, pluginPath string) (string, error) {
	//TODO
	return "", nil
}

func (test *e2eTest) listPlugin(t *testing.T, expectedPlugins []string) {
	//TODO
}

func (test *e2eTest) runPlugin(t *testing.T, pluginName string, args []string, expectedOutput string) {
	//TODO
}
