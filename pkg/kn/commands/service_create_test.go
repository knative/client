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
	"strings"
	"testing"

	servinglib "github.com/knative/client/pkg/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeServiceCreate(args []string) (
	action client_testing.Action,
	created *v1alpha1.Service,
	output string,
	err error) {

	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	cmd := NewKnCommand(KnParams{
		Output:         buf,
		ServingFactory: func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil },
	})
	fakeServing.AddReactor("*", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			createAction, ok := a.(client_testing.CreateAction)
			action = createAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", action)
			}
			created, ok = createAction.GetObject().(*v1alpha1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, created, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestServiceCreateImage(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz"})
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	conf, err := servinglib.GetConfiguration(created)
	if err != nil {
		t.Fatal(err)
	} else if conf.RevisionTemplate.Spec.Container.Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", conf.RevisionTemplate.Spec.Container.Image)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "created") ||
		!strings.Contains(output, "default") {
		t.Fatalf("wrong stdout message: %v", output)
	}
}

func TestServiceCreateEnv(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "-e", "A=DOGS", "--env", "B=WOLVES"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedEnvVars := map[string]string{
		"A": "DOGS",
		"B": "WOLVES"}

	conf, err := servinglib.GetConfiguration(created)
	actualEnvVars, err := servinglib.EnvToMap(conf.RevisionTemplate.Spec.Container.Env)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	} else if conf.RevisionTemplate.Spec.Container.Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", conf.RevisionTemplate.Spec.Container.Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars %v", conf.RevisionTemplate.Spec.Container.Env)
	}
}

func TestServiceCreateWithRequests(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--requests-cpu", "250m", "--requests-memory", "64Mi"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "250m"),
		corev1.ResourceMemory: parseQuantity(t, "64Mi"),
	}

	conf, err := servinglib.GetConfiguration(created)

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		conf.RevisionTemplate.Spec.Container.Resources.Requests,
		expectedRequestsVars) {
		t.Fatalf("wrong requests vars %v", conf.RevisionTemplate.Spec.Container.Resources.Requests)
	}
}

func TestServiceCreateWithLimits(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--limits-cpu", "1000m", "--limits-memory", "1024Mi"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "1000m"),
		corev1.ResourceMemory: parseQuantity(t, "1024Mi"),
	}

	conf, err := servinglib.GetConfiguration(created)

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		conf.RevisionTemplate.Spec.Container.Resources.Limits,
		expectedLimitsVars) {
		t.Fatalf("wrong limits vars %v", conf.RevisionTemplate.Spec.Container.Resources.Limits)
	}
}

func TestServiceCreateRequestsLimitsCPU(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--requests-cpu", "250m", "--limits-cpu", "1000m"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU: parseQuantity(t, "250m"),
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU: parseQuantity(t, "1000m"),
	}

	conf, err := servinglib.GetConfiguration(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", conf.RevisionTemplate.Spec.Container.Resources.Requests)
		}

		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", conf.RevisionTemplate.Spec.Container.Resources.Limits)
		}
	}
}

func TestServiceCreateRequestsLimitsMemory(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--requests-memory", "64Mi", "--limits-memory", "1024Mi"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceMemory: parseQuantity(t, "64Mi"),
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceMemory: parseQuantity(t, "1024Mi"),
	}

	conf, err := servinglib.GetConfiguration(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", conf.RevisionTemplate.Spec.Container.Resources.Requests)
		}

		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", conf.RevisionTemplate.Spec.Container.Resources.Limits)
		}
	}
}

func TestServiceCreateRequestsLimitsCPUMemory(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--requests-cpu", "250m", "--limits-cpu", "1000m",
		"--requests-memory", "64Mi", "--limits-memory", "1024Mi"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "250m"),
		corev1.ResourceMemory: parseQuantity(t, "64Mi"),
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "1000m"),
		corev1.ResourceMemory: parseQuantity(t, "1024Mi"),
	}

	conf, err := servinglib.GetConfiguration(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", conf.RevisionTemplate.Spec.Container.Resources.Requests)
		}

		if !reflect.DeepEqual(
			conf.RevisionTemplate.Spec.Container.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", conf.RevisionTemplate.Spec.Container.Resources.Limits)
		}
	}
}

func parseQuantity(t *testing.T, quantityString string) resource.Quantity {
	quantity, err := resource.ParseQuantity(quantityString)
	if err != nil {
		t.Fatal(err)
	}
	return quantity
}

func TestServiceCreateImageForce(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:v1"})
	if err != nil {
		t.Fatal(err)
	}
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--force", "--image", "gcr.io/foo/bar:v2"})
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	conf, err := servinglib.GetConfiguration(created)
	if err != nil {
		t.Fatal(err)
	} else if conf.RevisionTemplate.Spec.Container.Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", conf.RevisionTemplate.Spec.Container.Image)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "default") {
		t.Fatalf("wrong output: %s", output)
	}
}

func TestServiceCreateEnvForce(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:v1", "-e", "A=DOGS", "--env", "B=WOLVES"})
	if err != nil {
		t.Fatal(err)
	}
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--force", "--image", "gcr.io/foo/bar:v2", "-e", "A=CATS", "--env", "B=LIONS"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedEnvVars := map[string]string{
		"A": "CATS",
		"B": "LIONS"}

	conf, err := servinglib.GetConfiguration(created)
	actualEnvVars, err := servinglib.EnvToMap(conf.RevisionTemplate.Spec.Container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	} else if conf.RevisionTemplate.Spec.Container.Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", conf.RevisionTemplate.Spec.Container.Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars:%v", conf.RevisionTemplate.Spec.Container.Env)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "default") {
		t.Fatalf("wrong output: %s", output)
	}
}
