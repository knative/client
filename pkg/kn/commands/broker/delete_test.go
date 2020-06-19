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
	"fmt"
	"testing"

	"gotest.tools/assert"

	clienteventingv1beta1 "knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/util"
)

func TestBrokerDelete(t *testing.T) {
	brokerName := "foo"
	eventingClient := clienteventingv1beta1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteBroker(brokerName, nil)

	out, err := executeBrokerCommand(eventingClient, "delete", brokerName)
	assert.NilError(t, err, "Broker should be deleted")
	util.ContainsAll(out, "Broker", brokerName, "deleted", "namespace", "default")

	eventingRecorder.Validate()
}

func TestBrokerWithDelete(t *testing.T) {
	brokerName := "foo"
	eventingClient := clienteventingv1beta1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.DeleteBroker(brokerName, fmt.Errorf("broker %s not found", brokerName))

	out, err := executeBrokerCommand(eventingClient, "delete", brokerName)
	assert.ErrorContains(t, err, brokerName)
	util.ContainsAll(out, "broker", brokerName, "not found")

	eventingRecorder.Validate()
}
