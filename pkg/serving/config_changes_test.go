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

	"gotest.tools/v3/assert"

	"knative.dev/pkg/ptr"
	"knative.dev/serving/pkg/apis/autoscaling"

	"knative.dev/client/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestUpdateScalingAnnotations(t *testing.T) {
	template := &servingv1.RevisionTemplateSpec{}
	updateConcurrencyConfiguration(template, 10, 100, 1000, 1000, 50)
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
	if *template.Spec.ContainerConcurrency != int64(1000) {
		t.Error("limit failed")
	}
}

func TestUpdateInvalidScalingAnnotations(t *testing.T) {
	template := &servingv1.RevisionTemplateSpec{}
	updateConcurrencyConfiguration(template, 10, 100, 1000, 1000, 50)
	// Update with invalid concurrency options
	updateConcurrencyConfiguration(template, -1, -1, 0, -1, 200)
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
	if annos[autoscaling.TargetUtilizationPercentageKey] != "50" {
		t.Error("concurrency utilization failed")
	}
	if *template.Spec.ContainerConcurrency != 1000 {
		t.Error("limit failed")
	}
}

type userImageAnnotCase struct {
	image  string
	annot  string
	result string
	set    bool
}

func TestSetUserImageAnnotation(t *testing.T) {
	cases := []userImageAnnotCase{
		{"foo/bar", "", "foo/bar", true},
		{"foo/bar@sha256:asdfsf", "", "foo/bar@sha256:asdfsf", true},
		{"foo/bar@sha256:asdf", "foo/bar", "foo/bar", true},
		{"foo/bar", "baz/quux", "foo/bar", true},
		{"foo/bar", "", "", false},
		{"foo/bar", "baz/quux", "", false},
	}
	for _, c := range cases {
		template, container := getRevisionTemplate()
		if c.annot == "" {
			template.Annotations = nil
		} else {
			template.Annotations = map[string]string{
				UserImageAnnotationKey: c.annot,
			}
		}
		container.Image = c.image
		if c.set {
			UpdateUserImageAnnotation(template)
		} else {
			UnsetUserImageAnnotation(template)
		}
		assert.Equal(t, template.Annotations[UserImageAnnotationKey], c.result)
	}
}

func TestPinImageToDigest(t *testing.T) {
	template, container := getRevisionTemplate()
	revision := &servingv1.Revision{}
	revision.Spec = template.Spec
	revision.ObjectMeta = template.ObjectMeta
	revision.Status.ContainerStatuses = []servingv1.ContainerStatus{
		{Name: "user-container", ImageDigest: "gcr.io/foo/bar@sha256:deadbeef"},
	}
	container.Image = "gcr.io/foo/bar:latest"
	err := PinImageToDigest(template, revision)
	assert.NilError(t, err)
	assert.Equal(t, container.Image, "gcr.io/foo/bar@sha256:deadbeef")

	// No base revision --> no-op
	err = PinImageToDigest(template, nil)
	assert.NilError(t, err)
}

func TestPinImageToDigestInvalidImages(t *testing.T) {
	template, container := getRevisionTemplate()
	container.Image = "gcr.io/A"
	revision := &servingv1.Revision{
		Spec: servingv1.RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "gcr.io/B"},
				},
			},
		},
	}
	err := PinImageToDigest(template, revision)
	assert.ErrorContains(t, err, "unexpected image")
}

func TestUpdateTimestampAnnotation(t *testing.T) {
	template, _ := getRevisionTemplate()
	UpdateTimestampAnnotation(template)
	assert.Assert(t, template.Annotations[UpdateTimestampAnnotationKey] != "")
}

