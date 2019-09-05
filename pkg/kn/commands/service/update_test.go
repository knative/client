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

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"knative.dev/client/pkg/kn/commands"
	servinglib "knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

var exampleImageByDigest = "gcr.io/foo/bar@sha256:deadbeefdeadbeef"
var exampleRevisionName = "foo-asdf"
var exampleRevisionName2 = "foo-xyzzy"

func fakeServiceUpdate(original *v1alpha1.Service, args []string, sync bool) (
	action client_testing.Action,
	updated *v1alpha1.Service,
	output string,
	err error) {
	var reconciled v1alpha1.Service
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
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
	fakeServing.AddReactor("get", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
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
		func(a client_testing.Action) (bool, runtime.Object, error) {
			rev := &v1alpha1.Revision{}
			rev.Spec = original.Spec.Template.Spec
			rev.ObjectMeta = original.Spec.Template.ObjectMeta
			rev.Name = original.Status.LatestCreatedRevisionName
			rev.Status.ImageDigest = exampleImageByDigest
			return true, rev, nil
		})
	if sync {
		fakeServing.AddWatchReactor("services",
			func(a client_testing.Action) (bool, watch.Interface, error) {
				watchAction := a.(client_testing.WatchAction)
				_, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
				if !found {
					return true, nil, errors.New("no field selector on metadata.name found")
				}
				w := wait.NewFakeWatch(getServiceEvents("test-service"))
				w.Start()
				return true, w, nil
			})
		fakeServing.AddReactor("get", "services",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				return true, &v1alpha1.Service{}, nil
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

func TestServcieUpdateNoFlags(t *testing.T) {
	orig := newEmptyService()

	action, _, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo"}, false)

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

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, output, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar"}, true)

	assert.NilError(t, err)
	assert.Assert(t, action.Matches("update", "services"))

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)

	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/quux:xyzzy")
	assert.Assert(t, util.ContainsAll(strings.ToLower(output), "update", "foo", "service", "namespace", "bar", "ok", "waiting"))
}

func TestServiceUpdateImage(t *testing.T) {
	for _, orig := range []*v1alpha1.Service{newEmptyServiceBetaAPIStyle(), newEmptyServiceAlphaAPIStyle()} {

		template, err := servinglib.RevisionTemplateOfService(orig)
		if err != nil {
			t.Fatal(err)
		}

		servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

		action, updated, output, err := fakeServiceUpdate(orig, []string{
			"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--async"}, false)

		if err != nil {
			t.Fatal(err)
		} else if !action.Matches("update", "services") {
			t.Fatalf("Bad action %v", action)
		}

		template, err = servinglib.RevisionTemplateOfService(updated)
		if err != nil {
			t.Fatal(err)
		}
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
}

func TestServiceUpdateRevisionNameExplicit(t *testing.T) {
	orig := newEmptyServiceBetaAPIStyle()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	template.Name = "foo-asdf"

	// Test user provides prefix
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--revision-name", "foo-dogs", "--namespace", "bar", "--async"}, false)
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	assert.Equal(t, "foo-dogs", template.Name)

}

func TestServiceUpdateRevisionNameGenerated(t *testing.T) {
	orig := newEmptyServiceBetaAPIStyle()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	template.Name = "foo-asdf"

	// Test prefix added by command
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--async"}, false)
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	assert.Assert(t, strings.HasPrefix(template.Name, "foo-"))
	assert.Assert(t, !(template.Name == "foo-asdf"))
}

func TestServiceUpdateRevisionNameCleared(t *testing.T) {
	orig := newEmptyServiceBetaAPIStyle()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}
	template.Name = "foo-asdf"

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy", "--namespace", "bar", "--revision-name=", "--async"}, false)

	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	assert.Assert(t, cmp.Equal(template.Name, ""))
}

func TestServiceUpdateRevisionNameNoMutationNoChange(t *testing.T) {
	orig := newEmptyServiceBetaAPIStyle()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	template.Name = "foo-asdf"

	// Test prefix added by command
	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--namespace", "bar", "--async"}, false)
	assert.NilError(t, err)
	if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	assert.Equal(t, template.Name, "foo-asdf")
}

func TestServiceUpdateMaxMinScale(t *testing.T) {
	original := newEmptyService()

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo",
		"--min-scale", "1", "--max-scale", "5", "--concurrency-target", "10", "--concurrency-limit", "100", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err := servinglib.RevisionTemplateOfService(updated)
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

func TestServiceUpdateEnv(t *testing.T) {
	orig := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}
	template.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "EXISTING", Value: "thing"},
		{Name: "OTHEREXISTING"},
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--env", "EXISTING-", "--env=OTHEREXISTING-=whatever", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	expectedEnvVar := corev1.EnvVar{
		Name:  "TARGET",
		Value: "Awesome",
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
	assert.Equal(t, template.Spec.Containers[0].Env[0], expectedEnvVar)
}

func TestServiceUpdatePinsToDigestWhenAsked(t *testing.T) {
	orig := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(orig)
	delete(template.Annotations, servinglib.UserImageAnnotationKey)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--lock-to-digest", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
}

func TestServiceUpdatePinsToDigestWhenPreviouslyDidSo(t *testing.T) {
	orig := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, exampleImageByDigest)
}

