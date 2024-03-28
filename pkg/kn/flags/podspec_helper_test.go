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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/client/lib/test"

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
	expected := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	argsEnv := []string{
		"a=foo",
		"b=bar",
	}
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	args := append([]string{"command"}, argsEnv...)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, util.NewOrderedMap(), []string{}, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, spec.Containers[0].Env)
}

func TestUpdateEnvVarsMixedEnvOrder(t *testing.T) {
	spec, _ := getPodSpec()
	expected := []corev1.EnvVar{
		{Name: "z", Value: "foo"},
		{Name: "a", Value: "bar"},
		{Name: "x", Value: "baz"},
	}
	argsEnv := []string{
		"z=foo",
		"a=bar",
		"x=baz",
	}
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	args := append([]string{"command"}, argsEnv...)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, util.NewOrderedMap(), []string{}, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, spec.Containers[0].Env)
}

func TestUpdateEnvVarsValueFromNew(t *testing.T) {
	spec, _ := getPodSpec()
	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "b", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
		{Name: "c", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "baz",
				},
				Key: "key3",
			},
		}},
		{Name: "d", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "goo",
				},
				Key: "key4",
			},
		}},
	}
	argsEnvValueFrom := []string{
		"a=secret:foo:key",
		"b=sc:bar:key2",
		"c=config-map:baz:key3",
		"d=cm:goo:key4",
	}
	args := append([]string{"command"}, argsEnvValueFrom...)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnvValueFrom, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, util.NewOrderedMap(), []string{}, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, spec.Containers[0].Env)
}

func TestUpdateEnvVarsAllNew(t *testing.T) {
	spec, _ := getPodSpec()
	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "b", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
		{Name: "c", Value: "baz"},
		{Name: "d", Value: "goo"},
	}
	argsEnvValueFrom := []string{
		"a=secret:foo:key",
		"b=sc:bar:key2",
	}
	argsEnv := []string{
		"c=baz",
		"d=goo",
	}

	args := append([]string{"command"}, append(argsEnvValueFrom, argsEnv...)...)
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnvValueFrom, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, spec.Containers[0].Env)
}

