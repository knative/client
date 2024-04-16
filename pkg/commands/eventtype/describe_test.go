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
	"fmt"
	"testing"

	dynamicfake "knative.dev/client-pkg/pkg/dynamic/fake"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client-pkg/pkg/eventing/v1beta2"
	"knative.dev/client-pkg/pkg/util"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestEventtypeDescribe(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.GetEventtype(eventtypeName, getEventtype(eventtypeName, testNs), nil)

	out, err := executeEventtypeCommand(eventingClient, dynamicClient, "describe", eventtypeName, "--namespace", testNs)

	assert.NilError(t, err)

	assert.Assert(t, cmp.Regexp(fmt.Sprintf("Name:\\s+%s", eventtypeName), out))
	assert.Assert(t, cmp.Regexp(fmt.Sprintf("Namespace:\\s+%s", testNs), out))
	assert.Assert(t, cmp.Regexp(fmt.Sprintf("Source:\\s+%s", testSource), out))
	assert.Assert(t, cmp.Regexp("Reference:\\s+\n", out))
	assert.Assert(t, cmp.Regexp(fmt.Sprintf("Name:\\s+%s", testBroker), out))

	assert.Assert(t, util.ContainsAll(out, "Conditions:", "Ready", "BrokerReady", "BrokerExists"))

	eventingRecorder.Validate()
}

func TestEventtypeDescribeError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.GetEventtype(eventtypeName, getEventtype(eventtypeName, testNs), fmt.Errorf("mock-error"))

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "describe", eventtypeName, "--namespace", testNs)

	assert.Error(t, err, "mock-error")

	eventingRecorder.Validate()
}

func TestEventtypeDescribeWithNameMissingWithError(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	_, err := executeEventtypeCommand(eventingClient, dynamicClient, "describe", "--namespace", testNs)

	assert.ErrorContains(t, err, "eventtype describe")
	assert.Assert(t, util.ContainsAll(err.Error(), "requires", "eventtype", "name"))
}

func TestEventtypeDescribeMachineReadable(t *testing.T) {
	eventingClient := v1beta2.NewMockKnEventingV1beta2Client(t, testNs)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(testNs)

	eventingRecorder := eventingClient.Recorder()

	eventtype := getEventtype(eventtypeName, testNs)

	// json
	eventingRecorder.GetEventtype(eventtypeName, eventtype, nil)
	out, err := executeEventtypeCommand(eventingClient, dynamicClient, "describe", eventtypeName, "--namespace", testNs, "-o", "json")

	assert.NilError(t, err)
	result := &eventingv1beta2.EventType{}
	err = json.Unmarshal([]byte(out), result)
	assert.NilError(t, err)
	assert.DeepEqual(t, eventtype, result)

	// yaml
	eventingRecorder.GetEventtype(eventtypeName, eventtype, nil)
	out, err = executeEventtypeCommand(eventingClient, dynamicClient, "describe", eventtypeName, "--namespace", testNs, "-o", "yaml")

	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "kind: EventType", "spec:", "status:", "metadata:"))

	eventingRecorder.Validate()
}

func getEventtype(name string, ns string) *eventingv1beta2.EventType {
	source, _ := apis.ParseURL(testSource)
	return &eventingv1beta2.EventType{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventType",
			APIVersion: eventingv1beta2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: eventingv1beta2.EventTypeSpec{
			Type:   cetype,
			Source: source,
			Reference: &duckv1.KReference{
				APIVersion: eventingv1.SchemeGroupVersion.String(),
				Kind:       "Broker",
				Name:       testBroker,
			},
		},
		Status: eventingv1beta2.EventTypeStatus{
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					apis.Condition{
						Type:   "BrokerExists",
						Status: "True",
					},
					apis.Condition{
						Type:   "BrokerReady",
						Status: "True",
					},
					apis.Condition{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
		},
	}
}
