// Copyright © 2021 The Knative Authors
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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clienttesting "k8s.io/client-go/testing"
	v1 "knative.dev/client/pkg/serving/v1"
	v12 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"
)

type testType struct {
	name       string
	namespace  string
	p          *KnParams
	args       []string
	toComplete string
	resource   string
}

const (
	testNs  = "test-ns"
	errorNs = "error-ns"
)

var (
	testSvc1 = v12.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-1", Namespace: testNs},
	}
	testSvc2 = v12.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-2", Namespace: testNs},
	}
	testSvc3 = v12.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-3", Namespace: testNs},
	}
	testNsServices = []v12.Service{testSvc1, testSvc2, testSvc3}

	fakeServing = &servingv1fake.FakeServingV1{Fake: &clienttesting.Fake{}}
	knParams    = &KnParams{
		NewServingClient: func(namespace string) (v1.KnServingClient, error) {
			return v1.NewKnServingClient(fakeServing, namespace), nil
		},
		NewGitopsServingClient: func(namespace string, dir string) (v1.KnServingClient, error) {
			return v1.NewKnServingGitOpsClient(namespace, dir), nil
		},
	}
)

func TestResourceNameCompletionFuncService(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServing.AddReactor("list", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list services"))
			}
			return true, &v12.ServiceList{Items: testNsServices}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when no parent command found",
			testNs,
			knParams,
			nil,
			"",
			"no-parent",
		},
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"service",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"service",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"service",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"service",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"service",
		},
	}
	for _, tt := range tests {
		cmd := getResourceCommandWithTestSubcommand(tt.resource, tt.namespace != "", tt.resource != "no-parent")
		t.Run(tt.name, func(t *testing.T) {
			config := &completionConfig{
				params:     tt.p,
				command:    cmd,
				args:       tt.args,
				toComplete: tt.toComplete,
			}
			expectedFunc := resourceToFuncMap[tt.resource]
			if expectedFunc == nil {
				expectedFunc = func(config *completionConfig) []string {
					return []string{}
				}
			}
			cmd.Flags().Set("namespace", tt.namespace)
			actualSuggestions, actualDirective := completionFunc(cmd, tt.args, tt.toComplete)
			expectedSuggestions := expectedFunc(config)
			expectedDirective := cobra.ShellCompDirectiveNoFileComp
			assert.DeepEqual(t, actualSuggestions, expectedSuggestions)
			assert.Equal(t, actualDirective, expectedDirective)
		})
	}
}

func TestResourceNameCompletionFuncGitOps(t *testing.T) {
	tempDir := setupTempDir(t)
	assert.Assert(t, tempDir != "")
	defer os.RemoveAll(tempDir)

	completionFunc := ResourceNameCompletionFunc(knParams)

	tests := []testType{
		{
			"Empty suggestions when no parent command found",
			testNs,
			knParams,
			nil,
			"",
			"service",
		},
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"service",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"service",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"service",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"service",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"service",
		},
	}

	for _, tt := range tests {
		cmd := getResourceCommandWithTestSubcommand(tt.resource, tt.namespace != "", tt.resource != "no-parent")
		t.Run(tt.name, func(t *testing.T) {
			config := &completionConfig{
				params:     tt.p,
				command:    cmd,
				args:       tt.args,
				toComplete: tt.toComplete,
			}
			expectedFunc := resourceToFuncMap[tt.resource]
			cmd.Flags().String("target", tempDir, "target directory")
			cmd.Flags().Set("namespace", tt.namespace)

			expectedSuggestions := expectedFunc(config)
			expectedDirective := cobra.ShellCompDirectiveNoFileComp
			actualSuggestions, actualDirective := completionFunc(cmd, tt.args, tt.toComplete)
			assert.DeepEqual(t, actualSuggestions, expectedSuggestions)
			assert.Equal(t, actualDirective, expectedDirective)
		})
	}
}

func getResourceCommandWithTestSubcommand(resource string, addNamespace, addSubcommand bool) *cobra.Command {
	testCommand := &cobra.Command{
		Use: resource,
	}
	testSubCommand := &cobra.Command{
		Use: "test",
	}
	if addSubcommand {
		testCommand.AddCommand(testSubCommand)
	}
	if addNamespace {
		AddNamespaceFlags(testCommand.Flags(), true)
		AddNamespaceFlags(testSubCommand.Flags(), true)
	}
	return testSubCommand
}

func setupTempDir(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "test-dir")
	assert.NilError(t, err)

	svcPath := path.Join(tempDir, "test-ns", "ksvc")
	err = os.MkdirAll(svcPath, 0700)
	assert.NilError(t, err)

	for i, testSvc := range []v12.Service{testSvc1, testSvc2, testSvc3} {
		tempFile, err := os.Create(path.Join(svcPath, fmt.Sprintf("test-svc-%d.yaml", i+1)))
		assert.NilError(t, err)
		writeToFile(t, testSvc, tempFile)
	}

	return tempDir
}

func writeToFile(t *testing.T, testSvc v12.Service, tempFile *os.File) {
	yamlPrinter, err := genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
	assert.NilError(t, err)

	err = yamlPrinter.PrintObj(&testSvc, tempFile)
	assert.NilError(t, err)

	defer tempFile.Close()
}
