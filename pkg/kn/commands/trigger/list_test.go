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
	"strings"
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clienteventingv1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
)

func TestTriggerList(t *testing.T) {
	servingClient := clientservingv1.NewMockKnServiceClient(t)
	servingRecorder := servingClient.Recorder()
	servingRecorder.GetService("mysink", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysink"},
	}, nil)

	eventingClient := clienteventingv1alpha1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	trigger1 := createTriggerWithStatus("default", "trigger1", map[string]string{"type": "dev.knative.foo"}, "mybroker1", "mysink")
	trigger2 := createTriggerWithStatus("default", "trigger2", map[string]string{"source": "svc.service.knative"}, "mybroker2", "mysink")
	trigger3 := createTriggerWithStatus("default", "trigger3", map[string]string{"type": "src.eventing.knative"}, "mybroker3", "mysink")
	triggerList := &eventingv1alpha1.TriggerList{Items: []eventingv1alpha1.Trigger{*trigger1, *trigger2, *trigger3}}
	eventingRecorder.ListTriggers(triggerList, nil)

	output, err := executeTriggerCommand(eventingClient, nil, "list")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAME", "BROKER", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "trigger1", "mybroker1", "mysink"))
	assert.Check(t, util.ContainsAll(outputLines[2], "trigger2", "mybroker2", "mysink"))
	assert.Check(t, util.ContainsAll(outputLines[3], "trigger3", "mybroker3", "mysink"))

	eventingRecorder.Validate()
}

func TestTriggerListEmpty(t *testing.T) {
	eventingClient := clienteventingv1alpha1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	eventingRecorder.ListTriggers(&eventingv1alpha1.TriggerList{}, nil)
	output, err := executeTriggerCommand(eventingClient, nil, "list")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No", "triggers", "found"))

	eventingRecorder.Validate()
}

func TestTriggerListAllNamespace(t *testing.T) {
	servingClient := clientservingv1.NewMockKnServiceClient(t)
	servingRecorder := servingClient.Recorder()
	servingRecorder.GetService("mysink", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysink"},
	}, nil)

	eventingClient := clienteventingv1alpha1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()

	trigger1 := createTriggerWithStatus("default1", "trigger1", map[string]string{"type": "dev.knative.foo"}, "mybroker1", "mysink")
	trigger2 := createTriggerWithStatus("default2", "trigger2", map[string]string{"source": "svc.service.knative"}, "mybroker2", "mysink")
	trigger3 := createTriggerWithStatus("default3", "trigger3", map[string]string{"type": "src.eventing.knative"}, "mybroker3", "mysink")
	triggerList := &eventingv1alpha1.TriggerList{Items: []eventingv1alpha1.Trigger{*trigger1, *trigger2, *trigger3}}
	eventingRecorder.ListTriggers(triggerList, nil)

	output, err := executeTriggerCommand(eventingClient, nil, "list", "--all-namespaces")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")
	assert.Check(t, util.ContainsAll(outputLines[0], "NAMESPACE", "NAME", "BROKER", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[1], "default1", "trigger1", "mybroker1", "mysink"))
	assert.Check(t, util.ContainsAll(outputLines[2], "default2", "trigger2", "mybroker2", "mysink"))
	assert.Check(t, util.ContainsAll(outputLines[3], "default3", "trigger3", "mybroker3", "mysink"))

	eventingRecorder.Validate()
}
