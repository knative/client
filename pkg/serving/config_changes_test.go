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

package serving

import (
	"reflect"
	"testing"

	"github.com/knative/serving/pkg/apis/serving/v1beta1"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestUpdateAutoscalingAnnotations(t *testing.T) {
	template := &servingv1alpha1.RevisionTemplateSpec{}
	UpdateConcurrencyConfiguration(template, 10, 100, 1000, 1000)
	annos := template.Annotations
	if annos["autoscaling.knative.dev/minScale"] != "10" {
		t.Error("minScale failed")
	}
	if annos["autoscaling.knative.dev/maxScale"] != "100" {
		t.Error("maxScale failed")
	}
	if annos["autoscaling.knative.dev/target"] != "1000" {
		t.Error("target failed")
	}
	if template.Spec.ContainerConcurrency != 1000 {
		t.Error("limit failed")
	}
}

func TestUpdateEnvVarsNew(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsNew(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsNew(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsNew(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	env := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateEnvVars(template, env)
	if err != nil {
		t.Fatal(err)
	}
	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(env, found) {
		t.Fatalf("Env did not match expected %v found %v", env, found)
	}
}

func TestUpdateEnvVarsAppendOld(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsAppendOld(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsAppendOld(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsAppendOld(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
	}
	env := map[string]string{
		"b": "bar",
	}
	err := UpdateEnvVars(template, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "foo",
		"b": "bar",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v, found %v", env, found)
	}
}

func TestUpdateEnvVarsModify(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsModify(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsModify(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsModify(t *testing.T, revision *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		corev1.EnvVar{Name: "a", Value: "foo"}}
	env := map[string]string{
		"a": "fancy",
	}
	err := UpdateEnvVars(revision, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "fancy",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v, found %v", env, found)
	}
}

func TestUpdateEnvVarsBoth(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsBoth(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsBoth(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsBoth(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		corev1.EnvVar{Name: "a", Value: "foo"},
		corev1.EnvVar{Name: "c", Value: "caroline"}}
	env := map[string]string{
		"a": "fancy",
		"b": "boo",
	}
	err := UpdateEnvVars(template, env)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"a": "fancy",
		"b": "boo",
		"c": "caroline",
	}

	found, err := EnvToMap(container.Env)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v, found %v", env, found)
	}
}

// =========================================================================================================

func getV1alpha1RevisionTemplateWithOldFields() (*servingv1alpha1.RevisionTemplateSpec, *corev1.Container) {
	container := &corev1.Container{}
	template := &servingv1alpha1.RevisionTemplateSpec{
		Spec: servingv1alpha1.RevisionSpec{
			DeprecatedContainer: container,
		},
	}
	return template, container
}

func getV1alpha1Config() (*servingv1alpha1.RevisionTemplateSpec, *corev1.Container) {
	containers := []corev1.Container{{}}
	template := &servingv1alpha1.RevisionTemplateSpec{
		Spec: servingv1alpha1.RevisionSpec{
			RevisionSpec: v1beta1.RevisionSpec{
				PodSpec: v1beta1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
	return template, &containers[0]
}

func assertNoV1alpha1Old(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec) {
	if template.Spec.DeprecatedContainer != nil {
		t.Error("Assuming only new v1alpha1 fields but found spec.container")
	}
}

func assertNoV1alpha1(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec) {
	if template.Spec.Containers != nil {
		t.Error("Assuming only old v1alpha1 fields but found spec.template")
	}
}
