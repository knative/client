// Copyright 2020 The Knative Authors
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

package flags

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/ptr"
)

func TestPodSpecFlags(t *testing.T) {
	args := []string{"--image", "repo/user/imageID:tag", "--env", "b=c"}
	wantedPod := &PodSpecFlags{
		Image:        "repo/user/imageID:tag",
		Env:          []string{"b=c"},
		EnvFrom:      []string{},
		EnvValueFrom: []string{},
		Mount:        []string{},
		Volume:       []string{},
		Arg:          []string{},
		Command:      []string{},
	}
	flags := &PodSpecFlags{}
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			assert.DeepEqual(t, wantedPod, flags)
		},
	}
	testCmd.SetArgs(args)
	flags.AddFlags(testCmd.Flags())
	testCmd.Execute()
}

func TestUniqueStringArg(t *testing.T) {
	var a uniqueStringArg
	a.Set("test")
	assert.Equal(t, "test", a.String())
	assert.Equal(t, "string", a.Type())
}

func TestPodSpecResolve(t *testing.T) {
	inputArgs := []string{"--image", "repo/user/imageID:tag", "--env", "b=c",
		"--port", "8080", "--limit", "cpu=1000m", "--limit", "memory=1024Mi",
		"--cmd", "/app/start", "--arg", "myArg1", "--service-account", "foo-bar-account",
		"--mount", "/mount/path=volume-name", "--volume", "volume-name=cm:config-map-name",
		"--env-from", "config-map:config-map-name", "--user", "1001"}
	expectedPodSpec := corev1.PodSpec{
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
				Env: []corev1.EnvVar{{Name: "b", Value: "c"}},
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "config-map-name",
							},
						},
					},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    parseQuantity("1000m"),
						corev1.ResourceMemory: parseQuantity("1024Mi"),
					},
					Requests: corev1.ResourceList{},
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "volume-name", ReadOnly: true, MountPath: "/mount/path"}},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: ptr.Int64(int64(1001)),
				},
			},
		},
		ServiceAccountName: "foo-bar-account",
		Volumes: []corev1.Volume{
			{
				Name: "volume-name",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "config-map-name",
						},
					},
				},
			},
		},
	}
	flags := &PodSpecFlags{}
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			podSpec := &corev1.PodSpec{Containers: []corev1.Container{{}}}
			err := flags.ResolvePodSpec(podSpec, cmd.Flags(), inputArgs)
			assert.NilError(t, err, "PodSpec cannot be resolved.")
			assert.DeepEqual(t, expectedPodSpec, *podSpec)
		},
	}
	testCmd.SetArgs(inputArgs)
	flags.AddFlags(testCmd.Flags())
	testCmd.Execute()
}

func TestPodSpecResolveReturnError(t *testing.T) {
	outBuf := bytes.Buffer{}
	flags := &PodSpecFlags{}
	inputArgs := []string{"--mount", "123456"}
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			podSpec := &corev1.PodSpec{Containers: []corev1.Container{{}}}
			err := flags.ResolvePodSpec(podSpec, cmd.Flags(), inputArgs)
			fmt.Fprint(cmd.OutOrStdout(), "Return error: ", err)
		},
	}
	testCmd.SetOut(&outBuf)

	testCmd.SetArgs(inputArgs)
	flags.AddFlags(testCmd.Flags())
	testCmd.Execute()
	out := outBuf.String()
	assert.Assert(t, util.ContainsAll(out, "Invalid", "mount"))
}