func TestUpdateMinScale(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateMinScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValueInt(t, template, autoscaling.MinScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMinScale(template, -1)
	assert.ErrorContains(t, err, "minScale")
}

func TestUpdateMaxScale(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateMaxScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValueInt(t, template, autoscaling.MaxScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMaxScale(template, -1)
	assert.ErrorContains(t, err, "maxScale")
}

func TestScaleWindow(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateScaleWindow(template, "10s")
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.WindowAnnotationKey, "10s")
	// Update with invalid value
	err = UpdateScaleWindow(template, "blub")
	assert.Check(t, util.ContainsAll(err.Error(), "invalid duration", "scale-window"))
}

func TestUpdateConcurrencyTarget(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateConcurrencyTarget(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValueInt(t, template, autoscaling.TargetAnnotationKey, 10)
	// Update with invalid value
	err = UpdateConcurrencyTarget(template, -1)
	assert.ErrorContains(t, err, "should be at least 0.01")
}

func TestUpdateConcurrencyLimit(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateConcurrencyLimit(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerConcurrency(t, template, ptr.Int64(int64(10)))
	// Update with invalid value
	err = UpdateConcurrencyLimit(template, -1)
	assert.ErrorContains(t, err, "invalid")
}

// func TestUpdateEnvVarsBoth(t *testing.T) {
// 	template, container := getRevisionTemplate()
// 	container.Env = []corev1.EnvVar{
// 		{Name: "a", Value: "foo"},
// 		{Name: "c", Value: "caroline"},
// 		{Name: "d", Value: "byebye"},
// 	}
// 	env := map[string]string{
// 		"a": "fancy",
// 		"b": "boo",
// 	}
// 	remove := []string{"d"}
// 	err := UpdateEnvVars(template, env, remove)
// 	assert.NilError(t, err)

// 	expected := []corev1.EnvVar{
// 		{Name: "a", Value: "fancy"},
// 		{Name: "b", Value: "boo"},
// 		{Name: "c", Value: "caroline"},
// 	}

// 	assert.DeepEqual(t, expected, container.Env)
// }

func TestUpdateLabelsNew(t *testing.T) {
	service, template, _ := getService()

	labels := map[string]string{
		"a": "foo",
		"b": "bar",
	}

	service.ObjectMeta.Labels = UpdateLabels(service.ObjectMeta.Labels, labels, []string{})
	template.ObjectMeta.Labels = UpdateLabels(template.ObjectMeta.Labels, labels, []string{})

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
	service, template, _ := getService()
	service.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}

	labels := map[string]string{
		"a": "notfoo",
		"c": "bat",
		"d": "",
	}
	tlabels := map[string]string{
		"a": "notfoo",
		"c": "bat",
		"d": "",
		"r": "poo",
	}

	service.ObjectMeta.Labels = UpdateLabels(service.ObjectMeta.Labels, labels, []string{})
	template.ObjectMeta.Labels = UpdateLabels(template.ObjectMeta.Labels, tlabels, []string{})

	expectedServiceLabel := map[string]string{
		"a": "notfoo",
		"b": "bar",
		"c": "bat",
		"d": "",
	}
	expectedRevLabel := map[string]string{
		"a": "notfoo",
		"b": "bar",
		"c": "bat",
		"d": "",
		"r": "poo",
	}

	actual := service.ObjectMeta.Labels
	assert.DeepEqual(t, expectedServiceLabel, actual)

	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expectedRevLabel, actual)
}

func TestUpdateLabelsRemoveExisting(t *testing.T) {
	service, template, _ := getService()
	service.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Labels = map[string]string{"a": "foo", "b": "bar"}

	remove := []string{"b"}
	service.ObjectMeta.Labels = UpdateLabels(service.ObjectMeta.Labels, map[string]string{}, remove)
	template.ObjectMeta.Labels = UpdateLabels(template.ObjectMeta.Labels, map[string]string{}, remove)

	expected := map[string]string{
		"a": "foo",
	}

	actual := service.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)

	actual = template.ObjectMeta.Labels
	assert.DeepEqual(t, expected, actual)
}

func TestUpdateRevisionTemplateAnnotationsNew(t *testing.T) {
	_, template, _ := getService()

	annotations := map[string]string{
		autoscaling.InitialScaleAnnotationKey: "1",
		autoscaling.MaxScaleAnnotationKey:     "2",
	}
	err := UpdateRevisionTemplateAnnotations(template, annotations, []string{})
	assert.NilError(t, err)

	actual := template.ObjectMeta.Annotations
	assert.DeepEqual(t, annotations, actual)
}

