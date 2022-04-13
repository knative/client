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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"knative.dev/client/lib/test"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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
		"--env-from", "config-map:config-map-name", "--user", "1001", "--pull-policy", "always"}
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
	flags.AddUpdateFlags(testCmd.Flags())
	testCmd.Execute()
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
				tempDir, err := ioutil.TempDir("", "kn-file")
				assert.NilError(t, err)
				fileName := filepath.Join(tempDir, "container.yaml")
				ioutil.WriteFile(fileName, []byte(data), test.FileModeReadWrite)
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
	file, err := ioutil.TempFile("", "envfile.env")
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
				Resources: v1.ResourceRequirements{
					Limits:   v1.ResourceList{},
					Requests: v1.ResourceList{},
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
