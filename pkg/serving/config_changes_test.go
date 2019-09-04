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

	"knative.dev/pkg/ptr"
	"knative.dev/serving/pkg/apis/autoscaling"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
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
	if *template.Spec.ContainerConcurrency != int64(1000) {
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
	if *template.Spec.ContainerConcurrency != 1000 {
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
	env := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	found, err := EnvToMap(env)
	assert.NilError(t, err)
	err = UpdateEnvVars(template, found, []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, env, container.Env)
}

func TestUpdateEnvVarsAppendOld(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	testUpdateEnvVarsAppendOld(t, template, container)
	assertNoV1alpha1(t, template)

	template, container = getV1alpha1Config()
	testUpdateEnvVarsAppendOld(t, template, container)
	assertNoV1alpha1Old(t, template)
}

type userImageAnnotCase struct {
	image  string
	annot  string
	result string
	set    bool
}

func TestSetUserImageAnnot(t *testing.T) {
	cases := []userImageAnnotCase{
		{"foo/bar", "", "foo/bar", true},
		{"foo/bar@sha256:asdfsf", "", "foo/bar@sha256:asdfsf", true},
		{"foo/bar@sha256:asdf", "foo/bar", "foo/bar", true},
		{"foo/bar", "baz/quux", "foo/bar", true},
		{"foo/bar", "", "", false},
		{"foo/bar", "baz/quux", "", false},
	}
	for _, c := range cases {
		template, container := getV1alpha1Config()
		if c.annot == "" {
			template.Annotations = nil
		} else {
			template.Annotations = map[string]string{
				UserImageAnnotationKey: c.annot,
			}
		}
		container.Image = c.image
		if c.set {
			SetUserImageAnnot(template)
		} else {
			UnsetUserImageAnnot(template)
		}
		assert.Equal(t, template.Annotations[UserImageAnnotationKey], c.result)
	}
}

func TestFreezeImageToDigest(t *testing.T) {
	template, container := getV1alpha1Config()
	revision := &servingv1alpha1.Revision{}
	revision.Spec = template.Spec
	revision.ObjectMeta = template.ObjectMeta
	revision.Status.ImageDigest = "gcr.io/foo/bar@sha256:deadbeef"
	container.Image = "gcr.io/foo/bar:latest"
	FreezeImageToDigest(template, revision)
	assert.Equal(t, container.Image, "gcr.io/foo/bar@sha256:deadbeef")
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

	expected := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}

	assert.DeepEqual(t, expected, container.Env)
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

	expected := []corev1.EnvVar{
		{"a", "foo", nil},
	}

	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateMinScale(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateMinScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.MinScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMinScale(template, -1)
	assert.ErrorContains(t, err, "minScale")
}

func TestUpdateMaxScale(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateMaxScale(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.MaxScaleAnnotationKey, 10)
	// Update with invalid value
	err = UpdateMaxScale(template, -1)
	assert.ErrorContains(t, err, "maxScale")
}

func TestUpdateConcurrencyTarget(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateConcurrencyTarget(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.TargetAnnotationKey, 10)
	// Update with invalid value
	err = UpdateConcurrencyTarget(template, -1)
	assert.ErrorContains(t, err, "invalid")
}

func TestUpdateConcurrencyLimit(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	err := UpdateConcurrencyLimit(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerConcurrency(t, template, ptr.Int64(int64(10)))
	// Update with invalid value
	err = UpdateConcurrencyLimit(template, -1)
	assert.ErrorContains(t, err, "invalid")
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

	expected := []corev1.EnvVar{
		{Name: "a", Value: "fancy"},
		{Name: "b", Value: "boo"},
		{Name: "c", Value: "caroline"},
	}

	assert.DeepEqual(t, expected, container.Env)
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

func TestUpdateEnvFrom(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	container.EnvFrom = append(container.EnvFrom,
		corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "config-map-existing-name",
				}}},
		corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "secret-existing-name",
				}}},
	)
	UpdateEnvFrom(template,
		[]string{"config-map:config-map-new-name-1", "secret:secret-new-name-1"},
		[]string{"config-map:config-map-existing-name", "secret:secret-existing-name"})
	assert.Equal(t, len(container.EnvFrom), 2)
	assert.Equal(t, container.EnvFrom[0].ConfigMapRef.Name, "config-map-new-name-1")
	assert.Equal(t, container.EnvFrom[1].SecretRef.Name, "secret-new-name-1")
}