func TestUpdateEnvVarsValueFromValidate(t *testing.T) {
	spec, _ := getPodSpec()
	envValueFromWrongInput := [][]string{
		{"foo=foo"},
		{"foo=bar:"},
		{"foo=foo:bar"},
		{"foo=foo:bar:"},
		{"foo=foo:bar:baz"},
		{"foo=secret"},
		{"foo=sec"},
		{"foo=secret:"},
		{"foo=sec:"},
		{"foo=secret:name"},
		{"foo=sec:name"},
		{"foo=secret:name"},
		{"foo=sec:name:"},
		{"foo=config-map"},
		{"foo=cm"},
		{"foo=config-map:"},
		{"foo=cm:"},
		{"foo=config-map:name"},
		{"foo=cm:name"},
		{"foo=config-map:name:"},
		{"foo=cm:name:"},
	}

	for _, input := range envValueFromWrongInput {
		args := append([]string{"command"}, input...)
		envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(input, "=")
		assert.NilError(t, err)
		err = UpdateEnvVars(spec, args, util.NewOrderedMap(), []string{}, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
		fmt.Println()
		msg := fmt.Sprintf("input \"%s\" should fail, as it is not valid entry for containers.env.valueFrom", input[0])
		assert.ErrorContains(t, err, " ", msg)
	}

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
	quantity := resource.MustParse("10Gi")
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
				}}},
		corev1.Volume{
			Name: "new-empty-dir-volume-name-1",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    "",
					SizeLimit: nil,
				},
			},
		},
		corev1.Volume{
			Name: "new-empty-dir-volume-name-2",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    "Memory",
					SizeLimit: &quantity,
				},
			},
		},
		corev1.Volume{
			Name: "new-empty-dir-volume-name-3",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		},
		corev1.Volume{
			Name: "new-empty-dir-volume-name-4",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &quantity,
				},
			},
		},
		corev1.Volume{
			Name: "new-pvc-volume-name-1",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "pvc1",
				},
			},
		})

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
		corev1.VolumeMount{
			Name:      "new-empty-dir-volume-name-1",
			ReadOnly:  false,
			MountPath: "/empty-dir-1/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-empty-dir-volume-name-2",
			ReadOnly:  false,
			MountPath: "/empty-dir-2/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-empty-dir-volume-name-3",
			ReadOnly:  false,
			MountPath: "/empty-dir-3/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-empty-dir-volume-name-4",
			ReadOnly:  false,
			MountPath: "/empty-dir-4/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-empty-dir-volume-name-5",
			ReadOnly:  false,
			MountPath: "/empty-dir-5/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-pvc-volume-name-1",
			MountPath: "/pvc-1/mount/path",
		},
		corev1.VolumeMount{
			Name:      "new-pvc-volume-name-2",
			MountPath: "/pvc-2/mount/path",
		},
	)

	err := UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/new-config-map/mount/path", "config-map:new-config-map:readOnly=false"}}),
		[]string{},
		util.NewOrderedMap(),
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

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/empty-dir-1/mount/path", "new-empty-dir-volume-name-1"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-empty-dir-volume-name-1", "emptyDir:new-empty-dir-volume-name-1"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/empty-dir-2/mount/path", "new-empty-dir-volume-name-2"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-empty-dir-volume-name-2", "emptyDir:new-empty-dir-volume-name-2:type=Memory,size=10Gi"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/empty-dir-3/mount/path", "new-empty-dir-volume-name-3"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-empty-dir-volume-name-3", "emptyDir:new-empty-dir-volume-name-3:type=Memory"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/empty-dir-4/mount/path", "new-empty-dir-volume-name-4"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-empty-dir-volume-name-4", "emptyDir:new-empty-dir-volume-name-4:size=10Gi"}}),
		[]string{})
	assert.NilError(t, err)

	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/empty-dir-5/mount/path", "emptyDir:new-empty-dir-volume-name-5"}}),
		[]string{},
		util.NewOrderedMap(),
		[]string{})
	assert.NilError(t, err)
	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/pvc-1/mount/path", "new-pvc-volume-name-1"}}),
		[]string{},
		util.NewOrderedMapWithKVStrings([][]string{{"new-pvc-volume-name-1", "pvc:pvc1"}}),
		[]string{})
	assert.NilError(t, err)
	err = UpdateVolumeMountsAndVolumes(spec,
		util.NewOrderedMapWithKVStrings([][]string{{"/pvc-2/mount/path", "pvc:pvc2:readOnly=true"}}),
		[]string{},
		util.NewOrderedMap(),
		[]string{})
	assert.NilError(t, err)

	assert.Equal(t, len(spec.Volumes), 11)
	assert.Equal(t, len(container.VolumeMounts), 13)

	assert.Equal(t, spec.Volumes[0].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, spec.Volumes[0].ConfigMap.Name, "updated-config-map")
	assert.Equal(t, spec.Volumes[1].Name, "existing-secret-volume-name-2")
	assert.Equal(t, spec.Volumes[1].Secret.SecretName, "updated-secret")
	assert.Equal(t, spec.Volumes[2].Name, "new-empty-dir-volume-name-1")
	assert.Equal(t, spec.Volumes[2].EmptyDir.Medium, corev1.StorageMediumDefault)
	assert.Assert(t, spec.Volumes[2].EmptyDir.SizeLimit == nil)
	assert.Equal(t, spec.Volumes[3].Name, "new-empty-dir-volume-name-2")
	assert.Equal(t, spec.Volumes[3].EmptyDir.Medium, corev1.StorageMediumMemory)
	assert.DeepEqual(t, spec.Volumes[3].EmptyDir.SizeLimit, &quantity)
	assert.Equal(t, spec.Volumes[4].Name, "new-empty-dir-volume-name-3")
	assert.Equal(t, spec.Volumes[4].EmptyDir.Medium, corev1.StorageMediumMemory)
	assert.Equal(t, spec.Volumes[5].Name, "new-empty-dir-volume-name-4")
	assert.DeepEqual(t, spec.Volumes[5].EmptyDir.SizeLimit, &quantity)
	assert.Equal(t, spec.Volumes[6].Name, "new-pvc-volume-name-1")
	assert.Equal(t, spec.Volumes[6].PersistentVolumeClaim.ClaimName, "pvc1")
	assert.Equal(t, spec.Volumes[8].Name, "new-secret-volume-name")
	assert.Equal(t, spec.Volumes[8].Secret.SecretName, "new-secret")
	assert.Assert(t, strings.Contains(spec.Volumes[9].Name, "empty-dir-5"))
	assert.Equal(t, spec.Volumes[9].EmptyDir.Medium, corev1.StorageMediumDefault)
	assert.Assert(t, spec.Volumes[9].EmptyDir.SizeLimit == nil)

	assert.Equal(t, container.VolumeMounts[0].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, container.VolumeMounts[0].MountPath, "/existing-config-map-2/mount/path")
	assert.Equal(t, container.VolumeMounts[0].ReadOnly, true)
	assert.Equal(t, container.VolumeMounts[1].Name, "existing-secret-volume-name-2")
	assert.Equal(t, container.VolumeMounts[1].MountPath, "/existing-secret-2/mount/path")
	assert.Equal(t, container.VolumeMounts[1].ReadOnly, true)
	assert.Equal(t, container.VolumeMounts[2].Name, "new-empty-dir-volume-name-1")
	assert.Equal(t, container.VolumeMounts[2].MountPath, "/empty-dir-1/mount/path")
	assert.Equal(t, container.VolumeMounts[2].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[3].Name, "new-empty-dir-volume-name-2")
	assert.Equal(t, container.VolumeMounts[3].MountPath, "/empty-dir-2/mount/path")
	assert.Equal(t, container.VolumeMounts[3].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[4].Name, "new-empty-dir-volume-name-3")
	assert.Equal(t, container.VolumeMounts[4].MountPath, "/empty-dir-3/mount/path")
	assert.Equal(t, container.VolumeMounts[4].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[5].Name, "new-empty-dir-volume-name-4")
	assert.Equal(t, container.VolumeMounts[5].MountPath, "/empty-dir-4/mount/path")
	assert.Equal(t, container.VolumeMounts[5].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[6].MountPath, "/empty-dir-5/mount/path")
	assert.Equal(t, container.VolumeMounts[6].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[7].Name, "new-pvc-volume-name-1")
	assert.Equal(t, container.VolumeMounts[7].MountPath, "/pvc-1/mount/path")
	assert.Equal(t, container.VolumeMounts[7].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[8].MountPath, "/pvc-2/mount/path")
	assert.Equal(t, container.VolumeMounts[8].ReadOnly, true)
	assert.Equal(t, container.VolumeMounts[9].MountPath, "/new-config-map/mount/path")
	assert.Equal(t, container.VolumeMounts[9].ReadOnly, false)
	assert.Equal(t, container.VolumeMounts[10].Name, "existing-config-map-volume-name-2")
	assert.Equal(t, container.VolumeMounts[10].MountPath, "/updated-config-map/mount/path")
	assert.Equal(t, container.VolumeMounts[10].ReadOnly, true)
	assert.Equal(t, container.VolumeMounts[11].Name, "new-secret-volume-name")
	assert.Equal(t, container.VolumeMounts[11].MountPath, "/new-secret/mount/path")
	assert.Equal(t, container.VolumeMounts[11].ReadOnly, true)
	assert.Equal(t, container.VolumeMounts[12].Name, "existing-secret-volume-name-2")
	assert.Equal(t, container.VolumeMounts[12].MountPath, "/updated-secret/mount/path")
	assert.Equal(t, container.VolumeMounts[12].ReadOnly, true)
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
	err := UpdateContainerCommand(spec, []string{"/app/start"})
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Command, []string{"/app/start"})

	err = UpdateContainerCommand(spec, []string{"sh", "/app/latest.sh"})
	assert.NilError(t, err)
	assert.DeepEqual(t, spec.Containers[0].Command, []string{"sh", "/app/latest.sh"})
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

	expected := []corev1.EnvVar{
		{Name: "a", Value: "bar"},
	}
	argsEnv := []string{
		"a=bar",
	}
	args := append([]string{"command"}, argsEnv...)
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, util.NewOrderedMap(), []string{}, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateEnvVarsValueFromModify(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
	}

	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
	}
	argsEnvValueFrom := []string{
		"a=secret:bar:key2",
	}
	args := append([]string{"command"}, argsEnvValueFrom...)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnvValueFrom, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, util.NewOrderedMap(), []string{}, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateEnvVarsAllModify(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "b", Value: "bar"},
	}

	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
		{Name: "b", Value: "goo"},
	}
	argsEnvValueFrom := []string{
		"a=secret:bar:key2",
	}
	argsEnv := []string{
		"b=goo",
	}
	args := append([]string{"command"}, append(argsEnvValueFrom, argsEnv...)...)
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnvValueFrom, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateEnvVarsRemove(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", Value: "foo"},
		{Name: "b", Value: "bar"},
	}
	remove := []string{"b-"}
	args := append([]string{"command"}, remove...)
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(remove, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, util.NewOrderedMap(), []string{}, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)

	expected := []corev1.EnvVar{
		{Name: "a", Value: "foo"},
	}

	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateEnvVarsValueFromRemove(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "b", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
	}
	remove := []string{"b-"}
	args := append([]string{"command"}, remove...)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(remove, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, util.NewOrderedMap(), []string{}, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)

	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
	}

	assert.DeepEqual(t, expected, container.Env)
}