func TestServiceUpdateDoesntPinToDigestWhenUnAsked(t *testing.T) {
	orig := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(orig)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--no-lock-to-digest", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/bar:baz")
	_, present := template.Annotations[servinglib.UserImageAnnotationKey]
	assert.Assert(t, !present)
}
func TestServiceUpdateDoesntPinToDigestWhenPreviouslyDidnt(t *testing.T) {
	orig := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(orig)
	delete(template.Annotations, servinglib.UserImageAnnotationKey)
	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(template, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "-e", "TARGET=Awesome", "--async"}, false)

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}

	template, err = servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	// Test that we pinned to digest
	assert.Equal(t, template.Spec.Containers[0].Image, "gcr.io/foo/bar:baz")
	_, present := template.Annotations[servinglib.UserImageAnnotationKey]
	assert.Assert(t, !present)
}

func TestServiceUpdateRequestsLimitsCPU(t *testing.T) {
	service := createMockServiceWithResources(t, "250", "64Mi", "1000m", "1024Mi")

	action, updated, _, err := fakeServiceUpdate(service, []string{
		"service", "update", "foo", "--requests-cpu", "500m", "--limits-cpu", "1000m", "--async"}, false)
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

	newTemplate, err := servinglib.RevisionTemplateOfService(updated)
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
		"service", "update", "foo", "--requests-memory", "128Mi", "--limits-memory", "2048Mi", "--async"}, false)
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

	newTemplate, err := servinglib.RevisionTemplateOfService(updated)
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
		"--requests-memory", "128Mi", "--limits-memory", "2048Mi", "--async"}, false)
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

	newTemplate, err := servinglib.RevisionTemplateOfService(updated)
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
	origTemplate, err := servinglib.RevisionTemplateOfService(original)
	if err != nil {
		t.Fatal(err)
	}
	origContainer, err := servinglib.ContainerOfRevisionTemplate(origTemplate)
	if err != nil {
		t.Fatal(err)
	}
	origContainer.Image = "gcr.io/foo/bar:latest"

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "-l", "a=mouse", "--label", "b=cookie", "-l=single", "--async"}, false)

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

	template, err := servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
	container, err := servinglib.ContainerOfRevisionTemplate(template)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, container.Image, exampleImageByDigest)
}

func TestServiceUpdateLabelExisting(t *testing.T) {
	original := newEmptyService()
	original.ObjectMeta.Labels = map[string]string{"already": "here", "tobe": "removed"}
	originalTemplate, _ := servinglib.RevisionTemplateOfService(original)
	originalTemplate.ObjectMeta.Labels = map[string]string{"already": "here", "tobe": "removed"}

	action, updated, _, err := fakeServiceUpdate(original, []string{
		"service", "update", "foo", "-l", "already=gone", "--label=tobe-", "--label", "b=", "--async"}, false)

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

	template, err := servinglib.RevisionTemplateOfService(updated)
	assert.NilError(t, err)
	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

func newEmptyService() *v1alpha1.Service {
	return newEmptyServiceBetaAPIStyle()
}

func newEmptyServiceBetaAPIStyle() *v1alpha1.Service {
	ret := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{},
	}
	ret.Spec.Template = &v1alpha1.RevisionTemplateSpec{}
	ret.Spec.Template.Annotations = map[string]string{
		servinglib.UserImageAnnotationKey: "",
	}
	ret.Spec.Template.Spec.Containers = []corev1.Container{{}}
	return ret
}

func createMockServiceWithResources(t *testing.T, requestCPU, requestMemory, limitsCPU, limitsMemory string) *v1alpha1.Service {
	service := newEmptyService()

	template, err := servinglib.RevisionTemplateOfService(service)
	if err != nil {
		t.Fatal(err)
	}

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

func newEmptyServiceAlphaAPIStyle() *v1alpha1.Service {
	return &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
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
}
