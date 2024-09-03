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
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"knative.dev/client/pkg/printers"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientv1beta1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/util"
)

func TestSimpleDescribe(t *testing.T) {
	client := clientv1beta1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	trigger := getTriggerSinkRef()

	t.Run("default output", func(t *testing.T) {
		recorder.GetTrigger("testtrigger", trigger, nil)

		out, err := executeTriggerCommand(client, nil, "describe", "testtrigger")
		assert.NilError(t, err)

		assert.Assert(t, cmp.Regexp("Name:\\s+testtrigger", out))
		assert.Assert(t, cmp.Regexp("Namespace:\\s+default", out))

		assert.Assert(t, util.ContainsAll(out, "Broker:", "mybroker"))
		assert.Assert(t, util.ContainsAll(out, "Filter:", "type", "foo.type.knative", "source", "src.eventing.knative"))
		assert.Assert(t, util.ContainsAll(out, "Filters", "experimental", "cesql", "LOWER", "type"))
		assert.Assert(t, util.ContainsAll(out, "Sink:", "Service", "myservicenamespace", "mysvc"))
	})

	t.Run("json format output", func(t *testing.T) {
		recorder.GetTrigger("testtrigger", trigger, nil)

		out, err := executeTriggerCommand(client, nil, "describe", "testtrigger", "-o", "json")
		assert.NilError(t, err)

		result := &v1beta1.Trigger{}
		err = json.Unmarshal([]byte(out), result)
		assert.NilError(t, err)
		assert.DeepEqual(t, trigger, result)
	})

	// Validate that all recorded API methods have been called
	recorder.Validate()
}

func TestDescribeError(t *testing.T) {
	client := clientv1beta1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetTrigger("testtrigger", nil, errors.New("triggers.eventing.knative.dev 'testtrigger' not found"))

	_, err := executeTriggerCommand(client, nil, "describe", "testtrigger")
	assert.ErrorContains(t, err, "testtrigger", "not found")

	recorder.Validate()
}
func TestDescribeTriggerWithSinkURI(t *testing.T) {
	client := clientv1beta1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetTrigger("testtrigger", getTriggerSinkURI(), nil)

	out, err := executeTriggerCommand(client, nil, "describe", "testtrigger")
	assert.NilError(t, err)

	assert.Assert(t, cmp.Regexp("Name:\\s+testtrigger", out))
	assert.Assert(t, cmp.Regexp("Namespace:\\s+default", out))

	assert.Assert(t, util.ContainsAll(out, "Broker:", "mybroker"))
	assert.Assert(t, util.ContainsAll(out, "Filter:", "type", "foo.type.knative", "source", "src.eventing.knative"))
	assert.Assert(t, util.ContainsAll(out, "Sink:", "URI", "https", "foo"))

	// Validate that all recorded API methods have been called
	recorder.Validate()
}

func TestDescribeTriggerMachineReadable(t *testing.T) {
	client := clientv1beta1.NewMockKnEventingClient(t, "mynamespace")

	recorder := client.Recorder()
	recorder.GetTrigger("testtrigger", getTriggerSinkRef(), nil)

	output, err := executeTriggerCommand(client, nil, "describe", "testtrigger", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "kind: Trigger", "spec:", "status:", "metadata:"))

	// Validate that all recorded API methods have been called
	recorder.Validate()
}