func TestUpdateVolumeMountsAndVolumes(t *testing.T) {
	template, container := getV1alpha1RevisionTemplateWithOldFields()
	template.Spec.Volumes = append(template.Spec.Volumes,
		corev1.Volume{
			Name: "existing-config-map-volume-name-1",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-1",
					}}}},
		corev1.Volume{
			Name: "existing-config-map-volume-name-2",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "existing-config-map-2",
					}}}},
		corev1.Volume{
			Name: "existing-secret-volume-name-1",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-1",
				}}},
		corev1.Volume{
			Name: "existing-secret-volume-name-2",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "existing-secret-2",
				}}})

	container.VolumeMounts = append(container.VolumeMounts,
		corev1.VolumeMount{
			Name:      "existing-config-map-volume-name-1",
			ReadOnly:  true,
			MountPath: "/existing-config-map-1/mount/path",
		},
		corev1.VolumeMount{
			Name:      "existing-config-map-volume-name-2",
			ReadOnly:  true,
			MountPath: "/existing-config-map-2/mount/path",
		},
		corev1.VolumeMount{
			Name:      "existing-secret-volume-name-1",
			ReadOnly:  true,
			MountPath: "/existing-secret-1/mount/path",
		},
		corev1.VolumeMount{
			Name:      "existing-secret-volume-name-2",
			ReadOnly:  true,
			MountPath: "/existing-secret-2/mount/path",
		},
	)

	err := UpdateVolumeMountsAndVolumes(template,
		map[string]string{"/new-config-map/mount/path": "new-config-map-volume-name"},
		[]string{},
		map[string]string{"new-config-map-volume-name": "config-map:new-config-map"},
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		map[string]string{"/updated-config-map/mount/path": "existing-config-map-volume-name-2"},
		[]string{},
		map[string]string{"existing-config-map-volume-name-2": "config-map:updated-config-map"},
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		map[string]string{"/new-secret/mount/path": "new-secret-volume-name"},
		[]string{},
		map[string]string{"new-secret-volume-name": "secret:new-secret"},
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		map[string]string{"/updated-secret/mount/path": "existing-secret-volume-name-2"},
		[]string{"/existing-config-map-1/mount/path",
			"/existing-secret-1/mount/path"},
		map[string]string{"existing-secret-volume-name-2": "secret:updated-secret"},
		[]string{"existing-config-map-volume-name-1",
			"existing-secret-volume-name-1"})
	assert.NilError(t, err)

	assert.Equal(t, len(template.Spec.Volumes), 4)
	assert.Equal(t, len(container.VolumeMounts), 6)
	assert.Equal(t, template.Spec.Volumes[0].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, template.Spec.Volumes[0].ConfigMap.Name, "updated-config-map")
	assert.Equal(t, template.Spec.Volumes[1].Name, "existing-secret-volume-name-2")
	assert.Equal(t, template.Spec.Volumes[1].Secret.SecretName, "updated-secret")
	assert.Equal(t, template.Spec.Volumes[2].Name, "new-config-map-volume-name")
	assert.Equal(t, template.Spec.Volumes[2].ConfigMap.Name, "new-config-map")
	assert.Equal(t, template.Spec.Volumes[3].Name, "new-secret-volume-name")
	assert.Equal(t, template.Spec.Volumes[3].Secret.SecretName, "new-secret")

	assert.Equal(t, container.VolumeMounts[0].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, container.VolumeMounts[0].MountPath, "/existing-config-map-2/mount/path")
	assert.Equal(t, container.VolumeMounts[1].Name, "existing-secret-volume-name-2")
	assert.Equal(t, container.VolumeMounts[1].MountPath, "/existing-secret-2/mount/path")
	assert.Equal(t, container.VolumeMounts[2].Name, "new-config-map-volume-name")
	assert.Equal(t, container.VolumeMounts[2].MountPath, "/new-config-map/mount/path")
	assert.Equal(t, container.VolumeMounts[3].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, container.VolumeMounts[3].MountPath, "/updated-config-map/mount/path")
	assert.Equal(t, container.VolumeMounts[4].Name, "new-secret-volume-name")
	assert.Equal(t, container.VolumeMounts[4].MountPath, "/new-secret/mount/path")
	assert.Equal(t, container.VolumeMounts[5].Name, "existing-secret-volume-name-2")
	assert.Equal(t, container.VolumeMounts[5].MountPath, "/updated-secret/mount/path")
}

