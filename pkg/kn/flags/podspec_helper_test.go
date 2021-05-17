/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flags

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/ptr"
)

func getPodSpec() (*corev1.PodSpec, *corev1.Container) {
	spec := &corev1.PodSpec{
		Containers: []corev1.Container{{}},
	}
	return spec, &spec.Containers[0]
}

func TestUpdateEnvVarsNew(t *testing.T) {
	spec, _ := getPodSpec()
	env := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	found, err := util.EnvToMap(env)
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, found, []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, env, spec.Containers[0].Env)
}

func TestUpdateEnvFrom(t *testing.T) {
	spec, container := getPodSpec()
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
	UpdateEnvFrom(spec,
		[]string{"config-map:config-map-new-name-1", "secret:secret-new-name-1"},
		[]string{"config-map:config-map-existing-name", "secret:secret-existing-name"})
	assert.Equal(t, len(container.EnvFrom), 2)
	assert.Equal(t, container.EnvFrom[0].ConfigMapRef.Name, "config-map-new-name-1")
	assert.Equal(t, container.EnvFrom[1].SecretRef.Name, "secret-new-name-1")
}

func TestUpdateVolumeMountsAndVolumes(t *testing.T) {
	spec, container := getPodSpec()
	spec.Volumes = append(spec.Volumes,
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

	err := UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/new-config-map/mount/path", "new-config-map-volume-name"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-config-map-volume-name", "config-map:new-config-map"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/updated-config-map/mount/path", "existing-config-map-volume-name-2"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"existing-config-map-volume-name-2", "config-map:updated-config-map"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/new-secret/mount/path", "new-secret-volume-name"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-secret-volume-name", "secret:new-secret"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/updated-secret/mount/path", "existing-secret-volume-name-2"}}),
		[]string{"/existing-config-map-1/mount/path",
			"/existing-secret-1/mount/path"},
		util.NewOrderedMapWithKVStrings([][]string{{"existing-secret-volume-name-2", "secret:updated-secret"}}),
		[]string{"existing-config-map-volume-name-1",
			"existing-secret-volume-name-1"})
	assert.NilError(t, err)

	assert.Equal(t, len(spec.Volumes), 4)
	assert.Equal(t, len(container.VolumeMounts), 6)
	assert.Equal(t, spec.Volumes[0].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, spec.Volumes[0].ConfigMap.Name, "updated-config-map")
	assert.Equal(t, spec.Volumes[1].Name, "existing-secret-volume-name-2")
	assert.Equal(t, spec.Volumes[1].Secret.SecretName, "updated-secret")
	assert.Equal(t, spec.Volumes[2].Name, "new-config-map-volume-name")
	assert.Equal(t, spec.Volumes[2].ConfigMap.Name, "new-config-map")
	assert.Equal(t, spec.Volumes[3].Name, "new-secret-volume-name")
	assert.Equal(t, spec.Volumes[3].Secret.SecretName, "new-secret")

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

func TestUpdateContainerImage(t *testing.T) {
	spec, _ := getPodSpec()
	err := UpdateImage(spec, "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify update is successful or not
	checkContainerImage(t, spec, "gcr.io/foo/bar:baz")
	// Update spec with container image info
	spec.Containers[0].Image = "docker.io/foo/bar:baz"
	err = UpdateImage(spec, "query.io/foo/bar:baz")
	assert.NilError(t, err)
	// Verify that given image overrides the existing container image
	checkContainerImage(t, spec, "query.io/foo/bar:baz")
}

func checkContainerImage(t *testing.T, spec *corev1.PodSpec, image string) {
	if got, want := spec.Containers[0].Image, image; got != want {
		t.Errorf("Failed to update the container image: got=%s, want=%s", got, want)
	}
}

func TestUpdateContainerCommand(t *testing.T) {
	spec, _ := getPodSpec()
	err := UpdateContainerCommand(spec, "/app/start")
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Command, []string{"/app/start"})

	err = UpdateContainerCommand(spec, "/app/latest")
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Command, []string{"/app/latest"})
}

