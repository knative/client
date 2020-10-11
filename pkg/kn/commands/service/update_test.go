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
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
	network "knative.dev/networking/pkg"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

var exampleImageByDigest = "gcr.io/foo/bar@sha256:deadbeefdeadbeef"
var exampleRevisionName = "foo-asdf"
var exampleRevisionName2 = "foo-xyzzy"

func fakeServiceUpdate(original *servingv1.Service, args []string) (
	action clienttesting.Action,
	updated *servingv1.Service,
	output string,
	err error) {
	var reconciled servingv1.Service
	knParams := &commands.KnParams{}
	sync := !noWait(args)
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("update", "*",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			updateAction, ok := a.(clienttesting.UpdateAction)
			action = updateAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", action)
			}
			updated, ok = updateAction.GetObject().(*servingv1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, updated, nil
		})
	fakeServing.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if updated == nil {
				original.Status.LatestCreatedRevisionName = exampleRevisionName
				return true, original, nil
			}
			reconciled = *updated
			if updated.Spec.Template.Name == "" {
				reconciled.Status.LatestCreatedRevisionName = exampleRevisionName2
			} else {
				reconciled.Status.LatestCreatedRevisionName = updated.Spec.Template.Name
			}

			return true, &reconciled, nil
		})
	fakeServing.AddReactor("get", "revisions",
		// This is important for the way we set images to their image digest
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			rev := &servingv1.Revision{}
			rev.Spec = original.Spec.Template.Spec
			rev.ObjectMeta = original.Spec.Template.ObjectMeta
			rev.Name = original.Status.LatestCreatedRevisionName
			rev.Status.DeprecatedImageDigest = exampleImageByDigest
			return true, rev, nil
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
		return
	}
	output = buf.String()
	return
}

func TestServiceUpdateNoFlags(t *testing.T) {
	orig := newEmptyService()

	action, _, _, err := fakeServiceUpdate(orig, []string{"service", "update", "foo"})
	if action != nil {
		t.Errorf("Unexpected action if no flag(s) set")
	}

	if err == nil {
		t.Fatal(err)
	}

	expectedErrMsg := "flag(s) not set"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Fatalf("Missing %s in %s", expectedErrMsg, err.Error())
	}
}

func TestServiceUpdateImageSync(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	err := servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")
	if err != nil {
		t.Fatal(err)
	}

	action, updated, output, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar"})

	assert.NilError(t, err)
	assert.Assert(t, action.Matches("update", "services"))

	template = &updated.Spec.Template
	assert.NilError(t, err)

	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/quux:xyzzy")
	assert.Assert(t, util.ContainsAll(strings.ToLower(output), "updating", "foo", "service", "namespace", "bar", "ready"))
}

func TestServiceUpdateImage(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	err := servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")
	if err != nil {
		t.Fatal(err)
	}

	action, updated, output, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = &updated.Spec.Template
	container, err := servinglib.ContainerOfRevisionTemplate(template)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, container.Image, "gcr.io/foo/quux:xyzzy")

	if !strings.Contains(strings.ToLower(output), "update") ||
		!strings.Contains(output, "foo") ||
		!strings.Contains(strings.ToLower(output), "service") ||
		!strings.Contains(strings.ToLower(output), "namespace") ||
		!strings.Contains(output, "bar") {
		t.Fatalf("wrong or no success message: %s", output)
	}
}

func TestServiceUpdateWithMultipleImages(t *testing.T) {
	orig := newEmptyService()
	_, _, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/bar:baz", "--image", "gcr.io/bar/foo:baz", "--no-wait"})

	assert.Assert(t, util.ContainsAll(err.Error(), "\"--image\"", "\"gcr.io/bar/foo:baz\"", "flag", "once"))
}

func TestServiceUpdateWithMultipleNames(t *testing.T) {
	orig := newEmptyService()
	_, _, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "foo1", "--image", "gcr.io/foo/bar:baz", "--no-wait"})

	assert.Assert(t, util.ContainsAll(err.Error(), "'service update' requires the service name given as single argument"))
}

func TestServiceUpdateCommand(t *testing.T) {
	orig := newEmptyService()

	origTemplate := &orig.Spec.Template

	err := servinglib.UpdateContainerCommand(origTemplate, "./start")
	assert.NilError(t, err)

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--cmd", "/app/start", "--no-wait"})
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("update", "services"))

	updatedTemplate := updated.Spec.Template
	assert.DeepEqual(t, updatedTemplate.Spec.Containers[0].Command, []string{"/app/start"})
}

func TestServiceUpdateArg(t *testing.T) {
	orig := newEmptyService()

	origTemplate := orig.Spec.Template

	err := servinglib.UpdateContainerArg(&origTemplate, []string{"myArg0"})
	assert.NilError(t, err)

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--arg", "myArg1", "--arg", "--myArg2", "--arg", "--myArg3=3", "--no-wait"})
	assert.NilError(t, err)
	assert.Assert(t, action.Matches("update", "services"))

	updatedTemplate := updated.Spec.Template
	assert.DeepEqual(t, updatedTemplate.Spec.Containers[0].Args, []string{"myArg1", "--myArg2", "--myArg3=3"})
}

