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

	"knative.dev/client/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestUpdateAutoscalingAnnotations(t *testing.T) {
	template := &servingv1.RevisionTemplateSpec{}
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
	template := &servingv1.RevisionTemplateSpec{}
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
	template, container := getRevisionTemplate()
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
			SetUserImageAnnot(template)
		} else {
			UnsetUserImageAnnot(template)
		}
		assert.Equal(t, template.Annotations[UserImageAnnotationKey], c.result)
	}
}

func TestFreezeImageToDigest(t *testing.T) {
	template, container := getRevisionTemplate()
	revision := &servingv1.Revision{}
	revision.Spec = template.Spec
	revision.ObjectMeta = template.ObjectMeta
	revision.Status.ImageDigest = "gcr.io/foo/bar@sha256:deadbeef"
	container.Image = "gcr.io/foo/bar:latest"
	FreezeImageToDigest(template, revision)
	assert.Equal(t, container.Image, "gcr.io/foo/bar@sha256:deadbeef")
}

func testUpdateEnvVarsAppendOld(t *testing.T, template *servingv1.RevisionTemplateSpec, container *corev1.Container) {
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
	revision, container := getRevisionTemplate()
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
	revision, container := getRevisionTemplate()
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

func TestAutoscaleWindow(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateAutoscaleWindow(template, "10s")
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValue(t, template, autoscaling.WindowAnnotationKey, "10s")
	// Update with invalid value
	err = UpdateAutoscaleWindow(template, "blub")
	assert.Check(t, util.ContainsAll(err.Error(), "invalid duration", "autoscale-window"))
}

func TestUpdateConcurrencyTarget(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateConcurrencyTarget(template, 10)
	assert.NilError(t, err)
	// Verify update is successful or not
	checkAnnotationValueInt(t, template, autoscaling.TargetAnnotationKey, 10)
	// Update with invalid value
	err = UpdateConcurrencyTarget(template, -1)
	assert.ErrorContains(t, err, "invalid")
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

func TestUpdateContainerImage(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateImage(template, "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerImage(t, template, "gcr.io/foo/bar:baz")
	// Update template with container image info
	template.Spec.Containers[0].Image = "docker.io/foo/bar:baz"
	err = UpdateImage(template, "query.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify that given image overrides the existing container image
	checkContainerImage(t, template, "query.io/foo/bar:baz")
}

func checkContainerImage(t *testing.T, template *servingv1.RevisionTemplateSpec, image string) {
	if got, want := template.Spec.Containers[0].Image, image; got != want {
		t.Errorf("Failed to update the container image: got=%s, want=%s", got, want)
	}
}

func TestUpdateContainerCommand(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateContainerCommand(template, "/app/start")
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Command, []string{"/app/start"})

	err = UpdateContainerCommand(template, "/app/latest")
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Command, []string{"/app/latest"})
}

func TestUpdateContainerArg(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateContainerArg(template, []string{"--myArg"})
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Args, []string{"--myArg"})

	err = UpdateContainerArg(template, []string{"myArg1", "--myArg2"})
	assert.NilError(t, err)
	assert.DeepEqual(t, template.Spec.Containers[0].Args, []string{"myArg1", "--myArg2"})
}

func TestUpdateContainerPort(t *testing.T) {
	template, _ := getRevisionTemplate()
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

func checkPortUpdate(t *testing.T, template *servingv1.RevisionTemplateSpec, port int32) {
	if template.Spec.Containers[0].Ports[0].ContainerPort != port {
		t.Error("Failed to update the container port")
	}
}

func checkUserUpdate(t *testing.T, template *servingv1.RevisionTemplateSpec, user *int64) {
	assert.DeepEqual(t, template.Spec.Containers[0].SecurityContext.RunAsUser, user)
}

func TestUpdateEnvVarsBoth(t *testing.T) {
	template, container := getRevisionTemplate()
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

func TestUpdateEnvFrom(t *testing.T) {
	template, container := getRevisionTemplate()
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
	template, container := getRevisionTemplate()
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
		util.NewOrderedMapWithKVStrings([][]string{{"/new-config-map/mount/path", "new-config-map-volume-name"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-config-map-volume-name", "config-map:new-config-map"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		util.NewOrderedMapWithKVStrings([][]string{{"/updated-config-map/mount/path", "existing-config-map-volume-name-2"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"existing-config-map-volume-name-2", "config-map:updated-config-map"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		util.NewOrderedMapWithKVStrings([][]string{{"/new-secret/mount/path", "new-secret-volume-name"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-secret-volume-name", "secret:new-secret"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(template,
		util.NewOrderedMapWithKVStrings([][]string{{"/updated-secret/mount/path", "existing-secret-volume-name-2"}}),
		[]string{"/existing-config-map-1/mount/path",
			"/existing-secret-1/mount/path"},
		util.NewOrderedMapWithKVStrings([][]string{{"existing-secret-volume-name-2", "secret:updated-secret"}}),
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
	template, _ := getRevisionTemplate()
	template.Spec.ServiceAccountName = ""

	UpdateServiceAccountName(template, "foo-bar")
	assert.Equal(t, template.Spec.ServiceAccountName, "foo-bar")

	UpdateServiceAccountName(template, "")
	assert.Equal(t, template.Spec.ServiceAccountName, "")
}

func TestUpdateImagePullSecrets(t *testing.T) {
	template, _ := getRevisionTemplate()
	template.Spec.ImagePullSecrets = nil

	UpdateImagePullSecrets(template, "quay")
	assert.Equal(t, template.Spec.ImagePullSecrets[0].Name, "quay")

	UpdateImagePullSecrets(template, " ")
	assert.Check(t, template.Spec.ImagePullSecrets == nil)
}

func TestUpdateAnnotationsNew(t *testing.T) {
	service, template, _ := getService()

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
	service, template, _ := getService()
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
	service, template, _ := getService()
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

func TestGenerateVolumeName(t *testing.T) {
	actual := []string{
		"Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ",
		"/Ab12~`!@#$%^&*()-=_+[]{}|/\\<>,./?:;\"'xZ/",
		"",
		"/",
	}

	expected := []string{
		"ab12---------------------.----..-----xz",
		"ab12---------------------.----..-----xz.",
		"",
		"",
	}

	for i := range actual {
		actualName := GenerateVolumeName(actual[i])
		expectedName := appendCheckSum(expected[i], actual[i])
		assert.Equal(t, actualName, expectedName)
	}
}

func TestUpdateUser(t *testing.T) {
	template, _ := getRevisionTemplate()
	err := UpdateUser(template, int64(1001))
	assert.NilError(t, err)

	checkUserUpdate(t, template, ptr.Int64(int64(1001)))

	template.Spec.Containers[0].SecurityContext.RunAsUser = ptr.Int64(int64(1002))
	err = UpdateUser(template, int64(1002))
	assert.NilError(t, err)

	checkUserUpdate(t, template, ptr.Int64(int64(1002)))
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

func updateConcurrencyConfiguration(template *servingv1.RevisionTemplateSpec, minScale int, maxScale int, target int, limit int) {
	UpdateMinScale(template, minScale)
	UpdateMaxScale(template, maxScale)
	UpdateConcurrencyTarget(template, target)
	UpdateConcurrencyLimit(template, int64(limit))
}
