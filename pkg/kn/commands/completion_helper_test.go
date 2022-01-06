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
	clientv1alpha1 "knative.dev/client/pkg/serving/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clienttesting "k8s.io/client-go/testing"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	v1 "knative.dev/client/pkg/serving/v1"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1/fake"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"
	servingv1alpha1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
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
	testSvc1 = servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-1", Namespace: testNs},
	}
	testSvc2 = servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-2", Namespace: testNs},
	}
	testSvc3 = servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-svc-3", Namespace: testNs},
	}
	testNsServices = []servingv1.Service{testSvc1, testSvc2, testSvc3}

	fakeServing      = &servingv1fake.FakeServingV1{Fake: &clienttesting.Fake{}}
	fakeServingAlpha = &servingv1alpha1fake.FakeServingV1alpha1{Fake: &clienttesting.Fake{}}
)

var (
	testBroker1 = eventingv1.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-broker-1", Namespace: testNs},
	}
	testBroker2 = eventingv1.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-broker-2", Namespace: testNs},
	}
	testBroker3 = eventingv1.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-broker-3", Namespace: testNs},
	}
	testNsBrokers = []eventingv1.Broker{testBroker1, testBroker2, testBroker3}

	fakeEventing = &fake.FakeEventingV1{Fake: &clienttesting.Fake{}}
)

var (
	testRev1 = servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-rev-1", Namespace: testNs},
	}
	testRev2 = servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-rev-2", Namespace: testNs},
	}
	testRev3 = servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-rev-3", Namespace: testNs},
	}
	testNsRevs = []servingv1.Revision{testRev1, testRev2, testRev3}
)

var (
	testRoute1 = servingv1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-route-1", Namespace: testNs},
	}
	testRoute2 = servingv1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-route-2", Namespace: testNs},
	}
	testRoute3 = servingv1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-route-3", Namespace: testNs},
	}
	testNsRoutes = []servingv1.Route{testRoute1, testRoute2, testRoute3}
)

var (
	testDomain1 = v1alpha1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-1", Namespace: testNs},
	}
	testDomain2 = v1alpha1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-2", Namespace: testNs},
	}
	testDomain3 = v1alpha1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-3", Namespace: testNs},
	}
	testNsDomains = []v1alpha1.DomainMapping{testDomain1, testDomain2, testDomain3}
)

var knParams = initialiseKnParams()

func initialiseKnParams() *KnParams {
	return &KnParams{
		NewServingClient: func(namespace string) (v1.KnServingClient, error) {
			return v1.NewKnServingClient(fakeServing, namespace), nil
		},
		NewGitopsServingClient: func(namespace string, dir string) (v1.KnServingClient, error) {
			return v1.NewKnServingGitOpsClient(namespace, dir), nil
		},
		NewEventingClient: func(namespace string) (clienteventingv1.KnEventingClient, error) {
			return clienteventingv1.NewKnEventingClient(fakeEventing, namespace), nil
		},
		NewServingV1alpha1Client: func(namespace string) (clientv1alpha1.KnServingClient, error) {
			return clientv1alpha1.NewKnServingClient(fakeServingAlpha, namespace), nil
		},
	}
}

func TestResourceNameCompletionFuncService(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServing.AddReactor("list", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list services"))
			}
			return true, &servingv1.ServiceList{Items: testNsServices}, nil
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

func TestResourceNameCompletionFuncBroker(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeEventing.AddReactor("list", "brokers", func(action clienttesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() == errorNs {
			return true, nil, errors.NewInternalError(fmt.Errorf("unable to list services"))
		}
		return true, &eventingv1.BrokerList{Items: testNsBrokers}, nil
	})
	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"broker",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"broker",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"broker",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"broker",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"broker",
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

func TestResourceNameCompletionFuncRevision(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServing.AddReactor("list", "revisions",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list revisions"))
			}
			return true, &servingv1.RevisionList{Items: testNsRevs}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"revision",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"revision",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"revision",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"revision",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"revision",
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

func TestResourceNameCompletionFuncRoute(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServing.AddReactor("list", "routes",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list services"))
			}
			return true, &servingv1.RouteList{Items: testNsRoutes}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"route",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"route",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"route",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"route",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"route",
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

func TestResourceNameCompletionFuncDomain(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServingAlpha.AddReactor("list", "domainmappings",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list services"))
			}
			return true, &v1alpha1.DomainMappingList{Items: testNsDomains}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"domain",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"domain",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"domain",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"domain",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"domain",
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

	for i, testSvc := range []servingv1.Service{testSvc1, testSvc2, testSvc3} {
		tempFile, err := os.Create(path.Join(svcPath, fmt.Sprintf("test-svc-%d.yaml", i+1)))
		assert.NilError(t, err)
		writeToFile(t, testSvc, tempFile)
	}

	return tempDir
}

func writeToFile(t *testing.T, testSvc servingv1.Service, tempFile *os.File) {
	yamlPrinter, err := genericclioptions.NewJSONYamlPrintFlags().ToPrinter("yaml")
	assert.NilError(t, err)

	err = yamlPrinter.PrintObj(&testSvc, tempFile)
	assert.NilError(t, err)

	defer tempFile.Close()
}