func TestServiceUpdateRevisionNameExplicit(t *testing.T) {
	orig := newEmptyService()

	template := orig.Spec.Template
	template.Name = "foo-asdf"

	// Test user provides prefix
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--revision-name", "foo-dogs", "--namespace", "bar", "--no-wait"})
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template = updated.Spec.Template
	assert.Equal(t, "foo-dogs", template.Name)

}

func TestServiceUpdateRevisionNameGenerated(t *testing.T) {
	orig := newEmptyService()

	template := orig.Spec.Template
	template.Name = "foo-asdf"

	// Test prefix added by command
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--no-wait"})
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = updated.Spec.Template
	assert.Assert(t, strings.HasPrefix(template.Name, "foo-"))
	assert.Assert(t, !(template.Name == "foo-asdf"))
}

func TestServiceUpdateRevisionNameCleared(t *testing.T) {
	orig := newEmptyService()

	template := orig.Spec.Template
	template.Name = "foo-asdf"

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--revision-name=", "--no-wait"})

	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = updated.Spec.Template
	assert.Assert(t, cmp.Equal(template.Name, ""))
}

func TestServiceUpdateRevisionNameNoMutationNoChange(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	template.Name = "foo-asdf"

	// Test prefix added by command
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--namespace", "bar", "--no-wait"})
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = &updated.Spec.Template
	assert.Equal(t, template.Name, "foo-asdf")
}

func TestServiceUpdateMaxMinScale(t *testing.T) {
	original := newEmptyService()

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--scale-min", "1", "--scale-max", "5", "--concurrency-target", "10", "--concurrency-limit", "100", "--concurrency-utilization", "50", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	}

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

func TestServiceUpdateScale(t *testing.T) {
	original := newEmptyService()

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--scale", "5", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	}

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

func TestServiceUpdateScaleWithNegativeValue(t *testing.T) {
	original := newEmptyService()

	_, _, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--scale", "-1", "--no-wait"})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErrMsg := "expected 0 <= -1 <= 2147483647: autoscaling.knative.dev/maxScale"

	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceUpdateScaleWithMaxScaleSet(t *testing.T) {
	original := newEmptyService()

	_, _, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--scale", "5", "--scale-max", "2", "--no-wait"})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErrMsg := "only --scale or --scale-max can be specified"

	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}

func TestServiceUpdateScaleWithMinScaleSet(t *testing.T) {
	original := newEmptyService()

	_, _, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--scale", "5", "--scale-min", "2", "--no-wait"})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErrMsg := "only --scale or --scale-min can be specified"

	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Invalid error output, expected: %s, got : '%s'", expectedErrMsg, err)
	}

}
func TestServiceUpdateEnv(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	template.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "EXISTING", Value: "thing"},
		{Name: "OTHEREXISTING"},
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--env", "EXISTING-", "--env=OTHEREXISTING-=whatever", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	expectedEnvVar := corev1.EnvVar{
		Name:  "TARGET",
		Value: "Awesome",
	}

	template = &updated.Spec.Template
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
	assert.Equal(t, template.Spec.Containers[0].Env[0], expectedEnvVar)
}

func TestServiceUpdatePinsToDigestWhenAsked(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	delete(template.Annotations, servinglib.UserImageAnnotationKey)
	err := servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")
	if err != nil {
		t.Fatal(err)
	}

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--lock-to-digest", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = &updated.Spec.Template
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
}

func TestServiceUpdatePinsToDigestWhenPreviouslyDidSo(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	err := servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")
	if err != nil {
		t.Fatal(err)
	}

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = &updated.Spec.Template
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
}

func TestServiceUpdateDoesntPinToDigestWhenUnAsked(t *testing.T) {
	orig := newEmptyService()

	template := orig.Spec.Template
	err := servinglib.UpdateImage(&template, "gcr.io/foo/bar:baz")
	assert.NilError(t, err)

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--no-lock-to-digest", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = updated.Spec.Template
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/bar:baz")
	_, present := template.Annotations[servinglib.UserImageAnnotationKey]
	assert.Assert(t, !present)
}
func TestServiceUpdateDoesntPinToDigestWhenPreviouslyDidnt(t *testing.T) {
	orig := newEmptyService()

	template := &orig.Spec.Template
	delete(template.Annotations, servinglib.UserImageAnnotationKey)

	err := servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")
	if err != nil {
		t.Fatal(err)
	}

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template = &updated.Spec.Template
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/bar:baz")
	_, present := template.Annotations[servinglib.UserImageAnnotationKey]
	assert.Assert(t, !present)
}

