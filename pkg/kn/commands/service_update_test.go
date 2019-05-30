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

package commands

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	servinglib "github.com/knative/client/pkg/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeServiceUpdate(original *v1alpha1.Service, args []string) (
	action client_testing.Action,
	updated *v1alpha1.Service,
	output string,
	err error) {

	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	cmd := NewKnCommand(KnParams{
		Output:         buf,
		ServingFactory: func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil },
	})
	fakeServing.AddReactor("update", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			updateAction, ok := a.(client_testing.UpdateAction)
			action = updateAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", action)
			}
			updated, ok = updateAction.GetObject().(*v1alpha1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, updated, nil
		})
	fakeServing.AddReactor("get", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			return true, original, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestServiceUpdateImage(t *testing.T) {
	orig := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			DeprecatedRunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					DeprecatedRevisionTemplate: &v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							DeprecatedContainer: &corev1.Container{},
						},
					},
				},
			},
		},
	}

	config, err := servinglib.GetConfiguration(orig)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(config, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	conf, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else if conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Image != "gcr.io/foo/quux:xyzzy" {
		t.Fatalf("wrong image set: %v", conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Image)
	}
}

func TestServiceUpdateEnv(t *testing.T) {
	orig := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			DeprecatedRunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					DeprecatedRevisionTemplate: &v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							DeprecatedContainer: &corev1.Container{},
						},
					},
				},
			},
		},
	}

	config, err := servinglib.GetConfiguration(orig)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(config, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	expectedEnvVar := corev1.EnvVar{
		Name:  "TARGET",
		Value: "Awesome",
	}

	conf, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else if conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Image)
	} else if conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env[0] != expectedEnvVar {
		t.Fatalf("wrong env set: %v", conf.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Env)
	}
}

func TestServiceUpdateRequestsLimitsCPU(t *testing.T) {
	service := createMockServiceWithResources(t, "250", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo", "--requests-cpu", "500m", "--limits-cpu", "1000m"})
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("64Mi"),
	}
	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1000m"),
		corev1.ResourceMemory: resource.MustParse("1024Mi"),
	}

	newConfig, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits)
		}
	}
}

func TestServiceUpdateRequestsLimitsMemory(t *testing.T) {
	service := createMockServiceWithResources(t, "100m", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo", "--requests-memory", "128Mi", "--limits-memory", "2048Mi"})
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("128Mi"),
	}
	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("1000m"),
		corev1.ResourceMemory: resource.MustParse("2048Mi"),
	}

	newConfig, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits)
		}
	}
}

func TestServiceUpdateRequestsLimitsCPU_and_Memory(t *testing.T) {
	service := createMockServiceWithResources(t, "250m", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo",
		"--requests-cpu", "500m", "--limits-cpu", "2000m",
		"--requests-memory", "128Mi", "--limits-memory", "2048Mi"})
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("500m"),
		corev1.ResourceMemory: resource.MustParse("128Mi"),
	}
	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("2000m"),
		corev1.ResourceMemory: resource.MustParse("2048Mi"),
	}

	newConfig, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newConfig.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources.Limits)
		}
	}
}

func createMockServiceWithResources(t *testing.T, requestCPU, requestMemory, limitsCPU, limitsMemory string) *v1alpha1.Service {
	service := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			DeprecatedRunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					DeprecatedRevisionTemplate: &v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							DeprecatedContainer: &corev1.Container{},
						},
					},
				},
			},
		},
	}

	config, err := servinglib.GetConfiguration(service)
	if err != nil {
		t.Fatal(err)
	}

	config.DeprecatedRevisionTemplate.Spec.DeprecatedContainer.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(requestCPU),
			corev1.ResourceMemory: resource.MustParse(requestMemory),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(limitsCPU),
			corev1.ResourceMemory: resource.MustParse(limitsMemory),
		},
	}

	return service
}
