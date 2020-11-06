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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gotest.tools/assert"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
	network "knative.dev/networking/pkg"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func fakeServiceCreate(args []string, withExistingService bool) (
	action clienttesting.Action,
	service *servingv1.Service,
	output string,
	err error) {
	knParams := &commands.KnParams{}
	nrGetCalled := 0
	sync := !noWait(args)
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			nrGetCalled++
			if withExistingService || (sync && nrGetCalled > 1) {
				return true, &servingv1.Service{}, nil
			}
			return true, nil, api_errors.NewNotFound(schema.GroupResource{}, "")
		})
	fakeServing.AddReactor("create", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			createAction, ok := a.(clienttesting.CreateAction)
			action = createAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			service, ok = createAction.GetObject().(*servingv1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, service, nil
		})
	fakeServing.AddReactor("update", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			updateAction, ok := a.(clienttesting.UpdateAction)
			action = updateAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			service, ok = updateAction.GetObject().(*servingv1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, service, nil
		})
	if sync {
		fakeServing.AddWatchReactor("services",
			func(a clienttesting.Action) (bool, watch.Interface, error) {
				watchAction := a.(clienttesting.WatchAction)
				_, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
				if !found {
					return true, nil, errors.New("no field selector on metadata.name found")
				}
				w := wait.NewFakeWatch(getServiceEvents("test-service"))
				w.Start()
				return true, w, nil
			})
		fakeServing.AddReactor("get", "services",
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				return true, &servingv1.Service{}, nil
			})
	}
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		output = err.Error()
		return
	}
	output = buf.String()
	return
}

func getServiceEvents(name string) []watch.Event {
	return []watch.Event{
		{Type: watch.Added, Object: wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{Type: watch.Modified, Object: wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", "msg2")},
		{Type: watch.Modified, Object: wait.CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}
}

func TestServiceCreateImage(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--no-wait"}, false)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template := &created.Spec.Template
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.Containers[0].Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "created") ||
		!strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong stdout message: %v", output)
	}
}

func TestServiceCreateWithMultipleImages(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--image", "gcr.io/bar/foo:baz", "--no-wait"}, false)

	assert.Assert(t, util.ContainsAll(err.Error(), "\"--image\"", "\"gcr.io/bar/foo:baz\"", "flag", "once"))
}

func TestServiceCreateWithMultipleNames(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "foo1", "--image", "gcr.io/foo/bar:baz", "--no-wait"}, false)

	assert.Assert(t, util.ContainsAll(err.Error(), "'service create' requires the service name given as single argument"))
}

func TestServiceCreateImageSync(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz"}, false)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template := &created.Spec.Template
	if err != nil {
		t.Fatal(err)
	}
	if template.Spec.Containers[0].Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	}
	if !strings.Contains(output, "foo") || !strings.Contains(output, "Creating") ||
		!strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong stdout message: %v", output)
	}
	if !strings.Contains(output, "Ready") {
		t.Fatalf("not running in sync mode")
	}
}

func TestServiceCreateCommand(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--cmd", "/app/start", "--no-wait"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	template := &created.Spec.Template
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Command, []string{"/app/start"})
}

func TestServiceCreateArg(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--arg", "myArg1", "--arg", "--myArg2", "--arg", "--myArg3=3",
		"--no-wait"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	expectedArg := []string{"myArg1", "--myArg2", "--myArg3=3"}

	template := &created.Spec.Template
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Args, expectedArg)
}

func TestServiceCreateEnv(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"-e", "A=DOGS", "--env", "B=WOLVES", "--env=EMPTY", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedEnvVars := map[string]string{
		"A":     "DOGS",
		"B":     "WOLVES",
		"EMPTY": "",
	}

	template := &created.Spec.Template
	if err != nil {
		t.Fatal(err)
	}
	actualEnvVars, err := servinglib.EnvToMap(template.Spec.Containers[0].Env)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	} else if template.Spec.Containers[0].Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars %v", template.Spec.Containers[0].Env)
	}
}

