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

package dynamic

import (
	"testing"

	"github.com/magiconair/properties/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const testNamespace = "testns"

func newUnstructured(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       crdKind,
			"metadata": map[string]interface{}{
				"namespace": testNamespace,
				"name":      name,
				"labels": map[string]interface{}{
					sourcesLabelKey: sourcesLabelValue,
				},
			},
		},
	}
}

func TestNamespace(t *testing.T) {
	client := createFakeKnDynamicClient(testNamespace, newUnstructured("foo"))
	assert.Equal(t, client.Namespace(), testNamespace)
}

func TestListCRDs(t *testing.T) {
	client := createFakeKnDynamicClient(
		testNamespace,
		newUnstructured("foo"),
		newUnstructured("bar"),
	)

	t.Run("List CRDs with match", func(t *testing.T) {
		options := metav1.ListOptions{}
		uList, err := client.ListCRDs(options)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(uList.Items), 2)
	})

	t.Run("List CRDs without match", func(t *testing.T) {
		options := metav1.ListOptions{}
		sourcesLabels := labels.Set{"duck.knative.dev/source": "true1"}
		options.LabelSelector = sourcesLabels.String()
		uList, err := client.ListCRDs(options)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(uList.Items), 0)
	})
}

func TestListSourceTypes(t *testing.T) {
	client := createFakeKnDynamicClient(
		testNamespace,
		newUnstructured("foo"),
		newUnstructured("bar"),
	)

	t.Run("List source types", func(t *testing.T) {
		uList, err := client.ListSourcesTypes()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(uList.Items), 2)
		assert.Equal(t, uList.Items[0].GetName(), "foo")
		assert.Equal(t, uList.Items[1].GetName(), "bar")
	})
}

// createFakeKnDynamicClient gives you a dynamic client for testing contianing the given objects.
// See also the one in the fake package. Duplicated here to avoid a dependency loop.
func createFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) KnDynamicClient {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1alpha1", Kind: "Service"}, &servingv1.Service{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "eventing.knative.dev", Version: "v1alpha1", Kind: "Broker"}, &eventingv1alpha1.Broker{})
	client := dynamicfake.NewSimpleDynamicClient(scheme, objects...)
	return NewKnDynamicClient(client, testNamespace)
}
