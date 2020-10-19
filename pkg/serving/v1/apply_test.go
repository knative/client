package v1

import (
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util"
)

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

func TestApplyServiceWithNoImage(t *testing.T) {
	_, client := setup()
	serviceFaulty := newService("faulty-service")
	_, err := client.ApplyService(serviceFaulty)
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

	hasChanged, err := client.ApplyService(serviceNew)
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

	hasChanged, err := client.ApplyService(serviceNew)
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
