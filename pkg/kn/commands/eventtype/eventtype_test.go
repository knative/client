/*
Copyright 2022 The Knative Authors

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

package eventtype

import (
	"bytes"

	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/kn/commands"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/pkg/apis"
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

func createEventtype(eventtypeName, ceType, namespace string) *eventingv1beta1.EventType {
	return v1beta1.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(ceType).Build()
}

func createEventtypeWithSource(eventtypeName, ceType, namespace string, source *apis.URL) *eventingv1beta1.EventType {
	return v1beta1.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(ceType).Source(source).Build()
}

func createEventtypeWithBroker(name, cetype, broker, namespace string) *eventingv1beta1.EventType {
	return v1beta1.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(cetype).Broker(broker).Build()
}

func executeEventtypeCommand(client *v1beta1.MockKnEventingV1beta1Client, args ...string) (string, error) {

	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	knParams.NewEventingV1beta1Client = func(namespace string) (v1beta1.KnEventingV1Beta1Client, error) {
		return client, nil
	}

	cmd := NewEventTypeCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOut(output)

	err := cmd.Execute()

	return output.String(), err
}
