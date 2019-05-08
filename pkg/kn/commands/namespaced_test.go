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
	"github.com/spf13/cobra"
	"testing"
)

func CommandGenerator(allNamespaceFlag bool) *cobra.Command {
	var testCmd = &cobra.Command{
		Use:   "ns-test",
		Short: "Namespace test command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	AddNamespaceFlags(testCmd.Flags(), allNamespaceFlag)
	return testCmd
}

// test by setting some namespace
func TestGetNamespaceSample(t *testing.T) {
	testCmd := CommandGenerator(true)
	expectedNamespace := "test1"
	testCmd.SetArgs([]string{"--namespace", expectedNamespace})
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

// test without setting any namespace
func TestGetNamespaceDefault(t *testing.T) {
	testCmd := CommandGenerator(true)
	expectedNamespace := "default"
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
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
	testCmd := CommandGenerator(true)
	expectedNamespace := ""
	sampleNamespace := "test1"
	testCmd.SetArgs([]string{"--namespace", sampleNamespace, "--all-namespaces"})
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

// test with all-namespace flag set without any namespace flag set
// all-namespace flag takes precendence
func TestGetNamespaceDefaultAllNamespacesUnset(t *testing.T) {
	testCmd := CommandGenerator(true)
	expectedNamespace := ""
	testCmd.SetArgs([]string{"--all-namespaces"})
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

// test with all-namespaces flag not defined for command
func TestGetNamespaceAllNamespacesNotDefined(t *testing.T) {
	testCmd := CommandGenerator(false)
	expectedNamespace := "test1"
	testCmd.SetArgs([]string{"--namespace", expectedNamespace})
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}

// test with all-namespace flag not defined and no namespace given
func TestGetNamespaceDefaultAllNamespacesNotDefined(t *testing.T) {
	testCmd := CommandGenerator(false)
	expectedNamespace := "default"
	testCmd.Execute()
	actualNamespace, err := GetNamespace(testCmd)
	if err != nil {
		t.Fatal(err)
	}
	if actualNamespace != expectedNamespace {
		t.Fatalf("Incorrect namespace retrieved: %v, expected: %v", actualNamespace, expectedNamespace)
	}
}
