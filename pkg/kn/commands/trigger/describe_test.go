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
	"errors"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	client_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/util"
)

func TestSimpleDescribe(t *testing.T) {
	client := client_v1alpha1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetTrigger("testtrigger", getTrigger(), nil)

	out, err := executeTriggerCommand(client, nil, "describe", "testtrigger")
	assert.NilError(t, err)

	assert.Assert(t, cmp.Regexp("Name:\\s+testtrigger", out))
	assert.Assert(t, cmp.Regexp("Namespace:\\s+default", out))

	assert.Assert(t, util.ContainsAll(out, "Broker:", "mybroker"))
	assert.Assert(t, util.ContainsAll(out, "Filter:", "type", "foo.type.knative", "source", "src.eventing.knative"))
	assert.Assert(t, util.ContainsAll(out, "Sink:", "Service", "myservicenamespace", "mysvc"))

	// Validate that all recorded API methods have been called
	recorder.Validate()
}

func TestDescribeError(t *testing.T) {
	client := client_v1alpha1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetTrigger("testtrigger", nil, errors.New("triggers.eventing.knative.dev 'testtrigger' not found"))

	_, err := executeTriggerCommand(client, nil, "describe", "testtrigger")
	assert.ErrorContains(t, err, "testtrigger", "not found")

	recorder.Validate()
}

func getTrigger() *v1alpha1.Trigger {
	return &v1alpha1.Trigger{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testtrigger",
			Namespace: "default",
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: "mybroker",
			Filter: &v1alpha1.TriggerFilter{
				Attributes: &v1alpha1.TriggerFilterAttributes{
					"type":   "foo.type.knative",
					"source": "src.eventing.knative",
				},
			},
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:      "Service",
					Namespace: "myservicenamespace",
					Name:      "mysvc",
				},
			},
		},
		Status: v1alpha1.TriggerStatus{},
	}
}
