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

package service

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	servinglib "github.com/knative/client/pkg/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
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
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
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
	template, err := servinglib.GetRevisionTemplate(created)
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.DeprecatedContainer.Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.DeprecatedContainer.Image)
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

	template, err := servinglib.GetRevisionTemplate(created)
	actualEnvVars, err := servinglib.EnvToMap(template.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	} else if template.Spec.DeprecatedContainer.Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.DeprecatedContainer.Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars %v", template.Spec.DeprecatedContainer.Env)
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

	template, err := servinglib.GetRevisionTemplate(created)

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.DeprecatedContainer.Resources.Requests,
		expectedRequestsVars) {
		t.Fatalf("wrong requests vars %v", template.Spec.DeprecatedContainer.Resources.Requests)
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

	template, err := servinglib.GetRevisionTemplate(created)

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.DeprecatedContainer.Resources.Limits,
		expectedLimitsVars) {
		t.Fatalf("wrong limits vars %v", template.Spec.DeprecatedContainer.Resources.Limits)
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

	template, err := servinglib.GetRevisionTemplate(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.DeprecatedContainer.Resources.Limits)
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

	template, err := servinglib.GetRevisionTemplate(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.DeprecatedContainer.Resources.Limits)
		}
	}
}

func TestServiceCreateMaxMinScale(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--min-scale", "1", "--max-scale", "5", "--concurrency-target", "10", "--concurrency-limit", "100"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err := servinglib.GetRevisionTemplate(created)
	if err != nil {
		t.Fatal(err)
	}

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/minScale", "1",
		"autoscaling.knative.dev/maxScale", "5",
		"autoscaling.knative.dev/target", "10",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}

	if template.Spec.ContainerConcurrency != 100 {
		t.Fatalf("container concurrency not set to given value 1000")
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

	template, err := servinglib.GetRevisionTemplate(created)

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.DeprecatedContainer.Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.DeprecatedContainer.Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.DeprecatedContainer.Resources.Limits)
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
	template, err := servinglib.GetRevisionTemplate(created)
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.DeprecatedContainer.Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", template.Spec.DeprecatedContainer.Image)
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

	template, err := servinglib.GetRevisionTemplate(created)
	actualEnvVars, err := servinglib.EnvToMap(template.Spec.DeprecatedContainer.Env)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.DeprecatedContainer.Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", template.Spec.DeprecatedContainer.Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars:%v", template.Spec.DeprecatedContainer.Env)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "default") {
		t.Fatalf("wrong output: %s", output)
	}
}