func TestServiceCreateWithDeprecatedRequests(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--requests-cpu", "250m", "--requests-memory", "64Mi",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "250m"),
		corev1.ResourceMemory: parseQuantity(t, "64Mi"),
	}

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.Containers[0].Resources.Requests,
		expectedRequestsVars) {
		t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
	}
}

func TestServiceCreateWithRequests(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--request", "cpu=250m,memory=64Mi",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedRequestsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "250m"),
		corev1.ResourceMemory: parseQuantity(t, "64Mi"),
	}

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.Containers[0].Resources.Requests,
		expectedRequestsVars) {
		t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
	}
}

func TestServiceCreateWithDeprecatedLimits(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--limits-cpu", "1000m", "--limits-memory", "1024Mi",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "1000m"),
		corev1.ResourceMemory: parseQuantity(t, "1024Mi"),
	}

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.Containers[0].Resources.Limits,
		expectedLimitsVars) {
		t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
	}
}

func TestServiceCreateWithLimits(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--limit", "cpu=1000m", "--limit", "memory=1024Mi",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedLimitsVars := corev1.ResourceList{
		corev1.ResourceCPU:    parseQuantity(t, "1000m"),
		corev1.ResourceMemory: parseQuantity(t, "1024Mi"),
	}

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(
		template.Spec.Containers[0].Resources.Limits,
		expectedLimitsVars) {
		t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
	}
}

func TestServiceCreateDeprecatedRequestsLimitsCPU(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--requests-cpu", "250m", "--limits-cpu", "1000m",
		"--no-wait"}, false)

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

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceCreateRequestsLimitsCPU(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--request", "cpu=250m", "--limit", "cpu=1000m",
		"--no-wait"}, false)

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

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceCreateDeprecatedRequestsLimitsMemory(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo",
		"--image", "gcr.io/foo/bar:baz",
		"--requests-memory", "64Mi",
		"--limits-memory", "1024Mi", "--no-wait"}, false)

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

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceCreateRequestsLimitsMemory(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo",
		"--image", "gcr.io/foo/bar:baz",
		"--request", "memory=64Mi",
		"--limit", "memory=1024Mi", "--no-wait"}, false)

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

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceCreateMaxMinScale(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale-min", "1", "--scale-max", "5",
		"--concurrency-target", "10", "--concurrency-limit", "100",
		"--concurrency-utilization", "50",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	}
	if !action.Matches("create", "services") {
		t.Fatal("Bad action ", action)
	}

	template := &created.Spec.Template

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/minScale", "1",
		"autoscaling.knative.dev/maxScale", "5",
		"autoscaling.knative.dev/target", "10",
		"autoscaling.knative.dev/targetUtilizationPercentage", "50",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}

	if *template.Spec.ContainerConcurrency != int64(100) {
		t.Fatalf("container concurrency not set to given value 100")
	}
}

func TestServiceCreateScale(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "5", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	}
	if !action.Matches("create", "services") {
		t.Fatal("Bad action ", action)
	}

	template := &created.Spec.Template

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/minScale", "5",
		"autoscaling.knative.dev/maxScale", "5",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}
}

func TestServiceCreateScaleWithNegativeValue(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "-1", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	expectedErrMsg := "expected 0 <= -1 <= 2147483647: autoscaling.knative.dev/maxScale"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceCreateScaleWithMaxScaleSet(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "5", "--scale-max", "2", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	expectedErrMsg := "only --scale or --scale-max can be specified"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceCreateScaleWithMinScaleSet(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "5", "--scale-min", "2", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	expectedErrMsg := "only --scale or --scale-min can be specified"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceCreateScaleRange(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "1..5", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	}
	if !action.Matches("create", "services") {
		t.Fatal("Bad action ", action)
	}

	template := &created.Spec.Template

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/minScale", "1",
		"autoscaling.knative.dev/maxScale", "5",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}
}

func TestServiceCreateScaleRangeOnlyMin(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "1..", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	}
	if !action.Matches("create", "services") {
		t.Fatal("Bad action ", action)
	}

	template := &created.Spec.Template

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/minScale", "1",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}
}

