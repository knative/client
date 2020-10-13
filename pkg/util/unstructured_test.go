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

package util

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestToUnstructuredList(t *testing.T) {
	serviceList := servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{createService("s1"), createService("s2")},
	}
	expectedList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": string("v1"),
			"kind":       string("List"),
		},
	}
	expectedList.Items = []unstructured.Unstructured{createUnstructured("s1"), createUnstructured("s2")}
	unstructuredList, err := ToUnstructuredList(&serviceList)
	assert.NilError(t, err)
	assert.DeepEqual(t, unstructuredList, expectedList)

	service1 := createService("s3")
	expectedList = &unstructured.UnstructuredList{}
	expectedList.Items = []unstructured.Unstructured{createUnstructured("s3")}
	unstructuredList, err = ToUnstructuredList(&service1)
	assert.NilError(t, err)
	assert.DeepEqual(t, unstructuredList, expectedList)
}

func createService(name string) servingv1.Service {
	service := servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
	return service
}

func createUnstructured(name string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": nil,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{"creationTimestamp": nil},
					"spec":     map[string]interface{}{"containers": nil},
				},
			},
			"status": map[string]interface{}{},
		},
	}
}