func TestUpdateEnvVarsAllRemove(t *testing.T) {
	spec, container := getPodSpec()
	container.Env = []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "b", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "bar",
				},
				Key: "key2",
			},
		}},
		{Name: "c", Value: "baz"},
		{Name: "d", Value: "goo"},
	}
	argsEnvValueFrom := []string{
		"a=secret:foo:key",
		"b-",
	}
	argsEnv := []string{
		"c=baz",
		"d-",
	}

	args := append([]string{"command"}, append(argsEnvValueFrom, argsEnv...)...)
	envToUpdate, envToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnv, "=")
	assert.NilError(t, err)
	envValueFromToUpdate, envValueFromToRemove, err := util.OrderedMapAndRemovalListFromArray(argsEnvValueFrom, "=")
	assert.NilError(t, err)
	err = UpdateEnvVars(spec, args, envToUpdate, envToRemove, envValueFromToUpdate, envValueFromToRemove, "", util.NewOrderedMap(), []string{})
	assert.NilError(t, err)

	expected := []corev1.EnvVar{
		{Name: "a", ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "foo",
				},
				Key: "key",
			},
		}},
		{Name: "c", Value: "baz"},
	}

	assert.DeepEqual(t, expected, container.Env)
}

func Test_isValidEnvArg(t *testing.T) {
	for _, tc := range []struct {
		name     string
		arg      string
		envKey   string
		envValue string
		isValid  bool
	}{{
		name:     "valid env arg specified",
		arg:      "FOO=bar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  true,
	}, {
		name:     "invalid env arg specified",
		arg:      "FOObar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  false,
	}, {
		name:     "valid env arg specified: -e",
		arg:      "-e=FOO=bar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  true,
	}, {
		name:     "invalid env arg specified: -e",
		arg:      "-e=FOObar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  false,
	}, {
		name:     "valid env arg specified: --env",
		arg:      "--env=FOO=bar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  true,
	}, {
		name:     "invalid env arg specified: --env",
		arg:      "--env=FOObar",
		envKey:   "FOO",
		envValue: "bar",
		isValid:  false,
	},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidEnvArg(tc.arg, tc.envKey, tc.envValue)
			assert.Equal(t, result, tc.isValid)
		})
	}
}

