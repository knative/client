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
	"os"
	"path"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clienteventingv1beta2 "knative.dev/client/pkg/eventing/v1beta2"
	v1beta1 "knative.dev/client/pkg/messaging/v1"
	clientv1beta1 "knative.dev/client/pkg/serving/v1beta1"
	clientsourcesv1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/sources/v1beta2"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"
	sourcesv1fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1/fake"
	sourcesv1beta2fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1beta2/fake"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clienttesting "k8s.io/client-go/testing"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	v1 "knative.dev/client/pkg/serving/v1"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1/fake"
	beta2fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta2/fake"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"
	servingv1beta1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1beta1/fake"
)

type testType struct {
	name       string
	namespace  string
	p          *KnParams
	args       []string
	toComplete string
	resource   string
}

type mockMessagingClient struct {
	channelsClient      v1beta1.KnChannelsClient
	subscriptionsClient v1beta1.KnSubscriptionsClient
}

func (m *mockMessagingClient) ChannelsClient() v1beta1.KnChannelsClient {
	return m.channelsClient
}

func (m *mockMessagingClient) SubscriptionsClient() v1beta1.KnSubscriptionsClient {
	return m.subscriptionsClient
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
	fakeServingAlpha = &servingv1beta1fake.FakeServingV1beta1{Fake: &clienttesting.Fake{}}
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
	testDomain1 = servingv1beta1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-1", Namespace: testNs},
	}
	testDomain2 = servingv1beta1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-2", Namespace: testNs},
	}
	testDomain3 = servingv1beta1.DomainMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DomainMapping",
			APIVersion: "serving.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-domain-3", Namespace: testNs},
	}
	testNsDomains = []servingv1beta1.DomainMapping{testDomain1, testDomain2, testDomain3}
)

var (
	testTrigger1 = eventingv1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-trigger-1", Namespace: testNs},
	}
	testTrigger2 = eventingv1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-trigger-2", Namespace: testNs},
	}
	testTrigger3 = eventingv1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-trigger-3", Namespace: testNs},
	}
	testNsTriggers = []eventingv1.Trigger{testTrigger1, testTrigger2, testTrigger3}
)

var (
	testContainerSource1 = sourcesv1.ContainerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-container-source-1", Namespace: testNs},
	}
	testContainerSource2 = sourcesv1.ContainerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-container-source-2", Namespace: testNs},
	}
	testContainerSource3 = sourcesv1.ContainerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ContainerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-container-source-3", Namespace: testNs},
	}
	testNsContainerSources = []sourcesv1.ContainerSource{testContainerSource1, testContainerSource2, testContainerSource3}
	fakeSources            = &sourcesv1fake.FakeSourcesV1{Fake: &clienttesting.Fake{}}
)

var (
	testApiServerSource1 = sourcesv1.ApiServerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApiServerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ApiServer-source-1", Namespace: testNs},
	}
	testApiServerSource2 = sourcesv1.ApiServerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApiServerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ApiServer-source-2", Namespace: testNs},
	}
	testApiServerSource3 = sourcesv1.ApiServerSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApiServerSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ApiServer-source-3", Namespace: testNs},
	}
	testNsApiServerSources = []sourcesv1.ApiServerSource{testApiServerSource1, testApiServerSource2, testApiServerSource3}
)

var (
	testSinkBinding1 = sourcesv1.SinkBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SinkBinding",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-sink-binding-1", Namespace: testNs},
	}
	testSinkBinding2 = sourcesv1.SinkBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SinkBinding",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-sink-binding-2", Namespace: testNs},
	}
	testSinkBinding3 = sourcesv1.SinkBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SinkBinding",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-sink-binding-3", Namespace: testNs},
	}
	testNsSinkBindings = []sourcesv1.SinkBinding{testSinkBinding1, testSinkBinding2, testSinkBinding3}
)

