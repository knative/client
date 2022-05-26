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

package broker

import (
	"testing"

	"gotest.tools/v3/assert"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/util"
)

func TestBrokerUpdateWithDlSink(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithDlSink("test-broker", testSvc)
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--dl-sink", testSvc)
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateWithTimeout(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithTimeout("test-broker", "10")
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--timeout", "10")
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateWithRetry(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithRetry("test-broker", 5)
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--retry", "5")
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateWithBackoffPolicy(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithBackoffPolicy("test-broker", "linear")
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--backoff-policy", "linear")
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateWithBackoffDelay(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithBackoffDelay("test-broker", "PT10S")
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--backoff-delay", "PT10S")
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateWithRetryAfterMax(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	updated := createBrokerWithRetryAfterMax("test-broker", "PT10S")
	eventingRecorder.GetBroker("test-broker", present, nil)
	eventingRecorder.UpdateBroker(updated, nil)

	out, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--retry-after-max", "PT10S")
	assert.NilError(t, err, "Broker should be updated")
	assert.Assert(t, util.ContainsAll(out, "Broker", "test-broker", "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestBrokerUpdateError(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createBroker("test-broker")
	eventingRecorder.GetBroker("test-broker", present, nil)

	_, err := executeBrokerCommand(eventingClient, "update", "test-broker",
		"--dl-sink", "absent-svc")
	assert.ErrorContains(t, err, "not found")
	eventingRecorder.Validate()
}

func TestBrokerUpdateErrorNoFlags(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()

	_, err := executeBrokerCommand(eventingClient, "update", "test-broker")
	assert.ErrorContains(t, err, "flag(s) not set")
	eventingRecorder.Validate()
}

func TestBrokerUpdateErrorNoName(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()

	_, err := executeBrokerCommand(eventingClient, "update", "--dl-sink", testSvc)
	assert.ErrorContains(t, err, "requires the broker name")
	eventingRecorder.Validate()
}
