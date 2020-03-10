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
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kn_dynamic "knative.dev/client/pkg/dynamic"
	eventc_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
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

func executeTriggerCommand(triggerClient eventc_v1alpha1.KnEventingClient, dynamicClient kn_dynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kn_dynamic.KnDynamicClient, error) {
		return dynamicClient, nil
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

func createTrigger(namespace string, name string, filters map[string]string, broker string, svcname string) *v1alpha1.Trigger {
	return eventc_v1alpha1.NewTriggerBuilder(name).
		Namespace(namespace).
		Broker(broker).
		Filters(filters).
		Subscriber(createServiceSink(svcname)).
		Build()
}

func createTriggerWithInject(namespace string, name string, filters map[string]string, broker string, svcname string) *v1alpha1.Trigger {
	t := createTrigger(namespace, name, filters, broker, svcname)
	return eventc_v1alpha1.NewTriggerBuilderFromExisting(t).InjectBroker(true).Build()
}

func createTriggerWithStatus(namespace string, name string, filters map[string]string, broker string, svcname string) *v1alpha1.Trigger {
	wanted := createTrigger(namespace, name, filters, broker, svcname)
	wanted.Status = v1alpha1.TriggerStatus{
		Status: duckv1.Status{
			Conditions: []apis.Condition{{
				Type:   "Ready",
				Status: "True",
			}},
		},
		SubscriberURI: apis.HTTP(svcname),
	}
	return wanted
}

func createServiceSink(service string) *duckv1.Destination {
	return &duckv1.Destination{
		Ref: &duckv1.KReference{Name: service, Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
	}
}