func TestServiceCreateScaleRangeOnlyMax(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "..5", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	}
	if !action.Matches("create", "services") {
		t.Fatal("Bad action ", action)
	}

	template := &created.Spec.Template

	actualAnnos := template.Annotations
	expectedAnnos := []string{
		"autoscaling.knative.dev/maxScale", "5",
	}

	for i := 0; i < len(expectedAnnos); i += 2 {
		anno := expectedAnnos[i]
		if actualAnnos[anno] != expectedAnnos[i+1] {
			t.Fatalf("Unexpected annotation value for %s : %s (actual) != %s (expected)",
				anno, actualAnnos[anno], expectedAnnos[i+1])
		}
	}
}

func TestServiceCreateScaleRangeOnlyMinWrongSeparator(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "1--", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	expectedErrMsg := "Scale must be of the format x..y or x"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceCreateScaleRangeOnlyMaxWrongSeparator(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--scale", "--1", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	expectedErrMsg := "Scale must be of the format x..y or x"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceCreateRequestsLimitsCPUMemory(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--requests-cpu", "250m", "--limits-cpu", "1000m",
		"--requests-memory", "64Mi", "--limits-memory", "1024Mi",
		"--no-wait"}, false)

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

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", template.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			template.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", template.Spec.Containers[0].Resources.Limits)
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

func TestServiceCreateImageExistsAndNoForce(t *testing.T) {
	_, _, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:v2", "--no-wait"}, true)
	if err == nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "foo") ||
		!strings.Contains(output, commands.FakeNamespace) ||
		!strings.Contains(output, "create") ||
		!strings.Contains(output, "--force") {
		t.Errorf("Invalid error output: '%s'", output)
	}

}

func TestServiceCreateImageForce(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--force", "--image", "gcr.io/foo/bar:v2", "--no-wait"}, true)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template := &created.Spec.Template
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.Containers[0].Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong output: %s", output)
	}
}

func TestServiceCreateEnvForce(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:v1",
		"-e", "A=DOGS", "--env", "B=WOLVES", "--no-wait"}, false)
	if err != nil {
		t.Fatal(err)
	}
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--force", "--image", "gcr.io/foo/bar:v2",
		"-e", "A=CATS", "--env", "B=LIONS", "--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expectedEnvVars := map[string]string{
		"A": "CATS",
		"B": "LIONS"}

	template := &created.Spec.Template
	if err != nil {
		t.Fatal(err)
	}
	actualEnvVars, err := servinglib.EnvToMap(template.Spec.Containers[0].Env)
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.Containers[0].Image != "gcr.io/foo/bar:v2" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	} else if !reflect.DeepEqual(
		actualEnvVars,
		expectedEnvVars) {
		t.Fatalf("wrong env vars:%v", template.Spec.Containers[0].Env)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong output: %s", output)
	}
}

func TestServiceCreateWithServiceAccountName(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--service-account", "foo-bar-account",
		"--no-wait"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template := &created.Spec.Template

	if err != nil {
		t.Fatal(err)
	} else if template.Spec.ServiceAccountName != "foo-bar-account" {
		t.Fatalf("wrong service account name:%v", template.Spec.ServiceAccountName)
	}
}

func TestServiceCreateWithClusterLocal(t *testing.T) {
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz",
		"--cluster-local"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}

	labels := created.ObjectMeta.Labels

	labelValue, present := labels[network.VisibilityLabelKey]
	assert.Assert(t, present)

	if labelValue != serving.VisibilityClusterLocal {
		t.Fatalf("Incorrect VisibilityClusterLocal value '%s'", labelValue)
	}
}