func Test_isValidEnvValueFromArg(t *testing.T) {
	for _, tc := range []struct {
		name              string
		arg               string
		envValueFromKey   string
		envValueFromValue string
		isValid           bool
	}{{
		name:              "valid env value from arg specified",
		arg:               "FOO=secret:sercretName:key",
		envValueFromKey:   "FOO",
		envValueFromValue: "secret:sercretName:key",
		isValid:           true,
	}, {
		name:              "invalid env value from arg specified",
		arg:               "FOOsecret:sercretName:key",
		envValueFromKey:   "FOO",
		envValueFromValue: "secret:sercretName:key",
		isValid:           false,
	}, {
		name:              "valid env value from arg specified: --env-value-from",
		arg:               "--env-value-from=FOO=secret:sercretName:key",
		envValueFromKey:   "FOO",
		envValueFromValue: "secret:sercretName:key",
		isValid:           true,
	}, {
		name:              "invalid env value from arg specified: --env-value-from",
		arg:               "--env-value-from=FOOsecret:sercretName:key",
		envValueFromKey:   "FOO",
		envValueFromValue: "secret:sercretName:key",
		isValid:           false,
	},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidEnvValueFromArg(tc.arg, tc.envValueFromKey, tc.envValueFromValue)
			assert.Equal(t, result, tc.isValid)
		})
	}
}

