// Copyright © 2021 The Knative Authors
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

package container

import (
	"os"
	"testing"

	"knative.dev/client/pkg/util"

	"gotest.tools/v3/assert"
)

func TestContainerAdd(t *testing.T) {
	output, err := executeContainerCommand("add", "foo", "--image", "registry.foo:bar")
	assert.NilError(t, err)
	assert.Assert(t, len(output) > 0)
	assert.Assert(t, util.ContainsAllIgnoreCase(output, "containers", "image", "foo", "registry.foo:bar"))
}

func TestContainerAddWithContainers(t *testing.T) {
	rawInput := `
containers:
- image: bar:bar
  name: bar
  resources: {}`

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	for _, command := range []string{"containers", "extra-containers"} {
		stdinReader, stdinWriter, err := os.Pipe()
		assert.NilError(t, err)
		stdinWriter.Chmod(os.ModeCharDevice)
		_, err = stdinWriter.Write([]byte(rawInput))
		assert.NilError(t, err)
		stdinWriter.Close()

		os.Stdin = stdinReader

		output, err := executeContainerCommand("add", "foo", "--image", "registry.foo:bar", "--"+command, "-")
		assert.NilError(t, err)
		assert.Assert(t, len(output) > 0)
		assert.Assert(t, util.ContainsAllIgnoreCase(output, "containers", "image", "foo", "registry.foo:bar", "bar", "bar:bar"))
	}
}

func TestContainerAddError(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedErrors []string
	}{
		{
			"MissingName",
			[]string{"add", "--image", "registry.foo:bar"},
			[]string{"container", "name", "single", "argument"},
		},
		{
			"MissingImage",
			[]string{"add", "foo"},
			[]string{"'container add'", "requires", "--image"},
		},
		{
			"WrongEnvFormat",
			[]string{"add", "foo", "--image", "registry.foo:bar", "--env", "a b"},
			[]string{"--env", "argument", "requires", "\"=\""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := executeContainerCommand(tc.args...)
			assert.Assert(t, err != nil)
			assert.Assert(t, util.ContainsAll(err.Error(), tc.expectedErrors...))
		})
	}
}

func TestDetectPipeInput(t *testing.T) {
	stdinReader, stdinWriter, err := os.Pipe()
	assert.NilError(t, err)
	stdinWriter.Chmod(os.ModeCharDevice)
	_, err = stdinWriter.Write([]byte("test"))
	assert.NilError(t, err)
	stdinWriter.Close()

	// Mock piped input
	piped := IsPipeInput(stdinReader)
	assert.Assert(t, piped == true)

	// os.Stdin is not piped
	piped = IsPipeInput(os.Stdin)
	assert.Assert(t, piped == false)

	piped = IsPipeInput(nil)
	assert.Assert(t, piped == false)
}
