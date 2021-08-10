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
	"testing"

	"gotest.tools/v3/assert"

	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/util"
)

var (
	brokerName = "foo"
	className  = "foo-class"
)

func TestBrokerCreate(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateBroker(createBroker(brokerName), nil)

	out, err := executeBrokerCommand(eventingClient, "create", brokerName)
	assert.NilError(t, err, "Broker should be created")
	assert.Assert(t, util.ContainsAll(out, "Broker", brokerName, "created", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerCreateWithClass(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateBroker(createBrokerWithClass(brokerName, className), nil)

	out, err := executeBrokerCommand(eventingClient, "create", brokerName, "--class", className)
	assert.NilError(t, err, "Broker should be created")
	assert.Assert(t, util.ContainsAll(out, "Broker", brokerName, "created", "namespace", "default"))

	eventingRecorder.CreateBroker(createBrokerWithClass(brokerName, ""), nil)
	out, err = executeBrokerCommand(eventingClient, "create", brokerName, "--class", "")
	assert.NilError(t, err, "Broker should be created")
	assert.Assert(t, util.ContainsAll(out, "Broker", brokerName, "created", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerCreateWithError(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	_, err := executeBrokerCommand(eventingClient, "create")
	assert.ErrorContains(t, err, "broker create")
	assert.Assert(t, util.ContainsAll(err.Error(), "broker create", "requires", "name", "argument"))
}
