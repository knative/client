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
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/commands"
)

func TestContainerCommand(t *testing.T) {
	knParams := &commands.KnParams{}
	containerCmd := NewContainerCommand(knParams)
	assert.Equal(t, containerCmd.Name(), "container")
	assert.Equal(t, containerCmd.Use, "container COMMAND")
	subCommands := make([]string, 0, len(containerCmd.Commands()))
	for _, cmd := range containerCmd.Commands() {
		subCommands = append(subCommands, cmd.Name())
	}
	expectedSubCommands := []string{"add"}
	assert.DeepEqual(t, subCommands, expectedSubCommands)
}

func executeContainerCommand(args ...string) (string, error) {
	knParams := &commands.KnParams{}

	output := new(bytes.Buffer)
	knParams.Output = output

	cmd := NewContainerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOut(output)
	err := cmd.Execute()
	return output.String(), err
}