var (
	testPingSource1 = sourcesv1beta2.PingSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PingSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ping-source-1", Namespace: testNs},
	}
	testPingSource2 = sourcesv1beta2.PingSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PingSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ping-source-2", Namespace: testNs},
	}
	testPingSource3 = sourcesv1beta2.PingSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PingSource",
			APIVersion: "sources.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ping-source-3", Namespace: testNs},
	}
	testNsPingSources  = []sourcesv1beta2.PingSource{testPingSource1, testPingSource2, testPingSource3}
	fakeSourcesV1Beta2 = &sourcesv1beta2fake.FakeSourcesV1beta2{Fake: &clienttesting.Fake{}}
)

var (
	testChannel1 = messagingv1.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-channel-1", Namespace: testNs},
	}
	testChannel2 = messagingv1.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-channel-2", Namespace: testNs},
	}
	testChannel3 = messagingv1.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-channel-3", Namespace: testNs},
	}
	testNsChannels = []messagingv1.Channel{testChannel1, testChannel2, testChannel3}
)

var (
	testSubscription1 = messagingv1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-subscription-1", Namespace: testNs},
	}
	testSubscription2 = messagingv1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-subscription-2", Namespace: testNs},
	}
	testSubscription3 = messagingv1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "messaging.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-subscription-3", Namespace: testNs},
	}
	testNsSubscriptions = []messagingv1.Subscription{testSubscription1, testSubscription2, testSubscription3}
)

var (
	testEventtype1 = eventingv1beta2.EventType{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventType",
			APIVersion: "eventing.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-eventtype-1", Namespace: testNs},
	}
	testEventtype2 = eventingv1beta2.EventType{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventType",
			APIVersion: "eventing.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-eventtype-2", Namespace: testNs},
	}
	testEventtype3 = eventingv1beta2.EventType{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventType",
			APIVersion: "eventing.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-eventtype-3", Namespace: testNs},
	}
	testEventtypes          = []eventingv1beta2.EventType{testEventtype1, testEventtype2, testEventtype3}
	fakeEventingBeta2Client = &beta2fake.FakeEventingV1beta2{Fake: &clienttesting.Fake{}}
)

var knParams = initialiseKnParams()

func initialiseKnParams() *KnParams {
	blankConfig, err := clientcmd.NewClientConfigFromBytes([]byte(`kind: Config
version: v1beta2
users:
- name: u
clusters:
- name: c
  cluster:
    server: example.com
contexts:
- name: x
  context:
    user: u
    cluster: c
current-context: x
`))
	if err != nil {
		panic(err)
	}
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
		NewServingV1beta1Client: func(namespace string) (clientv1beta1.KnServingClient, error) {
			return clientv1beta1.NewKnServingClient(fakeServingAlpha, namespace), nil
		},
		NewSourcesClient: func(namespace string) (clientsourcesv1.KnSourcesClient, error) {
			return clientsourcesv1.NewKnSourcesClient(fakeSources, namespace), nil
		},
		NewSourcesV1beta2Client: func(namespace string) (v1beta2.KnSourcesClient, error) {
			return v1beta2.NewKnSourcesClient(fakeSourcesV1Beta2, namespace), nil
		},
		NewEventingV1beta2Client: func(namespace string) (clienteventingv1beta2.KnEventingV1Beta2Client, error) {
			return clienteventingv1beta2.NewKnEventingV1Beta2Client(fakeEventingBeta2Client, namespace), nil
		},
		ClientConfig: blankConfig,
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
			return true, nil, errors.NewInternalError(fmt.Errorf("unable to list brokers"))
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
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list routes"))
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
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list domains"))
			}
			return true, &servingv1beta1.DomainMappingList{Items: testNsDomains}, nil
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

