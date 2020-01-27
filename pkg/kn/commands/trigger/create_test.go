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

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	dynamic_fake "knative.dev/client/pkg/dynamic/fake"
	eventing_client "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/util"
)

var (
	triggerName = "foo"
)

func TestTriggerCreate(t *testing.T) {
	eventingClient := eventing_client.NewMockKnEventingClient(t)
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	})

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateTrigger(createTrigger("default", triggerName, map[string]string{"type": "dev.knative.foo"}, "mybroker", "mysvc"), nil)

	out, err := executeTriggerCommand(eventingClient, dynamicClient, "create", triggerName, "--broker", "mybroker",
		"--filter", "type=dev.knative.foo", "--sink", "svc:mysvc")
	assert.NilError(t, err, "Trigger should be created")
	util.ContainsAll(out, "Trigger", triggerName, "created", "namespace", "default")

	eventingRecorder.Validate()
}

func TestSinkNotFoundError(t *testing.T) {
	eventingClient := eventing_client.NewMockKnEventingClient(t)
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default")

	errorMsg := fmt.Sprintf("cannot create trigger '%s' in namespace 'default' because: services.serving.knative.dev \"mysvc\" not found", triggerName)

	out, err := executeTriggerCommand(eventingClient, dynamicClient, "create", triggerName, "--broker", "mybroker",
		"--filter", "type=dev.knative.foo", "--sink", "svc:mysvc")
	assert.Error(t, err, errorMsg)
	assert.Assert(t, util.ContainsAll(out, errorMsg, "Usage"))
}

func TestNoSinkError(t *testing.T) {
	eventingClient := eventing_client.NewMockKnEventingClient(t)
	_, err := executeTriggerCommand(eventingClient, nil, "create", triggerName, "--broker", "mybroker",
		"--filter", "type=dev.knative.foo")
	assert.ErrorContains(t, err, "required flag(s)", "sink", "not set")
}

func TestTriggerCreateMultipleFilter(t *testing.T) {
	eventingClient := eventing_client.NewMockKnEventingClient(t)
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	})

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateTrigger(createTrigger("default", triggerName, map[string]string{"type": "dev.knative.foo", "source": "event.host"}, "mybroker", "mysvc"), nil)

	out, err := executeTriggerCommand(eventingClient, dynamicClient, "create", triggerName, "--broker", "mybroker",
		"--filter", "type=dev.knative.foo", "--filter", "source=event.host", "--sink", "svc:mysvc")
	assert.NilError(t, err, "Trigger should be created")
	util.ContainsAll(out, "Trigger", triggerName, "created", "namespace", "default")

	eventingRecorder.Validate()
}

func TestTriggerCreateWithoutFilter(t *testing.T) {
	eventingClient := eventing_client.NewMockKnEventingClient(t)
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	})

	eventingRecorder := eventingClient.Recorder()
	eventingRecorder.CreateTrigger(createTrigger("default", triggerName, nil, "mybroker", "mysvc"), nil)

	out, err := executeTriggerCommand(eventingClient, dynamicClient, "create", triggerName, "--broker", "mybroker", "--sink", "svc:mysvc")
	assert.NilError(t, err, "Trigger should be created")
	util.ContainsAll(out, "Trigger", triggerName, "created", "namespace", "default")

	eventingRecorder.Validate()
}