func TestUpdateServiceAccountName(t *testing.T) {
	template, _ := getV1alpha1RevisionTemplateWithOldFields()
	template.Spec.ServiceAccountName = ""

	UpdateServiceAccountName(template, "foo-bar")
	assert.Equal(t, template.Spec.ServiceAccountName, "foo-bar")

	UpdateServiceAccountName(template, "")
	assert.Equal(t, template.Spec.ServiceAccountName, "")
}

func TestUpdateAnnotationsNew(t *testing.T) {
	service, template, _ := getV1alpha1Service()

	annotations := map[string]string{
		"a": "foo",
		"b": "bar",
	}
	err := UpdateAnnotations(service, template, annotations, []string{})
	assert.NilError(t, err)

	actual := service.ObjectMeta.Annotations
	if !reflect.DeepEqual(annotations, actual) {
		t.Fatalf("Service annotations did not match expected %v found %v", annotations, actual)
	}

	actual = template.ObjectMeta.Annotations
	if !reflect.DeepEqual(annotations, actual) {
		t.Fatalf("Template annotations did not match expected %v found %v", annotations, actual)
	}
}

func TestUpdateAnnotationsExisting(t *testing.T) {
	service, template, _ := getV1alpha1Service()
	service.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}

	annotations := map[string]string{
		"a": "notfoo",
		"c": "bat",
		"d": "",
	}
	err := UpdateAnnotations(service, template, annotations, []string{})
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "notfoo",
		"b": "bar",
		"c": "bat",
		"d": "",
	}

	actual := service.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)

	actual = template.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)
}

func TestUpdateAnnotationsRemoveExisting(t *testing.T) {
	service, template, _ := getV1alpha1Service()
	service.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}
	template.ObjectMeta.Annotations = map[string]string{"a": "foo", "b": "bar"}

	remove := []string{"b"}
	err := UpdateAnnotations(service, template, map[string]string{}, remove)
	assert.NilError(t, err)
	expected := map[string]string{
		"a": "foo",
	}

	actual := service.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)

	actual = template.ObjectMeta.Annotations
	assert.DeepEqual(t, expected, actual)
}

//
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
			RevisionSpec: servingv1.RevisionSpec{
				PodSpec: corev1.PodSpec{
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

func checkContainerConcurrency(t *testing.T, template *servingv1alpha1.RevisionTemplateSpec, value *int64) {
	if got, want := *template.Spec.ContainerConcurrency, *value; got != want {
		t.Errorf("Failed to update containerConcurrency value: got=%d, want=%d", got, want)
	}
}

func updateConcurrencyConfiguration(template *servingv1alpha1.RevisionTemplateSpec, minScale int, maxScale int, target int, limit int) {
	UpdateMinScale(template, minScale)
	UpdateMaxScale(template, maxScale)
	UpdateConcurrencyTarget(template, target)
	UpdateConcurrencyLimit(template, int64(limit))
}