func TestUpdateContainers(t *testing.T) {
	podSpec, _ := getPodSpec()
	containers := []corev1.Container{
		{
			Name:  "foo",
			Image: "foo:bar",
		},
		{
			Name:  "bar",
			Image: "foo:bar",
		},
	}
	assert.Equal(t, len(podSpec.Containers), 1)
	UpdateContainers(podSpec, containers)
	assert.Equal(t, len(podSpec.Containers), 3)

	updatedContainer := corev1.Container{Name: "bar", Image: "bar:bar"}
	UpdateContainers(podSpec, []corev1.Container{updatedContainer})
	assert.Equal(t, len(podSpec.Containers), 3)
	for _, container := range podSpec.Containers {
		if container.Name == updatedContainer.Name {
			assert.DeepEqual(t, container, updatedContainer)
		}
	}

	// Verify that containers aren't multiplied
	UpdateContainers(podSpec, containers)
	assert.Equal(t, len(podSpec.Containers), 3)

	podSpec, _ = getPodSpec()
	assert.Equal(t, len(podSpec.Containers), 1)
	UpdateContainers(podSpec, []corev1.Container{})
	assert.Equal(t, len(podSpec.Containers), 1)
}

func TestUpdateContainerWithName(t *testing.T) {
	for _, tc := range []struct {
		name               string
		updateArg          []corev1.Container
		expectedContainers []corev1.Container
	}{{
		"One Container Image",
		[]corev1.Container{
			{Name: "bar", Image: "bar:bar"},
		},
		[]corev1.Container{
			{},
			{Name: "foo", Image: "foo:bar"},
			{Name: "bar", Image: "bar:bar"},
		}},
		{
			"One Container Env Var",
			[]corev1.Container{
				{Name: "bar", Image: "foo:bar", Env: []corev1.EnvVar{{Name: "A", Value: "B"}}},
			},
			[]corev1.Container{
				{},
				{Name: "foo", Image: "foo:bar"},
				{Name: "bar", Image: "foo:bar", Env: []corev1.EnvVar{{Name: "A", Value: "B"}}},
			}},
		{
			"New container",
			[]corev1.Container{
				{Name: "new", Image: "foo:new"},
			},
			[]corev1.Container{
				{},
				{Name: "foo", Image: "foo:bar"},
				{Name: "bar", Image: "foo:bar"},
				{Name: "new", Image: "foo:new"},
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initialPodSpec, _ := getPodSpec()
			initialContainers := []corev1.Container{
				{Name: "foo", Image: "foo:bar"},
				{Name: "bar", Image: "foo:bar"},
			}
			initialPodSpec.Containers = append(initialPodSpec.Containers, initialContainers...)

			UpdateContainers(initialPodSpec, tc.updateArg)
			assert.DeepEqual(t, initialPodSpec.Containers, tc.expectedContainers)
		})
	}
}

func TestParseContainers(t *testing.T) {
	rawInput := `
containers:
- image: first
  name: foo
  resources: {}
- image: second
  name: bar
  resources: {}`

	stdinReader, stdinWriter, err := os.Pipe()
	assert.NilError(t, err)
	_, err = stdinWriter.Write([]byte(rawInput))
	assert.NilError(t, err)
	stdinWriter.Close()

	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()
	os.Stdin = stdinReader

	fromFile, err := decodeContainersFromFile("-")
	assert.NilError(t, err)
	assert.Equal(t, len(fromFile.Containers), 2)

	tempDir := t.TempDir()
	fileName := filepath.Join(tempDir, "container.yaml")
	os.WriteFile(fileName, []byte(rawInput), test.FileModeReadWrite)
	fromFile, err = decodeContainersFromFile(fileName)
	assert.NilError(t, err)
	assert.Equal(t, len(fromFile.Containers), 2)

	_, err = decodeContainersFromFile("non-existing")
	assert.Assert(t, err != nil)
	assert.Assert(t, util.ContainsAll(err.Error(), "no", "file", "directory"))
}

func TestUpdateImagePullPolicy(t *testing.T) {
	policyMap := make(map[string][]string)
	policyMap["Always"] = []string{"always", "ALWAYS", "Always"}
	policyMap["Never"] = []string{"never", "NEVER", "Never"}
	policyMap["IfNotPresent"] = []string{"ifnotpresent", "IFNOTPRESENT", "IfNotPresent"}

	for k, values := range policyMap {
		expectedPodSpec := &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:           "repo/user/imageID:tag",
					ImagePullPolicy: corev1.PullPolicy(k),
					Command:         []string{"/app/start"},
					Args:            []string{"myArg1"},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8080,
						},
					},
				},
			},
		}
		for _, v := range values {
			podSpec := &corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:   "repo/user/imageID:tag",
						Command: []string{"/app/start"},
						Args:    []string{"myArg1"},
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
					},
				},
			}
			err := UpdateImagePullPolicy(podSpec, v)
			assert.NilError(t, err, "update pull policy failed")
			assert.DeepEqual(t, expectedPodSpec, podSpec)
		}
	}
}

