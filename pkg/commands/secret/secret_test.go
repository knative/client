// Copyright © 2023 The Knative Authors
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

package secret

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/commands"
	knflags "knative.dev/client/pkg/flags"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

const kubeConfig = `kind: Config
version: v1
users:
- name: u
clusters:
- name: c
  cluster:
    server: example.com
contexts:
- name: x
  context:
    user: u
    cluster: c
current-context: x`

func init() {
	var err error
	blankConfig, err = clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
	if err != nil {
		panic(err)
	}
}

func TestSecretCommand(t *testing.T) {
	knParams := &commands.KnParams{}
	secretCommand := NewSecretCommand(knParams)
	assert.Equal(t, secretCommand.Name(), "secret")
	assert.Equal(t, secretCommand.Use, "secret COMMAND")
	subCommands := make([]string, 0, len(secretCommand.Commands()))
	for _, cmd := range secretCommand.Commands() {
		subCommands = append(subCommands, cmd.Name())
	}
	expectedSubCommands := []string{"create", "delete", "list"}
	assert.DeepEqual(t, subCommands, expectedSubCommands)
}

func executeSecretCommand(client kubernetes.Interface, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	knParams.NewKubeClient = func() (kubernetes.Interface, error) {
		return client, nil
	}

	cmd := NewSecretCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOut(output)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return knflags.ReconcileBoolFlags(cmd.Flags())
	}
	err := cmd.Execute()
	return output.String(), err
}
