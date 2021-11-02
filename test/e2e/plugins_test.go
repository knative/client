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

//go:build e2e && !eventing && !serving
// +build e2e,!eventing,!serving

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

const (
	TestPluginCodeBat string = `echo "Hello Knative, I'm a Kn plugin"
echo "  My plugin file is %0"
echo "  I received arguments: %1 %2 %3 %4"
`
	TestPluginCodeBash string = `#!/bin/bash
echo "Hello Knative, I'm a Kn plugin"
echo "  My plugin file is $0"
echo "  I received arguments: $1 $2 $3 $4"
`

	TestPluginCodeErrBash string = `#!/bin/bash
exit 1`

	TestPluginCodeErrBat string = `exit 1`

	delim = string(os.PathListSeparator)
)

var pluginBin, pluginBin2, pluginBin3, pluginBinErr, pluginCode string

type pluginTestConfig struct {
	knConfigDir, knPluginsDir, knPluginsDir2  string
	knConfigPath, knPluginPath, knPluginPath2 string
}

func (pc *pluginTestConfig) setup() error {
	var err error
	pc.knConfigDir, err = ioutil.TempDir("", "kn-config")
	if err != nil {
		return err
	}

	pc.knPluginsDir = filepath.Join(pc.knConfigDir, "plugins")
	err = os.MkdirAll(pc.knPluginsDir, test.FileModeExecutable)
	if err != nil {
		return err
	}

	pc.knPluginsDir2 = filepath.Join(pc.knConfigDir, "plugins2")
	err = os.MkdirAll(pc.knPluginsDir2, test.FileModeExecutable)
	if err != nil {
		return err
	}

	pc.knConfigPath, err = test.CreateFile("config.yaml", "", pc.knConfigDir, test.FileModeReadWrite)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		pluginBin = "kn-helloe2e.bat"
		pluginBin2 = "kn-hello2e2e.bat"
		pluginBin3 = "kn-hello3e2e.bat"
		pluginCode = TestPluginCodeBat
		pluginBinErr = TestPluginCodeErrBat
	default:
		pluginBin = "kn-helloe2e"
		pluginBin2 = "kn-hello2e2e"
		pluginBin3 = "kn-hello3e2e"
		pluginBinErr = TestPluginCodeErrBash
		pluginCode = TestPluginCodeBash
	}
	pc.knPluginPath, err = test.CreateFile(pluginBin, pluginCode, pc.knPluginsDir, test.FileModeExecutable)
	if err != nil {
		return err
	}
	pc.knPluginPath2, err = test.CreateFile(pluginBin2, pluginCode, pc.knPluginsDir2, test.FileModeExecutable)
	if err != nil {
		return err
	}
	return nil
}

func (pc *pluginTestConfig) teardown() {
	os.RemoveAll(pc.knConfigDir)
}

func TestPluginWithoutLookup(t *testing.T) {
	t.Parallel()

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	it, err := test.NewKnTest()
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir)}

	t.Log("list plugin in --plugins-dir")
	listPlugin(r, knFlags, []string{pc.knPluginPath}, []string{})

	t.Log("execute plugin in --plugins-dir")
	runPlugin(r, knFlags, "helloe2e", []string{"e2e", "test"}, []string{"Hello Knative, I'm a Kn plugin", "I received arguments", "e2e"})
}

func TestPluginInHelpMessage(t *testing.T) {
	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	defer pc.teardown()

	result := test.Kn{}.Run("--plugins-dir", pc.knPluginsDir, "--help")
	assert.NilError(t, result.Error)
	assert.Assert(t, util.ContainsAll(result.Stdout, "Plugins:", "helloe2e", "kn-helloe2e"))

	result = test.Kn{}.Run("--plugins-dir", pc.knPluginsDir, "service", "--help")
	assert.NilError(t, result.Error)
	assert.Assert(t, util.ContainsNone(result.Stdout, "Plugins:", "helloe2e", "kn-helloe2e"))
}

func TestPluginWithLookup(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	defer pc.teardown()

	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir)}

	t.Log("list plugin in --plugins-dir")
	listPlugin(r, knFlags, []string{pc.knPluginPath}, []string{pc.knPluginPath2})

	t.Log("execute plugin in --plugins-dir")
	runPlugin(r, knFlags, "helloe2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
}

func TestListPluginInPath(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	t.Log("list plugin in $PATH")
	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir)}
	listPlugin(r, knFlags, []string{pc.knPluginPath, pc.knPluginPath2}, []string{})

	r.DumpIfFailed()
}

func TestExecutePluginInPath(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	t.Log("execute plugin in $PATH")
	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir)}
	runPlugin(r, knFlags, "hello2e2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
}

func TestExecutePluginInPathWithError(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	oldPath := os.Getenv("PATH")

	t.Log("execute plugin in $PATH that returns error")
	pluginsDir := filepath.Join(pc.knConfigDir, "plugins3")
	err = os.MkdirAll(pluginsDir, test.FileModeExecutable)
	assert.NilError(t, err)
	_, err = test.CreateFile(pluginBin3, pluginBinErr, pluginsDir, test.FileModeExecutable)
	assert.NilError(t, err)
	assert.NilError(t, os.Setenv("PATH", fmt.Sprintf("%s%s%s", oldPath, delim, pluginsDir)))
	defer tearDownWithPath(pc, oldPath)
}

// Private
func setupPluginTestConfigWithNewPath(t *testing.T) (pluginTestConfig, string) {
	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	oldPath := os.Getenv("PATH")
	assert.NilError(t, os.Setenv("PATH", fmt.Sprintf("%s%s%s", oldPath, delim, pc.knPluginsDir2)))
	return pc, oldPath
}

func tearDownWithPath(pc pluginTestConfig, oldPath string) {
	os.Setenv("PATH", oldPath)
	pc.teardown()
}

func listPlugin(r *test.KnRunResultCollector, knFlags []string, expectedPlugins []string, unexpectedPlugins []string) {
	knArgs := append(knFlags, "plugin", "list")

	out := test.Kn{}.Run(knArgs...)
	r.AssertNoError(out)
	for _, p := range expectedPlugins {
		assert.Check(r.T(), util.ContainsAll(out.Stdout, filepath.Base(p)))
	}
	for _, p := range unexpectedPlugins {
		assert.Check(r.T(), util.ContainsNone(out.Stdout, filepath.Base(p)))
	}
}

func runPlugin(r *test.KnRunResultCollector, knFlags []string, pluginName string, args []string, expectedOutput []string) {
	knArgs := append([]string{}, knFlags...)
	knArgs = append(knArgs, pluginName)
	knArgs = append(knArgs, args...)

	out := test.Kn{}.Run(knArgs...)
	r.AssertNoError(out)
	for _, output := range expectedOutput {
		assert.Check(r.T(), util.ContainsAll(out.Stdout, output))
	}
}
