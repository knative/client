// Copyright © 2020 The Knative Authors
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

package duck

import (
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	//"knative.dev/client/pkg/util"
)

func TestToSource(t *testing.T) {
	s := toSource(newSourceUnstructuredObjWithSink("a1",
		"sources.knative.dev/v1alpha1", "ApiServerSource"))
	assert.Check(t, s.Name == "a1")
	s = toSource(newSourceUnstructuredObjWithSink("s1",
		"sources.knative.dev/v1alpha1", "SinkBinding"))
	assert.Check(t, s.SourceKind == "SinkBinding")
	s = toSource(newSourceUnstructuredObjWithSink("p1",
		"sources.knative.dev/v1alpha1", "PingSource"))
	assert.Check(t, s.Sink == "svc:foo")
	s = toSource(newSourceUnstructuredObjWithSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.Check(t, s.Sink == "svc:foo")
	s = toSource(newSourceUnstructuredObjWithoutSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.Check(t, s.Sink == "")
}

func TestSinkFromUnstructured(t *testing.T) {
	s, e := sinkFromUnstructured(newSourceUnstructuredObjWithSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.NilError(t, e)
	assert.Check(t, s != nil)

	s, e = sinkFromUnstructured(newSourceUnstructuredObjWithoutSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.NilError(t, e)
	assert.Check(t, s == nil)

	s, e = sinkFromUnstructured(newSourceUnstructuredObjWithIncorrectSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.Check(t, e != nil)
	assert.Check(t, s == nil)

}

func TestConditionsFromUnstructured(t *testing.T) {
	_, e := conditionsFromUnstructured(newSourceUnstructuredObjWithSink("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.NilError(t, e)

	_, e = conditionsFromUnstructured(newSourceUnstructuredObjWithoutConditions("k1",
		"sources.knative.dev/v1alpha1", "KafkaSource"))
	assert.Check(t, e != nil)

}

func newSourceUnstructuredObjWithSink(name, apiVersion, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
			},
			"spec": map[string]interface{}{
				"sink": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind": "Service",
						"name": "foo",
					},
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"Type":   "Ready",
						"Status": "True",
					},
				},
			},
		},
	}
}

func newSourceUnstructuredObjWithoutSink(name, apiVersion, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
			},
			"spec": map[string]interface{}{},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"Type":   "Ready",
						"Status": "False",
						"Reason": "SinkMissing",
					},
				},
			},
		},
	}
}

func newSourceUnstructuredObjWithIncorrectSink(name, apiVersion, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
			},
			"spec": map[string]interface{}{"sink": "incorrect"},
		},
	}
}

func newSourceUnstructuredObjWithoutConditions(name, apiVersion, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
			},
			"spec": map[string]interface{}{
				"sink": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind": "Service",
						"name": "foo",
					},
				},
			},
			"status": map[string]interface{}{},
		},
	}
}
