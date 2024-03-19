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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"knative.dev/client/lib/test"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/ptr"
)

func TestPodSpecFlags(t *testing.T) {
	args := []string{"--image", "repo/user/imageID:tag", "--env", "b=c"}
	wantedPod := &PodSpecFlags{
		Image:           "repo/user/imageID:tag",
		Env:             []string{"b=c"},
		EnvFrom:         []string{},
		EnvValueFrom:    []string{},
		Mount:           []string{},
		Volume:          []string{},
		NodeSelector:    []string{},
		Toleration:      []string{},
		NodeAffinity:    []string{},
		Arg:             []string{},
		Command:         []string{},
		SecurityContext: "none",
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
	flags.AddCreateFlags(testCmd.Flags())
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
		"--env-from", "config-map:config-map-name", "--user", "1001", "--pull-policy", "always",
		"--probe-readiness", "http::8080:/path", "--probe-liveness", "http::8080:/path",
		"--node-selector", "kubernetes.io/hostname=test-clusterw1-123",
		"--toleration", "Key=node-role.kubernetes.io/master,effect=NoSchedule,operator=Equal,Value=",
		"--node-affinity", "Type=Required,Key=topology.kubernetes.io/zone,Operator=In,Values=antarctica-east1"}
	expectedPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image:           "repo/user/imageID:tag",
				ImagePullPolicy: "Always",
				Command:         []string{"/app/start"},
				Args:            []string{"myArg1"},
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8080,
					},
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{Port: intstr.Parse("8080"), Path: "/path"},
					},
				},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{Port: intstr.Parse("8080"), Path: "/path"},
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
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": "test-clusterw1-123",
		},
		Tolerations: []corev1.Toleration{
			{
				Operator: corev1.TolerationOpEqual,
				Key:      "node-role.kubernetes.io/master",
				Effect:   corev1.TaintEffectNoSchedule,
				Value:    "",
			},
		},
		Affinity: &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Operator: corev1.NodeSelectorOpIn,
									Key:      "topology.kubernetes.io/zone",
									Values:   []string{"antarctica-east1"},
								},
							},
						},
					},
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
	flags.AddUpdateFlags(testCmd.Flags())
	testCmd.Execute()
}

func TestPodSpecMountResolveSubPath(t *testing.T) {
	args := map[string]*MountInfo{
		"/mydir=cm:my-cm": {
			VolumeName: "my-cm",
		},
		"/mydir=sc:my-sec": {
			VolumeName: "my-sec",
		},
		"/mydir=myvol": {
			VolumeName: "myvol",
		},

		"/mydir=cm:my-cm/subpath/to/mount": {
			VolumeName: "my-cm",
			SubPath:    "subpath/to/mount",
		},
		"/mydir=sc:my-sec/subpath/to/mount": {
			VolumeName: "my-sec",
			SubPath:    "subpath/to/mount",
		},
		"/mydir=myvol/subpath/to/mount": {
			VolumeName: "myvol",
			SubPath:    "subpath/to/mount",
		},
	}

	for arg, mountInfo := range args {
		outBuf := bytes.Buffer{}
		flags := &PodSpecFlags{}
		inputArgs := append([]string{"--image", "repo/user/imageID:tag", "--mount"}, arg)
		testCmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {
				podSpec := &corev1.PodSpec{Containers: []corev1.Container{{}}}
				err := flags.ResolvePodSpec(podSpec, cmd.Flags(), inputArgs)
				assert.NilError(t, err)
				assert.Equal(t, podSpec.Containers[0].VolumeMounts[0].SubPath, mountInfo.SubPath)
			},
		}
		testCmd.SetOut(&outBuf)

		testCmd.SetArgs(inputArgs)
		flags.AddFlags(testCmd.Flags())
		flags.AddCreateFlags(testCmd.Flags())
		testCmd.Execute()
	}
}

