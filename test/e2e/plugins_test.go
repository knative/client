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
echo "  I received arguments: $1 $2 $3 $4"`

	FileModeReadWrite  = 0666
	FileModeExecutable = 0777
)

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
	err = os.MkdirAll(pc.knPluginsDir, FileModeExecutable)
	if err != nil {
		return err
	}

	pc.knPluginsDir2 = filepath.Join(pc.knConfigDir, "plugins2")
	err = os.MkdirAll(pc.knPluginsDir2, FileModeExecutable)
	if err != nil {
		return err
	}

	pc.knConfigPath, err = createPluginFile("config.yaml", "", pc.knConfigDir, FileModeReadWrite)
	if err != nil {
		return err
	}

	pc.knPluginPath, err = createPluginFile("kn-helloe2e", TestPluginCode, pc.knPluginsDir, FileModeExecutable)
	if err != nil {
		return err
	}
	pc.knPluginPath2, err = createPluginFile("kn-hello2e2e", TestPluginCode, pc.knPluginsDir2, FileModeExecutable)
	if err != nil {
		return err
	}
	return nil
}

func (pc *pluginTestConfig) teardown() {
	os.RemoveAll(pc.knConfigDir)
}

func createPluginFile(fileName, fileContent, filePath string, fileMode os.FileMode) (string, error) {
	file := filepath.Join(filePath, fileName)
	err := ioutil.WriteFile(file, []byte(fileContent), fileMode)
	return file, err
}

func TestPluginWithoutLookup(t *testing.T) {
	t.Parallel()

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir), "--lookup-plugins=false"}

	t.Log("list plugin in --plugins-dir")
	listPlugin(t, r, knFlags, []string{pc.knPluginPath}, []string{})

	t.Log("execute plugin in --plugins-dir")
	runPlugin(t, r, knFlags, "helloe2e", []string{"e2e", "test"}, []string{"Hello Knative, I'm a Kn plugin", "I received arguments: e2e"})

	t.Log("does not list any other plugin in $PATH")
	listPlugin(t, r, knFlags, []string{pc.knPluginPath}, []string{pc.knPluginPath2})

	t.Log("with --lookup-plugins is true")
}

func TestPluginWithLookup(t *testing.T) {

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	defer pc.teardown()

	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir), "--lookup-plugins=true"}

	t.Log("list plugin in --plugins-dir")
	listPlugin(t, r, knFlags, []string{pc.knPluginPath}, []string{pc.knPluginPath2})

	t.Log("execute plugin in --plugins-dir")
	runPlugin(t, r, knFlags, "helloe2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
}

func TestListPluginInPath(t *testing.T) {

	r := NewKnRunResultCollector(t)

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	t.Log("list plugin in $PATH")
	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir), "--lookup-plugins=true"}
	listPlugin(t, r, knFlags, []string{pc.knPluginPath, pc.knPluginPath2}, []string{})

	r.DumpIfFailed()
}

func TestExecutePluginInPath(t *testing.T) {
	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	pc, oldPath := setupPluginTestConfigWithNewPath(t)
	defer tearDownWithPath(pc, oldPath)

	t.Log("execute plugin in $PATH")
	knFlags := []string{fmt.Sprintf("--plugins-dir=%s", pc.knPluginsDir), "--lookup-plugins=true"}
	runPlugin(t, r, knFlags, "hello2e2e", []string{}, []string{"Hello Knative, I'm a Kn plugin"})
}

func setupPluginTestConfigWithNewPath(t *testing.T) (pluginTestConfig, string) {
	pc := pluginTestConfig{}
	assert.NilError(t, pc.setup())
	oldPath := os.Getenv("PATH")
	assert.NilError(t, os.Setenv("PATH", fmt.Sprintf("%s:%s", oldPath, pc.knPluginsDir2)))
	return pc, oldPath
}

func tearDownWithPath(pc pluginTestConfig, oldPath string) {
	os.Setenv("PATH", oldPath)
	pc.teardown()
}

// Private

func listPlugin(t *testing.T, r *KnRunResultCollector, knFlags []string, expectedPlugins []string, unexpectedPlugins []string) {
	knArgs := append(knFlags, "plugin", "list")

	out := kn{}.Run(knArgs...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, expectedPlugins...))
	assert.Check(t, util.ContainsNone(out.Stdout, unexpectedPlugins...))
}

func runPlugin(t *testing.T, r *KnRunResultCollector, knFlags []string, pluginName string, args []string, expectedOutput []string) {
	knArgs := append([]string{}, knFlags...)
	knArgs = append(knArgs, pluginName)
	knArgs = append(knArgs, args...)

	out := kn{}.Run(knArgs...)
	r.AssertNoError(out)
	for _, output := range expectedOutput {
		assert.Check(t, util.ContainsAll(out.Stdout, output))
	}
}
