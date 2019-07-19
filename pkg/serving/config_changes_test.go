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
	"strconv"
	"testing"

	"gotest.tools/assert"

	"github.com/knative/serving/pkg/apis/autoscaling"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateAutoscalingAnnotations(t *testing.T) {
	template := &servingv1alpha1.RevisionTemplateSpec{}
	updateConcurrencyConfiguration(template, 10, 100, 1000, 1000)
	annos := template.Annotations
	if annos[autoscaling.MinScaleAnnotationKey] != "10" {
		t.Error("minScale failed")
	}
	if annos[autoscaling.MaxScaleAnnotationKey] != "100" {
		t.Error("maxScale failed")
	}
	if annos[autoscaling.TargetAnnotationKey] != "1000" {
		t.Error("target failed")
	}
	if template.Spec.ContainerConcurrency != 1000 {
		t.Error("limit failed")
	}
}

func TestUpdateInvalidAutoscalingAnnotations(t *testing.T) {
	template := &servingv1alpha1.RevisionTemplateSpec{}
	updateConcurrencyConfiguration(template, 10, 100, 1000, 1000)
	// Update with invalid concurrency options
	updateConcurrencyConfiguration(template, -1, -1, 0, -1)
	annos := template.Annotations
	if annos[autoscaling.MinScaleAnnotationKey] != "10" {
		t.Error("minScale failed")
	}
	if annos[autoscaling.MaxScaleAnnotationKey] != "100" {
		t.Error("maxScale failed")
	}
	if annos[autoscaling.TargetAnnotationKey] != "1000" {
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
	err := UpdateEnvVars(template, env, []string{})
	assert.NilError(t, err)
	found, err := EnvToMap(container.Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, env, found)
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
	err := UpdateEnvVars(template, env, []string{})
	assert.NilError(t, err)

	expected := map[string]string{
		"a": "foo",
		"b": "bar",
	}

	found, err := EnvToMap(container.Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, found)
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
		{Name: "a", Value: "foo"}}
	env := map[string]string{
		"a": "fancy",
	}
	err := UpdateEnvVars(revision, env, []string{})
	assert.NilError(t, err)

	expected := map[string]string{
		"a": "fancy",
	}

	found, err := EnvToMap(container.Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, found)
}

