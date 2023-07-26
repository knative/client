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
	"fmt"
	"testing"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/eventing/v1beta2"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/apis"
)

func TestEventTypeCreate(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateEventtype(createEventtype(eventtypeName, cetype, testNs), nil)

	out, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype, "--namespace", testNs)
	assert.NilError(t, err, "Eventtype should be created")
	assert.Assert(t, util.ContainsAll(out, "Eventtype", eventtypeName, "created", "namespace", testNs))

	// Create eventtype without namespace flag set
	eventingRecorder.CreateEventtype(createEventtype(eventtypeName, cetype, "default"), nil)
	out, err = executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype)

	assert.NilError(t, err, "Eventtype should be created")
	assert.Assert(t, util.ContainsAll(out, "Eventtype", eventtypeName, "created", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestEventTypeCreateWithoutTypeError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--namespace", testNs)
	assert.Assert(t, util.ContainsAll(err.Error(), "required", "flag(s)", "type", "not", "set"))
}

func TestEventTypeCreateWithoutNameError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", "--namespace", testNs, "--type", cetype)
	assert.Assert(t, util.ContainsAll(err.Error(), "requires", "eventtype", "name"))
}

func TestEventTypeCreateWithSource(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	url, _ := apis.ParseURL(testSource)
	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateEventtype(createEventtypeWithSource(eventtypeName, cetype, testNs, url), nil)

	out, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype, "--source", testSource, "--namespace", testNs)

	assert.NilError(t, err, "Eventtype should be created")
	assert.Assert(t, util.ContainsAll(out, "Eventtype", eventtypeName, "created", "namespace", testNs))

	eventingRecorder.Validate()

}

func TestEventTypeCreateWithSourceError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype, "--source", testSourceError, "--namespace", testNs)

	assert.ErrorContains(t, err, "cannot create eventtype")
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "character", "URL"))
}

func TestEventTypeCreateWithBroker(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateEventtype(createEventtypeWithBroker(eventtypeName, cetype, testBroker, testNs), nil)

	out, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype, "--namespace", testNs, "--broker", testBroker)
	assert.NilError(t, err, "Eventtype should be created")
	assert.Assert(t, util.ContainsAll(out, "Eventtype", eventtypeName, "created", "namespace", testNs))

	eventingRecorder.Validate()
}

func TestEventTypeCreateWithError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateEventtype(createEventtype(eventtypeName, cetype, testNs), fmt.Errorf("mock-error"))

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "create", eventtypeName, "--type", cetype, "--namespace", testNs)

	assert.ErrorContains(t, err, "cannot create eventtype")
	assert.Assert(t, util.ContainsAll(err.Error(), "mock-error"))

	eventingRecorder.Validate()
}
