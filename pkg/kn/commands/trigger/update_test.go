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
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/util"
)

func TestTriggerUpdate(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	})

	eventingRecorder := eventingClient.Recorder()
	present := createTrigger("default", triggerName, map[string]string{"type": "dev.knative.foo"}, "mybroker", "mysvc")
	updated := createTrigger("default", triggerName, map[string]string{"type": "dev.knative.new"}, "mybroker", "mysvc")
	eventingRecorder.GetTrigger(triggerName, present, nil)
	eventingRecorder.UpdateTrigger(updated, nil)

	out, err := executeTriggerCommand(eventingClient, dynamicClient, "update", triggerName,
		"--filter", "type=dev.knative.new", "--sink", "ksvc:mysvc")
	assert.NilError(t, err, "Trigger should be updated")
	assert.Assert(t, util.ContainsAll(out, "Trigger", triggerName, "updated", "namespace", "default"))

	eventingRecorder.Validate()
}

func TestTriggerUpdateWithError(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)
	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.GetTrigger(triggerName, nil, fmt.Errorf("trigger not found"))

	out, err := executeTriggerCommand(eventingClient, nil, "update", triggerName,
		"--filter", "type=dev.knative.new", "--sink", "ksvc:newsvc")
	assert.ErrorContains(t, err, "trigger not found")
	assert.Assert(t, util.ContainsAll(out, "Usage", triggerName))

	eventingRecorder.Validate()
}

func TestTriggerUpdateDeletionTimestampNotNil(t *testing.T) {
	eventingClient := clienteventingv1.NewMockKnEventingClient(t)

	eventingRecorder := eventingClient.Recorder()
	present := createTrigger("default", triggerName, map[string]string{"type": "dev.knative.foo"}, "mybroker", "mysvc")
	present.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	eventingRecorder.GetTrigger(triggerName, present, nil)

	_, err := executeTriggerCommand(eventingClient, nil, "update", triggerName,
		"--filter", "type=dev.knative.new", "--sink", "ksvc:mysvc")
	assert.ErrorContains(t, err, present.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "trigger")
}