func TestUpdateImagePullPolicyError(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image:   "repo/user/imageID:tag",
				Command: []string{"/app/start"},
				Args:    []string{"myArg1"},
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8080,
					},
				},
			},
		},
	}
	err := UpdateImagePullPolicy(podSpec, "InvalidPolicy")
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid --pull-policy", "Valid arguments", "Always | Never | IfNotPresent"))
}

func TestUpdateProbes(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{{}},
	}
	expected := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Port: intstr.Parse("8080"), Path: "/path"}}},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Port: intstr.Parse("8080"), Path: "/path"}}},
			},
		},
	}
	t.Run("Update readiness & liveness", func(t *testing.T) {
		err := UpdateLivenessProbe(podSpec, "http::8080:/path")
		assert.NilError(t, err)
		err = UpdateReadinessProbe(podSpec, "http::8080:/path")
		assert.NilError(t, err)
		assert.DeepEqual(t, podSpec, expected)
	})
	t.Run("Update readiness with error", func(t *testing.T) {
		err := UpdateReadinessProbe(podSpec, "http-probe::8080:/path")
		assert.Assert(t, err != nil)
		assert.ErrorContains(t, err, "unsupported probe type")
	})
	t.Run("Update liveness with error", func(t *testing.T) {
		err := UpdateLivenessProbe(podSpec, "http-probe::8080:/path")
		assert.Assert(t, err != nil)
		assert.ErrorContains(t, err, "unsupported probe type")
	})
}

func TestUpdateProbesOpts(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{{}},
	}
	expected := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				LivenessProbe: &corev1.Probe{
					InitialDelaySeconds: 5,
					TimeoutSeconds:      10},
				ReadinessProbe: &corev1.Probe{
					InitialDelaySeconds: 5,
					TimeoutSeconds:      10},
			},
		},
	}
	t.Run("Update readiness & liveness", func(t *testing.T) {
		err := UpdateLivenessProbeOpts(podSpec, "initialdelayseconds=5")
		assert.NilError(t, err)
		err = UpdateLivenessProbeOpts(podSpec, "timeoutseconds=10")
		assert.NilError(t, err)
		err = UpdateReadinessProbeOpts(podSpec, "initialdelayseconds=5,timeoutseconds=10")
		assert.NilError(t, err)
		assert.DeepEqual(t, podSpec, expected)
	})
	t.Run("Update readiness with error", func(t *testing.T) {
		err := UpdateReadinessProbeOpts(podSpec, "timeout=10")
		assert.Assert(t, err != nil)
		assert.ErrorContains(t, err, "not a valid probe parameter")
	})
	t.Run("Update liveness with error", func(t *testing.T) {
		err := UpdateLivenessProbeOpts(podSpec, "initdelay=5")
		assert.Assert(t, err != nil)
		assert.ErrorContains(t, err, "not a valid probe parameter")
	})
}

func TestUpdateProbesWithOpts(t *testing.T) {
	podSpec := &corev1.PodSpec{
		Containers: []corev1.Container{{}},
	}
	expected := &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Port: intstr.Parse("8080"), Path: "/path"}},
					InitialDelaySeconds: 10},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Port: intstr.Parse("8080"), Path: "/path"}},
					TimeoutSeconds: 10},
			},
		},
	}
	t.Run("Update readiness & liveness", func(t *testing.T) {
		err := UpdateLivenessProbe(podSpec, "http::8080:/path")
		assert.NilError(t, err)
		err = UpdateLivenessProbeOpts(podSpec, "initialdelayseconds=10")
		assert.NilError(t, err)
		err = UpdateReadinessProbeOpts(podSpec, "timeoutseconds=10")
		assert.NilError(t, err)
		err = UpdateReadinessProbe(podSpec, "http::8080:/path")
		assert.NilError(t, err)
		assert.DeepEqual(t, podSpec, expected)
	})
}

func TestResolveProbeHandlerError(t *testing.T) {
	for _, tc := range []struct {
		name        string
		probeString string
		err         error
	}{
		{
			name:        "Probe string empty",
			probeString: "",
			err:         errors.New("no probe parameters detected"),
		},
		{
			name:        "Probe string too many parameters",
			probeString: "http:too-many:test-host:8080:/",
			err:         errors.New("too many probe parameters provided, please check the format"),
		},
		{
			name:        "Probe invalid prefix",
			probeString: "http-probe:test-host:8080:/",
			err:         errors.New("unsupported probe type 'http-probe'; supported types: http, https, exec, tcp"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := resolveProbeHandler(tc.probeString)
			assert.Assert(t, actual == nil)
			assert.Error(t, err, tc.err.Error())
			assert.ErrorType(t, err, tc.err)

		})
	}
}

