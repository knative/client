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

package commands

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"

	"knative.dev/client-pkg/pkg/util/test"
)

func TestCreateTestKnCommand(t *testing.T) {
	knParams := &KnParams{}
	knCmd, serving, buffer := CreateTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
	assert.Assert(t, knCmd != nil)
	assert.Assert(t, serving != nil)
	assert.Assert(t, buffer != nil)
	assert.Assert(t, len(knCmd.Commands()) == 1)
	assert.Assert(t, knCmd.Commands()[0].Use == "fake")
}

func TestCreateSourcesTestKnCommand(t *testing.T) {
	knParams := &KnParams{}
	knCmd, sources, buffer := CreateSourcesTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
	assert.Assert(t, knCmd != nil)
	assert.Assert(t, sources != nil)
	assert.Assert(t, buffer != nil)
	assert.Assert(t, len(knCmd.Commands()) == 1)
	assert.Assert(t, knCmd.Commands()[0].Use == "fake")
}

func TestCreateDynamicTestKnCommand(t *testing.T) {
	knParams := &KnParams{}
	knCmd, dynamic, buffer := CreateDynamicTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
	assert.Assert(t, knCmd != nil)
	assert.Assert(t, dynamic != nil)
	assert.Assert(t, buffer != nil)
	assert.Assert(t, len(knCmd.Commands()) == 1)
	assert.Assert(t, knCmd.Commands()[0].Use == "fake")
	client, err := knParams.NewDynamicClient("foo")
	assert.NilError(t, err)
	assert.Assert(t, client != nil)
}

func TestCaptureStdout(t *testing.T) {
	c := test.CaptureOutput(t)
	fmt.Print("Hello World !")
	stdOut, stdErr := c.Close()
	assert.Equal(t, stdErr, "")
	assert.Equal(t, stdOut, "Hello World !")
}