func TestPodSpecResolveContainers(t *testing.T) {
	rawInput := `
containers:
- image: foo:bar
  name: foo
  env:
  - name: a
    value: b
  resources: {}
- image: bar:bar
  name: bar
  resources: {}`

	rawInvalidFormatInput := `
containers:
  image: foo:bar
  name: foo	
  resources: {}`

	expectedPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image: "repo/user/imageID:tag",
				Resources: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
				},
			},
			{
				Name:  "foo",
				Image: "foo:bar",
				Env:   []corev1.EnvVar{{Name: "a", Value: "b"}},
			},
			{
				Name:  "bar",
				Image: "bar:bar",
			},
		},
	}

	testCases := []struct {
		name            string
		rawInput        string
		mockInput       func(data string) string
		expectedPodSpec corev1.PodSpec
		expectedError   error
	}{
		{
			"Input:stdin",
			rawInput,
			func(data string) string {
				return mockDataToStdin(t, data)
			},
			expectedPodSpec,
			nil,
		},
		{
			"Input:file",
			rawInput,
			func(data string) string {
				tempDir := t.TempDir()
				fileName := filepath.Join(tempDir, "container.yaml")
				os.WriteFile(fileName, []byte(data), test.FileModeReadWrite)
				return fileName
			},
			expectedPodSpec,
			nil,
		},
		{
			"Input:error",
			rawInput,
			func(data string) string {
				return "-"
			},
			corev1.PodSpec{Containers: []corev1.Container{{}}},
			errors.New("EOF"),
		},
		{
			"Input:invalidFormat",
			rawInvalidFormatInput,
			func(data string) string {
				return mockDataToStdin(t, data)
			},
			corev1.PodSpec{Containers: []corev1.Container{{}}},
			errors.New("cannot unmarshal object into Go struct field PodSpec.containers"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origStdin := os.Stdin

			fileName := tc.mockInput(tc.rawInput)
			if fileName == "-" {
				defer func() { os.Stdin = origStdin }()
			} else {
				defer os.RemoveAll(fileName)
			}

			inputArgs := []string{"--image", "repo/user/imageID:tag", "--containers", fileName}

			flags := &PodSpecFlags{}
			testCmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {
					podSpec := &corev1.PodSpec{Containers: []corev1.Container{{}}}
					err := flags.ResolvePodSpec(podSpec, cmd.Flags(), inputArgs)
					if tc.expectedError == nil {
						assert.NilError(t, err, "PodSpec cannot be resolved.")
						assert.DeepEqual(t, tc.expectedPodSpec, *podSpec)
					} else {
						assert.ErrorContains(t, err, tc.expectedError.Error())
					}
				},
			}
			testCmd.SetArgs(inputArgs)
			flags.AddFlags(testCmd.Flags())
			flags.AddUpdateFlags(testCmd.Flags())
			testCmd.Execute()
		})
	}
}

func mockDataToStdin(t *testing.T, data string) string {
	stdinReader, stdinWriter, err := os.Pipe()
	assert.NilError(t, err)
	_, err = stdinWriter.Write([]byte(data))
	assert.NilError(t, err)
	stdinWriter.Close()
	os.Stdin = stdinReader
	return "-"
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
	flags.AddCreateFlags(testCmd.Flags())
	testCmd.Execute()
	out := outBuf.String()
	assert.Assert(t, util.ContainsAll(out, "Invalid", "mount"))
}

func TestPodSpecResolveReturnErrorPullPolicy(t *testing.T) {
	outBuf := bytes.Buffer{}
	flags := &PodSpecFlags{}
	inputArgs := []string{"--pull-policy", "invalidPolicy"}
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
	flags.AddCreateFlags(testCmd.Flags())
	testCmd.Execute()
	out := outBuf.String()
	assert.Assert(t, util.ContainsAll(out, "invalid --pull-policy", "Always | Never | IfNotPresent"))
}

func TestPodSpecResolveWithEnvFile(t *testing.T) {
	file, err := os.CreateTemp("", "envfile.env")
	assert.NilError(t, err)
	file.WriteString("svcOwner=James\nsvcAuthor=James")
	defer os.Remove(file.Name())

	inputArgs := []string{"--image", "repo/user/imageID:tag", "--env", "svcOwner=David",
		"--env-file", file.Name(), "--port", "8080", "--cmd", "/app/start", "--arg", "myArg1"}
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
				Env: []corev1.EnvVar{{Name: "svcOwner", Value: "James"}, {Name: "svcAuthor", Value: "James"}},
				Resources: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
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
	flags.AddUpdateFlags(testCmd.Flags())
	testCmd.Execute()
}