func TestResolveProbeHandlerHTTP(t *testing.T) {
	for _, tc := range []struct {
		name        string
		probeString string
		expected    *corev1.ProbeHandler
		err         error
	}{
		{
			name:        "HTTP probe empty params",
			probeString: "http:::",
			expected:    &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{}},
			err:         nil,
		},
		{
			name:        "HTTP probe  all params",
			probeString: "http:test-host:8080:/",
			expected: &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
				Host: "test-host", Port: intstr.Parse("8080"), Path: "/",
			}},
			err: nil,
		},
		{
			name:        "HTTPS probe  all params",
			probeString: "https:test-host:8080:/",
			expected: &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
				Host: "test-host", Scheme: corev1.URISchemeHTTPS, Port: intstr.Parse("8080"), Path: "/",
			}},
			err: nil,
		},
		{
			name:        "HTTP probe port path params",
			probeString: "http::8080:/",
			expected: &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.Parse("8080"), Path: "/",
			}},
			err: nil,
		},
		{
			name:        "HTTP probe path params",
			probeString: "http:::/path",
			expected: &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
				Path: "/path",
			}},
			err: nil,
		},
		{
			name:        "HTTP probe named port path params",
			probeString: "http::namedPort:/",
			expected: &corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.Parse("namedPort"), Path: "/",
			}},
			err: nil,
		},
		{
			name:        "HTTP probe not enough params",
			probeString: "http::",
			expected:    nil,
			err:         errors.New("unexpected probe format, please use 'http:host:port:path'"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := resolveProbeHandler(tc.probeString)
			if tc.err == nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			} else {
				assert.Assert(t, actual == nil)
				assert.Error(t, err, tc.err.Error())
				assert.ErrorType(t, err, tc.err)
			}

		})
	}
}

func TestResolveProbeHandlerExec(t *testing.T) {
	for _, tc := range []struct {
		name        string
		probeString string
		expected    *corev1.ProbeHandler
		err         error
	}{
		{
			name:        "Exec probe single cmd params",
			probeString: "exec:/bin/cmd/with/slashes",
			expected: &corev1.ProbeHandler{Exec: &corev1.ExecAction{
				Command: []string{"/bin/cmd/with/slashes"},
			}},
			err: nil,
		},
		{
			name:        "Exec probe multiple cmd params",
			probeString: "exec:/bin/cmd,arg,arg",
			expected: &corev1.ProbeHandler{Exec: &corev1.ExecAction{
				Command: []string{"/bin/cmd", "arg", "arg"},
			}},
			err: nil,
		},
		{
			name:        "Exec probe empty command param",
			probeString: "exec:",
			expected:    nil,
			err:         errors.New("at least one command parameter is required for Exec probe"),
		},
		{
			name:        "Exec probe too many params",
			probeString: "exec:cmd:cmd",
			expected:    nil,
			err:         errors.New("unexpected probe format, please use 'exec:<exec_command>[,<exec_command>,...]'"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := resolveProbeHandler(tc.probeString)
			if tc.err == nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			} else {
				assert.Assert(t, actual == nil)
				assert.Error(t, err, tc.err.Error())
				assert.ErrorType(t, err, tc.err)
			}

		})
	}
}

func TestResolveProbeHandlerTCP(t *testing.T) {
	for _, tc := range []struct {
		name        string
		probeString string
		expected    *corev1.ProbeHandler
		err         error
	}{
		{
			name:        "TCPSocket probe host port",
			probeString: "tcp:test-host:8080",
			expected: &corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{
				Host: "test-host", Port: intstr.Parse("8080"),
			}},
			err: nil,
		},
		{
			name:        "TCPSocket probe host named port",
			probeString: "tcp:test-host:myPort",
			expected: &corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{
				Host: "test-host", Port: intstr.Parse("myPort"),
			}},
			err: nil,
		},
		{
			name:        "TCPSocket probe empty command param",
			probeString: "tcp:",
			expected:    nil,
			err:         errors.New("unexpected probe format, please use 'tcp:host:port"),
		},
		{
			name:        "TCPSocket probe too many params",
			probeString: "tcp:host:port:more",
			expected:    nil,
			err:         errors.New("unexpected probe format, please use 'tcp:host:port"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := resolveProbeHandler(tc.probeString)
			if tc.err == nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			} else {
				assert.Assert(t, actual == nil)
				assert.Error(t, err, tc.err.Error())
				assert.ErrorType(t, err, tc.err)
			}

		})
	}
}

