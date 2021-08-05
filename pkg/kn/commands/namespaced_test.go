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

package commands

import (
	"io/ioutil"
	"knative.dev/client/lib/test"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
)

// testCommandGenerator generates a test cobra command
func testCommandGenerator(allNamespaceFlag bool) *cobra.Command {
	var testCmd = &cobra.Command{
		Use:   "kn",
		Short: "Namespace test kn command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	AddNamespaceFlags(testCmd.Flags(), allNamespaceFlag)
	return testCmd
}

// test by setting some namespace
func TestGetNamespaceSample(t *testing.T) {
	testCmd := testCommandGenerator(true)
	expectedNamespace := "test1"
	testCmd.SetArgs([]string{"--namespace", expectedNamespace})
	testCmd.Execute()
	kp := &KnParams{fixedCurrentNamespace: FakeNamespace}
	actualNamespace, err := kp.GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
	kp = &KnParams{}
	// Mock ClientConfig to avoid clash with real kubeconfig on host OS
	kp.ClientConfig, _ = clientcmd.NewClientConfigFromBytes([]byte(BASIC_KUBECONFIG))
	testCmd = testCommandGenerator(false)
	actualNamespace, err = kp.GetNamespace(testCmd)
	assert.NilError(t, err)
	assert.Equal(t, "default", actualNamespace)
}

// test current namespace without setting any namespace
func TestGetNamespaceDefault(t *testing.T) {
	testCmd := testCommandGenerator(true)
	expectedNamespace := "current"
	testCmd.Execute()
	kp := &KnParams{fixedCurrentNamespace: FakeNamespace}
	actualNamespace, err := kp.GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

// test with all-namespaces flag set with sample namespace
// all-namespaces flag takes the precedence
func TestGetNamespaceAllNamespacesSet(t *testing.T) {
	testCmd := testCommandGenerator(true)
	expectedNamespace := ""
	sampleNamespace := "test1"

	// Test both variants of the "all namespaces" flag
	for _, arg := range []string{"--all-namespaces", "-A"} {
		testCmd.SetArgs([]string{"--namespace", sampleNamespace, arg})
		testCmd.Execute()
		kp := &KnParams{fixedCurrentNamespace: FakeNamespace}
		actualNamespace, err := kp.GetNamespace(testCmd)
		if err != nil {
			t.Fatal(err)
		}
		if actualNamespace != expectedNamespace {
			t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
		}
	}

}

// test with all-namespace flag set without any namespace flag set
// all-namespace flag takes precedence
func TestGetNamespaceDefaultAllNamespacesUnset(t *testing.T) {
	testCmd := testCommandGenerator(true)
	expectedNamespace := ""

	// Test both variants of the "all namespaces" flag
	for _, arg := range []string{"--all-namespaces", "-A"} {
		testCmd.SetArgs([]string{arg})
		testCmd.Execute()
		kp := &KnParams{fixedCurrentNamespace: FakeNamespace}
		actualNamespace, err := kp.GetNamespace(testCmd)
		if err != nil {
			t.Fatal(err)
		}
		if actualNamespace != expectedNamespace {
			t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
		}
	}

}

// test with all-namespaces flag not defined for command
func TestGetNamespaceAllNamespacesNotDefined(t *testing.T) {
	testCmd := testCommandGenerator(false)
	expectedNamespace := "test1"
	testCmd.SetArgs([]string{"--namespace", expectedNamespace})
	testCmd.Execute()
	kp := &KnParams{fixedCurrentNamespace: FakeNamespace}
	actualNamespace, err := kp.GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

func TestGetNamespaceFallback(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-unit-tests")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "mock")
	err = ioutil.WriteFile(tempFile, []byte(BASIC_KUBECONFIG), test.FileModeReadWrite)
	assert.NilError(t, err)

	testCmd := testCommandGenerator(true)
	kp := &KnParams{KubeCfgPath: tempFile}
	testCmd.Execute()
	actual, err := kp.GetNamespace(testCmd)
	assert.NilError(t, err)
	assert.Assert(t, actual == "default")
}

func TestCurrentNamespace(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kn-unit-tests")
	defer os.RemoveAll(tempDir)
	assert.NilError(t, err)

	tempFile := filepath.Join(tempDir, "empty")
	err = ioutil.WriteFile(tempFile, []byte(""), test.FileModeReadWrite)
	assert.NilError(t, err)

	// Invalid kubeconfig
	kp := &KnParams{KubeCfgPath: tempFile}
	actual, err := kp.CurrentNamespace()
	assert.Assert(t, err != nil)
	assert.Assert(t, clientcmd.IsConfigurationInvalid(err))
	assert.Assert(t, actual == "")

	// Missing kubeconfig
	kp = &KnParams{KubeCfgPath: filepath.Join(tempDir, "missing")}
	actual, err = kp.CurrentNamespace()
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "can not be found")
	assert.Assert(t, actual == "")

	// Variable namespace override
	kp = &KnParams{fixedCurrentNamespace: FakeNamespace}
	actual, err = kp.CurrentNamespace()
	assert.NilError(t, err)
	assert.Equal(t, actual, FakeNamespace)

	// Fallback to "default" namespace from mock kubeconfig
	tempFile = filepath.Join(tempDir, "mock")
	err = ioutil.WriteFile(tempFile, []byte(BASIC_KUBECONFIG), test.FileModeReadWrite)
	assert.NilError(t, err)
	kp = &KnParams{KubeCfgPath: tempFile}
	actual, err = kp.CurrentNamespace()
	assert.NilError(t, err)
	assert.Equal(t, actual, "default")
}
