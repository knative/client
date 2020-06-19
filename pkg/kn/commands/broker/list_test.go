// Copyright © 2020 The Knative Authors
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

package broker

import (
	"strings"
	"testing"

	"gotest.tools/assert"

	clienteventingv1beta1 "knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/util"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
)

func TestBrokerList(t *testing.T) {
	eventingClient := clienteventingv1beta1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	broker1 := createBroker("foo1")
	broker2 := createBroker("foo2")
	broker3 := createBroker("foo3")
	brokerList := &v1beta1.BrokerList{Items: []v1beta1.Broker{*broker1, *broker2, *broker3}}
	eventingRecorder.ListBrokers(brokerList, nil)

	output, err := executeBrokerCommand(eventingClient, "list")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "URL", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "foo1"))
	assert.Check(t, util.ContainsAll(outputLines[2], "foo2"))
	assert.Check(t, util.ContainsAll(outputLines[3], "foo3"))

	eventingRecorder.Validate()
}

func TestBrokerListEmpty(t *testing.T) {
	eventingClient := clienteventingv1beta1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	eventingRecorder.ListBrokers(&v1beta1.BrokerList{}, nil)
	output, err := executeBrokerCommand(eventingClient, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "brokers", "found"))

	eventingRecorder.Validate()
}

func TestTriggerListAllNamespace(t *testing.T) {
	eventingClient := clienteventingv1beta1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	broker1 := createBrokerWithNamespace("foo1", "default1")
	broker2 := createBrokerWithNamespace("foo2", "default2")
	broker3 := createBrokerWithNamespace("foo3", "default3")
	brokerList := &v1beta1.BrokerList{Items: []v1beta1.Broker{*broker1, *broker2, *broker3}}
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