func TestResolveProbeOptions(t *testing.T) {
	for _, tc := range []struct {
		name        string
		probeString string
		expected    *corev1.Probe
		err         error
	}{
		{
			name:        "Common options all",
			probeString: "InitialDelaySeconds=1,TimeoutSeconds=2,PeriodSeconds=3,SuccessThreshold=4,FailureThreshold=5",
			expected: &corev1.Probe{
				InitialDelaySeconds: 1,
				TimeoutSeconds:      2,
				PeriodSeconds:       3,
				SuccessThreshold:    4,
				FailureThreshold:    5,
			},
			err: nil,
		},
		{
			name:        "Error duplicate value",
			probeString: "InitialDelaySeconds=2,InitialDelaySeconds=3",
			expected:    nil,
			err:         errors.New("The key \"InitialDelaySeconds\" has been duplicate in [InitialDelaySeconds=2 InitialDelaySeconds=3]"),
		},
		{
			name:        "Error not a numeric value",
			probeString: "InitialDelaySeconds=v",
			expected:    nil,
			err:         errors.New("not a nummeric value for parameter 'InitialDelaySeconds'"),
		},
		{
			name:        "Error invalid parameter name",
			probeString: "InitialDelay=5",
			expected:    nil,
			err:         errors.New("not a valid probe parameter name 'InitialDelay'"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := &corev1.Probe{}
			err := resolveProbeOptions(actual, tc.probeString)
			if tc.err == nil {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			} else {
				assert.Error(t, err, tc.err.Error())
				assert.ErrorType(t, err, tc.err)
			}

		})
	}
}

func TestUpdateSecurityContext(t *testing.T) {
	testCases := []struct {
		name string

		expected      *corev1.PodSpec
		expectedError error
	}{
		{
			name: "strict",
			expected: &corev1.PodSpec{
				Containers: []corev1.Container{
					{SecurityContext: DefaultStrictSecCon()}},
			},
			expectedError: nil,
		},
		{
			name: "none",
			expected: &corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
			expectedError: nil,
		},
		{
			name: "unknown",
			expected: &corev1.PodSpec{
				Containers: []corev1.Container{
					{SecurityContext: DefaultStrictSecCon()}},
			},
			expectedError: errors.New("invalid --security-context unknown. Valid arguments: strict | none"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &corev1.PodSpec{}
			err := UpdateSecurityContext(actual, tc.name)
			if tc.expectedError != nil {
				assert.Error(t, err, tc.expectedError.Error())
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			}
		})
	}
}

func TestUpdateNodeSelector(t *testing.T) {
	testCases := []struct {
		name               string
		nodeSelectorString []string
		expected           *corev1.PodSpec
		expectedError      error
	}{
		{
			name:               "Single node selector",
			nodeSelectorString: []string{"foo=bar"},
			expected: &corev1.PodSpec{
				NodeSelector: map[string]string{
					"k8s.io/hostname": "test",
					"foo":             "bar",
				},
			},
			expectedError: nil,
		},
		{
			name:               "Multiple node selectors",
			nodeSelectorString: []string{"foo1=bar1", "foo2=bar2"},
			expected: &corev1.PodSpec{
				NodeSelector: map[string]string{
					"k8s.io/hostname": "test",
					"foo1":            "bar1",
					"foo2":            "bar2",
				},
			},
			expectedError: nil,
		},
		{
			name:               "Removing a node selector",
			nodeSelectorString: []string{"k8s.io/hostname-"},
			expected: &corev1.PodSpec{
				NodeSelector: map[string]string{},
			},
			expectedError: nil,
		},
		{
			name:               "Passing empty key in node selector",
			nodeSelectorString: []string{"=test"},
			expected: &corev1.PodSpec{
				NodeSelector: map[string]string{},
			},
			expectedError: errors.New("The key is empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &corev1.PodSpec{
				NodeSelector: map[string]string{
					"k8s.io/hostname": "test",
				},
			}
			err := UpdateNodeSelector(actual, tc.nodeSelectorString)
			if tc.expectedError != nil {
				assert.Error(t, err, tc.expectedError.Error())
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, actual, tc.expected)
			}
		})
	}
}
