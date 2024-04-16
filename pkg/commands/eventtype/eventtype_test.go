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

	kndynamic "knative.dev/client-pkg/pkg/dynamic"

	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client-pkg/pkg/eventing/v1beta2"
	"knative.dev/client/pkg/commands"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	"knative.dev/pkg/apis"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

const (
	eventtypeName   = "foo"
	testNs          = "test-ns"
	cetype          = "foo.type"
	testSource      = "https://test-source.com"
	testSourceError = "bad-source\b"
	testBroker      = "test-broker"
)

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

func createEventtype(eventtypeName, ceType, namespace string) *eventingv1beta2.EventType {
	return v1beta2.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(ceType).Build()
}

func createEventtypeWithSource(eventtypeName, ceType, namespace string, source *apis.URL) *eventingv1beta2.EventType {
	return v1beta2.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(ceType).Source(source).Build()
}

func createEventtypeWithBroker(name, cetype, broker, namespace string) *eventingv1beta2.EventType {
	return v1beta2.NewEventtypeBuilder(eventtypeName).Namespace(namespace).Type(cetype).Broker(broker).Build()
}

func executeEventtypeCommand(client *v1beta2.MockKnEventingV1beta2Client, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {

	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	knParams.NewEventingV1beta2Client = func(namespace string) (v1beta2.KnEventingV1Beta2Client, error) {
		return client, nil
	}

	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewEventTypeCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOut(output)

	err := cmd.Execute()

	return output.String(), err
}
