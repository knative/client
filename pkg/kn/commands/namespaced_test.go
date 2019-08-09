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
	"testing"

	"github.com/spf13/cobra"
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
// all-namespaces flag takes the precendence
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
// all-namespace flag takes precendence
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
