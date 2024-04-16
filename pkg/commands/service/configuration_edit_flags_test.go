// Copyright © 2022 The Knative Authors
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

package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/v3/assert"

	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/config"
	"knative.dev/serving/pkg/apis/autoscaling"
)

func TestApplyPullPolicyFlag(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001", "test-svc-00002"}, goodConditions())
	cmd.SetArgs([]string{"--pull-policy", "Always"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.NilError(t, err)
}

func TestApplyPullPolicyFlagError(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001", "test-svc-00002"}, goodConditions())
	cmd.SetArgs([]string{"--pull-policy", "InvalidPolicy"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.Assert(t, util.ContainsAll(err.Error(), "invalid", "InvalidPolicy", "Valid arguments (case insensitive): Always | Never | IfNotPresent"))
}

func TestScaleMetric(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001", "test-svc-00002"}, goodConditions())
	cmd.SetArgs([]string{"--scale-metric", "rps"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.NilError(t, err)
}

func TestScaleActivation(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)
	svc := createTestService("test-svc", []string{"test-svc-00001"}, goodConditions())
	cmd.SetArgs([]string{"--scale-activation", "2"})
	cmd.Execute()
	err := editFlags.Apply(&svc, nil, cmd)
	assert.NilError(t, err)
	assert.Equal(t, svc.Spec.Template.Annotations[autoscaling.ActivationScaleKey], "2")
}

func TestApplyDefaultProfileFlag(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)

	editFlags.AddCreateFlags(cmd)

	err := config.BootstrapConfig()
	assert.NilError(t, err)

	svc := createTestService("test-svc", []string{"test-svc-00001"}, goodConditions())
	cmd.SetArgs([]string{"--profile", "istio"})
	cmd.Execute()
	editFlags.Apply(&svc, nil, cmd)
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.istio.io/inject"], "true")
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.istio.io/rewriteAppHTTPProbers"], "true")
	assert.Equal(t, svc.Spec.Template.Annotations["serving.knative.openshift.io/enablePassthrough"], "true")
}

func TestApplyProfileFlag(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)
	configYaml := `
profiles:
  testprofile:
      labels:
        - name: environment
          value: "test"
      annotations:
        - name: sidecar.testprofile.io/inject
          value: "true"
        - name: sidecar.testprofile.io/rewriteAppHTTPProbers
          value: "true"
`
	_, cleanup := setupConfig(t, configYaml)
	defer cleanup()

	editFlags.AddCreateFlags(cmd)

	err := config.BootstrapConfig()
	assert.NilError(t, err)

	svc := createTestService("test-svc", []string{"test-svc-00001"}, goodConditions())
	cmd.SetArgs([]string{"--profile", "testprofile"})
	cmd.Execute()
	editFlags.Apply(&svc, nil, cmd)
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.testprofile.io/inject"], "true")
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.testprofile.io/rewriteAppHTTPProbers"], "true")
	assert.Equal(t, svc.ObjectMeta.Labels["environment"], "test")
}

func TestDeleteProfileFlag(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)
	configYaml := `
profiles:
  testprofile:
    labels:
      - name: environment
        value: "test"
    annotations:
      - name: sidecar.testprofile.io/inject
        value: "true"
      - name: sidecar.testprofile.io/rewriteAppHTTPProbers
        value: "true"
`
	_, cleanup := setupConfig(t, configYaml)
	defer cleanup()

	editFlags.AddCreateFlags(cmd)

	err := config.BootstrapConfig()
	assert.NilError(t, err)

	svc := createTestService("test-svc", []string{"test-svc-00001"}, goodConditions())
	cmd.SetArgs([]string{"--profile", "testprofile"})
	cmd.Execute()
	editFlags.Apply(&svc, nil, cmd)
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.testprofile.io/inject"], "true")
	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.testprofile.io/rewriteAppHTTPProbers"], "true")
	assert.Equal(t, svc.ObjectMeta.Labels["environment"], "test")

	cmd.SetArgs([]string{"--profile", "testprofile-"})
	cmd.Execute()
	editFlags.Apply(&svc, nil, cmd)

	assert.Equal(t, svc.Spec.Template.Annotations["sidecar.istio.io/inject"], "")
	assert.Equal(t, len(svc.Spec.Template.Annotations), 1)

	assert.Equal(t, svc.ObjectMeta.Labels["environment"], "")
}

func TestApplyProfileFlagError(t *testing.T) {
	var editFlags ConfigurationEditFlags
	knParams := &commands.KnParams{}
	cmd, _, _ := commands.CreateTestKnCommand(NewServiceCreateCommand(knParams), knParams)
	editFlags.AddCreateFlags(cmd)
	err := config.BootstrapConfig()
	assert.NilError(t, err)

	svc := createTestService("test-svc", []string{"test-svc-00001"}, goodConditions())
	cmd.SetArgs([]string{"--profile", "invalidprofile"})
	cmd.Execute()
	err = editFlags.Apply(&svc, nil, cmd)
	assert.Assert(t, util.ContainsAll(err.Error(), "profile", "invalidprofile"))
}

func setupConfig(t *testing.T, configContent string) (string, func()) {
	tmpDir := t.TempDir()

	// Avoid to be fooled by the things in the the real homedir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Save old args
	backupArgs := os.Args

	// WriteCache out a temporary configContent file
	var cfgFile string
	if configContent != "" {
		cfgFile = filepath.Join(tmpDir, "config.yaml")
		os.Args = []string{"kn", "--config", cfgFile}
		err := os.WriteFile(cfgFile, []byte(configContent), 0644)
		assert.NilError(t, err)
	}
	return cfgFile, func() {
		// Cleanup everything
		os.Setenv("HOME", oldHome)
		os.Args = backupArgs
		viper.Reset()
	}
}