var serviceYAML = `
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: foo
spec:
  template:
    spec:
      containers:
        - image: gcr.io/foo/bar:baz
          env:
            - name: TARGET
              value: "Go Sample v1"
`

var serviceJSON = `
{
  "apiVersion": "serving.knative.dev/v1",
  "kind": "Service",
  "metadata": {
    "name": "foo"
  },
  "spec": {
    "template": {
		"spec": {
		"containers": [
			{
			"image": "gcr.io/foo/bar:baz",
			"env": [
				{
				"name": "TARGET",
				"value": "Go Sample v1"
				}
			]
		  }
		]
	  }
	}
  }
}`

func TestServiceCreateFromYAML(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.yaml")
	err = ioutil.WriteFile(tempFile, []byte(serviceYAML), os.FileMode(0666))
	assert.NilError(t, err)

	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	assert.Equal(t, created.Name, "foo")
	assert.Equal(t, created.Spec.Template.Spec.GetContainer().Image, "gcr.io/foo/bar:baz")
}

func TestServiceCreateFromJSON(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.json")
	err = ioutil.WriteFile(tempFile, []byte(serviceJSON), os.FileMode(0666))
	assert.NilError(t, err)

	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	assert.Equal(t, created.Name, "foo")
	assert.Equal(t, created.Spec.Template.Spec.GetContainer().Image, "gcr.io/foo/bar:baz")
}

func TestServiceCreateFromFileWithName(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.yaml")
	err = ioutil.WriteFile(tempFile, []byte(serviceYAML), os.FileMode(0666))
	assert.NilError(t, err)

	t.Log("no NAME param provided")
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "--filename", tempFile}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	assert.Equal(t, created.Name, "foo")
	assert.Equal(t, created.Spec.Template.Spec.GetContainer().Image, "gcr.io/foo/bar:baz")

	t.Log("no service.Name provided in file")
	err = ioutil.WriteFile(tempFile, []byte(strings.ReplaceAll(serviceYAML, "name: foo", "")), os.FileMode(0666))
	assert.NilError(t, err)
	action, created, _, err = fakeServiceCreate([]string{
		"service", "create", "cli-foo", "--filename", tempFile}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))

	assert.Equal(t, created.Name, "cli-foo")
	assert.Equal(t, created.Spec.Template.Spec.GetContainer().Image, "gcr.io/foo/bar:baz")
}

func TestServiceCreateFileNameMismatch(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.json")
	err = ioutil.WriteFile(tempFile, []byte(serviceJSON), os.FileMode(0666))
	assert.NilError(t, err)

	t.Log("NAME param nad service.Name differ")
	_, _, _, err = fakeServiceCreate([]string{
		"service", "create", "anotherFoo", "--filename", tempFile}, false)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAllIgnoreCase(err.Error(), "provided", "'anotherFoo'", "name", "match", "from", "file", "'foo'"))

	t.Log("no NAME param & no service.Name provided in file")
	err = ioutil.WriteFile(tempFile, []byte(strings.ReplaceAll(serviceYAML, "name: foo", "")), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{
		"service", "create", "--filename", tempFile}, false)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAllIgnoreCase(err.Error(), "no", "service", "name", "provided", "parameter", "file"))
}

func TestServiceCreateFileError(t *testing.T) {
	_, _, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", "filepath"}, false)

	assert.Assert(t, util.ContainsAll(err.Error(), "no", "such", "file", "directory", "filepath"))
}

func TestServiceCreateInvalidDataJSON(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)
	tempFile := filepath.Join(tempDir, "invalid.json")

	// Double curly bracket at the beginning of file
	invalidData := strings.Replace(serviceJSON, "{\n", "{{\n", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "character", "'{'", "beginning"))

	// Remove closing quote on key
	invalidData = strings.Replace(serviceJSON, "metadata\"", "metadata", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "character", "'\\n'", "string", "literal"))

	// Remove opening square bracket
	invalidData = strings.Replace(serviceJSON, " [", "", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "character", "']'", "after", "key:value"))
}

