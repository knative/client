/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package broker

import (
	"bytes"

	"k8s.io/client-go/tools/clientcmd"

	clientv1beta1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/kn/commands"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1"
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

func executeBrokerCommand(brokerClient clientv1beta1.KnEventingClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	knParams.NewEventingClient = func(namespace string) (clientv1beta1.KnEventingClient, error) {
		return brokerClient, nil
	}

	cmd := NewBrokerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	err := cmd.Execute()

	return output.String(), err
}

func createBroker(brokerName string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Build()
}

func createBrokerWithNamespace(brokerName, namespace string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace(namespace).Build()
}