func TestUpdateRevisionTemplateAnnotationsExisting(t *testing.T) {
	_, template, _ := getService()
	template.ObjectMeta.Annotations = map[string]string{
		autoscaling.InitialScaleAnnotationKey: "1",
		autoscaling.MaxScaleAnnotationKey:     "2",
	}

	annotations := map[string]string{
		autoscaling.InitialScaleAnnotationKey: "5",
		autoscaling.MaxScaleAnnotationKey:     "10",
	}
	err := UpdateRevisionTemplateAnnotations(template, annotations, []string{})
	assert.NilError(t, err)

	actual := template.ObjectMeta.Annotations
	assert.DeepEqual(t, annotations, actual)
}

func TestUpdateRevisionTemplateAnnotationsRemoveExisting(t *testing.T) {
	_, template, _ := getService()
	template.ObjectMeta.Annotations = map[string]string{
		autoscaling.InitialScaleAnnotationKey: "1",
		autoscaling.MaxScaleAnnotationKey:     "2",
	}
	expectedAnnotations := map[string]string{
		autoscaling.InitialScaleAnnotationKey: "1",
	}
	remove := []string{autoscaling.MaxScaleAnnotationKey}
	err := UpdateRevisionTemplateAnnotations(template, map[string]string{}, remove)
	assert.NilError(t, err)

	actual := template.ObjectMeta.Annotations
	assert.DeepEqual(t, expectedAnnotations, actual)
}

func TestUpdateAnnotationsNew(t *testing.T) {
	service, _, _ := getService()

	annotations := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateServiceAnnotations(service, annotations, []string{})
	assert.NilError(t, err)

	actual := service.ObjectMeta.Annotations
	if !reflect.DeepEqual(annotations, actual) {
		t.Fatalf("Service annotations did not match expected %v found %v", annotations, actual)
	}
}

func TestUpdateAnnotationsExisting(t *testing.T) {
	service, _, _ := getService()
	service.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}

	annotations := map[string]string{
		"a": "notfoo",
		"c": "bat",
		"d": "",
	}
	err := UpdateServiceAnnotations(service, annotations, []string{})
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "notfoo",
		"b": "bar",
		"c": "bat",
		"d": "",
	}

	actual := service.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)
}

func TestUpdateAnnotationsRemoveExisting(t *testing.T) {
	service, _, _ := getService()
	service.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}

	remove := []string{"b"}
	err := UpdateServiceAnnotations(service, map[string]string{}, remove)
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "foo",
	}

	actual := service.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)
}

func TestString(t *testing.T) {
	vt := ConfigMapVolumeSourceType
	assert.Equal(t, "config-map", vt.String())
	vt = -1
	assert.Equal(t, "unknown", vt.String())
}

//
// =========================================================================================================

func getRevisionTemplate() (*servingv1.RevisionTemplateSpec, *corev1.Container) {
	template := &servingv1.RevisionTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "template-foo",
			Namespace: "default",
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
		},
	}
	return template, &template.Spec.Containers[0]
}

func getService() (*servingv1.Service, *servingv1.RevisionTemplateSpec, *corev1.Container) {
	template, container := getRevisionTemplate()
	service := &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: *template,
			},
		},
	}
	return service, template, container
}

func checkAnnotationValueInt(t *testing.T, template *servingv1.RevisionTemplateSpec, key string, value int) {
	anno := template.GetAnnotations()
	if v, ok := anno[key]; !ok && v != strconv.Itoa(value) {
		t.Errorf("Failed to update %s annotation key: got=%s, want=%d", key, v, value)
	}
}

func checkAnnotationValue(t *testing.T, template *servingv1.RevisionTemplateSpec, key string, value string) {
	anno := template.GetAnnotations()
	if v, ok := anno[key]; !ok && v != value {
		t.Errorf("Failed to update %s annotation key: got=%s, want=%s", key, v, value)
	}
}

func checkContainerConcurrency(t *testing.T, template *servingv1.RevisionTemplateSpec, value *int64) {
	if got, want := *template.Spec.ContainerConcurrency, *value; got != want {
		t.Errorf("Failed to update containerConcurrency value: got=%d, want=%d", got, want)
	}
}

func updateConcurrencyConfiguration(template *servingv1.RevisionTemplateSpec, minScale, maxScale, target, limit, utilization int) {
	UpdateMinScale(template, minScale)
	UpdateMaxScale(template, maxScale)
	UpdateConcurrencyTarget(template, target)
	UpdateConcurrencyLimit(template, int64(limit))
	UpdateConcurrencyUtilization(template, utilization)
}
