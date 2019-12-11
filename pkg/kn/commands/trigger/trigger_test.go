// Copyright © 2019 The Knative Authors
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

package trigger

import (
	"bytes"

	"k8s.io/client-go/tools/clientcmd"

	eventc_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	serving_client_v1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

func init() {
	var err error
	blankConfig, err = clientcmd.NewClientConfigFromBytes([]byte(`kind: Config
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
current-context: x
`))
	if err != nil {
		panic(err)
	}
}

func executeTriggerCommand(triggerClient eventc_v1alpha1.KnEventingClient, servingClient serving_client_v1alpha1.KnServingClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewServingClient = func(namespace string) (serving_client_v1alpha1.KnServingClient, error) {
		return servingClient, nil
	}
	knParams.NewEventingClient = func(namespace string) (eventc_v1alpha1.KnEventingClient, error) {
		return triggerClient, nil
	}

	cmd := NewTriggerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	err := cmd.Execute()

	return output.String(), err
}