func TestResourceNameCompletionFuncTrigger(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeServing.AddReactor("list", "triggers",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list triggers"))
			}
			return true, &eventingv1.TriggerList{Items: testNsTriggers}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"trigger",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"trigger",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"trigger",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"trigger",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"trigger",
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

func TestResourceNameCompletionFuncContainerSource(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeSources.AddReactor("list", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list container sources"))
			}
			return true, &sourcesv1.ContainerSourceList{Items: testNsContainerSources}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"container",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"container",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"container",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"container",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"container",
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

func TestResourceNameCompletionFuncApiserverSource(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeSources.AddReactor("list", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list apiserver sources"))
			}
			return true, &sourcesv1.ApiServerSourceList{Items: testNsApiServerSources}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"apiserver",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"apiserver",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"apiserver",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"apiserver",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"apiserver",
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

func TestResourceNameCompletionFuncBindingSource(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeSources.AddReactor("list", "sinkbindings",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list binding sources"))
			}
			return true, &sourcesv1.SinkBindingList{Items: testNsSinkBindings}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"binding",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"binding",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"binding",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"binding",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"binding",
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

func TestResourceNameCompletionFuncPingSource(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeSourcesV1Beta2.AddReactor("list", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			if a.GetNamespace() == errorNs {
				return true, nil, errors.NewInternalError(fmt.Errorf("unable to list ping sources"))
			}
			return true, &sourcesv1beta2.PingSourceList{Items: testNsPingSources}, nil
		})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"ping",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"ping",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"ping",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"ping",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"ping",
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

func TestResourceNameCompletionFuncChannel(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	channelClient := v1beta1.NewMockKnChannelsClient(t)
	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{Items: testNsChannels}, nil)
	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{Items: testNsChannels}, nil)

	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{Items: testNsChannels}, nil)
	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{Items: testNsChannels}, nil)

	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{}, fmt.Errorf("error listing channels"))
	channelClient.Recorder().ListChannel(&messagingv1.ChannelList{}, fmt.Errorf("error listing channels"))

	messagingClient := &mockMessagingClient{channelClient, nil}

	knParams.NewMessagingClient = func(namespace string) (v1beta1.KnMessagingClient, error) {
		return messagingClient, nil
	}
	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"channel",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"channel",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"channel",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"channel",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"channel",
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
	channelClient.Recorder().Validate()
}

func TestResourceNameCompletionFuncSubscription(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	subscriptionsClient := v1beta1.NewMockKnSubscriptionsClient(t)
	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{Items: testNsSubscriptions}, nil)
	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{Items: testNsSubscriptions}, nil)

	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{Items: testNsSubscriptions}, nil)
	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{Items: testNsSubscriptions}, nil)

	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{}, fmt.Errorf("error listing channels"))
	subscriptionsClient.Recorder().ListSubscription(&messagingv1.SubscriptionList{}, fmt.Errorf("error listing channels"))

	messagingClient := &mockMessagingClient{nil, subscriptionsClient}

	knParams.NewMessagingClient = func(namespace string) (v1beta1.KnMessagingClient, error) {
		return messagingClient, nil
	}
	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"subscription",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"subscription",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"subscription",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"subscription",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"subscription",
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
	subscriptionsClient.Recorder().Validate()
}

func TestResourceNameCompletionFuncEventtype(t *testing.T) {
	completionFunc := ResourceNameCompletionFunc(knParams)

	fakeEventingBeta2Client.AddReactor("list", "eventtypes", func(a clienttesting.Action) (bool, runtime.Object, error) {
		if a.GetNamespace() == errorNs {
			return true, nil, errors.NewInternalError(fmt.Errorf("unable to list eventtypes"))
		}
		return true, &eventingv1beta2.EventTypeList{Items: testEventtypes}, nil
	})

	tests := []testType{
		{
			"Empty suggestions when non-zero args",
			testNs,
			knParams,
			[]string{"xyz"},
			"",
			"eventtype",
		},
		{
			"Empty suggestions when no namespace flag",
			"",
			knParams,
			nil,
			"",
			"eventtype",
		},
		{
			"Suggestions when test-ns namespace set",
			testNs,
			knParams,
			nil,
			"",
			"eventtype",
		},
		{
			"Empty suggestions when toComplete is not a prefix",
			testNs,
			knParams,
			nil,
			"xyz",
			"eventtype",
		},
		{
			"Empty suggestions when error during list operation",
			errorNs,
			knParams,
			nil,
			"",
			"eventtype",
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
	tempDir := t.TempDir()

	svcPath := path.Join(tempDir, "test-ns", "ksvc")
	err := os.MkdirAll(svcPath, 0700)
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