func TestUpdateContainerArg(t *testing.T) {
	spec, _ := getPodSpec()
	err := UpdateContainerArg(spec, []string{"--myArg"})
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Args, []string{"--myArg"})

	err = UpdateContainerArg(spec, []string{"myArg1", "--myArg2"})
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Args, []string{"myArg1", "--myArg2"})
}

func TestUpdateContainerPort(t *testing.T) {
	spec, _ := getPodSpec()
	for _, tc := range []struct {
		name    string
		input   string
		isErr   bool
		expPort int32
		expName string
	}{{
		name:    "only port 8888",
		input:   "8888",
		expPort: int32(8888),
	}, {
		name:    "name and port h2c:8080",
		input:   "h2c:8080",
		expPort: int32(8080),
		expName: "h2c",
	}, {
		name:  "error case - not correct format",
		input: "h2c:800000000000000000",
		isErr: true,
	}, {
		name:  "error case - empty port",
		input: "h2c:",
		isErr: true,
	}, {
		name:  "error case - wrong format",
		input: "8080:h2c",
		isErr: true,
	}, {
		name:  "error case - multiple :",
		input: "h2c:8080:proto",
		isErr: true,
	}, {
		name:    "empty name no error",
		input:   ":8888",
		expPort: int32(8888),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			err := UpdateContainerPort(spec, tc.input)
			if tc.isErr {
				assert.Error(t, err, fmt.Sprintf(PortFormatErr, tc.input))
			} else {
				assert.NilError(t, err)
				assert.Equal(t, spec.Containers[0].Ports[0].ContainerPort, tc.expPort)
				assert.Equal(t, spec.Containers[0].Ports[0].Name, tc.expName)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	spec, _ := getPodSpec()
	err := UpdateUser(spec, int64(1001))
	assert.NilError(t, err)

	checkUserUpdate(t, spec, ptr.Int64(int64(1001)))

	spec.Containers[0].SecurityContext.RunAsUser = ptr.Int64(int64(1002))
	err = UpdateUser(spec, int64(1002))
	assert.NilError(t, err)

	checkUserUpdate(t, spec, ptr.Int64(int64(1002)))
}

func checkUserUpdate(t *testing.T, spec *corev1.PodSpec, user *int64) {
	assert.DeepEqual(t, spec.Containers[0].SecurityContext.RunAsUser, user)
}

func TestUpdateServiceAccountName(t *testing.T) {
	spec, _ := getPodSpec()
	spec.ServiceAccountName = ""

	UpdateServiceAccountName(spec, "foo-bar")
	assert.Equal(t, spec.ServiceAccountName, "foo-bar")

	UpdateServiceAccountName(spec, "")
	assert.Equal(t, spec.ServiceAccountName, "")
}

func TestUpdateImagePullSecrets(t *testing.T) {
	spec, _ := getPodSpec()
	spec.ImagePullSecrets = nil

	UpdateImagePullSecrets(spec, "quay")
	assert.Equal(t, spec.ImagePullSecrets[0].Name, "quay")

	UpdateImagePullSecrets(spec, " ")
	assert.Check(t, spec.ImagePullSecrets == nil)
}

func TestUpdateEnvVarsModify(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"}}
	env := map[string]string{
		"a": "fancy",
	}
	err := UpdateEnvVars(spec, env, []string{})
	assert.NilError(t, err)

	expected := map[string]string{
		"a": "fancy",
	}

	found, err := util.EnvToMap(container.Env)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, found)
}

func TestUpdateEnvVarsRemove(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	remove := []string{"b"}
	err := UpdateEnvVars(spec, map[string]string{}, remove)
	assert.NilError(t, err)

	expected := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
	}

	assert.DeepEqual(t, expected, container.Env)
}
