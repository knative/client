// Copyright © 2018 The Knative Authors
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
	"io"
	"os"

	"k8s.io/client-go/kubernetes"
	"knative.dev/client/pkg/k8s"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	eventingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1"
	eventingv1beta2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta2"
	messagingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1"
	sourcesv1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
	servingv1client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	servingv1beta1client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1beta1"

	"knative.dev/client/pkg/util"

	clientdynamic "knative.dev/client/pkg/dynamic"
	knerrors "knative.dev/client/pkg/errors"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	clienteventingv1beta2 "knative.dev/client/pkg/eventing/v1beta2"
	clientmessagingv1 "knative.dev/client/pkg/messaging/v1"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	clientservingv1beta1 "knative.dev/client/pkg/serving/v1beta1"
	clientsourcesv1 "knative.dev/client/pkg/sources/v1"
)

// KnParams for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	k8s.Params
	Output                   io.Writer
	NewKubeClient            func() (kubernetes.Interface, error)
	NewServingClient         func(namespace string) (clientservingv1.KnServingClient, error)
	NewServingV1beta1Client  func(namespace string) (clientservingv1beta1.KnServingClient, error)
	NewGitopsServingClient   func(namespace string, dir string) (clientservingv1.KnServingClient, error)
	NewSourcesClient         func(namespace string) (clientsourcesv1.KnSourcesClient, error)
	NewEventingClient        func(namespace string) (clienteventingv1.KnEventingClient, error)
	NewMessagingClient       func(namespace string) (clientmessagingv1.KnMessagingClient, error)
	NewDynamicClient         func(namespace string) (clientdynamic.KnDynamicClient, error)
	NewEventingV1beta2Client func(namespace string) (clienteventingv1beta2.KnEventingV1Beta2Client, error)

	// General global options
	LogHTTP bool

	// Set this if you want to nail down the namespace
	fixedCurrentNamespace string

	// Memorizes the loaded config
	clientcmd.ClientConfig
}

// Initialize will initialize the default factories for the clients.
func (params *KnParams) Initialize() {
	if params.NewKubeClient == nil {
		params.NewKubeClient = params.newKubeClient
	}

	if params.NewServingClient == nil {
		params.NewServingClient = params.newServingClient
	}

	if params.NewServingV1beta1Client == nil {
		params.NewServingV1beta1Client = params.newServingClientV1beta1
	}

	if params.NewGitopsServingClient == nil {
		params.NewGitopsServingClient = params.newGitopsServingClient
	}

	if params.NewSourcesClient == nil {
		params.NewSourcesClient = params.newSourcesClient
	}

	if params.NewEventingClient == nil {
		params.NewEventingClient = params.newEventingClient
	}

	if params.NewMessagingClient == nil {
		params.NewMessagingClient = params.newMessagingClient
	}

	if params.NewDynamicClient == nil {
		params.NewDynamicClient = params.newDynamicClient
	}

	if params.NewEventingV1beta2Client == nil {
		params.NewEventingV1beta2Client = params.newEventingV1Beta2Client
	}
}

func (params *KnParams) newKubeClient() (kubernetes.Interface, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (params *KnParams) newServingClient(namespace string) (clientservingv1.KnServingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := servingv1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientservingv1.NewKnServingClient(client, namespace), nil
}

func (params *KnParams) newServingClientV1beta1(namespace string) (clientservingv1beta1.KnServingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := servingv1beta1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientservingv1beta1.NewKnServingClient(client, namespace), nil
}

func (params *KnParams) newGitopsServingClient(namespace string, dir string) (clientservingv1.KnServingClient, error) {
	return clientservingv1.NewKnServingGitOpsClient(namespace, dir), nil
}

func (params *KnParams) newSourcesClient(namespace string) (clientsourcesv1.KnSourcesClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := sourcesv1client.NewForConfig(restConfig)
	return clientsourcesv1.NewKnSourcesClient(client, namespace), nil
}

func (params *KnParams) newEventingClient(namespace string) (clienteventingv1.KnEventingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := eventingv1.NewForConfig(restConfig)
	return clienteventingv1.NewKnEventingClient(client, namespace), nil
}

func (params *KnParams) newEventingV1Beta2Client(namespace string) (clienteventingv1beta2.KnEventingV1Beta2Client, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := eventingv1beta2.NewForConfig(restConfig)
	return clienteventingv1beta2.NewKnEventingV1Beta2Client(client, namespace), nil
}

func (params *KnParams) newMessagingClient(namespace string) (clientmessagingv1.KnMessagingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := messagingv1.NewForConfig(restConfig)
	return clientmessagingv1.NewKnMessagingClient(client, namespace), nil
}

func (params *KnParams) newDynamicClient(namespace string) (clientdynamic.KnDynamicClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := dynamic.NewForConfig(restConfig)
	return clientdynamic.NewKnDynamicClient(client, namespace), nil
}

// RestConfig returns REST config, which can be to use to create specific clientset
func (params *KnParams) RestConfig() (*rest.Config, error) {
	var err error

	if params.ClientConfig == nil {
		params.ClientConfig, err = params.GetClientConfig()
		if err != nil {
			return nil, knerrors.GetError(err)
		}
	}

	config, err := params.ClientConfig.ClientConfig()
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	if params.LogHTTP {
		config.Wrap(util.NewLoggingTransport)
	}
	// Override client-go's warning handler to give us nicely printed warnings.
	config.WarningHandler = rest.NewWarningWriter(os.Stderr, rest.WarningWriterOptions{
		// only print a given warning the first time we receive it
		Deduplicate: true,
	})

	return config, nil
}
