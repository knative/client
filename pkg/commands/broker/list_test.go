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
	"encoding/json"
	"strings"
	"testing"

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"

	clientv1 "knative.dev/client-pkg/pkg/eventing/v1"
	"knative.dev/client-pkg/pkg/util"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

func TestBrokerList(t *testing.T) {
	eventingClient := clientv1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	broker1 := createBrokerWithGvk("foo1")
	broker2 := createBrokerWithGvk("foo2")
	broker3 := createBrokerWithGvk("foo3")
	brokerList := &eventingv1.BrokerList{Items: []eventingv1.Broker{*broker1, *broker2, *broker3}}
	_ = util.UpdateGroupVersionKindWithScheme(brokerList, eventingv1.SchemeGroupVersion, scheme.Scheme)

	t.Run("default output", func(t *testing.T) {
		eventingRecorder.ListBrokers(brokerList, nil)

		output, err := executeBrokerCommand(eventingClient, "list")
		assert.NilError(t, err)

		outputLines := strings.Split(output, "\n")
		assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "AGE", "CONDITIONS", "READY", "REASON"))
		assert.Check(t, util.ContainsAll(outputLines[1], "foo1"))
		assert.Check(t, util.ContainsAll(outputLines[2], "foo2"))
		assert.Check(t, util.ContainsAll(outputLines[3], "foo3"))
	})

	t.Run("json format output", func(t *testing.T) {
		eventingRecorder.ListBrokers(brokerList, nil)

		output, err := executeBrokerCommand(eventingClient, "list", "-o", "json")
		assert.NilError(t, err)

		result := eventingv1.BrokerList{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NilError(t, err)
		assert.DeepEqual(t, brokerList.Items, result.Items)
	})

	eventingRecorder.Validate()
}

func TestBrokerListEmpty(t *testing.T) {
	eventingClient := clientv1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	eventingRecorder.ListBrokers(&eventingv1.BrokerList{}, nil)
	output, err := executeBrokerCommand(eventingClient, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "brokers", "found"))

	eventingRecorder.Validate()
}

func TestBrokerListEmptyWithJSON(t *testing.T) {
	eventingClient := clientv1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()
	brokerList := &eventingv1.BrokerList{}
	brokerList.APIVersion = "eventing.knative.dev/v1"
	brokerList.Kind = "BrokerList"
	eventingRecorder.ListBrokers(brokerList, nil)
	output, err := executeBrokerCommand(eventingClient, "list", "-o", "json")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "\"apiVersion\": \"eventing.knative.dev/v1\"", "\"items\": [],", "\"kind\": \"BrokerList\""))

	eventingRecorder.Validate()
}

func TestTriggerListAllNamespace(t *testing.T) {
	eventingClient := clientv1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	broker1 := createBrokerWithNamespace("foo1", "default1")
	broker2 := createBrokerWithNamespace("foo2", "default2")
	broker3 := createBrokerWithNamespace("foo3", "default3")
	brokerList := &eventingv1.BrokerList{Items: []eventingv1.Broker{*broker1, *broker2, *broker3}}
	eventingRecorder.ListBrokers(brokerList, nil)

	output, err := executeBrokerCommand(eventingClient, "list", "--all-namespaces")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "default1", "foo1"))
	assert.Check(t, util.ContainsAll(outputLines[2], "default2", "foo2"))
	assert.Check(t, util.ContainsAll(outputLines[3], "default3", "foo3"))

	eventingRecorder.Validate()
}