func TestServiceCreateInvalidDataYAML(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)
	tempFile := filepath.Join(tempDir, "invalid.yaml")

	// Remove dash
	invalidData := strings.Replace(serviceYAML, "- image", "image", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "mapping", "values", "not", "allowed"))

	// Remove name key
	invalidData = strings.Replace(serviceYAML, "name:", "", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "cannot", "unmarshal", "Go", "struct", "Service.metadata"))

	// Remove opening square bracket
	invalidData = strings.Replace(serviceYAML, "env", "\tenv", 1)
	err = ioutil.WriteFile(tempFile, []byte(invalidData), os.FileMode(0666))
	assert.NilError(t, err)
	_, _, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile}, false)
	assert.Assert(t, util.ContainsAll(err.Error(), "found", "tab", "violates", "indentation"))
}

func TestServiceCreateFromYAMLWithOverride(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.yaml")
	err = ioutil.WriteFile(tempFile, []byte(serviceYAML), os.FileMode(0666))
	assert.NilError(t, err)
	// Merge env vars
	expectedEnvVars := map[string]string{
		"TARGET": "Go Sample v1",
		"FOO":    "BAR"}
	action, created, _, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile, "--env", "FOO=BAR"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))
	assert.Equal(t, created.Name, "foo")

	actualEnvVar, err := servinglib.EnvToMap(created.Spec.Template.Spec.GetContainer().Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, actualEnvVar, expectedEnvVars)

	// Override env vars
	expectedEnvVars = map[string]string{
		"TARGET": "FOOBAR",
		"FOO":    "BAR"}
	action, created, _, err = fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile, "--env", "TARGET=FOOBAR", "--env", "FOO=BAR"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))
	assert.Equal(t, created.Name, "foo")

	actualEnvVar, err = servinglib.EnvToMap(created.Spec.Template.Spec.GetContainer().Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, actualEnvVar, expectedEnvVars)

	// Remove existing env vars
	expectedEnvVars = map[string]string{
		"FOO": "BAR"}
	action, created, _, err = fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile, "--env", "TARGET-", "--env", "FOO=BAR"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))
	assert.Equal(t, created.Name, "foo")

	actualEnvVar, err = servinglib.EnvToMap(created.Spec.Template.Spec.GetContainer().Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, actualEnvVar, expectedEnvVars)

	// Multiple edit flags
	expectedAnnotations := map[string]string{
		"foo": "bar"}
	action, created, _, err = fakeServiceCreate([]string{"service", "create", "foo", "--filename", tempFile,
		"--service-account", "foo", "--cmd", "/foo/bar", "-a", "foo=bar"}, false)
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("create", "services"))
	assert.Equal(t, created.Name, "foo")
	assert.DeepEqual(t, created.Spec.Template.Spec.GetContainer().Command, []string{"/foo/bar"})
	assert.Equal(t, created.Spec.Template.Spec.ServiceAccountName, "foo")
	assert.DeepEqual(t, created.ObjectMeta.Annotations, expectedAnnotations)
}

func TestServiceCreateFromYAMLWithOverrideError(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-file")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "service.yaml")
	err = ioutil.WriteFile(tempFile, []byte(serviceYAML), os.FileMode(0666))
	assert.NilError(t, err)

	_, _, _, err = fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile, "--image", "foo/bar", "--image", "bar/foo"}, false)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "\"--image\"", "invalid", "argument", "only", "once"))

	_, _, _, err = fakeServiceCreate([]string{
		"service", "create", "foo", "--filename", tempFile, "--scale", "-1"}, false)
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "expected 0 <= -1 <= 2147483647: autoscaling.knative.dev/maxScale"))
}
