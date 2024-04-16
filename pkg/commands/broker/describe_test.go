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
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientv1 "knative.dev/client-pkg/pkg/eventing/v1"
	"knative.dev/client-pkg/pkg/util"
)

func TestBrokerDescribe(t *testing.T) {
	client := clientv1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	broker := getBroker()

	t.Run("default output", func(t *testing.T) {
		recorder.GetBroker("foo", broker, nil)

		out, err := executeBrokerCommand(client, "describe", "foo")
		assert.NilError(t, err)

		assert.Assert(t, cmp.Regexp("Name:\\s+foo", out))
		assert.Assert(t, cmp.Regexp("Namespace:\\s+default", out))

		assert.Assert(t, util.ContainsAll(out, "Address:", "URL:", "http://foo-broker.test"))
		assert.Assert(t, util.ContainsAll(out, "Conditions:", "Ready"))

		// There're 2 empty lines used in the "describe" formatting
		lineCounter := 0
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if line == "" {
				lineCounter++
			}
		}
		assert.Equal(t, lineCounter, 2)
	})

	t.Run("json format output", func(t *testing.T) {
		recorder.GetBroker("foo", broker, nil)

		out, err := executeBrokerCommand(client, "describe", "foo", "-o", "json")
		assert.NilError(t, err)

		result := &eventingv1.Broker{}
		err = json.Unmarshal([]byte(out), result)
		assert.NilError(t, err)
		assert.DeepEqual(t, broker, result)
	})

	// Validate that all recorded API methods have been called
	recorder.Validate()
}

func TestDescribeError(t *testing.T) {
	client := clientv1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetBroker("foo", nil, errors.New("brokers.eventing.knative.dev 'foo' not found"))

	_, err := executeBrokerCommand(client, "describe", "foo")
	assert.ErrorContains(t, err, "foo", "not found")

	recorder.Validate()
}

func TestBrokerDescribeURL(t *testing.T) {
	client := clientv1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetBroker("foo", getBroker(), nil)

	out, err := executeBrokerCommand(client, "describe", "foo", "-o", "url")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "http://foo-broker.test"))

	recorder.Validate()
}

func TestTriggerDescribeMachineReadable(t *testing.T) {
	client := clientv1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetBroker("foo", getBroker(), nil)

	out, err := executeBrokerCommand(client, "describe", "foo", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "kind: Broker", "spec:", "status:", "metadata:"))

	recorder.Validate()

}
func getBroker() *eventingv1.Broker {
	return &eventingv1.Broker{
		TypeMeta: v1.TypeMeta{
			Kind:       "Broker",
			APIVersion: eventingv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Status: eventingv1.BrokerStatus{
			AddressStatus: duckv1.AddressStatus{
				Address: &duckv1.Addressable{
					URL: &apis.URL{Scheme: "http", Host: "foo-broker.test"},
				},
			},
			Status: duckv1.Status{
				Conditions: duckv1.Conditions{
					apis.Condition{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
		},
	}
}
