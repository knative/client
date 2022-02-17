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
	"encoding/json"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/util"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
)

func TestEventtypeList(t *testing.T) {
	eventingClient := v1beta1.NewMockKnEventingV1beta1Client(t, testNs)

	eventingRecorder := eventingClient.Recorder()

	eventtype1 := getEventtype("foo1", testNs)
	eventtype2 := getEventtype("foo2", testNs)
	eventtype3 := getEventtype("foo3", testNs)

	eventtypeList := &eventingv1beta1.EventTypeList{Items: []eventingv1beta1.EventType{*eventtype1, *eventtype2, *eventtype3}}

	util.UpdateGroupVersionKindWithScheme(eventtypeList, eventingv1beta1.SchemeGroupVersion, scheme.Scheme)

	t.Run("default output", func(t *testing.T) {
		eventingRecorder.ListEventtypes(eventtypeList, nil)

		output, err := executeEventtypeCommand(eventingClient, "list")
		assert.NilError(t, err)

		outputLines := strings.Split(output, "\n")
		assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "TYPE", "SOURCE", "BROKER", "SCHEMA", "AGE", "CONDITIONS", "READY", "REASON"))
		assert.Check(t, util.ContainsAll(outputLines[1], "foo1", cetype, testBroker, testSource, "True"))
		assert.Check(t, util.ContainsAll(outputLines[2], "foo2", cetype, testBroker, testSource, "True"))
		assert.Check(t, util.ContainsAll(outputLines[3], "foo3", cetype, testBroker, testSource, "True"))

		eventingRecorder.Validate()
	})

	t.Run("json format output", func(t *testing.T) {
		eventingRecorder.ListEventtypes(eventtypeList, nil)

		output, err := executeEventtypeCommand(eventingClient, "list", "-o", "json")
		assert.NilError(t, err)

		result := eventingv1beta1.EventTypeList{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NilError(t, err)
		assert.DeepEqual(t, eventtypeList.Items, result.Items)

		eventingRecorder.Validate()
	})

	t.Run("all namespaces", func(t *testing.T) {
		eventingRecorder.ListEventtypes(eventtypeList, nil)

		output, err := executeEventtypeCommand(eventingClient, "list", "--all-namespaces")
		assert.NilError(t, err)

		outputLines := strings.Split(output, "\n")
		assert.Check(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "TYPE", "SOURCE", "BROKER", "SCHEMA", "AGE", "CONDITIONS", "READY", "REASON"))
		assert.Check(t, util.ContainsAll(outputLines[1], "foo1", testNs, cetype, testBroker, testSource, "True"))
		assert.Check(t, util.ContainsAll(outputLines[2], "foo2", testNs, cetype, testBroker, testSource, "True"))
		assert.Check(t, util.ContainsAll(outputLines[3], "foo3", testNs, cetype, testBroker, testSource, "True"))

		eventingRecorder.Validate()
	})
}

func TestEventtypeListEmpty(t *testing.T) {
	eventingClient := v1beta1.NewMockKnEventingV1beta1Client(t, testNs)

	eventingRecorder := eventingClient.Recorder()

	eventingRecorder.ListEventtypes(&eventingv1beta1.EventTypeList{}, nil)
	output, err := executeEventtypeCommand(eventingClient, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "eventtypes", "found"))

	eventingRecorder.Validate()
}
