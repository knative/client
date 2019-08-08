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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

const (
	KnTestPluginName   = "kn-test"
	KnTestPluginScript = `#!/bin/bash

echo "I am a test Kn plugin"
exit 0
`
	FileModeReadable   = 0644
	FileModeExecutable = 0777
)

// FindSubCommand return the sub-command by name
func FindSubCommand(t *testing.T, rootCmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Name() == name {
			return subCmd
		}
	}

	return nil
}

// CreateTestPlugin with name, script, and fileMode and return the tmp random path
func CreateTestPlugin(t *testing.T, name, script string, fileMode os.FileMode) string {
	path, err := ioutil.TempDir("", "plugin")
	assert.Assert(t, err == nil)

	return CreateTestPluginInPath(t, name, script, fileMode, path)
}

// CreateTestPluginInPath with name, path, script, and fileMode and return the tmp random path
func CreateTestPluginInPath(t *testing.T, name, script string, fileMode os.FileMode, path string) string {
	fullPath := filepath.Join(path, name)
	if runtime.GOOS == "windows" {
		fullPath += ".bat"
	}
	err := ioutil.WriteFile(fullPath, []byte(script), fileMode)
	assert.NilError(t, err)

	return filepath.Join(path, name)
}

// DeleteTestPlugin with path
func DeleteTestPlugin(t *testing.T, path string) {
	err := os.RemoveAll(filepath.Dir(path))
	assert.Assert(t, err == nil)
}