func TestWriteNestedFilters(t *testing.T) {
	testCases := []struct {
		name           string
		filter         v1beta1.SubscriptionsAPIFilter
		expectedOutput string
	}{
		{
			name: "Exact filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Exact: map[string]string{
					"type": "example"}},
			expectedOutput: "exact:   \n" +
				"  type:  example\n",
		},
		{
			name: "Prefix filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Prefix: map[string]string{
					"type": "foo.bar"}},
			expectedOutput: "" +
				"prefix:  \n" +
				"  type:  foo.bar\n",
		},
		{
			name: "Suffix filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Suffix: map[string]string{
					"type": "foo.bar"}},
			expectedOutput: "" +
				"suffix:  \n" +
				"  type:  foo.bar\n",
		},
		{
			name: "All filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				All: []v1beta1.SubscriptionsAPIFilter{
					{Exact: map[string]string{
						"type": "foo.bar"}},
					{Prefix: map[string]string{
						"source": "foo"}},
					{Suffix: map[string]string{
						"subject": "test"}}}},
			expectedOutput: "" +
				"all:          \n" +
				"  exact:      \n" +
				"    type:     foo.bar\n" +
				"  prefix:     \n" +
				"    source:   foo\n" +
				"  suffix:     \n" +
				"    subject:  test\n",
		},
		{
			name: "Any filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Any: []v1beta1.SubscriptionsAPIFilter{
					{Exact: map[string]string{
						"type": "foo.bar"}},
					{Prefix: map[string]string{
						"source": "foo"}},
					{Suffix: map[string]string{
						"subject": "test"}}},
			},
			expectedOutput: "" +
				"any:          \n" +
				"  exact:      \n" +
				"    type:     foo.bar\n" +
				"  prefix:     \n" +
				"    source:   foo\n" +
				"  suffix:     \n" +
				"    subject:  test\n",
		},
		{
			name: "Nested All filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				All: []v1beta1.SubscriptionsAPIFilter{
					{Exact: map[string]string{
						"type": "foo.bar"}},
					{All: []v1beta1.SubscriptionsAPIFilter{
						{Prefix: map[string]string{
							"source": "foo"}},
						{Suffix: map[string]string{
							"subject": "test"}}}}}},
			expectedOutput: "" +
				"all:            \n" +
				"  exact:        \n" +
				"    type:       foo.bar\n" +
				"  all:          \n" +
				"    prefix:     \n" +
				"      source:   foo\n" +
				"    suffix:     \n" +
				"      subject:  test\n",
		},
		{
			name: "Nested Any filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Any: []v1beta1.SubscriptionsAPIFilter{
					{Exact: map[string]string{
						"type": "foo.bar"}},
					{Any: []v1beta1.SubscriptionsAPIFilter{
						{Prefix: map[string]string{
							"source": "foo"}},
						{Suffix: map[string]string{
							"subject": "test"}}}}}},
			expectedOutput: "" +
				"any:            \n" +
				"  exact:        \n" +
				"    type:       foo.bar\n" +
				"  any:          \n" +
				"    prefix:     \n" +
				"      source:   foo\n" +
				"    suffix:     \n" +
				"      subject:  test\n",
		},
		{
			name: "Nested Not filter",
			filter: v1beta1.SubscriptionsAPIFilter{
				Not: &v1beta1.SubscriptionsAPIFilter{
					Exact: map[string]string{
						"type": "foo.bar",
					},
					Prefix: map[string]string{
						"type": "bar",
					},
					CESQL: "select bar",
					Not: &v1beta1.SubscriptionsAPIFilter{
						Suffix: map[string]string{
							"source": "foo"}}}},
			expectedOutput: "" +
				"not:           \n" +
				"  not:         \n" +
				"    suffix:    \n" +
				"      source:  foo\n" +
				"  exact:       \n" +
				"    type:      foo.bar\n" +
				"  prefix:      \n" +
				"    type:      bar\n" +
				"  cesql:       select bar\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			dw := printers.NewPrefixWriter(buf)
			writeNestedFilters(dw, tc.filter)
			err := dw.Flush()
			assert.NilError(t, err)
			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}

func getTriggerSinkRef() *v1beta1.Trigger {
	return &v1beta1.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testtrigger",
			Namespace: "default",
		},
		Spec: v1beta1.TriggerSpec{
			Broker: "mybroker",
			Filter: &v1beta1.TriggerFilter{
				Attributes: v1beta1.TriggerFilterAttributes{
					"type":   "foo.type.knative",
					"source": "src.eventing.knative",
				},
			},
			Filters: []v1beta1.SubscriptionsAPIFilter{
				{CESQL: "LOWER(type) = 'my-event-type'"},
			},
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:      "Service",
					Namespace: "myservicenamespace",
					Name:      "mysvc",
				},
			},
		},
		Status: v1beta1.TriggerStatus{},
	}
}

func getTriggerSinkURI() *v1beta1.Trigger {
	return &v1beta1.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testtrigger",
			Namespace: "default",
		},
		Spec: v1beta1.TriggerSpec{
			Broker: "mybroker",
			Filter: &v1beta1.TriggerFilter{
				Attributes: v1beta1.TriggerFilterAttributes{
					"type":   "foo.type.knative",
					"source": "src.eventing.knative",
				},
			},
			Subscriber: duckv1.Destination{
				URI: &apis.URL{
					Scheme: "https",
					Host:   "foo",
				},
			},
		},
		Status: v1beta1.TriggerStatus{},
	}
}
