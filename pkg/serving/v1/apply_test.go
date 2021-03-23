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

package v1

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/yaml"

	"knative.dev/client/pkg/util"
)

func TestApplyServiceWithNoImage(t *testing.T) {
	_, client := setup()
	serviceFaulty := newService("faulty-service")
	_, err := client.ApplyService(context.Background(), serviceFaulty)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "image name"))
}

func TestApplyServiceCreate(t *testing.T) {
	serving, client := setup()

	serviceNew := newServiceWithImage("new-service", "test/image")
	serving.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
		})

	serving.AddReactor("create", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			return true, serviceNew, nil
		})

	hasChanged, err := client.ApplyService(context.Background(), serviceNew)
	assert.NilError(t, err)
	assert.Assert(t, hasChanged, "service has changed")
}

func TestApplyServiceUpdate(t *testing.T) {
	serving, client := setup()

	serviceOld := newServiceWithImage("my-service", "test/image")
	serviceNew := newServiceWithImage("my-service", "test/new-image")
	serving.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			assert.Equal(t, name, "my-service")
			return true, serviceOld, nil
		})

	serving.AddReactor("patch", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			serviceNew.Generation = 2
			serviceNew.Status.ObservedGeneration = 1
			return true, serviceNew, nil
		})

	hasChanged, err := client.ApplyService(context.Background(), serviceNew)
	assert.NilError(t, err)
	assert.Assert(t, hasChanged, "service has changed")
}

func newServiceWithImage(name string, image string) *servingv1.Service {
	svc := newService(name)
	svc.Spec = servingv1.ServiceSpec{
		ConfigurationSpec: servingv1.ConfigurationSpec{
			Template: servingv1.RevisionTemplateSpec{
				Spec: servingv1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: image,
							},
						},
					},
				},
			},
		},
	}
	return svc
}

func TestExtractUserContainer(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{"Simple Service",
			`
spec:
  template:
    spec:
      containers:
        - image: gcr.io/foo/bar:baz
`,
			`
image: gcr.io/foo/bar:baz
`,
		},
		{
			"No template",
			`
spec:
`,
			"",
		}, {
			"No template spec",
			`
spec:
  template:
`,
			"",
		},
		{
			"No template spec containers",
			`
spec:
  template:
    spec:
`,
			"",
		},
		{
			"Empty template spec containers",
			`
spec:
  template:
    spec:
      containers: []
`,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceMap map[string]interface{}
			yaml.Unmarshal([]byte(tt.service), &serviceMap)

			got := extractUserContainer(serviceMap)

			if tt.want == "" {
				assert.Assert(t, got == nil)
			} else {
				var expectedMap map[string]interface{}
				yaml.Unmarshal([]byte(tt.want), &expectedMap)
				if !reflect.DeepEqual(got, expectedMap) {
					t.Errorf("extractUserContainer() = %v, want %v", got, expectedMap)
				}
			}
		})
	}
}

func TestCleanupServiceUnstructured(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
	}{
		{"Simple Service with fields to remove",
			`
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: foo
  creationTimestamp: "2020-10-22T08:16:37Z"
spec:
  template:
    metadata:
      name: "bar"
      creationTimestamp: null
    spec:
      containers:
      - image: gcr.io/foo/bar:baz
        name: "bla"
        resources: {}
status:
  observedGeneration: 1
`,
			`
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: foo
spec:
  template:
    metadata:
      name: "bar"
    spec:
      containers:
      - image: gcr.io/foo/bar:baz
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ud := &unstructured.Unstructured{}
			assert.NilError(t, yaml.Unmarshal([]byte(tt.service), ud))
			cleanupServiceUnstructured(ud)

			expectedMap := &unstructured.Unstructured{}
			yaml.Unmarshal([]byte(tt.want), &expectedMap)
			if !reflect.DeepEqual(ud, expectedMap) {
				t.Errorf("cleanupServiceUnstructured(): " + cmp.Diff(ud, expectedMap))
			}
		})
	}
}