func TestServiceUpdateRequestsLimitsCPU(t *testing.T) {
	service := createMockServiceWithResources(t, "250", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo", "--requests-cpu", "500m", "--limits-cpu", "1000m", "--no-wait"})
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

	newTemplate := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newTemplate.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newTemplate.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceUpdateRequestsLimitsMemory(t *testing.T) {
	service := createMockServiceWithResources(t, "100m", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo", "--requests-memory", "128Mi", "--limits-memory", "2048Mi", "--no-wait"})
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

	newTemplate := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newTemplate.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newTemplate.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceUpdateRequestsLimitsCPU_and_Memory(t *testing.T) {
	service := createMockServiceWithResources(t, "250m", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo",
		"--requests-cpu", "500m", "--limits-cpu", "2000m",
		"--requests-memory", "128Mi", "--limits-memory", "2048Mi", "--no-wait"})
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

	newTemplate := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Requests,
			expectedRequestsVars) {
			t.Fatalf("wrong requests vars %v", newTemplate.Spec.Containers[0].Resources.Requests)
		}

		if !reflect.DeepEqual(
			newTemplate.Spec.Containers[0].Resources.Limits,
			expectedLimitsVars) {
			t.Fatalf("wrong limits vars %v", newTemplate.Spec.Containers[0].Resources.Limits)
		}
	}
}

func TestServiceUpdateLabelWhenEmpty(t *testing.T) {
	original := newEmptyService()
	origTemplate := original.Spec.Template
	origContainer, err := servinglib.ContainerOfRevisionTemplate(&origTemplate)
	if err != nil {
		t.Fatal(err)
	}
	origContainer.Image = "gcr.io/foo/bar:latest"

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "-l", "a=mouse", "--label", "b=cookie", "-l=single", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expected := map[string]string{
		"a":      "mouse",
		"b":      "cookie",
		"single": "",
	}
	actual := updated.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	template := updated.Spec.Template
	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
	container, err := servinglib.ContainerOfRevisionTemplate(&template)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, container.Image, exampleImageByDigest)
}

func TestServiceUpdateLabelExisting(t *testing.T) {
	original := newEmptyService()
	original.ObjectMeta.Labels = map[string]string{"already": "here", "tobe": "removed"}
	originalTemplate := original.Spec.Template
	originalTemplate.ObjectMeta.Labels = map[string]string{"already": "here", "tobe": "removed"}

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "-l", "already=gone", "--label=tobe-", "--label", "b=", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expected := map[string]string{
		"already": "gone",
		"b":       "",
	}
	actual := updated.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	template := updated.Spec.Template
	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

func TestServiceUpdateNoClusterLocal(t *testing.T) {
	original := newEmptyService()
	original.ObjectMeta.Labels = map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}
	originalTemplate := &original.Spec.Template
	originalTemplate.ObjectMeta.Labels = map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "--no-cluster-local", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expected := map[string]string{}
	actual := updated.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

//TODO: add check for template name not changing when issue #646 solution is merged
func TestServiceUpdateNoClusterLocalOnPublicService(t *testing.T) {
	original := newEmptyService()
	original.ObjectMeta.Labels = map[string]string{}

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "--no-cluster-local", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expected := map[string]string{}
	actual := updated.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

//TODO: add check for template name not changing when issue #646 solution is merged
func TestServiceUpdateNoClusterLocalOnPrivateService(t *testing.T) {
	original := newEmptyService()
	original.ObjectMeta.Labels = map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}
	originalTemplate := &original.Spec.Template
	originalTemplate.ObjectMeta.Labels = map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "--cluster-local", "--no-wait"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	expected := map[string]string{network.VisibilityLabelKey: serving.VisibilityClusterLocal}
	actual := updated.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	newTemplate := updated.Spec.Template
	if err != nil {
		t.Fatal(err)
	}

	actual = newTemplate.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

func TestServiceUpdateDeletionTimestampNotNil(t *testing.T) {
	original := newEmptyService()
	original.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	_, _, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "--revision-name", "foo-v1"})
	assert.ErrorContains(t, err, original.Name)
	assert.ErrorContains(t, err, "deletion")
	assert.ErrorContains(t, err, "service")
}

func TestServiceUpdateTagDoesNotExist(t *testing.T) {
	orig := newEmptyService()

	_, _, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--untag", "foo", "--no-wait"})

	assert.Assert(t, util.ContainsAll(err.Error(), "tag(s)", "foo", "not present", "service", "foo"))
}

func newEmptyService() *servingv1.Service {
	ret := &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: servingv1.ServiceSpec{},
	}
	ret.Spec.Template = servingv1.RevisionTemplateSpec{}
	ret.Spec.Template.Annotations = map[string]string{
		servinglib.UserImageAnnotationKey: "",
	}
	ret.Spec.Template.Spec.Containers = []corev1.Container{{}}
	return ret
}

func createMockServiceWithResources(t *testing.T, requestCPU, requestMemory, limitsCPU, limitsMemory string) *servingv1.Service {
	service := newEmptyService()

	template := service.Spec.Template

	template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
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

func noWait(args []string) bool {
	for _, arg := range args {
		if arg == "--no-wait" {
			return true
		}
	}
	return false
}
