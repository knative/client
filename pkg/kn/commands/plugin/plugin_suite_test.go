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

package plugin_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

const (
	KnTestPluginName   = "kn-test"
	KnTestPluginScript = `#!/bin/bash

echo "I am a test Kn plugin"
`
	FileModeReadable   = 0644
	FileModeExecutable = 777
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Suite")
}

var _ = BeforeSuite(func() {
	removeGinkgoFlags()
	captureStdout()
})

var _ = AfterSuite(func() {
	restoreGinkgoFlags()
	releaseStdout()
})

var (
	oldStdout *os.File
	stdout    *os.File
	output    string

	readFile, writeFile *os.File

	origArgs []string
)

// ReadStdout collects the current content of os.Stdout
// into Output global
func ReadStdout() string {
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, readFile)
		outC <- buf.String()
	}()
	writeFile.Close()
	output = <-outC

	captureStdout()

	return output
}

// FindSubCommand return the sub-command by name
func FindSubCommand(rootCmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Name() == name {
			return subCmd
		}
	}

	return nil
}

// CreateTestPlugin with name, script, and fileMode and return the tmp random path
func CreateTestPlugin(name, script string, fileMode os.FileMode) string {
	path, err := ioutil.TempDir("", "plugin")
	Expect(err).NotTo(HaveOccurred())

	return CreateTestPluginInPath(name, script, fileMode, path)
}

// CreateTestPluginInPath with name, path, script, and fileMode and return the tmp random path
func CreateTestPluginInPath(name, script string, fileMode os.FileMode, path string) string {
	err := ioutil.WriteFile(filepath.Join(path, name), []byte(script), fileMode)
	Expect(err).NotTo(HaveOccurred())

	return filepath.Join(path, name)
}

// DeleteTestPlugin with path
func DeleteTestPlugin(path string) {
	err := os.RemoveAll(filepath.Dir(path))
	Expect(err).NotTo(HaveOccurred())
}

// Private

func captureStdout() {
	oldStdout = os.Stdout
	var err error
	readFile, writeFile, err = os.Pipe()
	Expect(err).NotTo(HaveOccurred())
	stdout = writeFile
	os.Stdout = writeFile
}

func releaseStdout() {
	output = ReadStdout()
	os.Stdout = oldStdout
}

// removeGinkgoFlags is needed since if we execute with ginkgo it's
// default flags are added to os.Args and they will be parsed when we
// execute kn commands and they will generate errors
func removeGinkgoFlags() {
	origArgs = os.Args[:]
	os.Args = os.Args[:1]
}

// restoreGinkgoFlags restores the os.Args flags removed in removeGinkgoFlags
func restoreGinkgoFlags() {
	os.Args = origArgs[:]
}