func TestUpdateEnvVarsRemove(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsRemove(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsRemove(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsRemove(t *testing.T, revision *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	remove := []string{"b"}
	err := UpdateEnvVars(revision, map[string]string{}, remove)
	assert.NilError(t, err)

	expected := map[string]string{
		"a": "foo",
	}

	found, err := EnvToMap(container.Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, found)
}

func TestUpdateMinScale(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateMinScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.MinScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMinScale(template, -1)
	assert.ErrorContains(t, err, "Invalid")
}

func TestUpdateMaxScale(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateMaxScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.MaxScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMaxScale(template, -1)
	assert.ErrorContains(t, err, "Invalid")
}

func TestUpdateConcurrencyTarget(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateConcurrencyTarget(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.TargetAnnotationKey, 10)
	// Update with invalid value
	err = UpdateConcurrencyTarget(template, -1)
	assert.ErrorContains(t, err, "Invalid")
}

func TestUpdateConcurrencyLimit(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateConcurrencyLimit(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerConcurrency(t, template, 10)
	// Update with invalid value
	err = UpdateConcurrencyLimit(template, -1)
	assert.ErrorContains(t, err, "Invalid")
}

func TestUpdateContainerImage(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateImage(template, "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerImage(t, template, "gcr.io/foo/bar:baz")
	// Update template with container image info
	template.Spec.GetContainer().Image = "docker.io/foo/bar:baz"
	err = UpdateImage(template, "query.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify that given image overrides the existing container image
	checkContainerImage(t, template, "query.io/foo/bar:baz")
}

func checkContainerImage(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, image string) {
	if got, want := template.Spec.GetContainer().Image, image; got != want {
		t.Errorf("Failed to update the container image: got=%s, want=%s", got, want)
	}
}

func TestUpdateContainerPort(t *testing.T) {
	template, _ := getV1alpha1Config()
	err := UpdateContainerPort(template, 8888)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkPortUpdate(t, template, 8888)
	// update template with container port info
	template.Spec.Containers[0].Ports[0].ContainerPort = 9090
	err = UpdateContainerPort(template, 80)
	assert.NilError(t, err)
	// Verify that given port overrides the existing container port
	checkPortUpdate(t, template, 80)
}

func TestUpdateName(t *testing.T) {
	template, _ := getV1alpha1Config()
	err := UpdateName(template, "foo-asdf")
	assert.NilError(t, err)
	assert.Equal(t, "foo-asdf", template.Name)
}

func checkPortUpdate(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, port int32) {
	if template.Spec.GetContainer().Ports[0].ContainerPort != port {
		t.Error("Failed to update the container port")
	}
}

func TestUpdateEnvVarsBoth(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsAll(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsAll(t, template, container)
	assertNoV1alpha1Old(t, template)
}

func testUpdateEnvVarsAll(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, container *corev1.Container) {
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "c", Value: "caroline"},
		{Name: "d", Value: "byebye"},
	}
	env := map[string]string{
		"a": "fancy",
		"b": "boo",
	}
	remove := []string{"d"}
	err := UpdateEnvVars(template, env, remove)
	assert.NilError(t, err)

	expected := map[string]string{
		"a": "fancy",
		"b": "boo",
		"c": "caroline",
	}

	found, err := EnvToMap(container.Env)
	assert.NilError(t, err)
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("Env did not match expected %v, found %v", env, found)
	}
}

func TestUpdateLabelsNew(t *testing.T) {
	service, template, _ := getV1alpha1Service()

	labels := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateLabels(service, template, labels, []string{})
	assert.NilError(t, err)

	actual := service.ObjectMeta.Labels
	if !reflect.DeepEqual(labels, actual) {
		t.Fatalf("Service labels did not match expected %v found %v", labels, actual)
	}

	actual = template.ObjectMeta.Labels
	if !reflect.DeepEqual(labels, actual) {
		t.Fatalf("Template labels did not match expected %v found %v", labels, actual)
	}
}

func TestUpdateLabelsExisting(t *testing.T) {
	service, template, _ := getV1alpha1Service()
	service.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}

	labels := map[string]string{
		"a": "notfoo",
		"c": "bat",
		"d": "",
	}
	err := UpdateLabels(service, template, labels, []string{})
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "notfoo",
		"b": "bar",
		"c": "bat",
		"d": "",
	}

	actual := service.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

func TestUpdateLabelsRemoveExisting(t *testing.T) {
	service, template, _ := getV1alpha1Service()
	service.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}

	remove := []string{"b"}
	err := UpdateLabels(service, template, map[string]string{}, remove)
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "foo",
	}

	actual := service.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

// =========================================================================================================

func getV1alpha1RevisionTemplateWithOldFields() (*servingv1alpha1.RevisionTemplateSpec, *corev1.Container) {
	container := &corev1.Container{}
	template := &servingv1alpha1.RevisionTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "template-foo",
			Namespace: "default",
		},
		Spec: servingv1alpha1.RevisionSpec{
			DeprecatedContainer: container,
		},
	}
	return template, container
}

func getV1alpha1Config() (*servingv1alpha1.RevisionTemplateSpec, *corev1.Container) {
	containers := []corev1.Container{{}}
	template := &servingv1alpha1.RevisionTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "template-foo",
			Namespace: "default",
		},
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

func getV1alpha1Service() (*servingv1alpha1.Service, *servingv1alpha1.RevisionTemplateSpec, *corev1.Container) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	service := &servingv1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alph1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: servingv1alpha1.ServiceSpec{
			DeprecatedRunLatest: &servingv1alpha1.RunLatestType{
				Configuration: servingv1alpha1.ConfigurationSpec{
					DeprecatedRevisionTemplate: template,
				},
			},
		},
	}
	return service, template, container
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

func checkAnnotationValue(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, key string, value int) {
	anno := template.GetAnnotations()
	if v, ok := anno[key]; !ok && v != strconv.Itoa(value) {
		t.Errorf("Failed to update %s annotation key: got=%s, want=%d", key, v, value)
	}
}

func checkContainerConcurrency(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, value int) {
	if got, want := template.Spec.ContainerConcurrency, value; got != v1beta1.RevisionContainerConcurrencyType(want) {
		t.Errorf("Failed to update containerConcurrency value: got=%d, want=%d", got, want)
	}
}

func updateConcurrencyConfiguration(template *servingv1alpha1.RevisionTemplateSpec, minScale int, maxScale int, target int, limit int) {
	UpdateMinScale(template, minScale)
	UpdateMaxScale(template, maxScale)
	UpdateConcurrencyTarget(template, target)
	UpdateConcurrencyLimit(template, limit)
}
